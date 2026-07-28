package service

import (
	"fmt"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"
)

func CalculateSalary(personID uint, month string, operatorID uint, operatorName string) error {
	var calc model.AttendanceCalculationMonthly
	err := dao.DB.Where("person_id = ? AND belong_month = ?", personID, month).First(&calc).Error
	if err != nil {
		return fmt.Errorf("未完成月度考勤核算，请先进行考勤核算")
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

	var totalDays float64
	var wPerfBase, wPost, wMeal, wHousing, wTransport, wHighTemp, wInsComp, wFundComp, wSSDeduct, wHFDeduct float64
	salaryDays := 0
	isActive := false

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
		if s.IsActive {
			isActive = true
		}
		salaryDays = s.SalaryDays
	}

	if totalDays == 0 {
		return fmt.Errorf("当月无在职记录")
	}

	allowanceRatio := 1.0
	if !isActive {
		allowanceRatio = totalDays / float64(salaryDays)
	}

	perfBase := wPerfBase / totalDays
	post := utils.RoundTwoDecimal(wPost / totalDays * allowanceRatio)
	meal := utils.RoundTwoDecimal(wMeal / totalDays * allowanceRatio)
	housing := utils.RoundTwoDecimal(wHousing / totalDays * allowanceRatio)
	transport := utils.RoundTwoDecimal(wTransport / totalDays * allowanceRatio)
	highTemp := utils.RoundTwoDecimal(wHighTemp / totalDays * allowanceRatio)
	if !isHighTempMonth(month) {
		highTemp = 0
	}
	insComp := utils.RoundTwoDecimal(wInsComp / totalDays * allowanceRatio)
	fundComp := utils.RoundTwoDecimal(wFundComp / totalDays * allowanceRatio)
	ssDeduct := utils.RoundTwoDecimal(wSSDeduct / totalDays)
	hfDeduct := utils.RoundTwoDecimal(wHFDeduct / totalDays)

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

	perfSalary := utils.RoundTwoDecimal(perfBase * perfCoeff * allowanceRatio)

	var carryoverDeductHours float64
	dao.DB.Model(&model.AnnualLeaveAccountEvent{}).
		Where("person_id = ? AND event_type = ? AND effective_date >= ? AND effective_date <= ?",
			personID, "carryover_deduct", monthStartD, monthEndD).
		Select("COALESCE(SUM(hours), 0)").Scan(&carryoverDeductHours)

	workHoursPerDay := getWorkHoursPerDay()
	holidayRatio := getOvertimeHolidayRatio()
	carryoverSalary := 0.0
	if salaryDays > 0 {
		carryoverSalary = utils.RoundTwoDecimal(carryoverDeductHours * (calc.WeightedBaseSalary + calc.WeightedMealAllowance) / float64(salaryDays) / workHoursPerDay * holidayRatio)
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
		SalaryDays:                    salaryDays,
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
		SalaryDays:                    salaryDays,
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
			if err.Error()[:6] == "未完成月度考勤核算" {
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

func GetSalarySummaries(month string, personID uint) ([]model.SalarySummary, error) {
	tx := dao.DB.Model(&model.SalarySummary{})
	if month != "" {
		tx = tx.Where("belong_month = ?", month)
	}
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	var list []model.SalarySummary
	err := tx.Order("belong_month DESC, person_id ASC").Find(&list).Error
	return list, err
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
	m := month[5:7]
	return len(v) > 0 && containsMonth(v, m)
}

func containsMonth(jsonList string, m string) bool {
	return len(jsonList) > len(m) && (jsonList[1:3] == m || jsonList[6:8] == m || jsonList[11:13] == m || jsonList[16:18] == m)
}
