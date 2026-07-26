package salary

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
)

func getClientIP(ip string) string {
	return ip
}

func CreateEvent(event *SalaryEvent, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		personName := getPersonNameFromDB(tx, event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "salary_event", event.ID, personName+"-"+event.EventName, "新增", nil, event, clientIP); err != nil {
			return err
		}
		return nil
	})
}

func UpdateEvent(id uint, req *SalaryEventUpdateRequest, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event SalaryEvent
		if err := tx.First(&event, id).Error; err != nil {
			return err
		}

		before := event

		if req.PersonID != nil {
			event.PersonID = *req.PersonID
		}
		if req.BelongMonth != nil {
			event.BelongMonth = *req.BelongMonth
		}
		if req.EventType != nil {
			event.EventType = *req.EventType
		}
		if req.Amount != nil {
			event.Amount = *req.Amount
		}
		if req.EventName != nil {
			event.EventName = *req.EventName
		}
		if req.Remark != nil {
			event.Remark = *req.Remark
		}

		if err := tx.Save(&event).Error; err != nil {
			return err
		}

		personName := getPersonNameFromDB(tx, event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "salary_event", event.ID, personName+"-"+event.EventName, "修改", before, event, clientIP); err != nil {
			return err
		}
		return nil
	})
}

func DeleteEvent(id uint, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event SalaryEvent
		if err := tx.First(&event, id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&event).Error; err != nil {
			return err
		}

		personName := getPersonNameFromDB(tx, event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "salary_event", event.ID, personName+"-"+event.EventName, "删除", event, nil, clientIP); err != nil {
			return err
		}
		return nil
	})
}

func RestoreEvent(id uint, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&SalaryEvent{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}

		var event SalaryEvent
		if err := tx.Unscoped().First(&event, id).Error; err != nil {
			return err
		}

		personName := getPersonNameFromDB(tx, event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "salary_event", event.ID, personName+"-"+event.EventName, "恢复", nil, event, clientIP); err != nil {
			return err
		}
		return nil
	})
}

func ListEvents(filter SalaryEventFilter) ([]SalaryEvent, int64, error) {
	return ListSalaryEvents(filter)
}

func ListTrashEvents(filter SalaryEventFilter) ([]SalaryEvent, int64, error) {
	return ListTrashSalaryEvents(filter)
}

func CalcSalarySummary(belongMonth string, personIDs []uint, operatorID uint, operatorName string) (succeeded int, failed int, errors []string) {
	if len(personIDs) == 0 {
		var persons []database.Person
		database.DB.Find(&persons)
		for _, p := range persons {
			personIDs = append(personIDs, p.ID)
		}
	}

	belongStart, _ := time.Parse("2006-01", belongMonth)
	year, month, _ := belongStart.Date()
	monthEnd := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

	for _, personID := range personIDs {
		err := calcOnePersonSummary(personID, belongMonth, belongStart, monthEnd, operatorID, operatorName)
		if err != nil {
			failed++
			personName := getPersonName(personID)
			errors = append(errors, personName+": "+err.Error())
		} else {
			succeeded++
		}
	}

	return
}

func calcOnePersonSummary(personID uint, belongMonth string, belongStart time.Time, monthEnd int, operatorID uint, operatorName string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var attendSalary database.AttendanceSalaryMonthly
		if err := tx.Where("person_id = ? AND belong_month = ?", personID, belongMonth).First(&attendSalary).Error; err != nil {
			return fmt.Errorf("假勤工资未核算，请先核算假勤工资")
		}

		effectiveSnapshots, err := getEffectiveSnapshots(tx, personID, belongStart, monthEnd)
		if err != nil {
			return err
		}

		perfBase := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.PerformanceSalary })
		postAllow := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.PostAllowance })
		housingAllow := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.HousingAllowance })
		transportAllow := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.TransportAllowance })
		insuranceComp := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.InsuranceCompensation })
		fundComp := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.FundCompensation })
		socialDeduct := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.SocialSecurityDeduct })
		fundDeduct := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.HousingFundDeduct })

		carryoverSalary := 0.0
		isEntryAnniversary := checkEntryAnniversary(tx, personID, belongStart.Month())
		if isEntryAnniversary {
			var balance database.LeaveAccountBalance
			if err := tx.Where("person_id = ? AND leave_type = ?", personID, "annual_leave").First(&balance).Error; err == nil {
				if balance.BalanceHours > 0 {
					overtimeHolidayRatio := config.GetFloat("attendance.overtime_holiday_ratio")
					workHoursPerDay := float64(config.GetInt("system.work_hours_per_day"))
					if workHoursPerDay == 0 {
						workHoursPerDay = 8
					}
					if attendSalary.SalaryDays > 0 {
						carryoverSalary = balance.BalanceHours * (attendSalary.WeightedBaseSalary + attendSalary.WeightedMealAllowance) / float64(attendSalary.SalaryDays) / workHoursPerDay * overtimeHolidayRatio
					}
				}
			}
		}

		coeff := 1.0
		perfEvent, err := GetLatestPerformanceCoefficientEvent(personID, belongMonth)
		if err == nil && perfEvent != nil {
			coeff = perfEvent.Amount
		}

		perfSalary := perfBase * coeff

		prorateRatio := 1.0
		if hasSnapshots(effectiveSnapshots) {
			firstSnap := effectiveSnapshots[0]
			lastSnap := effectiveSnapshots[len(effectiveSnapshots)-1]
			if firstSnap.EntryDate != nil {
				entryMonth := firstSnap.EntryDate.Format("2006-01")
				if entryMonth == belongMonth {
					entryDay := firstSnap.EntryDate.Day()
					prorateRatio = float64(monthEnd-entryDay+1) / float64(monthEnd)
				}
			}
			if lastSnap.LeaveDate != nil {
				leaveMonth := lastSnap.LeaveDate.Format("2006-01")
				if leaveMonth == belongMonth {
					leaveDay := lastSnap.LeaveDate.Day()
					prorateRatio = math.Min(prorateRatio, float64(leaveDay)/float64(monthEnd))
				}
			}
		}

		if prorateRatio < 1.0 {
			perfSalary = perfSalary * prorateRatio
			postAllow = postAllow * prorateRatio
			housingAllow = housingAllow * prorateRatio
			transportAllow = transportAllow * prorateRatio
			insuranceComp = insuranceComp * prorateRatio
			fundComp = fundComp * prorateRatio
		}

		highTempAllow := 0.0
		highTempMonths := config.Get("attendance.high_temp_months")
		if highTempMonths != "" {
			var months []string
			if err := json.Unmarshal([]byte(highTempMonths), &months); err == nil {
				currentMonth := belongStart.Format("01")
				for _, m := range months {
					if m == currentMonth {
						weightedHighTemp := weightedField(effectiveSnapshots, func(s *database.PositionSnapshot) float64 { return s.HighTempAllowance })
						highTempAllow = weightedHighTemp
						if prorateRatio < 1.0 {
							highTempAllow = highTempAllow * prorateRatio
						}
						break
					}
				}
			}
		}

		var salaryEvents []SalaryEvent
		tx.Where("person_id = ? AND belong_month = ?", personID, belongMonth).Find(&salaryEvents)

		totalAdj := 0.0
		taxDeduct := 0.0
		for _, e := range salaryEvents {
			if e.EventType == "个税扣除" {
				taxDeduct += e.Amount
			} else if e.EventType != "绩效系数" {
				totalAdj += e.Amount
			}
		}

		finalSalary := attendSalary.AttendanceSalary +
			attendSalary.OvertimeWorkdaySalary +
			attendSalary.OvertimeHolidaySalary +
			carryoverSalary +
			attendSalary.AttendanceBonus +
			perfSalary +
			postAllow +
			attendSalary.WeightedMealAllowance +
			housingAllow +
			transportAllow +
			highTempAllow +
			insuranceComp +
			fundComp +
			totalAdj -
			socialDeduct -
			fundDeduct -
			taxDeduct

		now := time.Now()
		summary := SalarySummary{
			PersonID:                    personID,
			BelongMonth:                 belongMonth,
			SalaryDays:                  attendSalary.SalaryDays,
			WeightedBaseSalary:          attendSalary.WeightedBaseSalary,
			TotalWorkHours:              attendSalary.TotalWorkHours,
			TotalOvertimeWorkdayHours:   attendSalary.TotalOvertimeWorkdayHours,
			TotalOvertimeHolidayHours:   attendSalary.TotalOvertimeHolidayHours,
			AttendanceSalary:            attendSalary.AttendanceSalary,
			OvertimeWorkdaySalary:       attendSalary.OvertimeWorkdaySalary,
			OvertimeHolidaySalary:       attendSalary.OvertimeHolidaySalary,
			AnnualLeaveCarryoverSalary:  carryoverSalary,
			AttendanceBonus:             attendSalary.AttendanceBonus,
			PerformanceSalary:           math.Round(perfSalary*100) / 100,
			PostAllowance:               math.Round(postAllow*100) / 100,
			MealAllowance:               attendSalary.WeightedMealAllowance,
			HousingAllowance:            math.Round(housingAllow*100) / 100,
			TransportAllowance:          math.Round(transportAllow*100) / 100,
			HighTempAllowance:           math.Round(highTempAllow*100) / 100,
			InsuranceCompensation:       math.Round(insuranceComp*100) / 100,
			FundCompensation:            math.Round(fundComp*100) / 100,
			TotalAdjustment:             math.Round(totalAdj*100) / 100,
			SocialSecurityDeduct:        math.Round(socialDeduct*100) / 100,
			HousingFundDeduct:           math.Round(fundDeduct*100) / 100,
			TaxDeduct:                   math.Round(taxDeduct*100) / 100,
			FinalSalary:                 math.Round(finalSalary*100) / 100,
			LastCalcAt:                  &now,
		}

		if err := UpsertSalarySummary(&summary); err != nil {
			return err
		}

		personName := getPersonNameFromDB(tx, personID)
		audit.CreateAuditLog(tx, operatorID, operatorName, "salary_summary", summary.ID, personName+"-"+belongMonth, "核算", nil, summary, "")

		return nil
	})
}

func getEffectiveSnapshots(tx *gorm.DB, personID uint, belongStart time.Time, monthEnd int) ([]database.PositionSnapshot, error) {
	year, month, _ := belongStart.Date()
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEndDate := time.Date(year, month, monthEnd, 0, 0, 0, 0, time.UTC)

	var snapshots []database.PositionSnapshot
	err := tx.Where("person_id = ? AND effective_start_date <= ? AND (effective_end_date >= ? OR effective_end_date IS NULL)",
		personID, monthEndDate, monthStart).Order("effective_start_date ASC").Find(&snapshots).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return snapshots, nil
}

func hasSnapshots(snapshots []database.PositionSnapshot) bool {
	return len(snapshots) > 0
}

func weightedField(snapshots []database.PositionSnapshot, getter func(*database.PositionSnapshot) float64) float64 {
	if len(snapshots) == 0 {
		return 0
	}

	totalWeight := 0.0
	totalValue := 0.0

	for i := range snapshots {
		s := &snapshots[i]
		if s.EffectiveStartDate == nil || s.EffectiveEndDate == nil {
			continue
		}
		weight := s.EffectiveEndDate.Sub(*s.EffectiveStartDate).Hours()/24 + 1
		totalWeight += weight
		totalValue += getter(s) * weight
	}

	if totalWeight == 0 {
		return 0
	}
	return totalValue / totalWeight
}

func checkEntryAnniversary(tx *gorm.DB, personID uint, currentMonth time.Month) bool {
	var snapshot database.PositionSnapshot
	err := tx.Where("person_id = ? AND entry_date IS NOT NULL", personID).
		Order("effective_start_date ASC").First(&snapshot).Error
	if err != nil {
		return false
	}
	if snapshot.EntryDate == nil {
		return false
	}
	return snapshot.EntryDate.Month() == currentMonth
}

func ListSummaries(filter SalarySummaryFilter) ([]SalarySummaryVO, int64, error) {
	return ListSalarySummaries(filter)
}

func GetAllSummaries(filter SalarySummaryFilter) ([]SalarySummaryVO, error) {
	return GetAllSalarySummaries(filter)
}

func GetSummaryDetail(id uint) (*SalarySummary, error) {
	return GetSummaryByID(id)
}

func getPersonNameFromDB(tx *gorm.DB, personID uint) string {
	var person database.Person
	if err := tx.Unscoped().Where("id = ?", personID).First(&person).Error; err != nil {
		return strconv.FormatUint(uint64(personID), 10)
	}
	return person.Name
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func isEntryOrLeaveMonth(snapshots []database.PositionSnapshot, belongMonth string) bool {
	if len(snapshots) == 0 {
		return false
	}
	firstSnap := snapshots[0]
	lastSnap := snapshots[len(snapshots)-1]

	if firstSnap.EntryDate != nil && firstSnap.EntryDate.Format("2006-01") == belongMonth {
		return true
	}
	if lastSnap.LeaveDate != nil && lastSnap.LeaveDate.Format("2006-01") == belongMonth {
		return true
	}
	return false
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseBool(s string) bool {
	return strings.ToLower(s) == "true" || s == "1"
}
