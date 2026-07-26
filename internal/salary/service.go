package salary

import (
	"strconv"
	"strings"
	"time"

	"probig/internal/pkg/config"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) ListEvents(pageNum, pageSize int, personID uint, belongMonth, eventType string) ([]SalaryEvent, int64, error) {
	var list []SalaryEvent
	var total int64
	db := s.DB.Model(&SalaryEvent{})
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if belongMonth != "" {
		db = db.Where("belong_month = ?", belongMonth)
	}
	if eventType != "" {
		db = db.Where("event_type = ?", eventType)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("belong_month desc, created_at desc").Find(&list).Error
	return list, total, err
}

func (s *Service) CreateEvent(req map[string]interface{}) (uint, error) {
	event := SalaryEvent{
		PersonID:    uint(getF(req, "person_id")),
		BelongMonth: getS(req, "belong_month"),
		EventType:   getS(req, "event_type"),
		Amount:      getF(req, "amount"),
		EventName:   getS(req, "event_name"),
		Remark:      getS(req, "remark"),
	}
	if err := s.DB.Create(&event).Error; err != nil {
		return 0, err
	}
	return event.ID, nil
}

func (s *Service) UpdateEvent(id uint, req map[string]interface{}) error {
	updates := map[string]interface{}{}
	for _, k := range []string{"belong_month", "event_type", "event_name", "remark"} {
		if v := getS(req, k); v != "" {
			updates[k] = v
		}
	}
	if v, ok := req["amount"]; ok {
		updates["amount"] = v
	}
	return s.DB.Model(&SalaryEvent{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) DeleteEvent(id uint) error {
	return s.DB.Delete(&SalaryEvent{}, id).Error
}

func (s *Service) CalculateSummary(personID uint, belongMonth string) error {
	s.DB.Where("person_id = ? AND belong_month = ?", personID, belongMonth).Delete(&SalarySummary{})

	var attSalary struct {
		AttendanceSalary       float64
		OvertimeWorkdaySalary  float64
		OvertimeHolidaySalary  float64
		AttendanceBonus        float64
		WeightedBaseSalary     float64
		TotalWorkHours         float64
		TotalOvertimeWorkdayH  float64
		TotalOvertimeHolidayH  float64
		SalaryDays             int
		HasPersonalLeaveMonth  bool
		TotalViolationCount    int
		WeightedMealAllowance  float64
	}
	err := s.DB.Table("attendance_salary").Where("person_id = ? AND belong_month = ?", personID, belongMonth).First(&attSalary).Error
	if err != nil {
		return err
	}

	var snaps []struct {
		PerformanceSalary    float64
		PostAllowance        float64
		HousingAllowance     float64
		TransportAllowance   float64
		HighTempAllowance    float64
		InsuranceComp        float64
		FundComp             float64
		SocialSecurityDeduct float64
		HousingFundDeduct    float64
		MealAllowance        float64
		EntryDate            string
		LeaveDate            *string
		EffectiveStartDate   string
		EffectiveEndDate     string
		HasAttendanceBonus   bool
	}

	monthStart := belongMonth + "-01"
	y, _ := strconv.Atoi(belongMonth[:4])
	m, _ := strconv.Atoi(belongMonth[5:7])
	monthEnd := time.Date(y, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	s.DB.Table("position_snapshot").Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
		personID, monthEnd, monthStart).Order("effective_start_date").Find(&snaps)

	if len(snaps) == 0 {
		return gorm.ErrRecordNotFound
	}

	totalDays := calcDays(monthStart, monthEnd)
	entryDate := snaps[0].EntryDate
	var leaveDateStr string
	for _, snap := range snaps {
		if snap.LeaveDate != nil && *snap.LeaveDate != "" {
			leaveDateStr = *snap.LeaveDate
			break
		}
	}

	var perfBase, postAll, housingAll, transportAll, highTempAll, insComp, fundComp, ssDeduct, hfDeduct, mealAll float64
	for _, snap := range snaps {
		sStart, _ := time.Parse("2006-01-02", snap.EffectiveStartDate)
		sEnd, _ := time.Parse("2006-01-02", snap.EffectiveEndDate)
		if sStart.Before(time.Now().Add(-10*365*24*time.Hour)) {
			ms, _ := time.Parse("2006-01-02", monthStart)
			sStart = ms
		}
		if sEnd.After(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)) {
			me, _ := time.Parse("2006-01-02", monthEnd)
			sEnd = me
		}
		if sStart, sEnd = clampRange(sStart, sEnd, monthStart, monthEnd); sStart.After(sEnd) || sEnd.Before(sStart) {
			continue
		}
		ratio := calcDays(sStart.Format("2006-01-02"), sEnd.Format("2006-01-02")) / totalDays
		perfBase += snap.PerformanceSalary * ratio
		postAll += snap.PostAllowance * ratio
		housingAll += snap.HousingAllowance * ratio
		transportAll += snap.TransportAllowance * ratio
		insComp += snap.InsuranceComp * ratio
		fundComp += snap.FundComp * ratio
		ssDeduct += snap.SocialSecurityDeduct * ratio
		hfDeduct += snap.HousingFundDeduct * ratio
		mealAll += snap.MealAllowance * ratio
	}

	highTempMonthStr := belongMonth[5:]
	isHighTempMonth := false
	if v := config.GetConfig("attendance.high_temp_months"); v != "" {
		highTempStr := strings.Trim(v, `[]"`)
		for _, m := range splitAndTrim(highTempStr, ",") {
			if m == highTempMonthStr {
				isHighTempMonth = true
				break
			}
		}
	}
	if isHighTempMonth {
		for _, snap := range snaps {
			sStart, _ := time.Parse("2006-01-02", snap.EffectiveStartDate)
			sEnd, _ := time.Parse("2006-01-02", snap.EffectiveEndDate)
			if sStart, sEnd = clampRange(sStart, sEnd, monthStart, monthEnd); sStart.Before(sEnd) || sStart.Equal(sEnd) {
				ratio := calcDays(sStart.Format("2006-01-02"), sEnd.Format("2006-01-02")) / totalDays
				highTempAll += snap.HighTempAllowance * ratio
			}
		}
	}

	inWorkDays := calcInWorkDays(monthStart, monthEnd, entryDate, leaveDateStr)

	var perfRatio float64 = 1
	var latestPerfEvent SalaryEvent
	s.DB.Where("person_id = ? AND belong_month = ? AND event_type = ?", personID, belongMonth, "绩效系数").
		Order("created_at desc").First(&latestPerfEvent)
	if latestPerfEvent.ID > 0 {
		perfRatio = latestPerfEvent.Amount
	}

	perfSalary := perfBase * perfRatio

	if inWorkDays > 0 && totalDays > 0 && inWorkDays < totalDays {
		ratio := inWorkDays / totalDays
		perfSalary = perfBase * perfRatio * ratio
		postAll *= ratio
		housingAll *= ratio
		transportAll *= ratio
		highTempAll *= ratio
		insComp *= ratio
		fundComp *= ratio
		mealAll *= ratio
	}

	var adjustTotal float64
	var taxTotal float64
	var allEvents []SalaryEvent
	s.DB.Where("person_id = ? AND belong_month = ?", personID, belongMonth).Find(&allEvents)
	for _, e := range allEvents {
		switch e.EventType {
		case "绩效系数":
		case "个税扣除":
			taxTotal += e.Amount
		default:
			adjustTotal += e.Amount
		}
	}

	var annualLeaveSal float64
	if isAnniversaryEveMonth(belongMonth, entryDate) {
		var bal struct{ BalanceHours float64 }
		s.DB.Table("leave_account_balance").
			Where("person_id = ? AND leave_type = ?", personID, "annual_leave").
			Select("balance_hours").First(&bal)
		if bal.BalanceHours > 0 {
			hoursPerDay := 8.0
			if v := config.GetConfig("system.work_hours_per_day"); v != "" {
				hoursPerDay = parseFloat(v)
			}
			otHolidayRatio := 2.0
			if v := config.GetConfig("attendance.overtime_holiday_ratio"); v != "" {
				otHolidayRatio = parseFloat(v)
			}
			salaryDays := attSalary.SalaryDays
			if salaryDays == 0 {
				salaryDays = 21
			}
			annualLeaveSal = bal.BalanceHours * (attSalary.WeightedBaseSalary / float64(salaryDays) / hoursPerDay) * otHolidayRatio
		}
	}

	now := time.Now()
	summary := SalarySummary{
		PersonID:                 personID,
		BelongMonth:              belongMonth,
		SalaryDays:               attSalary.SalaryDays,
		WeightedBaseSalary:       attSalary.WeightedBaseSalary,
		TotalWorkHours:           attSalary.TotalWorkHours,
		TotalOvertimeWorkdayH:    attSalary.TotalOvertimeWorkdayH,
		TotalOvertimeHolidayH:    attSalary.TotalOvertimeHolidayH,
		AttendanceSalary:         attSalary.AttendanceSalary,
		OvertimeWorkdaySalary:    attSalary.OvertimeWorkdaySalary,
		OvertimeHolidaySalary:    attSalary.OvertimeHolidaySalary,
		AnnualLeaveCarryoverSal:  annualLeaveSal,
		AttendanceBonus:          attSalary.AttendanceBonus,
		PerformanceSalary:        perfSalary,
		PostAllowance:            postAll,
		MealAllowance:            mealAll,
		HousingAllowance:         housingAll,
		TransportAllowance:       transportAll,
		HighTempAllowance:        highTempAll,
		InsuranceComp:            insComp,
		FundComp:                 fundComp,
		TotalAdjustment:          adjustTotal,
		SocialSecurityDeduct:     ssDeduct,
		HousingFundDeduct:        hfDeduct,
		TaxDeduct:                taxTotal,
		LastCalcAt:               &now,
	}

	summary.FinalSalary = summary.AttendanceSalary +
		summary.OvertimeWorkdaySalary + summary.OvertimeHolidaySalary +
		summary.AnnualLeaveCarryoverSal + summary.AttendanceBonus +
		summary.PerformanceSalary + summary.PostAllowance + summary.MealAllowance +
		summary.HousingAllowance + summary.TransportAllowance + summary.HighTempAllowance +
		summary.InsuranceComp + summary.FundComp + summary.TotalAdjustment -
		summary.SocialSecurityDeduct - summary.HousingFundDeduct - summary.TaxDeduct

	if inWorkDays > 0 && totalDays > 0 && inWorkDays < totalDays {
		ratio := inWorkDays / totalDays
		summary.SocialSecurityDeduct = ssDeduct * ratio
		summary.HousingFundDeduct = hfDeduct * ratio
		summary.FinalSalary = summary.AttendanceSalary +
			summary.OvertimeWorkdaySalary + summary.OvertimeHolidaySalary +
			summary.AnnualLeaveCarryoverSal + summary.AttendanceBonus +
			summary.PerformanceSalary + summary.PostAllowance + summary.MealAllowance +
			summary.HousingAllowance + summary.TransportAllowance + summary.HighTempAllowance +
			summary.InsuranceComp + summary.FundComp + summary.TotalAdjustment -
			summary.SocialSecurityDeduct - summary.HousingFundDeduct - summary.TaxDeduct
	}

	return s.DB.Create(&summary).Error
}

func (s *Service) GetSummaryList(belongMonth string, personID uint, pageNum, pageSize int) ([]SalarySummary, int64, error) {
	var list []SalarySummary
	var total int64
	db := s.DB.Model(&SalarySummary{})
	if belongMonth != "" {
		db = db.Where("belong_month = ?", belongMonth)
	}
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("belong_month desc").Find(&list).Error
	return list, total, err
}

func (s *Service) HasAttendaceSalary(personID uint, belongMonth string) bool {
	var count int64
	s.DB.Table("attendance_salary").Where("person_id = ? AND belong_month = ?", personID, belongMonth).Count(&count)
	return count > 0
}

func clampRange(snapStart, snapEnd time.Time, monthStart, monthEnd string) (time.Time, time.Time) {
	ms, _ := time.Parse("2006-01-02", monthStart)
	me, _ := time.Parse("2006-01-02", monthEnd)
	if snapStart.Before(ms) {
		snapStart = ms
	}
	if snapEnd.After(me) {
		snapEnd = me
	}
	return snapStart, snapEnd
}

func calcDays(startStr, endStr string) float64 {
	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)
	return end.Sub(start).Hours()/24 + 1
}

func getS(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getF(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case string:
			v, _ := strconv.ParseFloat(val, 64)
			return v
		}
	}
	return 0
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(p, `"`))
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func calcInWorkDays(monthStart, monthEnd, entryDate, leaveDate string) float64 {
	ms, _ := time.Parse("2006-01-02", monthStart)
	me, _ := time.Parse("2006-01-02", monthEnd)
	totalMonthDays := me.Sub(ms).Hours()/24 + 1

	workStart := ms
	if entryDate != "" {
		ed, _ := time.Parse("2006-01-02", entryDate)
		if ed.After(ms) {
			workStart = ed
		}
	}
	workEnd := me
	if leaveDate != "" {
		ld, _ := time.Parse("2006-01-02", leaveDate)
		if ld.Before(me) {
			workEnd = ld
		}
	}
	days := workEnd.Sub(workStart).Hours()/24 + 1
	if days > totalMonthDays {
		days = totalMonthDays
	}
	if days < 0 {
		days = 0
	}
	return days
}

func isAnniversaryEveMonth(belongMonth, entryDate string) bool {
	if entryDate == "" || len(entryDate) < 7 {
		return false
	}
	emo := entryDate[5:7]
	bmo := belongMonth[5:7]
	ey, _ := strconv.Atoi(entryDate[:4])
	by, _ := strconv.Atoi(belongMonth[:4])
	em, _ := strconv.Atoi(emo)
	bm, _ := strconv.Atoi(bmo)

	prevM := bm - 1
	prevY := by
	if prevM == 0 {
		prevM = 12
		prevY = by - 1
	}
	if prevY < ey {
		return false
	}
	return em == prevM
}
