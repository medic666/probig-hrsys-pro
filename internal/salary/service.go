package salary

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"

	"probig/internal/attendance"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/config"
	"probig/internal/position"
)

type Service struct {
	dao *DAO
}

var globalService *Service

func NewService(db *gorm.DB) *Service {
	svc := &Service{dao: NewDAO(db)}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) CreateEvent(event *SalaryEvent, operatorID uint, operatorName string) error {
	if err := s.dao.CreateEvent(event); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "salary_event", event.ID, "create", "", event)
	return nil
}

func (s *Service) UpdateEvent(event *SalaryEvent, operatorID uint, operatorName string) error {
	old, err := s.dao.GetEventByID(event.ID)
	if err != nil {
		return err
	}
	if err := s.dao.UpdateEvent(event); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "salary_event", event.ID, "update", old, event)
	return nil
}

func (s *Service) DeleteEvent(id uint, operatorID uint, operatorName string) error {
	event, err := s.dao.GetEventByID(id)
	if err != nil {
		return err
	}
	if err := s.dao.DeleteEvent(id); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "salary_event", id, "delete", event, nil)
	return nil
}

func (s *Service) GetEventByID(id uint) (*SalaryEvent, error) {
	return s.dao.GetEventByID(id)
}

func (s *Service) ListEvents(personID uint, belongMonth string, eventType string, page, pageSize int) ([]SalaryEvent, int64, error) {
	return s.dao.ListEvents(personID, belongMonth, eventType, page, pageSize)
}

func (s *Service) CalculateSalary(personID uint, belongMonth string) (*SalarySummary, error) {
	posSvc := position.GetService()
	attSvc := attendance.GetService()

	snapshots, err := posSvc.GetSnapshotsByMonth([]uint{personID}, belongMonth)
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	snapshot := snapshots[0]

	attEvents, err := attSvc.GetEventsByPersonAndMonth(personID, belongMonth)
	if err != nil {
		return nil, err
	}

	salaryEvents, err := s.dao.GetEventsByPersonAndMonth(personID, belongMonth)
	if err != nil {
		return nil, err
	}

	summary := s.calculateSummary(snapshot, attEvents, salaryEvents, belongMonth)
	summary.PersonID = personID
	summary.BelongMonth = belongMonth

	now := time.Now().Format("2006-01-02 15:04:05")
	summary.LastCalcAt = &now

	if err := s.dao.CreateOrUpdateSummary(summary); err != nil {
		return nil, err
	}

	return summary, nil
}

func (s *Service) CalculateSalaries(personIDs []uint, belongMonth string) ([]SalarySummary, error) {
	var summaries []SalarySummary
	for _, pid := range personIDs {
		summary, err := s.CalculateSalary(pid, belongMonth)
		if err != nil {
			continue
		}
		summaries = append(summaries, *summary)
	}
	return summaries, nil
}

func (s *Service) calculateSummary(snapshot position.PositionSnapshot, attEvents []attendance.AttendanceEvent, salaryEvents []SalaryEvent, belongMonth string) *SalarySummary {
	workDays := 0.0
	makeUpDays := 0.0
	sickLeaveDays := 0.0
	personalLeaveDays := 0.0
	annualLeaveDays := 0.0
	statutoryLeaveDays := 0.0
	welfareLeaveDays := 0.0
	overtimeWorkdayHours := 0.0
	overtimeHolidayHours := 0.0
	annualLeaveAdjustDays := 0.0
	violationCount := 0

	for _, e := range attEvents {
		hours := floatPtrVal(e.Hours)
		switch e.EventType {
		case "出勤":
			switch e.SubType {
			case "普通出勤":
				workDays += hours / 8.0
			case "补班出勤":
				makeUpDays += hours / 8.0
			}
		case "休假":
			switch e.SubType {
			case "病假":
				sickLeaveDays += hours / 8.0
			case "事假":
				personalLeaveDays += hours / 8.0
			case "年假":
				annualLeaveDays += hours / 8.0
			case "法定假":
				statutoryLeaveDays += hours / 8.0
			case "福利假":
				welfareLeaveDays += hours / 8.0
			}
		case "加班":
			switch e.SubType {
			case "工作日加班":
				overtimeWorkdayHours += hours
			case "节假日加班":
				overtimeHolidayHours += hours
			}
		case "违纪":
			violationCount++
		case "年假调整":
			if e.SubType == "年假结转" {
				annualLeaveAdjustDays += floatPtrVal(e.LeaveAdjustAmount)
			}
		}
	}

	salaryDays := snapshot.SalaryDays
	if salaryDays == 0 {
		salaryDays = 21
	}
	defaultSalaryDays := float64(salaryDays)

	dailySalary := snapshot.BaseSalary / defaultSalaryDays

	attendanceDays := workDays + makeUpDays + sickLeaveDays*0.6 + annualLeaveDays + statutoryLeaveDays + welfareLeaveDays
	attendanceSalary := attendanceDays * dailySalary

	attendanceBonus := 0.0
	if snapshot.HasAttendanceBonus {
		if personalLeaveDays > 0 {
			attendanceBonus = 0
		} else {
			bonusRate := parseFloatConfig("salary.attendance_bonus_rate", 0)
			bonus := (attendanceDays - float64(violationCount)) * bonusRate
			if bonus < 0 {
				bonus = 0
			}
			attendanceBonus = bonus
		}
	}

	overtimeBase := snapshot.BaseSalary + snapshot.MealAllowance
	overtimeWorkdayRate := parseFloatConfig("salary.overtime_workday_rate", 1.5)
	overtimeHolidayRate := parseFloatConfig("salary.overtime_holiday_rate", 2.0)
	overtimeWorkdaySalary := overtimeBase / defaultSalaryDays / 8 * overtimeWorkdayHours * overtimeWorkdayRate
	overtimeHolidaySalary := overtimeBase / defaultSalaryDays / 8 * overtimeHolidayHours * overtimeHolidayRate
	annualLeaveOvertimeSalary := dailySalary * annualLeaveAdjustDays * 2
	totalOvertimeSalary := overtimeWorkdaySalary + overtimeHolidaySalary + annualLeaveOvertimeSalary

	coefficient := 1.0
	latestPerfEvent, err := s.dao.GetLatestPerformanceEvent(personIDFromSnapshot(snapshot), belongMonth)
	if err == nil {
		coefficient = latestPerfEvent.Amount
	}

	performanceSalary := snapshot.PerformanceSalary * coefficient

	postAllowance := snapshot.PostAllowance
	transportAllowance := snapshot.TransportAllowance
	insuranceCompensation := snapshot.InsuranceCompensation
	fundCompensation := snapshot.FundCompensation
	housingAllowance := snapshot.HousingAllowance
	mealAllowance := snapshot.MealAllowance

	highTempAllowance := 0.0
	highTempMonths := parseHighTempMonths()
	currentMonth, _ := strconv.Atoi(belongMonth[5:7])
	for _, m := range highTempMonths {
		if m == currentMonth {
			highTempAllowance = snapshot.HighTempAllowance
			break
		}
	}

	totalAllowance := postAllowance + transportAllowance + insuranceCompensation + fundCompensation + housingAllowance + mealAllowance + highTempAllowance

	totalAdjustment := 0.0
	for _, e := range salaryEvents {
		if e.EventType != "绩效调整" {
			totalAdjustment += e.Amount
		}
	}

	totalDeduction := snapshot.SocialSecurityDeduct + snapshot.HousingFundDeduct

	if isPartialMonth(snapshot, belongMonth) {
		prorate := attendanceDays / defaultSalaryDays
		if prorate > 1.0 {
			prorate = 1.0
		}
		performanceSalary = performanceSalary * prorate
		totalAllowance = totalAllowance * prorate
	}

	finalSalary := attendanceSalary + totalOvertimeSalary + attendanceBonus + performanceSalary + totalAllowance + totalAdjustment - totalDeduction
	finalSalary = math.Round(finalSalary*100) / 100

	return &SalarySummary{
		AttendanceSalary:  math.Round(attendanceSalary*100) / 100,
		OvertimeSalary:    math.Round(totalOvertimeSalary*100) / 100,
		AttendanceBonus:   math.Round(attendanceBonus*100) / 100,
		PerformanceSalary: math.Round(performanceSalary*100) / 100,
		TotalAllowance:    math.Round(totalAllowance*100) / 100,
		TotalAdjustment:   math.Round(totalAdjustment*100) / 100,
		TotalDeduction:    math.Round(totalDeduction*100) / 100,
		FinalSalary:       finalSalary,
	}
}

func (s *Service) GetSummaryByPersonMonth(personID uint, belongMonth string) (*SalarySummary, error) {
	return s.dao.GetSummaryByPersonMonth(personID, belongMonth)
}

func (s *Service) ListSummaries(personIDs []uint, belongMonth string, isLocked *bool, page, pageSize int) ([]SalarySummary, int64, error) {
	return s.dao.ListSummaries(personIDs, belongMonth, isLocked, page, pageSize)
}

func (s *Service) LockSummary(personID uint, belongMonth string, operatorID uint, operatorName string) error {
	if err := s.dao.LockSummary(personID, belongMonth); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "salary_summary", 0, "lock", nil, map[string]interface{}{"personId": personID, "belongMonth": belongMonth})
	return nil
}

func (s *Service) UnlockSummary(personID uint, belongMonth string, operatorID uint, operatorName string) error {
	if err := s.dao.UnlockSummary(personID, belongMonth); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "salary_summary", 0, "unlock", nil, map[string]interface{}{"personId": personID, "belongMonth": belongMonth})
	return nil
}

func (s *Service) IsMonthLocked(personID uint, belongMonth string) (bool, error) {
	summary, err := s.dao.GetSummaryByPersonMonth(personID, belongMonth)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return summary.IsLocked, nil
}

func (s *Service) logAudit(operatorID uint, operatorName, targetType string, targetID uint, action string, before, after interface{}) {
	if audit.GlobalAuditService == nil {
		return
	}
	var beforeJSON, afterJSON string
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			beforeJSON = string(b)
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			afterJSON = string(b)
		}
	}
	audit.GlobalAuditService.Log(operatorID, operatorName, targetType, targetID, action, beforeJSON, afterJSON, "", "")
}

func parseFloatConfig(key string, defaultVal float64) float64 {
	val := config.GetSysConfig(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return f
}

func parseHighTempMonths() []int {
	val := config.GetSysConfig("salary.high_temp_months")
	if val == "" {
		return []int{6, 7, 8, 9}
	}
	var months []int
	if err := json.Unmarshal([]byte(val), &months); err != nil {
		return []int{6, 7, 8, 9}
	}
	return months
}

func isPartialMonth(snapshot position.PositionSnapshot, belongMonth string) bool {
	monthStart := belongMonth + "-01"
	year, _ := strconv.Atoi(belongMonth[0:4])
	month, _ := strconv.Atoi(belongMonth[5:7])
	monthEnd := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	if snapshot.EntryDate != nil {
		entryDateStr := *snapshot.EntryDate
		if entryDateStr >= monthStart && entryDateStr <= monthEnd {
			return true
		}
	}
	if snapshot.LeaveDate != nil {
		leaveDateStr := *snapshot.LeaveDate
		if leaveDateStr >= monthStart && leaveDateStr <= monthEnd {
			return true
		}
	}
	return false
}

func personIDFromSnapshot(snapshot position.PositionSnapshot) uint {
	return snapshot.PersonID
}

func floatPtrVal(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
