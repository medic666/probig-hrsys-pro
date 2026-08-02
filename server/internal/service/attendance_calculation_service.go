package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// GetActivePersonIDsInMonth 当月在职人员 ID 列表（批量核算人选）
func GetActivePersonIDsInMonth(month string) []uint {
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		return nil
	}
	monthEnd := monthStart.AddDate(0, 1, -1)
	var personIDs []uint
	dao.DB.Model(&model.PositionSnapshot{}).
		Select("DISTINCT person_id").
		Where("effective_start_date <= ? AND effective_end_date >= ? AND is_active = true",
			utils.DateOnlyFromTime(monthEnd), utils.DateOnlyFromTime(monthStart)).
		Pluck("person_id", &personIDs)
	return personIDs
}

func CalculateMonthlyAttendance(ctx context.Context, personID uint, month string) (*model.AttendanceCalculationMonthly, error) {
	var result *model.AttendanceCalculationMonthly
	var oldJSON, newJSON, personName string
	err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		monthStart, err := time.Parse("2006-01", month)
		if err != nil {
			return err
		}
		monthEnd := monthStart.AddDate(0, 1, -1)
		monthStartDate := utils.DateOnlyFromTime(monthStart)
		monthEndDate := utils.DateOnlyFromTime(monthEnd)

		var pendingCount int64
		tx.Model(&model.AttendanceDailyProjection{}).
			Where("person_id = ? AND work_date >= ? AND work_date <= ? AND status = ?",
				personID, monthStartDate, monthEndDate, "pending").Count(&pendingCount)
		if pendingCount > 0 {
			return fmt.Errorf("当月有 %d 天待确认的考勤记录，请先完成核实", pendingCount)
		}

		var snapshots []model.PositionSnapshot
		tx.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			personID, monthEndDate, monthStartDate).Find(&snapshots)

		if len(snapshots) == 0 {
			return fmt.Errorf("当月无在职记录")
		}

		var totalDays float64
		var weightedBase, weightedMeal float64
		var salaryDaysTotal float64
		var hasAttendanceBonus bool = true

		for _, s := range snapshots {
			segStart := s.EffectiveStartDate.Time()
			segEnd := s.EffectiveEndDate.Time()

			calcStart := monthStart
			if segStart.After(calcStart) { calcStart = segStart }
			calcEnd := monthEnd
			if segEnd.Before(calcEnd) { calcEnd = segEnd }

			segDays := calcEnd.Sub(calcStart).Hours()/24 + 1
			if segDays <= 0 { continue }
			if !s.IsActive || !s.HasAttendanceBonus {
				hasAttendanceBonus = false
			}
		}

		for _, s := range snapshots {
			segStart := s.EffectiveStartDate.Time()
			segEnd := s.EffectiveEndDate.Time()

			calcStart := monthStart
			if segStart.After(calcStart) { calcStart = segStart }
			calcEnd := monthEnd
			if segEnd.Before(calcEnd) { calcEnd = segEnd }

			segDays := calcEnd.Sub(calcStart).Hours()/24 + 1
			if segDays <= 0 { continue }
			if !s.IsActive { continue }

			totalDays += segDays
			weightedBase = utils.DecimalAdd(weightedBase, utils.DecimalMul(s.BaseSalary, segDays))
			weightedMeal = utils.DecimalAdd(weightedMeal, utils.DecimalMul(s.MealAllowance, segDays))
			salaryDaysTotal += s.SalaryDays * segDays
		}

		if totalDays == 0 {
			return fmt.Errorf("当月无在职记录")
		}

		weightedBase = (weightedBase / totalDays)
		weightedMeal = (weightedMeal / totalDays)
		avgSalaryDays := utils.RoundTwoDecimal(salaryDaysTotal / totalDays)

		var totalWorkHours, overtimeWorkday, overtimeHoliday float64
		var hasPersonalLeave bool
		var totalViolations int

		var dailyProjections []model.AttendanceDailyProjection
		tx.Where("person_id = ? AND work_date >= ? AND work_date <= ?",
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
			attendanceSalary = totalWorkHours * (weightedBase / avgSalaryDays / workHoursPerDay)
		}
		attendanceSalary = utils.RoundTwoDecimal(attendanceSalary)

		overtimeWorkdaySalary := 0.0
		if avgSalaryDays > 0 {
			overtimeWorkdaySalary = overtimeWorkday * (weightedBase + weightedMeal) / avgSalaryDays / workHoursPerDay * getOvertimeWorkdayRatio()
		}
		overtimeWorkdaySalary = utils.RoundTwoDecimal(overtimeWorkdaySalary)

		overtimeHolidaySalary := 0.0
		if avgSalaryDays > 0 {
			overtimeHolidaySalary = overtimeHoliday * (weightedBase + weightedMeal) / avgSalaryDays / workHoursPerDay * getOvertimeHolidayRatio()
		}
		overtimeHolidaySalary = utils.RoundTwoDecimal(overtimeHolidaySalary)

		attendanceBonus := 0.0
		if hasAttendanceBonus && !hasPersonalLeave {
			bonus := (totalWorkHours/workHoursPerDay - float64(totalViolations)) * getAttendanceBonusDaily()
			if bonus > 0 {
				attendanceBonus = utils.RoundTwoDecimal(bonus)
			}
		}

		calc := model.AttendanceCalculationMonthly{
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
			LastCalcAt: time.Now(),
		}

		// 核算前旧结果快照（审计 before）
		oldJSON = ""
		var oldCalc model.AttendanceCalculationMonthly
		if err := tx.Where("person_id = ? AND belong_month = ?", personID, month).First(&oldCalc).Error; err == nil {
			if b, err := json.Marshal(oldCalc); err == nil {
				oldJSON = string(b)
			}
		}

		if err := tx.Where("person_id = ? AND belong_month = ?", personID, month).Delete(&model.AttendanceCalculationMonthly{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&calc).Error; err != nil {
			return err
		}

		b, _ := json.Marshal(calc); newJSON = string(b)
		tx.Table("persons").Select("name").Where("id = ?", personID).Scan(&personName)
		result = &calc
		return nil
	})
	// 核算审计（每人一条），事务提交后写入，避免与业务事务的 SQLite 写锁竞争
	if err == nil {
		dao.WriteBusinessAudit(ctx, "核算", "attendance_calculation_monthly", personID, personName, oldJSON, newJSON)
	}
	return result, err
}

// MonthlyListQuery 月度考勤核算列表查询（列表与导出共用）
type MonthlyListQuery struct {
	PageNum  int
	PageSize int
	Month    string
	PersonID uint
}

func GetMonthlyList(q MonthlyListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.AttendanceCalculationMonthly{})
	if q.Month != "" {
		tx = tx.Where("belong_month = ?", q.Month)
	}
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}
	var total int64
	tx.Count(&total)
	var list []model.AttendanceCalculationMonthly
	offset := (q.PageNum - 1) * q.PageSize
	err := tx.Order("person_id ASC").Offset(offset).Limit(q.PageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	ids := make([]uint, len(list))
	for i, s := range list {
		ids[i] = s.PersonID
	}
	nameMap := PersonNameMap(ids)

	result := make([]map[string]interface{}, len(list))
	for i, s := range list {
		item := map[string]interface{}{
			"id":                        s.ID,
			"person_id":                 s.PersonID,
			"person_name":               nameMap[s.PersonID],
			"belong_month":              s.BelongMonth,
			"salary_days":               s.SalaryDays,
			"weighted_base_salary":      s.WeightedBaseSalary,
			"weighted_meal_allowance":   s.WeightedMealAllowance,
			"total_work_hours":          s.TotalWorkHours,
			"total_overtime_workday_hours": s.TotalOvertimeWorkdayHours,
			"total_overtime_holiday_hours": s.TotalOvertimeHolidayHours,
			"attendance_salary":         s.AttendanceSalary,
			"overtime_workday_salary":   s.OvertimeWorkdaySalary,
			"overtime_holiday_salary":   s.OvertimeHolidaySalary,
			"has_personal_leave_month":  s.HasPersonalLeaveMonth,
			"total_violation_count":     s.TotalViolationCount,
			"attendance_bonus":          s.AttendanceBonus,
			"last_calc_at":              s.LastCalcAt,
			"status":                    IsAttendanceMonthlyStale(&s),
		}
		result[i] = item
	}
	return result, total, nil
}

func latestTime(times []*time.Time) time.Time {
	var max time.Time
	for _, t := range times {
		if t != nil && t.After(max) {
			max = *t
		}
	}
	return max
}

func IsAttendanceMonthlyStale(calc *model.AttendanceCalculationMonthly) string {
	monthStart, _ := utils.MonthStart(calc.BelongMonth)
	monthEnd, _ := utils.MonthEnd(calc.BelongMonth)
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthEnd)

	var dailyTimes []*time.Time
	dao.DB.Model(&model.AttendanceDailyProjection{}).
		Where("person_id = ? AND work_date >= ? AND work_date <= ?",
			calc.PersonID, monthStartD, monthEndD).
		Pluck("last_calc_at", &dailyTimes)
	if latestTime(dailyTimes).After(calc.LastCalcAt) {
		return "data_changed"
	}

	var snapTimes []*time.Time
	dao.DB.Model(&model.PositionSnapshot{}).
		Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			calc.PersonID, monthEndD, monthStartD).
		Pluck("last_calc_at", &snapTimes)
	if latestTime(snapTimes).After(calc.LastCalcAt) {
		return "data_changed"
	}

	return "calculated"
}

func CalculateMonthlyBatch(ctx context.Context, month string, personIDs []uint) (int, int, error) {
	success, fail := 0, 0
	for _, pid := range personIDs {
		if _, err := CalculateMonthlyAttendance(ctx, pid, month); err != nil {
			fail++
		} else {
			success++
		}
	}
	return success, fail, nil
}
