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
			weightedBase += s.BaseSalary * segDays
			weightedMeal += s.MealAllowance * segDays
			salaryDaysTotal += float64(s.SalaryDays) * segDays
		}

		if totalDays == 0 {
			return fmt.Errorf("当月无在职记录")
		}

		weightedBase = (weightedBase / totalDays)
		weightedMeal = (weightedMeal / totalDays)
		avgSalaryDays := int(salaryDaysTotal / totalDays)

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
