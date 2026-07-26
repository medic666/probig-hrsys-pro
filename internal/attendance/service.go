package attendance

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
	"probig/internal/pkg/projection"
)

var OnLeaveAccountRecalc func(personID uint, leaveType string) error

func getSickLeaveCoefficient() float64 {
	v := config.Get("attendance.sick_leave_ratio")
	if v == "" {
		return 0.6
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0.6
	}
	return f
}

func getWorkdayOvertimeCoefficient() float64 {
	v := config.Get("attendance.overtime_workday_ratio")
	if v == "" {
		return 1.5
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 1.5
	}
	return f
}

func getHolidayOvertimeCoefficient() float64 {
	v := config.Get("attendance.overtime_holiday_ratio")
	if v == "" {
		return 3.0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 3.0
	}
	return f
}

func getAttendanceBonusDaily() float64 {
	v := config.Get("attendance.bonus_daily_standard")
	if v == "" {
		return 20.0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 20.0
	}
	return f
}

func getWorkHoursPerDay() int {
	v := config.Get("system.work_hours_per_day")
	if v == "" {
		return 8
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 8
	}
	return i
}

func RebuildDailyProjection(personID uint, workDate time.Time) error {
	events, err := FindEventsByPersonAndDate(personID, workDate)
	if err != nil {
		return fmt.Errorf("failed to find events: %w", err)
	}

	sickCoef := getSickLeaveCoefficient()

	var workHours float64
	var overtimeWorkdayHours float64
	var overtimeHolidayHours float64
	hasPersonalLeave := false
	violationCount := 0
	var punchTime string
	var remark string

	for _, e := range events {
		if e.PunchTime != "" {
			punchTime = e.PunchTime
		}
		if e.Remark != "" {
			if remark == "" {
				remark = e.Remark
			} else {
				remark += "; " + e.Remark
			}
		}

		switch e.EventType {
		case "出勤":
			switch e.SubType {
			case "普通出勤", "补班出勤", "外勤出勤":
				workHours += e.Hours
			}
		case "休假":
			switch e.SubType {
			case "病假":
				workHours += e.Hours * sickCoef
			case "事假":
				hasPersonalLeave = true
			case "年假", "调休", "法定假", "福利假":
				workHours += e.Hours
			}
		case "加班":
			switch e.SubType {
			case "工作日加班":
				overtimeWorkdayHours += e.Hours
			case "节假日加班":
				overtimeHolidayHours += e.Hours
			}
		case "违纪":
			violationCount++
		}
	}

	now := time.Now()
	wd := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 0, 0, 0, 0, time.UTC)

	proj := AttendanceDailyProjection{
		PersonID:             personID,
		WorkDate:             &wd,
		PunchTime:            punchTime,
		WorkHours:            math.Round(workHours*10) / 10,
		OvertimeWorkdayHours: math.Round(overtimeWorkdayHours*10) / 10,
		OvertimeHolidayHours: math.Round(overtimeHolidayHours*10) / 10,
		HasPersonalLeave:     hasPersonalLeave,
		ViolationCount:       violationCount,
		Remark:               remark,
		LastCalcAt:           &now,
	}

	return UpsertDailyProjection(&proj)
}

func CalcMonthlySalary(belongMonth string, personIDs []uint, operatorID uint, operatorName string, ip string) ([]MonthlySalaryWithName, []string, error) {
	var results []MonthlySalaryWithName
	var errors []string

	monthStart, err := time.Parse("2006-01-02", belongMonth+"-01")
	if err != nil {
		return nil, nil, fmt.Errorf("invalid month format: %s", belongMonth)
	}
	monthEnd := monthStart.AddDate(0, 1, -1)

	if len(personIDs) == 0 {
		rows, err := getActivePersonIDsInMonth(monthStart)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get active persons: %w", err)
		}
		for _, row := range rows {
			personIDs = append(personIDs, row.PersonID)
		}
	}

	for _, personID := range personIDs {
		ms, err := calcMonthlyForPerson(personID, belongMonth, monthStart, monthEnd)
		if err != nil {
			errors = append(errors, fmt.Sprintf("人员#%d: %s", personID, err.Error()))
			continue
		}

		if ms == nil {
			continue
		}

		if err := UpsertMonthlySalary(ms); err != nil {
			errors = append(errors, fmt.Sprintf("人员#%d 保存失败: %s", personID, err.Error()))
			continue
		}

		personName := GetPersonName(personID)
		results = append(results, MonthlySalaryWithName{
			AttendanceSalaryMonthly: *ms,
			PersonName:              personName,
		})
	}

	return results, errors, nil
}

func calcMonthlyForPerson(personID uint, belongMonth string, monthStart, monthEnd time.Time) (*AttendanceSalaryMonthly, error) {
	snapshots, err := getPositionSnapshotsInRange(personID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, nil
	}

	weightedBase, weightedMealAllowance, salaryDays := calcWeightedFromSnapshots(snapshots, monthStart, monthEnd)

	dailyProjections, err := GetDailyByPersonAndDateRange(personID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}

	var totalWorkHours float64
	var totalOvertimeWorkdayHours float64
	var totalOvertimeHolidayHours float64
	hasPersonalLeaveMonth := false
	totalViolationCount := 0

	for _, dp := range dailyProjections {
		totalWorkHours += dp.WorkHours
		totalOvertimeWorkdayHours += dp.OvertimeWorkdayHours
		totalOvertimeHolidayHours += dp.OvertimeHolidayHours
		if dp.HasPersonalLeave {
			hasPersonalLeaveMonth = true
		}
		totalViolationCount += dp.ViolationCount
	}

	totalWorkHours = math.Round(totalWorkHours*10) / 10
	totalOvertimeWorkdayHours = math.Round(totalOvertimeWorkdayHours*10) / 10
	totalOvertimeHolidayHours = math.Round(totalOvertimeHolidayHours*10) / 10

	if salaryDays <= 0 {
		salaryDays = getWorkHoursPerDay()
	}

	hourlyRate := 0.0
	hourlyRateOT := 0.0
	if salaryDays > 0 {
		workHoursPerDay := float64(getWorkHoursPerDay())
		hourlyRate = weightedBase / float64(salaryDays) / workHoursPerDay
		hourlyRateOT = (weightedBase + weightedMealAllowance) / float64(salaryDays) / workHoursPerDay
	}

	attendanceSalary := math.Round(totalWorkHours*hourlyRate*100) / 100
	overtimeWorkdaySalary := math.Round(totalOvertimeWorkdayHours*hourlyRateOT*getWorkdayOvertimeCoefficient()*100) / 100
	overtimeHolidaySalary := math.Round(totalOvertimeHolidayHours*hourlyRateOT*getHolidayOvertimeCoefficient()*100) / 100

	attendanceBonus := 0.0
	if !hasAttendanceBonusDisabled(snapshots) && !hasPersonalLeaveMonth {
		bonusDaily := getAttendanceBonusDaily()
		workHoursPerDay := float64(getWorkHoursPerDay())
		bonus := (totalWorkHours/workHoursPerDay - float64(totalViolationCount)) * bonusDaily
		if bonus > 0 {
			attendanceBonus = math.Round(bonus*100) / 100
		}
	}

	now := time.Now()
	ms := &AttendanceSalaryMonthly{
		PersonID:                 personID,
		BelongMonth:              belongMonth,
		SalaryDays:               salaryDays,
		WeightedBaseSalary:       math.Round(weightedBase*100) / 100,
		WeightedMealAllowance:    math.Round(weightedMealAllowance*100) / 100,
		TotalWorkHours:           totalWorkHours,
		TotalOvertimeWorkdayHours: totalOvertimeWorkdayHours,
		TotalOvertimeHolidayHours: totalOvertimeHolidayHours,
		AttendanceSalary:         attendanceSalary,
		OvertimeWorkdaySalary:    overtimeWorkdaySalary,
		OvertimeHolidaySalary:    overtimeHolidaySalary,
		HasPersonalLeaveMonth:    hasPersonalLeaveMonth,
		TotalViolationCount:      totalViolationCount,
		AttendanceBonus:          attendanceBonus,
		LastCalcAt:               &now,
	}

	return ms, nil
}

func getActivePersonIDsInMonth(monthStart time.Time) ([]struct{ PersonID uint }, error) {
	monthEnd := monthStart.AddDate(0, 1, -1)
	var rows []struct{ PersonID uint }
	err := db().Model(&PositionSnapshot{}).
		Where("effective_start_date <= ? AND effective_end_date >= ?", monthEnd, monthStart).
		Select("DISTINCT person_id").
		Find(&rows).Error
	return rows, err
}

func getPositionSnapshotsInRange(personID uint, start, end time.Time) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	err := db().Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
		personID, end, start).
		Order("effective_start_date ASC").
		Find(&snapshots).Error
	return snapshots, err
}

func calcWeightedFromSnapshots(snapshots []PositionSnapshot, monthStart, monthEnd time.Time) (baseSalary, mealAllowance float64, salaryDays int) {
	if len(snapshots) == 0 {
		return 0, 0, 0
	}

	totalDays := float64(monthEnd.Sub(monthStart).Hours()/24) + 1

	var weightedBase float64
	var weightedMeal float64
	var weightSum float64
	totalSalaryDays := 0

	for _, s := range snapshots {
		if s.EffectiveStartDate == nil || s.EffectiveEndDate == nil {
			continue
		}
		segStart := *s.EffectiveStartDate
		segEnd := *s.EffectiveEndDate

		if segStart.Before(monthStart) {
			segStart = monthStart
		}
		if segEnd.After(monthEnd) {
			segEnd = monthEnd
		}
		if segStart.After(segEnd) {
			continue
		}

		days := segEnd.Sub(segStart).Hours()/24 + 1
		weightedBase += s.BaseSalary * days
		weightedMeal += s.MealAllowance * days
		weightSum += days

		if s.SalaryDays > 0 {
			totalSalaryDays = s.SalaryDays
		}
	}

	if weightSum > 0 {
		_ = totalDays
		return weightedBase / weightSum, weightedMeal / weightSum, totalSalaryDays
	}

	s := snapshots[len(snapshots)-1]
	return s.BaseSalary, s.MealAllowance, s.SalaryDays
}

func hasAttendanceBonusDisabled(snapshots []PositionSnapshot) bool {
	for _, s := range snapshots {
		if !s.HasAttendanceBonus {
			return true
		}
	}
	return false
}

func CreateAttendanceEvent(event *AttendanceEvent, operatorID uint, operatorName string, ip string) error {
	if event.PersonID == 0 {
		return fmt.Errorf("person_id is required")
	}
	if event.EventDate == nil {
		return fmt.Errorf("event_date is required")
	}
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.Hours < 0 {
		return fmt.Errorf("hours cannot be negative")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)

		if err := createAuditInTx(tx, operatorID, operatorName, "attendance_event", event.ID, personName, "新增", nil, event, ip); err != nil {
			return err
		}

		if err := RebuildDailyProjection(event.PersonID, *event.EventDate); err != nil {
			return err
		}

		if needsLeaveRecalc(event) {
			triggerLeaveRecalc(event)
		}

		return nil
	})
}

func UpdateAttendanceEvent(event *AttendanceEvent, operatorID uint, operatorName string, ip string) error {
	if event.ID == 0 {
		return fmt.Errorf("id is required")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var oldEvent AttendanceEvent
		if err := tx.First(&oldEvent, event.ID).Error; err != nil {
			return err
		}

		if err := tx.Save(event).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)

		if err := createAuditInTx(tx, operatorID, operatorName, "attendance_event", event.ID, personName, "修改", oldEvent, event, ip); err != nil {
			return err
		}

		if err := RebuildDailyProjection(event.PersonID, *event.EventDate); err != nil {
			return err
		}

		if needsLeaveRecalc(event) || needsLeaveRecalc(&oldEvent) {
			triggerLeaveRecalc(event)
		}

		return nil
	})
}

func DeleteAttendanceEvent(id uint, operatorID uint, operatorName string, ip string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event AttendanceEvent
		if err := tx.First(&event, id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&event).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)

		if err := createAuditInTx(tx, operatorID, operatorName, "attendance_event", event.ID, personName, "删除", event, nil, ip); err != nil {
			return err
		}

		if event.EventDate != nil {
			if err := RebuildDailyProjection(event.PersonID, *event.EventDate); err != nil {
				return err
			}
		}

		if needsLeaveRecalc(&event) {
			triggerLeaveRecalc(&event)
		}

		return nil
	})
}

func RestoreAttendanceEvent(id uint, operatorID uint, operatorName string, ip string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event AttendanceEvent
		if err := tx.Unscoped().First(&event, id).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)

		if err := createAuditInTx(tx, operatorID, operatorName, "attendance_event", event.ID, personName, "恢复", nil, event, ip); err != nil {
			return err
		}

		if event.EventDate != nil {
			if err := RebuildDailyProjection(event.PersonID, *event.EventDate); err != nil {
				return err
			}
		}

		if needsLeaveRecalc(&event) {
			triggerLeaveRecalc(&event)
		}

		return nil
	})
}

func needsLeaveRecalc(event *AttendanceEvent) bool {
	return event.SubType == "年假" || event.SubType == "调休"
}

func triggerLeaveRecalc(event *AttendanceEvent) {
	if OnLeaveAccountRecalc != nil && event.PersonID > 0 {
		leaveType := ""
		switch event.SubType {
		case "年假":
			leaveType = "annual_leave"
		case "调休":
			leaveType = "time_off"
		}
		if leaveType != "" {
			_ = OnLeaveAccountRecalc(event.PersonID, leaveType)
		}
	}
}

func CreateCrossDayEvents(personID uint, startDate, endDate time.Time, eventType, subType string, dailyHours float64, punchTime, remark string, isSpecialApproval bool, operatorID uint, operatorName string, ip string) ([]AttendanceEvent, []string, error) {
	var events []AttendanceEvent
	var errors []string

	current := startDate
	for !current.After(endDate) {
		d := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
		event := AttendanceEvent{
			PersonID:          personID,
			EventDate:         &d,
			PunchTime:         punchTime,
			EventType:         eventType,
			SubType:           subType,
			Hours:             dailyHours,
			IsSpecialApproval: isSpecialApproval,
			Remark:            remark,
		}

		if err := CreateAttendanceEvent(&event, operatorID, operatorName, ip); err != nil {
			errors = append(errors, fmt.Sprintf("日期%s: %s", d.Format("2006-01-02"), err.Error()))
		} else {
			events = append(events, event)
		}

		current = current.AddDate(0, 0, 1)
	}

	return events, errors, nil
}

func CreateBatchEvents(personIDs []uint, eventDate time.Time, eventType, subType string, hours float64, punchTime, remark string, isSpecialApproval bool, operatorID uint, operatorName string, ip string) ([]AttendanceEvent, []string, error) {
	var events []AttendanceEvent
	var errors []string

	d := time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 0, 0, 0, 0, time.UTC)

	for _, personID := range personIDs {
		event := AttendanceEvent{
			PersonID:          personID,
			EventDate:         &d,
			PunchTime:         punchTime,
			EventType:         eventType,
			SubType:           subType,
			Hours:             hours,
			IsSpecialApproval: isSpecialApproval,
			Remark:            remark,
		}

		if err := CreateAttendanceEvent(&event, operatorID, operatorName, ip); err != nil {
			errors = append(errors, fmt.Sprintf("人员#%d: %s", personID, err.Error()))
		} else {
			events = append(events, event)
		}
	}

	return events, errors, nil
}

func GetMonthlyStatus(ms AttendanceSalaryMonthly, positionLastCalc *time.Time) projection.Status {
	return projection.CheckProjectionStatus(ms.LastCalcAt, positionLastCalc)
}

func createAuditInTx(tx *gorm.DB, operatorID uint, operatorName string, targetType string, targetID uint, targetName string, action string, beforeData, afterData interface{}, ip string) error {
	return audit.CreateAuditLog(tx, operatorID, operatorName, targetType, targetID, targetName, action, beforeData, afterData, ip)
}
