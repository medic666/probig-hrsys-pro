package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

var ErrAttendanceNotCalculated = errors.New("未完成月度考勤核算，请先进行考勤核算")

func CalculateSalary(ctx context.Context, personID uint, month string, operatorID uint, operatorName string) (*model.SalarySummary, error) {
	if err := EnsureOwnPerson(ctx, personID); err != nil {
		return nil, err
	}
	var result *model.SalarySummary
	var oldJSON, newJSON, personName string
	var audited bool
	err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		monthStart, _ := time.Parse("2006-01", month)
		monthEnd := monthStart.AddDate(0, 1, -1)
		monthStartD := utils.DateOnlyFromTime(monthStart)
		monthEndD := utils.DateOnlyFromTime(monthEnd)

		var calc model.AttendanceCalculationMonthly
		err := tx.Where("person_id = ? AND belong_month = ?", personID, month).First(&calc).Error
		if err != nil {
			// 考勤核算为空（无值）→ 空值传递：删除旧工资汇总，结果置空
			return clearSalarySummaryInTx(tx, personID, month, &oldJSON, &personName, &audited)
		}

		var snapshots []model.PositionSnapshot
		tx.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			personID, monthEndD, monthStartD).Find(&snapshots)

		if len(snapshots) == 0 {
			return fmt.Errorf("当月无在职记录")
		}

		var activeDays float64
		var wPerfBase, wPost, wMeal, wHousing, wTransport, wHighTemp, wInsComp, wFundComp, wSSDeduct, wHFDeduct float64

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
			if !s.IsActive {
				continue
			}
			activeDays += segDays
			wPerfBase = utils.DecimalAdd(wPerfBase, utils.DecimalMul(s.PerformanceSalary, segDays))
			wPost = utils.DecimalAdd(wPost, utils.DecimalMul(s.PostAllowance, segDays))
			wMeal = utils.DecimalAdd(wMeal, utils.DecimalMul(s.MealAllowance, segDays))
			wHousing = utils.DecimalAdd(wHousing, utils.DecimalMul(s.HousingAllowance, segDays))
			wTransport = utils.DecimalAdd(wTransport, utils.DecimalMul(s.TransportAllowance, segDays))
			wHighTemp = utils.DecimalAdd(wHighTemp, utils.DecimalMul(s.HighTempAllowance, segDays))
			wInsComp = utils.DecimalAdd(wInsComp, utils.DecimalMul(s.InsuranceCompensation, segDays))
			wFundComp = utils.DecimalAdd(wFundComp, utils.DecimalMul(s.FundCompensation, segDays))
			wSSDeduct = utils.DecimalAdd(wSSDeduct, utils.DecimalMul(s.SocialSecurityDeduct, segDays))
			wHFDeduct = utils.DecimalAdd(wHFDeduct, utils.DecimalMul(s.HousingFundDeduct, segDays))
		}

		if activeDays == 0 {
			return fmt.Errorf("当月无在职记录")
		}

		if calc.SalaryDays == 0 {
			return fmt.Errorf("计薪天数为0")
		}

		salaryDays := calc.SalaryDays
		totalCalendarDays := monthEnd.Sub(monthStart).Hours()/24 + 1
		isFullMonth := activeDays == totalCalendarDays
		attendanceDays := calc.TotalWorkHours / getWorkHoursPerDay()

		// 非全月（入职/离职月）：补贴与绩效统一按实际出勤比例折算
		// 全月在职：wXxx/activeDays 直取定额；入职/离职月：(wXxx/activeDays) × (attendanceDays/salaryDays)
		subsidyRatio := 1.0
		if !isFullMonth {
			subsidyRatio = attendanceDays / salaryDays
		}

		post := utils.RoundTwoDecimal(wPost / activeDays * subsidyRatio)
		meal := utils.RoundTwoDecimal(wMeal / activeDays * subsidyRatio)
		housing := utils.RoundTwoDecimal(wHousing / activeDays * subsidyRatio)
		transport := utils.RoundTwoDecimal(wTransport / activeDays * subsidyRatio)
		highTemp := utils.RoundTwoDecimal(wHighTemp / activeDays * subsidyRatio)
		if !isHighTempMonth(month) {
			highTemp = 0
		}
		insComp := utils.RoundTwoDecimal(wInsComp / activeDays * subsidyRatio)
		fundComp := utils.RoundTwoDecimal(wFundComp / activeDays * subsidyRatio)
		ssDeduct := utils.RoundTwoDecimal(wSSDeduct / activeDays)
		hfDeduct := utils.RoundTwoDecimal(wHFDeduct / activeDays)

		var perfCoeff float64 = 1
		var salesCommission, rewardPunishment, borrowingRepayment, taxDeduct float64
		var salaryEvents []model.SalaryEvent
		tx.Where("person_id = ? AND belong_month = ?", personID, month).Order("seq DESC").Find(&salaryEvents)

		maxCoeffSeq := 0
		for _, e := range salaryEvents {
			switch e.EventType {
			case "绩效系数":
				if e.Seq > maxCoeffSeq {
					perfCoeff = e.Amount
					maxCoeffSeq = e.Seq
				}
			case "提成":
				salesCommission = utils.DecimalAdd(salesCommission, e.Amount)
			case "奖惩":
				rewardPunishment = utils.DecimalAdd(rewardPunishment, e.Amount)
			case "预支还款":
				borrowingRepayment = utils.DecimalAdd(borrowingRepayment, e.Amount)
			case "个税扣除":
				taxDeduct = utils.DecimalAdd(taxDeduct, e.Amount)
			}
		}

		perfSalary := utils.RoundTwoDecimal(wPerfBase / activeDays * subsidyRatio * perfCoeff)

		var carryoverDeductHours float64
		tx.Model(&model.AnnualLeaveAccountEvent{}).
			Where("person_id = ? AND event_type = ? AND effective_date >= ? AND effective_date <= ?",
				personID, "carryover_deduct", monthStartD, monthEndD).
			Select("COALESCE(SUM(hours), 0)").Scan(&carryoverDeductHours)

		workHoursPerDay := getWorkHoursPerDay()
		holidayRatio := getOvertimeHolidayRatio()
		carryoverSalary := 0.0
		if calc.SalaryDays > 0 {
			carryoverSalary = utils.RoundTwoDecimal(carryoverDeductHours * (calc.WeightedBaseSalary + calc.WeightedMealAllowance) / calc.SalaryDays / workHoursPerDay * holidayRatio)
		}

		salesCommission = utils.RoundTwoDecimal(salesCommission)
		rewardPunishment = utils.RoundTwoDecimal(rewardPunishment)
		borrowingRepayment = utils.RoundTwoDecimal(borrowingRepayment)
		taxDeduct = utils.RoundTwoDecimal(taxDeduct)

		finalSalary := utils.RoundTwoDecimal(
			calc.AttendanceSalary + calc.OvertimeWorkdaySalary + calc.OvertimeHolidaySalary +
				carryoverSalary + calc.AttendanceBonus + perfSalary +
				post + meal + housing + transport + highTemp + insComp + fundComp +
				salesCommission + rewardPunishment + borrowingRepayment -
				ssDeduct - hfDeduct - taxDeduct,
		)

		summary := model.SalarySummary{
			PersonID:                   personID,
			BelongMonth:                month,
			SalaryDays:                 calc.SalaryDays,
			WeightedBaseSalary:         calc.WeightedBaseSalary,
			TotalWorkHours:             calc.TotalWorkHours,
			TotalOvertimeWorkdayHours:  calc.TotalOvertimeWorkdayHours,
			TotalOvertimeHolidayHours:  calc.TotalOvertimeHolidayHours,
			AttendanceSalary:           calc.AttendanceSalary,
			OvertimeWorkdaySalary:      calc.OvertimeWorkdaySalary,
			OvertimeHolidaySalary:      calc.OvertimeHolidaySalary,
			AnnualLeaveCarryoverDeduct: carryoverDeductHours,
			AnnualLeaveCarryoverSalary: carryoverSalary,
			AttendanceBonus:            calc.AttendanceBonus,
			PerformanceSalary:          perfSalary,
			PostAllowance:              post,
			MealAllowance:              meal,
			HousingAllowance:           housing,
			TransportAllowance:         transport,
			HighTempAllowance:          highTemp,
			InsuranceCompensation:      insComp,
			FundCompensation:           fundComp,
			SalesCommission:            salesCommission,
			RewardPunishment:           rewardPunishment,
			BorrowingRepayment:         borrowingRepayment,
			SocialSecurityDeduct:       ssDeduct,
			HousingFundDeduct:          hfDeduct,
			TaxDeduct:                  taxDeduct,
			FinalSalary:                finalSalary,
			LastCalcAt:                 time.Now(),
		}

		var maxVersion int
		tx.Model(&model.SalarySummaryVersion{}).
			Where("person_id = ? AND belong_month = ?", personID, month).
			Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

		batchNo := "SAL-" + month + "-" + strconv.FormatInt(time.Now().Unix(), 10)

		version := model.SalarySummaryVersion{
			PersonID:                   personID,
			BelongMonth:                month,
			Version:                    maxVersion + 1,
			CalcBatchNo:                batchNo,
			OperatorID:                 operatorID,
			OperatorName:               operatorName,
			SalaryDays:                 calc.SalaryDays,
			WeightedBaseSalary:         calc.WeightedBaseSalary,
			TotalWorkHours:             calc.TotalWorkHours,
			TotalOvertimeWorkdayHours:  calc.TotalOvertimeWorkdayHours,
			TotalOvertimeHolidayHours:  calc.TotalOvertimeHolidayHours,
			AttendanceSalary:           calc.AttendanceSalary,
			OvertimeWorkdaySalary:      calc.OvertimeWorkdaySalary,
			OvertimeHolidaySalary:      calc.OvertimeHolidaySalary,
			AnnualLeaveCarryoverDeduct: carryoverDeductHours,
			AnnualLeaveCarryoverSalary: carryoverSalary,
			AttendanceBonus:            calc.AttendanceBonus,
			PerformanceSalary:          perfSalary,
			PostAllowance:              post,
			MealAllowance:              meal,
			HousingAllowance:           housing,
			TransportAllowance:         transport,
			HighTempAllowance:          highTemp,
			InsuranceCompensation:      insComp,
			FundCompensation:           fundComp,
			SalesCommission:            salesCommission,
			RewardPunishment:           rewardPunishment,
			BorrowingRepayment:         borrowingRepayment,
			SocialSecurityDeduct:       ssDeduct,
			HousingFundDeduct:          hfDeduct,
			TaxDeduct:                  taxDeduct,
			FinalSalary:                finalSalary,
		}
		if err := tx.Create(&version).Error; err != nil {
			return err
		}

		// 核算前旧汇总快照（审计 before）
		oldJSON = ""
		var oldSummary model.SalarySummary
		if err := tx.Where("person_id = ? AND belong_month = ?", personID, month).First(&oldSummary).Error; err == nil {
			if b, err := json.Marshal(oldSummary); err == nil {
				oldJSON = string(b)
			}
		}

		if err := tx.Where("person_id = ? AND belong_month = ?", personID, month).Delete(&model.SalarySummary{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&summary).Error; err != nil {
			return err
		}

		b, _ := json.Marshal(summary)
		newJSON = string(b)
		tx.Table("persons").Select("name").Where("id = ?", personID).Scan(&personName)
		audited = true
		result = &summary
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 核算审计（每人一条），事务提交后写入，避免与业务事务的 SQLite 写锁竞争
	if audited {
		dao.WriteBusinessAudit(ctx, "核算", "salary_summaries", personID, personName, oldJSON, newJSON)
	}
	return result, nil
}

// clearSalarySummaryInTx 置空工资汇总（空值传递语义）：物理删除旧汇总（如有），
// 返回 nil 表示"核算完成但结果为空"；仅当存在旧记录时登记置空审计。
func clearSalarySummaryInTx(tx *gorm.DB, personID uint, month string, oldJSON, personName *string, audited *bool) error {
	var oldSummary model.SalarySummary
	hasOld := tx.Where("person_id = ? AND belong_month = ?", personID, month).First(&oldSummary).Error == nil
	if err := tx.Where("person_id = ? AND belong_month = ?", personID, month).Delete(&model.SalarySummary{}).Error; err != nil {
		return err
	}
	if hasOld {
		if b, err := json.Marshal(oldSummary); err == nil {
			*oldJSON = string(b)
		}
		tx.Table("persons").Select("name").Where("id = ?", personID).Scan(personName)
		*audited = true
	}
	return nil
}

// CalculateSalaryBatch 批量工资核算：三态计数——
// hasValue 有结果 / empty 空结果（考勤空→置空）/ fail 失败（需人工干预）
func CalculateSalaryBatch(ctx context.Context, month string, personIDs []uint, operatorID uint, operatorName string) (hasValue, empty, fail int, err error) {
	personIDs = ScopePersonIDs(ctx, personIDs)
	for _, pid := range personIDs {
		r, calcErr := CalculateSalary(ctx, pid, month, operatorID, operatorName)
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

// SalarySummaryListQuery 工资汇总列表查询（列表与导出共用）
type SalarySummaryListQuery struct {
	PageNum  int
	PageSize int
	Month    string
	Months   []string
	PersonID uint
	Status   string
}

func GetSalarySummaries(ctx context.Context, q SalarySummaryListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DBFromContext(ctx).Model(&model.SalarySummary{})
	tx = OwnFilter(ctx, tx, "person_id")
	if q.Month != "" {
		tx = tx.Where("belong_month = ?", q.Month)
	}
	if len(q.Months) > 0 {
		tx = tx.Where("belong_month IN ?", q.Months)
	}
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}

	// 状态筛选：status 为逐行计算的派生值，需全量拉取→算状态→过滤→内存分页（保证分页总数正确）
	if q.Status != "" {
		var all []model.SalarySummary
		if err := tx.Order("belong_month DESC, person_id ASC").Find(&all).Error; err != nil {
			return nil, 0, err
		}
		rows := buildSalarySummaryRows(all)
		filtered := rows[:0]
		for _, r := range rows {
			if r["status"] == q.Status {
				filtered = append(filtered, r)
			}
		}
		start := (q.PageNum - 1) * q.PageSize
		if start > len(filtered) {
			start = len(filtered)
		}
		end := start + q.PageSize
		if end > len(filtered) {
			end = len(filtered)
		}
		return filtered[start:end], int64(len(filtered)), nil
	}

	var total int64
	tx.Count(&total)
	var list []model.SalarySummary
	offset := (q.PageNum - 1) * q.PageSize
	err := tx.Order("belong_month DESC, person_id ASC").Offset(offset).Limit(q.PageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return buildSalarySummaryRows(list), total, nil
}

// buildSalarySummaryRows 工资汇总行构建（列表/导出/状态筛选共用）
func buildSalarySummaryRows(list []model.SalarySummary) []map[string]interface{} {
	ids := make([]uint, len(list))
	for i, s := range list {
		ids[i] = s.PersonID
	}
	nameMap := PersonNameMap(ids)

	result := make([]map[string]interface{}, len(list))
	for i, s := range list {
		data, _ := json.Marshal(s)
		var item map[string]interface{}
		json.Unmarshal(data, &item)
		item["person_name"] = nameMap[s.PersonID]
		item["status"] = IsSalarySummaryStale(&s)
		result[i] = item
	}
	return result
}

// salarySummaryStaleSources 月度工资汇总过期数据源（行级/批量徽章共用同一份定义）：
// 核算/快照 last_calc_at + 工资事件/年假事件/考勤日事件的 updated_at 与 deleted_at
// （L0 事件软删除时间纳入检测：事件删除致派生层清空时由事件表 deleted_at 兜底）
func salarySummaryStaleSources(start, end utils.DateOnly, month string) []StaleSource {
	return []StaleSource{
		{Model: &model.AttendanceCalculationMonthly{}, Column: "last_calc_at",
			Where: "belong_month = ?", Args: []interface{}{month}},
		{Model: &model.PositionSnapshot{}, Column: "last_calc_at",
			Where: "effective_start_date <= ? AND effective_end_date >= ?", Args: []interface{}{end, start}},
		{Model: &model.AttendanceDaily{}, Column: "updated_at", Unscoped: true,
			Where: "event_date >= ? AND event_date <= ?", Args: []interface{}{start, end}},
		{Model: &model.AttendanceDaily{}, Column: "deleted_at", Unscoped: true, Nullable: true,
			Where: "event_date >= ? AND event_date <= ?", Args: []interface{}{start, end}},
		{Model: &model.SalaryEvent{}, Column: "updated_at", Unscoped: true,
			Where: "belong_month = ?", Args: []interface{}{month}},
		{Model: &model.SalaryEvent{}, Column: "deleted_at", Unscoped: true, Nullable: true,
			Where: "belong_month = ?", Args: []interface{}{month}},
		{Model: &model.AnnualLeaveAccountEvent{}, Column: "updated_at", Unscoped: true,
			Where: "effective_date >= ? AND effective_date <= ?", Args: []interface{}{start, end}},
		{Model: &model.AnnualLeaveAccountEvent{}, Column: "deleted_at", Unscoped: true, Nullable: true,
			Where: "effective_date >= ? AND effective_date <= ?", Args: []interface{}{start, end}},
	}
}

func IsSalarySummaryStale(summary *model.SalarySummary) string {
	monthStart, _ := utils.MonthStart(summary.BelongMonth)
	monthEnd, _ := utils.MonthEnd(summary.BelongMonth)
	changed, err := RowDataChanged(summary.LastCalcAt, summary.PersonID,
		salarySummaryStaleSources(utils.DateOnlyFromTime(monthStart), utils.DateOnlyFromTime(monthEnd), summary.BelongMonth))
	if err != nil || changed {
		return "data_changed"
	}
	return "calculated"
}

func GetSalaryVersions(ctx context.Context, personID uint, month string) ([]model.SalarySummaryVersion, error) {
	if err := EnsureOwnPerson(ctx, personID); err != nil {
		return nil, err
	}
	var versions []model.SalarySummaryVersion
	err := dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).
		Order("version DESC").Find(&versions).Error
	if err != nil {
		return nil, err
	}
	// person_name 富化：版本展开展示契约统一（同人同月，单次查询填充）
	if len(versions) > 0 {
		name := PersonName(personID)
		for i := range versions {
			versions[i].PersonName = name
		}
	}
	return versions, err
}

func GetSalaryVersionByID(ctx context.Context, versionID uint) (*model.SalarySummaryVersion, error) {
	var version model.SalarySummaryVersion
	err := dao.DB.First(&version, versionID).Error
	if err != nil {
		return nil, err
	}
	if err := EnsureOwnPerson(ctx, version.PersonID); err != nil {
		return nil, err
	}
	version.PersonName = PersonName(version.PersonID)
	return &version, nil
}

func isHighTempMonth(month string) bool {
	v := GetConfigValueOrDefault("attendance.high_temp_months", `["06","07","08","09"]`)
	var months []string
	if err := json.Unmarshal([]byte(v), &months); err != nil {
		return false
	}
	m := month[5:7]
	for _, mm := range months {
		if mm == m {
			return true
		}
	}
	return false
}

type SalaryTrace struct {
	Summary              model.SalarySummary                `json:"summary"`
	AttendanceCalc       model.AttendanceCalculationMonthly `json:"attendance_calc"`
	DailyProjections     []model.AttendanceDailyProjection  `json:"daily_projections"`
	AttendanceDailies    []model.AttendanceDaily            `json:"attendance_dailies"`
	PositionSnapshots    []model.PositionSnapshot           `json:"position_snapshots"`
	SalaryEvents         []model.SalaryEvent                `json:"salary_events"`
	AnnualLeaveCarryover []model.AnnualLeaveAccountEvent    `json:"annual_leave_carryover"`
}

func GetSalaryTrace(ctx context.Context, personID uint, month string) (*SalaryTrace, error) {
	if err := EnsureOwnPerson(ctx, personID); err != nil {
		return nil, err
	}
	monthStart, _ := time.Parse("2006-01", month)
	monthEnd := monthStart.AddDate(0, 1, -1)
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthEnd)

	var summary model.SalarySummary
	if err := dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).First(&summary).Error; err != nil {
		return nil, fmt.Errorf("工资汇总不存在")
	}

	var calc model.AttendanceCalculationMonthly
	dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).First(&calc)

	// person_name 富化：展示契约统一，追溯数据自足（一次查询两处复用）
	name := PersonName(personID)
	summary.PersonName = name
	calc.PersonName = name

	var dailyProjections []model.AttendanceDailyProjection
	dao.DB.Where("person_id = ? AND work_date >= ? AND work_date <= ?",
		personID, monthStartD, monthEndD).Order("work_date ASC").Find(&dailyProjections)

	var attendanceDailies []model.AttendanceDaily
	dao.DB.Where("person_id = ? AND event_date >= ? AND event_date <= ?",
		personID, monthStartD, monthEndD).Order("event_date ASC").Preload("Details").Find(&attendanceDailies)

	var snapshots []model.PositionSnapshot
	dao.DB.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
		personID, monthEndD, monthStartD).Order("effective_start_date ASC").Find(&snapshots)

	var salaryEvents []model.SalaryEvent
	dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).Order("seq ASC").Find(&salaryEvents)

	var alCarryover []model.AnnualLeaveAccountEvent
	dao.DB.Where("person_id = ? AND event_type = ? AND effective_date >= ? AND effective_date <= ?",
		personID, "carryover_deduct", monthStartD, monthEndD).Find(&alCarryover)

	return &SalaryTrace{
		Summary:              summary,
		AttendanceCalc:       calc,
		DailyProjections:     dailyProjections,
		AttendanceDailies:    attendanceDailies,
		PositionSnapshots:    snapshots,
		SalaryEvents:         salaryEvents,
		AnnualLeaveCarryover: alCarryover,
	}, nil
}

// GetSalarySummariesBadges 月度工资汇总徽章（指定月份）：无汇总记录 gray；
// 汇总过期（IsSalarySummaryStale = data_changed 语义）orange；已核算未过期 green。
// stale 判定按人员批量聚合 4 类源（考勤核算/职务快照/工资事件/年假事件）的最后变更时间，
// 与 IsSalarySummaryStale 语义等价（Select 原列 + Go 聚合，规避 MAX 聚合 Scan 类型坑）。
func GetSalarySummariesBadges(ctx context.Context, month string) ([]PersonBadge, error) {
	monthStart, err := utils.MonthStart(month)
	if err != nil {
		return nil, err
	}
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthStart.AddDate(0, 1, -1))

	var summaries []model.SalarySummary
	if err := dao.DBFromContext(ctx).Where("belong_month = ?", month).Find(&summaries).Error; err != nil {
		return nil, err
	}
	sumMap := make(map[uint]model.SalarySummary, len(summaries))
	for _, s := range summaries {
		sumMap[s.PersonID] = s
	}

	latest, err := PersonLatestTimes(salarySummaryStaleSources(monthStartD, monthEndD, month))
	if err != nil {
		return nil, err
	}

	var personIDs []uint
	db := dao.DBFromContext(ctx).Table("persons").Where("deleted_at IS NULL")
	if pid, ok := dao.OwnPersonID(ctx); ok {
		db = db.Where("id = ?", pid)
	}
	if err := db.Pluck("id", &personIDs).Error; err != nil {
		return nil, err
	}
	result := make([]PersonBadge, 0, len(personIDs))
	for _, pid := range personIDs {
		s, ok := sumMap[pid]
		if !ok {
			result = append(result, PersonBadge{PersonID: pid, Level: "gray"})
			continue
		}
		level := "green"
		if latest[pid].After(s.LastCalcAt) {
			level = "orange"
		}
		result = append(result, PersonBadge{PersonID: pid, Level: level})
	}
	return result, nil
}
