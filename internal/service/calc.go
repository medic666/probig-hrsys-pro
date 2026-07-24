package service

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"probig/internal/dao"
	"probig/internal/middleware"
	"probig/internal/models"

	"github.com/gin-gonic/gin"
)

func GetAttendanceSummaryList(page, pageSize int, personID uint, belongMonth string) ([]models.AttendanceSummary, int64, error) {
	return dao.GetAttendanceSummaryList(page, pageSize, personID, belongMonth)
}

func CalculateAttendance(c *gin.Context, personID uint, belongMonth string) (*models.AttendanceSummary, error) {
	events, err := dao.GetAttendanceEventsByPersonAndMonth(personID, belongMonth)
	if err != nil {
		return nil, err
	}
	summary := &models.AttendanceSummary{
		PersonID:    personID,
		BelongMonth: belongMonth,
	}
	for _, e := range events {
		switch {
		case e.EventType == "出勤" && e.SubType == "普通出勤":
			summary.WorkDays += float64ValFromPtr(e.Hours) / 8
		case e.EventType == "出勤" && e.SubType == "补班出勤":
			summary.MakeUpDays += float64ValFromPtr(e.Hours) / 8
		case e.EventType == "休假" && e.SubType == "病假":
			summary.SickLeaveDays += float64ValFromPtr(e.Hours) / 8
		case e.EventType == "休假" && e.SubType == "事假":
			summary.PersonalLeaveDays += float64ValFromPtr(e.Hours) / 8
		case e.EventType == "休假" && e.SubType == "年假":
			summary.AnnualLeaveDays += float64ValFromPtr(e.Hours) / 8
		case e.EventType == "休假" && e.SubType == "法定假":
			summary.StatutoryLeaveDays += float64ValFromPtr(e.Hours) / 8
		case e.EventType == "休假" && e.SubType == "福利假":
			summary.WelfareLeaveDays += float64ValFromPtr(e.Hours) / 8
		case e.EventType == "加班" && e.SubType == "工作日加班":
			summary.OvertimeWorkdayHours += float64ValFromPtr(e.Hours)
		case e.EventType == "加班" && e.SubType == "节假日加班":
			summary.OvertimeHolidayHours += float64ValFromPtr(e.Hours)
		case e.EventType == "违纪":
			summary.ViolationCount++
		}
	}
	now := time.Now()
	summary.LastCalcAt = &now
	if err := dao.UpsertAttendanceSummary(summary); err != nil {
		return nil, err
	}
	middleware.RecordAudit(c, "核算", "attendance_summary", summary.ID, nil, summary, "")
	return summary, nil
}

func LockAttendanceSummary(c *gin.Context, personID uint, belongMonth string, locked bool) error {
	if err := dao.LockAttendanceSummary(personID, belongMonth, locked); err != nil {
		return err
	}
	action := "锁定"
	if !locked {
		action = "解锁"
	}
	middleware.RecordAudit(c, action, "attendance_summary", 0, nil, map[string]interface{}{
		"person_id":    personID,
		"belong_month": belongMonth,
		"is_locked":    locked,
	}, "")
	return nil
}

func CheckAttendanceOutdated(personID uint, belongMonth string) (bool, error) {
	lastEventTime, err := dao.GetLastAttendanceEventTimeInMonth(personID, belongMonth)
	if err != nil {
		return false, nil
	}
	var summary models.AttendanceSummary
	err = dao.DB().Where("person_id = ? AND belong_month = ?", personID, belongMonth).First(&summary).Error
	if err != nil {
		return false, nil
	}
	if summary.LastCalcAt == nil {
		return true, nil
	}
	return lastEventTime.After(*summary.LastCalcAt), nil
}

func CalculateSalary(c *gin.Context, personID uint, belongMonth string) (*models.SalarySummary, error) {
	snapshot, err := getSnapshotForMonth(personID, belongMonth)
	if err != nil {
		return nil, errors.New("未找到该员工该月的职务快照")
	}
	attEvents, _ := dao.GetAttendanceEventsByPersonAndMonth(personID, belongMonth)
	salaryEvents, _ := dao.GetSalaryEventsByPersonAndMonth(personID, belongMonth)

	daySalary := snapshot.BaseSalary / float64(snapshot.SalaryDays)

	attendanceSalary := calculateAttendanceSalary(snapshot, attEvents, daySalary, belongMonth)
	overtimeSalary := calculateOvertimeSalary(snapshot, attEvents, daySalary)
	attendanceBonus := calculateAttendanceBonus(snapshot, attEvents)
	performanceSalary := calculateFinalPerformanceSalary(snapshot, salaryEvents)
	totalAllowance := calculateTotalAllowance(snapshot, belongMonth)
	totalAdjustment := calculateTotalAdjustment(salaryEvents)
	totalDeduction := snapshot.SocialSecurityDeduct + snapshot.HousingFundDeduct

	finalSalary := attendanceSalary + overtimeSalary + attendanceBonus + performanceSalary + totalAllowance + totalAdjustment - totalDeduction
	if finalSalary < 0 {
		finalSalary = 0
	}

	now := time.Now()
	summary := &models.SalarySummary{
		PersonID:          personID,
		BelongMonth:       belongMonth,
		AttendanceSalary:  attendanceSalary,
		OvertimeSalary:    overtimeSalary,
		AttendanceBonus:   attendanceBonus,
		PerformanceSalary: performanceSalary,
		TotalAllowance:    totalAllowance,
		TotalAdjustment:   totalAdjustment,
		TotalDeduction:    totalDeduction,
		FinalSalary:       finalSalary,
		LastCalcAt:        &now,
	}
	if err := dao.UpsertSalarySummary(summary); err != nil {
		return nil, err
	}
	middleware.RecordAudit(c, "核算", "salary_summary", summary.ID, nil, summary, "")
	return summary, nil
}

func getSnapshotForMonth(personID uint, belongMonth string) (*models.PositionSnapshot, error) {
	parts := strings.Split(belongMonth, "-")
	if len(parts) != 2 {
		return nil, errors.New("月份格式错误")
	}
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	date := time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.UTC)
	return dao.GetPositionSnapshot(personID, date)
}

func calculateAttendanceSalary(snapshot *models.PositionSnapshot, events []models.AttendanceEvent, daySalary float64, belongMonth string) float64 {
	if snapshot.EntryDate != nil {
		entryMonth := snapshot.EntryDate.Format("2006-01")
		if entryMonth == belongMonth {
			daysInMonth := float64(daysInMonthOf(belongMonth))
			actualDays := daysInMonth - float64(snapshot.EntryDate.Day()) + 1
			daySalary = snapshot.BaseSalary / float64(snapshot.SalaryDays) * (actualDays / daysInMonth)
		}
	}
	if snapshot.LeaveDate != nil {
		leaveMonth := snapshot.LeaveDate.Format("2006-01")
		if leaveMonth == belongMonth {
			daysInMonth := float64(daysInMonthOf(belongMonth))
			actualDays := float64(snapshot.LeaveDate.Day())
			daySalary = snapshot.BaseSalary / float64(snapshot.SalaryDays) * (actualDays / daysInMonth)
		}
	}

	totalDays := 0.0
	for _, e := range events {
		hours := float64ValFromPtr(e.Hours)
		switch {
		case e.EventType == "出勤":
			totalDays += hours / 8
		case e.EventType == "休假" && e.SubType == "病假":
			ratio, _ := strconv.ParseFloat(ConfigCache["salary.sick_leave_ratio"], 64)
			totalDays += hours / 8 * ratio
		case e.EventType == "休假" && (e.SubType == "年假" || e.SubType == "法定假" || e.SubType == "福利假"):
			totalDays += hours / 8
		}
	}
	return totalDays * daySalary
}

func calculateOvertimeSalary(snapshot *models.PositionSnapshot, events []models.AttendanceEvent, daySalary float64) float64 {
	base := snapshot.BaseSalary + snapshot.MealAllowance
	if snapshot.SalaryDays == 0 {
		return 0
	}
	hourlyBase := base / float64(snapshot.SalaryDays) / 8

	workdayRatio, _ := strconv.ParseFloat(ConfigCache["salary.workday_overtime_ratio"], 64)
	holidayRatio, _ := strconv.ParseFloat(ConfigCache["salary.holiday_overtime_ratio"], 64)

	total := 0.0
	for _, e := range events {
		hours := float64ValFromPtr(e.Hours)
		if e.EventType == "加班" && e.SubType == "工作日加班" {
			total += hourlyBase * hours * workdayRatio
		}
		if e.EventType == "加班" && e.SubType == "节假日加班" {
			total += hourlyBase * hours * holidayRatio
		}
	}
	return total
}

func calculateAttendanceBonus(snapshot *models.PositionSnapshot, events []models.AttendanceEvent) float64 {
	if !snapshot.HasAttendanceBonus {
		return 0
	}
	hasPersonaLeave := false
	violationCount := 0
	workDays := 0.0
	for _, e := range events {
		if e.EventType == "休假" && e.SubType == "事假" {
			hasPersonaLeave = true
		}
		if e.EventType == "违纪" {
			violationCount++
		}
		if e.EventType == "出勤" {
			workDays += float64ValFromPtr(e.Hours) / 8
		}
	}
	if hasPersonaLeave {
		return 0
	}
	ratio, _ := strconv.ParseFloat(ConfigCache["salary.attendance_bonus_ratio"], 64)
	bonus := (workDays - float64(violationCount)) * ratio
	if bonus < 0 {
		bonus = 0
	}
	return bonus
}

func calculateFinalPerformanceSalary(snapshot *models.PositionSnapshot, salaryEvents []models.SalaryEvent) float64 {
	coefficient := 1.0
	var latestPerfEvent *models.SalaryEvent
	for i := range salaryEvents {
		if salaryEvents[i].EventType == "绩效调整" {
			if latestPerfEvent == nil || salaryEvents[i].CreatedAt.After(latestPerfEvent.CreatedAt) {
				latestPerfEvent = &salaryEvents[i]
			}
		}
	}
	// Convert to slice for easier iteration
	var perfEvents []models.SalaryEvent
	for _, e := range salaryEvents {
		if e.EventType == "绩效调整" {
			perfEvents = append(perfEvents, e)
		}
	}
	sort.Slice(perfEvents, func(i, j int) bool {
		return perfEvents[i].CreatedAt.After(perfEvents[j].CreatedAt)
	})
	if len(perfEvents) > 0 {
		coefficient = perfEvents[0].Amount
	}
	return snapshot.PerformanceSalary * coefficient
}

func calculateTotalAllowance(snapshot *models.PositionSnapshot, belongMonth string) float64 {
	total := snapshot.PostAllowance + snapshot.HousingAllowance + snapshot.TransportAllowance +
		snapshot.InsuranceCompensation + snapshot.FundCompensation

	highTempMonths := ConfigCache["salary.high_temp_months"]
	if highTempMonths != "" {
		parts := strings.Split(belongMonth, "-")
		if len(parts) == 2 {
			month := parts[1]
			if strings.Contains(highTempMonths, month) {
				total += snapshot.HighTempAllowance
			}
		}
	}
	return total
}

func calculateTotalAdjustment(salaryEvents []models.SalaryEvent) float64 {
	total := 0.0
	for _, e := range salaryEvents {
		if e.EventType != "绩效调整" {
			total += e.Amount
		}
	}
	return total
}

func float64ValFromPtr(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func daysInMonthOf(belongMonth string) int {
	parts := strings.Split(belongMonth, "-")
	if len(parts) != 2 {
		return 30
	}
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}

func GetSalarySummaryList(page, pageSize int, personID uint, belongMonth string) ([]models.SalarySummary, int64, error) {
	return dao.GetSalarySummaryList(page, pageSize, personID, belongMonth)
}

func LockSalarySummary(c *gin.Context, personID uint, belongMonth string, locked bool) error {
	if err := dao.LockSalarySummary(personID, belongMonth, locked); err != nil {
		return err
	}
	action := "锁定"
	if !locked {
		action = "解锁"
	}
	middleware.RecordAudit(c, action, "salary_summary", 0, nil, map[string]interface{}{
		"person_id":    personID,
		"belong_month": belongMonth,
		"is_locked":    locked,
	}, "")
	return nil
}
