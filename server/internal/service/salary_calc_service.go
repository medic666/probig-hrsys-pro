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

func CalculateSalary(ctx context.Context, personID uint, month string, operatorID uint, operatorName string) error {
	var oldJSON, newJSON, personName string
	err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		monthStart, _ := time.Parse("2006-01", month)
		monthEnd := monthStart.AddDate(0, 1, -1)
		monthStartD := utils.DateOnlyFromTime(monthStart)
		monthEndD := utils.DateOnlyFromTime(monthEnd)

		// L1 状态检查：当月存在待确认日记工时投影时禁止核算（与考勤核算一致，形成 L0→L1→L2→L3 完整控制链）
		var pendingCount int64
		tx.Model(&model.AttendanceDailyProjection{}).
			Where("person_id = ? AND work_date >= ? AND work_date <= ? AND status = ?",
				personID, monthStartD, monthEndD, "pending").Count(&pendingCount)
		if pendingCount > 0 {
			return fmt.Errorf("当月有 %d 天待确认的考勤记录，请先完成核实", pendingCount)
		}

		var calc model.AttendanceCalculationMonthly
		err := tx.Where("person_id = ? AND belong_month = ?", personID, month).First(&calc).Error
		if err != nil {
			return fmt.Errorf("%w", ErrAttendanceNotCalculated)
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
		case "借款还款":
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
		PersonID:                      personID,
		BelongMonth:                   month,
		SalaryDays:                    calc.SalaryDays,
		WeightedBaseSalary:            calc.WeightedBaseSalary,
		TotalWorkHours:                calc.TotalWorkHours,
		TotalOvertimeWorkdayHours:     calc.TotalOvertimeWorkdayHours,
		TotalOvertimeHolidayHours:     calc.TotalOvertimeHolidayHours,
		AttendanceSalary:              calc.AttendanceSalary,
		OvertimeWorkdaySalary:         calc.OvertimeWorkdaySalary,
		OvertimeHolidaySalary:         calc.OvertimeHolidaySalary,
		AnnualLeaveCarryoverDeduct:    carryoverDeductHours,
		AnnualLeaveCarryoverSalary:    carryoverSalary,
		AttendanceBonus:               calc.AttendanceBonus,
		PerformanceSalary:             perfSalary,
		PostAllowance:                 post,
		MealAllowance:                 meal,
		HousingAllowance:              housing,
		TransportAllowance:            transport,
		HighTempAllowance:             highTemp,
		InsuranceCompensation:         insComp,
		FundCompensation:              fundComp,
		SalesCommission:               salesCommission,
		RewardPunishment:              rewardPunishment,
		BorrowingRepayment:            borrowingRepayment,
		SocialSecurityDeduct:          ssDeduct,
		HousingFundDeduct:             hfDeduct,
		TaxDeduct:                     taxDeduct,
		FinalSalary:                   finalSalary,
		LastCalcAt: time.Now(),
	}

	var maxVersion int
	tx.Model(&model.SalarySummaryVersion{}).
		Where("person_id = ? AND belong_month = ?", personID, month).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	batchNo := "SAL-" + month + "-" + strconv.FormatInt(time.Now().Unix(), 10)

	version := model.SalarySummaryVersion{
		PersonID:                      personID,
		BelongMonth:                   month,
		Version:                       maxVersion + 1,
		CalcBatchNo:                   batchNo,
		OperatorID:                    operatorID,
		OperatorName:                  operatorName,
		SalaryDays:                    calc.SalaryDays,
		WeightedBaseSalary:            calc.WeightedBaseSalary,
		TotalWorkHours:                calc.TotalWorkHours,
		TotalOvertimeWorkdayHours:     calc.TotalOvertimeWorkdayHours,
		TotalOvertimeHolidayHours:     calc.TotalOvertimeHolidayHours,
		AttendanceSalary:              calc.AttendanceSalary,
		OvertimeWorkdaySalary:         calc.OvertimeWorkdaySalary,
		OvertimeHolidaySalary:         calc.OvertimeHolidaySalary,
		AnnualLeaveCarryoverDeduct:    carryoverDeductHours,
		AnnualLeaveCarryoverSalary:    carryoverSalary,
		AttendanceBonus:               calc.AttendanceBonus,
		PerformanceSalary:             perfSalary,
		PostAllowance:                 post,
		MealAllowance:                 meal,
		HousingAllowance:              housing,
		TransportAllowance:            transport,
		HighTempAllowance:             highTemp,
		InsuranceCompensation:         insComp,
		FundCompensation:              fundComp,
		SalesCommission:               salesCommission,
		RewardPunishment:              rewardPunishment,
		BorrowingRepayment:            borrowingRepayment,
		SocialSecurityDeduct:          ssDeduct,
		HousingFundDeduct:             hfDeduct,
		TaxDeduct:                     taxDeduct,
		FinalSalary:                   finalSalary,
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

	b, _ := json.Marshal(summary); newJSON = string(b)
	tx.Table("persons").Select("name").Where("id = ?", personID).Scan(&personName)
	return nil
	})
	if err != nil {
		return err
	}
	// 核算审计（每人一条），事务提交后写入，避免与业务事务的 SQLite 写锁竞争
	dao.WriteBusinessAudit(ctx, "核算", "salary_summaries", personID, personName, oldJSON, newJSON)
	return nil
}

func CalculateSalaryBatch(ctx context.Context, month string, personIDs []uint, operatorID uint, operatorName string) (int, int, int, error) {
	success, fail, skip := 0, 0, 0
	for _, pid := range personIDs {
		err := CalculateSalary(ctx, pid, month, operatorID, operatorName)
		if err != nil {
			if errors.Is(err, ErrAttendanceNotCalculated) {
				skip++
			} else {
				fail++
			}
		} else {
			success++
		}
	}
	return success, fail, skip, nil
}

func GetSalarySummaries(month string, personID uint, pageNum, pageSize int) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.SalarySummary{})
	if month != "" {
		tx = tx.Where("belong_month = ?", month)
	}
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	var total int64
	tx.Count(&total)
	var list []model.SalarySummary
	offset := (pageNum - 1) * pageSize
	err := tx.Order("belong_month DESC, person_id ASC").Offset(offset).Limit(pageSize).Find(&list).Error
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
		data, _ := json.Marshal(s)
		var item map[string]interface{}
		json.Unmarshal(data, &item)
		item["person_name"] = nameMap[s.PersonID]
		item["status"] = IsSalarySummaryStale(&s)
		result[i] = item
	}
	return result, total, nil
}

func IsSalarySummaryStale(summary *model.SalarySummary) string {
	monthStart, _ := utils.MonthStart(summary.BelongMonth)
	monthEnd, _ := utils.MonthEnd(summary.BelongMonth)
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthEnd)

	var calcTimes []*time.Time
	dao.DB.Model(&model.AttendanceCalculationMonthly{}).
		Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).
		Pluck("last_calc_at", &calcTimes)
	if latestTime(calcTimes).After(summary.LastCalcAt) {
		return "data_changed"
	}

	var snapTimes []*time.Time
	dao.DB.Model(&model.PositionSnapshot{}).
		Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			summary.PersonID, monthEndD, monthStartD).
		Pluck("last_calc_at", &snapTimes)
	if latestTime(snapTimes).After(summary.LastCalcAt) {
		return "data_changed"
	}

	// 工资事件：行级 max(updated_at, deleted_at) 再聚合取最大（软删除时间纳入检测）
	var evUpds, evDels []*time.Time
	dao.DB.Model(&model.SalaryEvent{}).Unscoped().
		Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).
		Pluck("updated_at", &evUpds)
	dao.DB.Model(&model.SalaryEvent{}).Unscoped().
		Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).
		Pluck("deleted_at", &evDels)
	if t := latestTime(evUpds); t.After(summary.LastCalcAt) {
		return "data_changed"
	}
	if t := latestTime(evDels); t.After(summary.LastCalcAt) {
		return "data_changed"
	}

	var alUpds, alDels []*time.Time
	dao.DB.Model(&model.AnnualLeaveAccountEvent{}).Unscoped().
		Where("person_id = ? AND effective_date >= ? AND effective_date <= ?",
			summary.PersonID, monthStartD, monthEndD).
		Pluck("updated_at", &alUpds)
	dao.DB.Model(&model.AnnualLeaveAccountEvent{}).Unscoped().
		Where("person_id = ? AND effective_date >= ? AND effective_date <= ?",
			summary.PersonID, monthStartD, monthEndD).
		Pluck("deleted_at", &alDels)
	if t := latestTime(alUpds); t.After(summary.LastCalcAt) {
		return "data_changed"
	}
	if t := latestTime(alDels); t.After(summary.LastCalcAt) {
		return "data_changed"
	}

	return "calculated"
}

func GetSalaryVersions(personID uint, month string) ([]model.SalarySummaryVersion, error) {
	var versions []model.SalarySummaryVersion
	err := dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).
		Order("version DESC").Find(&versions).Error
	return versions, err
}

func GetSalaryVersionByID(versionID uint) (*model.SalarySummaryVersion, error) {
	var version model.SalarySummaryVersion
	err := dao.DB.First(&version, versionID).Error
	if err != nil {
		return nil, err
	}
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
	Summary            model.SalarySummary               `json:"summary"`
	AttendanceCalc     model.AttendanceCalculationMonthly `json:"attendance_calc"`
	DailyProjections   []model.AttendanceDailyProjection  `json:"daily_projections"`
	AttendanceDailies   []model.AttendanceDaily              `json:"attendance_dailies"`
	PositionSnapshots  []model.PositionSnapshot           `json:"position_snapshots"`
	SalaryEvents       []model.SalaryEvent                `json:"salary_events"`
	AnnualLeaveCarryover []model.AnnualLeaveAccountEvent  `json:"annual_leave_carryover"`
}

func GetSalaryTrace(personID uint, month string) (*SalaryTrace, error) {
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
		AttendanceDailies:     attendanceDailies,
		PositionSnapshots:    snapshots,
		SalaryEvents:         salaryEvents,
		AnnualLeaveCarryover: alCarryover,
	}, nil
}
