package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"
)

var ErrAttendanceNotCalculated = errors.New("未完成月度考勤核算，请先进行考勤核算")

func CalculateSalary(personID uint, month string, operatorID uint, operatorName string) error {
	var calc model.AttendanceCalculationMonthly
	err := dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).First(&calc).Error
	if err != nil {
		return fmt.Errorf("%w", ErrAttendanceNotCalculated)
	}

	monthStart, _ := time.Parse("2006-01", month)
	monthEnd := monthStart.AddDate(0, 1, -1)
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthEnd)

	var snapshots []model.PositionSnapshot
	dao.DB.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
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
		wPerfBase += s.PerformanceSalary * segDays
		wPost += s.PostAllowance * segDays
		wMeal += s.MealAllowance * segDays
		wHousing += s.HousingAllowance * segDays
		wTransport += s.TransportAllowance * segDays
		wHighTemp += s.HighTempAllowance * segDays
		wInsComp += s.InsuranceCompensation * segDays
		wFundComp += s.FundCompensation * segDays
		wSSDeduct += s.SocialSecurityDeduct * segDays
		wHFDeduct += s.HousingFundDeduct * segDays
	}

	if activeDays == 0 {
		return fmt.Errorf("当月无在职记录")
	}

	if calc.SalaryDays == 0 {
		return fmt.Errorf("计薪天数为0")
	}

	salaryDays := float64(calc.SalaryDays)
	totalCalendarDays := monthEnd.Sub(monthStart).Hours()/24 + 1
	isFullMonth := activeDays == totalCalendarDays
	attendanceDays := calc.TotalWorkHours / getWorkHoursPerDay()

	subsidyDiv := activeDays
	if !isFullMonth {
		subsidyDiv = salaryDays
	}

	perfDiv := activeDays
	perfRatio := 1.0
	if !isFullMonth {
		perfDiv = salaryDays
		perfRatio = attendanceDays / salaryDays
	}

	post := utils.RoundTwoDecimal(wPost / subsidyDiv)
	meal := utils.RoundTwoDecimal(wMeal / subsidyDiv)
	housing := utils.RoundTwoDecimal(wHousing / subsidyDiv)
	transport := utils.RoundTwoDecimal(wTransport / subsidyDiv)
	highTemp := utils.RoundTwoDecimal(wHighTemp / subsidyDiv)
	if !isHighTempMonth(month) {
		highTemp = 0
	}
	insComp := utils.RoundTwoDecimal(wInsComp / subsidyDiv)
	fundComp := utils.RoundTwoDecimal(wFundComp / subsidyDiv)
	ssDeduct := utils.RoundTwoDecimal(wSSDeduct / activeDays)
	hfDeduct := utils.RoundTwoDecimal(wHFDeduct / activeDays)

	var perfCoeff float64 = 1
	var salesCommission, rewardPunishment, borrowingRepayment, taxDeduct float64
	var salaryEvents []model.SalaryEvent
	dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).Order("seq DESC").Find(&salaryEvents)

	maxCoeffSeq := 0
	for _, e := range salaryEvents {
		switch e.EventType {
		case "绩效系数":
			if e.Seq > maxCoeffSeq {
				perfCoeff = e.Amount
				maxCoeffSeq = e.Seq
			}
		case "提成":
			salesCommission += e.Amount
		case "奖惩":
			rewardPunishment += e.Amount
		case "借款还款":
			borrowingRepayment += e.Amount
		case "个税扣除":
			taxDeduct += e.Amount
		}
	}

	perfSalary := utils.RoundTwoDecimal(wPerfBase / perfDiv * perfRatio * perfCoeff)

	var carryoverDeductHours float64
	dao.DB.Model(&model.AnnualLeaveAccountEvent{}).
		Where("person_id = ? AND event_type = ? AND effective_date >= ? AND effective_date <= ?",
			personID, "carryover_deduct", monthStartD, monthEndD).
		Select("COALESCE(SUM(hours), 0)").Scan(&carryoverDeductHours)

	workHoursPerDay := getWorkHoursPerDay()
	holidayRatio := getOvertimeHolidayRatio()
	carryoverSalary := 0.0
	if calc.SalaryDays > 0 {
		carryoverSalary = utils.RoundTwoDecimal(carryoverDeductHours * (calc.WeightedBaseSalary + calc.WeightedMealAllowance) / float64(calc.SalaryDays) / workHoursPerDay * holidayRatio)
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
		LastCalcAt:                    utils.DateOnlyFromTime(time.Now()),
	}

	var maxVersion int
	dao.DB.Model(&model.SalarySummaryVersion{}).
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
	dao.DB.Create(&version)

	dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).Delete(&model.SalarySummary{})
	dao.DB.Create(&summary)

	return nil
}

func CalculateSalaryBatch(month string, personIDs []uint, operatorID uint, operatorName string) (int, int, int, error) {
	success, fail, skip := 0, 0, 0
	for _, pid := range personIDs {
		err := CalculateSalary(pid, month, operatorID, operatorName)
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

func GetSalarySummaries(month string, personID uint, pageNum, pageSize int) ([]model.SalarySummary, int64, error) {
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
	return list, total, err
}

func IsSalarySummaryStale(summary *model.SalarySummary) string {
	monthStart, _ := utils.MonthStart(summary.BelongMonth)
	monthEnd, _ := utils.MonthEnd(summary.BelongMonth)
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthEnd)

	var calcLastCalcAt utils.DateOnly
	dao.DB.Model(&model.AttendanceCalculationMonthly{}).
		Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).
		Select("COALESCE(MAX(last_calc_at), '0001-01-01')").Scan(&calcLastCalcAt)
	if calcLastCalcAt.Time().After(summary.LastCalcAt.Time()) {
		return "data_changed"
	}

	var maxSnapLastCalc utils.DateOnly
	dao.DB.Model(&model.PositionSnapshot{}).
		Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			summary.PersonID, monthEndD, monthStartD).
		Select("COALESCE(MAX(last_calc_at), '0001-01-01')").Scan(&maxSnapLastCalc)
	if maxSnapLastCalc.Time().After(summary.LastCalcAt.Time()) {
		return "data_changed"
	}

	var salaryEventMaxTime *time.Time
	dao.DB.Model(&model.SalaryEvent{}).Unscoped().
		Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).
		Select("COALESCE(MAX(updated_at), MAX(deleted_at))").Scan(&salaryEventMaxTime)
	if salaryEventMaxTime != nil && salaryEventMaxTime.After(summary.LastCalcAt.Time()) {
		return "data_changed"
	}

	var alEventMaxTime *time.Time
	dao.DB.Model(&model.AnnualLeaveAccountEvent{}).Unscoped().
		Where("person_id = ? AND effective_date >= ? AND effective_date <= ?",
			summary.PersonID, monthStartD, monthEndD).
		Select("COALESCE(MAX(updated_at), MAX(deleted_at))").Scan(&alEventMaxTime)
	if alEventMaxTime != nil && alEventMaxTime.After(summary.LastCalcAt.Time()) {
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
