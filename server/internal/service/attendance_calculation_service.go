package service

import (
	"fmt"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"
)

func CalculateMonthlyAttendance(personID uint, month string) (*model.AttendanceCalculationMonthly, error) {
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, err
	}
	monthEnd := monthStart.AddDate(0, 1, -1)
	monthStartDate := utils.DateOnlyFromTime(monthStart)
	monthEndDate := utils.DateOnlyFromTime(monthEnd)

	var snapshots []model.PositionSnapshot
	dao.DB.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
		personID, monthEndDate, monthStartDate).Find(&snapshots)

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("当月无在职记录")
	}

	var totalDays float64
	var weightedBase, weightedMeal float64
	var salaryDaysTotal float64
	var hasAttendanceBonus bool = true

	for _, s := range snapshots {
		segStart := s.EffectiveStartDate.Time()
		segEnd := s.EffectiveEndDate.Time()

		calcStart := monthStart
		if segStart.After(calcStart) {
			calcStart = segStart
		}
		calcEnd := monthEnd
		if segEnd.Before(calcEnd) {
			calcEnd = segEnd
		}

		segDays := calcEnd.Sub(calcStart).Hours()/24 + 1
		if segDays <= 0 {
			continue
		}

		totalDays += segDays
		weightedBase += s.BaseSalary * segDays
		weightedMeal += s.MealAllowance * segDays
		salaryDaysTotal += float64(s.SalaryDays) * segDays

		if !s.IsActive || !s.HasAttendanceBonus {
			hasAttendanceBonus = false
		}
	}

	if totalDays == 0 {
		return nil, fmt.Errorf("当月无在职记录")
	}

	weightedBase = (weightedBase / totalDays)
	weightedMeal = (weightedMeal / totalDays)
	avgSalaryDays := int(salaryDaysTotal / totalDays)

	var totalWorkHours, overtimeWorkday, overtimeHoliday float64
	var hasPersonalLeave bool
	var totalViolations int

	var dailyProjections []model.AttendanceDailyProjection
	dao.DB.Where("person_id = ? AND work_date >= ? AND work_date <= ?",
		personID, monthStartDate, monthEndDate).Find(&dailyProjections)

	for _, d := range dailyProjections {
		totalWorkHours += d.WorkHours
		overtimeWorkday += d.OvertimeWorkdayHours
		overtimeHoliday += d.OvertimeHolidayHours
		if d.HasPersonalLeave {
			hasPersonalLeave = true
		}
		totalViolations += d.ViolationCount
	}

	workHoursPerDay := getWorkHoursPerDay()

	attendanceSalary := 0.0
	if avgSalaryDays > 0 {
		attendanceSalary = totalWorkHours * (weightedBase / float64(avgSalaryDays) / workHoursPerDay)
	}
	attendanceSalary = utils.RoundTwoDecimal(attendanceSalary)

	overtimeWorkdaySalary := 0.0
	if avgSalaryDays > 0 {
		overtimeWorkdaySalary = overtimeWorkday * (weightedBase + weightedMeal) / float64(avgSalaryDays) / workHoursPerDay * getOvertimeWorkdayRatio()
	}
	overtimeWorkdaySalary = utils.RoundTwoDecimal(overtimeWorkdaySalary)

	overtimeHolidaySalary := 0.0
	if avgSalaryDays > 0 {
		overtimeHolidaySalary = overtimeHoliday * (weightedBase + weightedMeal) / float64(avgSalaryDays) / workHoursPerDay * getOvertimeHolidayRatio()
	}
	overtimeHolidaySalary = utils.RoundTwoDecimal(overtimeHolidaySalary)

	attendanceBonus := 0.0
	if hasAttendanceBonus && !hasPersonalLeave {
		bonus := (totalWorkHours/workHoursPerDay - float64(totalViolations)) * getAttendanceBonusDaily()
		if bonus > 0 {
			attendanceBonus = utils.RoundTwoDecimal(bonus)
		}
	}

	result := model.AttendanceCalculationMonthly{
		PersonID:                   personID,
		BelongMonth:                month,
		SalaryDays:                 avgSalaryDays,
		WeightedBaseSalary:         utils.RoundTwoDecimal(weightedBase),
		WeightedMealAllowance:      utils.RoundTwoDecimal(weightedMeal),
		TotalWorkHours:             totalWorkHours,
		TotalOvertimeWorkdayHours:  overtimeWorkday,
		TotalOvertimeHolidayHours:  overtimeHoliday,
		AttendanceSalary:           attendanceSalary,
		OvertimeWorkdaySalary:      overtimeWorkdaySalary,
		OvertimeHolidaySalary:      overtimeHolidaySalary,
		HasPersonalLeaveMonth:      hasPersonalLeave,
		TotalViolationCount:        totalViolations,
		AttendanceBonus:            attendanceBonus,
		LastCalcAt:                 utils.DateOnlyFromTime(time.Now()),
	}

	dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).Delete(&model.AttendanceCalculationMonthly{})
	dao.DB.Create(&result)

	return &result, nil
}

func GetMonthlyList(month string, personID uint, pageNum, pageSize int) ([]model.AttendanceCalculationMonthly, int64, error) {
	tx := dao.DB.Model(&model.AttendanceCalculationMonthly{})
	if month != "" {
		tx = tx.Where("belong_month = ?", month)
	}
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	var total int64
	tx.Count(&total)
	var list []model.AttendanceCalculationMonthly
	offset := (pageNum - 1) * pageSize
	err := tx.Order("person_id ASC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func IsAttendanceMonthlyStale(calc *model.AttendanceCalculationMonthly) string {
	monthStart, _ := utils.MonthStart(calc.BelongMonth)
	monthEnd, _ := utils.MonthEnd(calc.BelongMonth)
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthEnd)

	var maxDailyLastCalc utils.DateOnly
	dao.DB.Model(&model.AttendanceDailyProjection{}).
		Where("person_id = ? AND work_date >= ? AND work_date <= ?",
			calc.PersonID, monthStartD, monthEndD).
		Select("COALESCE(MAX(last_calc_at), '0001-01-01')").Scan(&maxDailyLastCalc)
	if maxDailyLastCalc.Time().After(calc.LastCalcAt.Time()) {
		return "data_changed"
	}

	var maxSnapLastCalc utils.DateOnly
	dao.DB.Model(&model.PositionSnapshot{}).
		Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			calc.PersonID, monthEndD, monthStartD).
		Select("COALESCE(MAX(last_calc_at), '0001-01-01')").Scan(&maxSnapLastCalc)
	if maxSnapLastCalc.Time().After(calc.LastCalcAt.Time()) {
		return "data_changed"
	}

	return "calculated"
}

func CalculateMonthlyBatch(month string, personIDs []uint) (int, int, error) {
	success, fail := 0, 0
	for _, pid := range personIDs {
		if _, err := CalculateMonthlyAttendance(pid, month); err != nil {
			fail++
		} else {
			success++
		}
	}
	return success, fail, nil
}
