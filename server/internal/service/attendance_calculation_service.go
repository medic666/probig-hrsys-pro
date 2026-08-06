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
	var audited bool
	err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		monthStart, err := time.Parse("2006-01", month)
		if err != nil {
			return err
		}
		monthEnd := monthStart.AddDate(0, 1, -1)
		monthStartDate := utils.DateOnlyFromTime(monthStart)
		monthEndDate := utils.DateOnlyFromTime(monthEnd)

		// ① 存在待确认日记工时 → 数据未定稿，跳过核算并置空（无值）
		var pendingCount int64
		tx.Model(&model.AttendanceDailyProjection{}).
			Where("person_id = ? AND work_date >= ? AND work_date <= ? AND status = ?",
				personID, monthStartDate, monthEndDate, "pending").Count(&pendingCount)
		if pendingCount > 0 {
			return clearMonthlyCalcInTx(tx, personID, month, &oldJSON, &personName, &audited)
		}

		var snapshots []model.PositionSnapshot
		tx.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			personID, monthEndDate, monthStartDate).Find(&snapshots)

		// ② 当月无职务快照 → 无核算对象，置空（无值）
		if len(snapshots) == 0 {
			return clearMonthlyCalcInTx(tx, personID, month, &oldJSON, &personName, &audited)
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

		// ③ 当月无活跃在职段 → 置空（无值）
		if totalDays == 0 {
			return clearMonthlyCalcInTx(tx, personID, month, &oldJSON, &personName, &audited)
		}

		weightedBase = (weightedBase / totalDays)
		weightedMeal = (weightedMeal / totalDays)
		avgSalaryDays := utils.RoundTwoDecimal(salaryDaysTotal / totalDays)

		// ④ 计薪天数未配置 → 失败（需人工设置职务计薪天数后重新核算）
		if avgSalaryDays == 0 {
			return fmt.Errorf("计薪天数未设置，请先配置该人员的职务计薪天数后重新核算")
		}

		var totalWorkHours, overtimeWorkday, overtimeHoliday float64
		var hasPersonalLeave bool
		var totalViolations int

		var dailyProjections []model.AttendanceDailyProjection
		tx.Where("person_id = ? AND work_date >= ? AND work_date <= ?",
			personID, monthStartDate, monthEndDate).Find(&dailyProjections)

		// ⑤ 当月无任何日记工时投影 → 空结果，置空（无值）
		if len(dailyProjections) == 0 {
			return clearMonthlyCalcInTx(tx, personID, month, &oldJSON, &personName, &audited)
		}

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

		// ⑥ 有值：物理删旧记录 + 新建（幂等覆盖，历史由审计快照留存）
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
		audited = true
		result = &calc
		return nil
	})
	// 核算审计（每人一条），事务提交后写入，避免与业务事务的 SQLite 写锁竞争
	if err == nil && audited {
		dao.WriteBusinessAudit(ctx, "核算", "attendance_calculation_monthly", personID, personName, oldJSON, newJSON)
	}
	return result, err
}

// clearMonthlyCalcInTx 置空月考勤核算（无值语义）：物理删除旧记录（如有），
// 返回 nil 表示"核算完成但结果为空"；仅当存在旧记录时登记置空审计。
func clearMonthlyCalcInTx(tx *gorm.DB, personID uint, month string, oldJSON, personName *string, audited *bool) error {
	var oldCalc model.AttendanceCalculationMonthly
	hasOld := tx.Where("person_id = ? AND belong_month = ?", personID, month).First(&oldCalc).Error == nil
	if err := tx.Where("person_id = ? AND belong_month = ?", personID, month).Delete(&model.AttendanceCalculationMonthly{}).Error; err != nil {
		return err
	}
	if hasOld {
		if b, err := json.Marshal(oldCalc); err == nil {
			*oldJSON = string(b)
		}
		tx.Table("persons").Select("name").Where("id = ?", personID).Scan(personName)
		*audited = true
	}
	return nil
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

	var snapTimes []*time.Time
	dao.DB.Model(&model.PositionSnapshot{}).
		Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			calc.PersonID, monthEndD, monthStartD).
		Pluck("last_calc_at", &snapTimes)

	if IsStaleAfter(calc.LastCalcAt, dailyTimes, snapTimes) {
		return "data_changed"
	}
	return "calculated"
}

// CalculateMonthlyBatch 批量考勤核算：三态计数——
// hasValue 有结果（含 0）/ empty 空结果（置空）/ fail 失败（需人工干预）
func CalculateMonthlyBatch(ctx context.Context, month string, personIDs []uint) (hasValue, empty, fail int, err error) {
	for _, pid := range personIDs {
		r, calcErr := CalculateMonthlyAttendance(ctx, pid, month)
		if calcErr != nil {
			fail++
		} else if r == nil {
			empty++
		} else {
			hasValue++
		}
	}
	return hasValue, empty, fail, nil
}

// GetAttendanceMonthlyBadges 月度考勤核算徽章（指定月份）：无核算记录 gray；
// 核算过期（投影/快照 last_calc_at 晚于核算时间）orange；已核算未过期 green。
// stale 判定与 IsAttendanceMonthlyStale 语义等价，但按人员批量聚合派生层最后计算时间，
// 消除逐人循环查询（N+1）。
func GetAttendanceMonthlyBadges(month string) ([]PersonBadge, error) {
	monthStart, err := utils.MonthStart(month)
	if err != nil {
		return nil, err
	}
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthStart.AddDate(0, 1, -1))

	var calcs []model.AttendanceCalculationMonthly
	if err := dao.DB.Where("belong_month = ?", month).Find(&calcs).Error; err != nil {
		return nil, err
	}
	calcMap := make(map[uint]model.AttendanceCalculationMonthly, len(calcs))
	for _, c := range calcs {
		calcMap[c.PersonID] = c
	}

	type calcTimeRow struct {
		PersonID   uint
		LastCalcAt time.Time
	}
	// 批量拉取当月投影 / 覆盖当月的职务快照最后计算时间（一次查询 + Go 聚合，消除 N+1）
	projLatest := make(map[uint]time.Time)
	var projRows []calcTimeRow
	if err := dao.DB.Model(&model.AttendanceDailyProjection{}).
		Select("person_id, last_calc_at").
		Where("work_date >= ? AND work_date <= ?", monthStartD, monthEndD).
		Scan(&projRows).Error; err != nil {
		return nil, err
	}
	for _, r := range projRows {
		if r.LastCalcAt.After(projLatest[r.PersonID]) {
			projLatest[r.PersonID] = r.LastCalcAt
		}
	}
	snapLatest := make(map[uint]time.Time)
	var snapRows []calcTimeRow
	if err := dao.DB.Model(&model.PositionSnapshot{}).
		Select("person_id, last_calc_at").
		Where("effective_start_date <= ? AND effective_end_date >= ?", monthEndD, monthStartD).
		Scan(&snapRows).Error; err != nil {
		return nil, err
	}
	for _, r := range snapRows {
		if r.LastCalcAt.After(snapLatest[r.PersonID]) {
			snapLatest[r.PersonID] = r.LastCalcAt
		}
	}

	var personIDs []uint
	if err := dao.DB.Table("persons").Where("deleted_at IS NULL").Pluck("id", &personIDs).Error; err != nil {
		return nil, err
	}
	result := make([]PersonBadge, 0, len(personIDs))
	for _, pid := range personIDs {
		calc, ok := calcMap[pid]
		if !ok {
			result = append(result, PersonBadge{PersonID: pid, Level: "gray"})
			continue
		}
		level := "green"
		latest := projLatest[pid]
		if t, ok := snapLatest[pid]; ok && t.After(latest) {
			latest = t
		}
		if latest.After(calc.LastCalcAt) {
			level = "orange"
		}
		result = append(result, PersonBadge{PersonID: pid, Level: level})
	}
	return result, nil
}
