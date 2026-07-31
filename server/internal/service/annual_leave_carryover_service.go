package service

import (
	"fmt"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func ExecuteCarryover(month string, operatorID uint) (map[string]interface{}, error) {
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("月份格式错误")
	}
	nextMonthStart := monthStart.AddDate(0, 1, 0)
	nextMonthEnd := nextMonthStart.AddDate(0, 1, -1)

	var snapshots []model.PositionSnapshot
	dao.DB.Where("effective_start_date <= ? AND effective_end_date >= ? AND is_active = true",
		utils.DateOnlyFromTime(nextMonthEnd), utils.DateOnlyFromTime(nextMonthStart)).Find(&snapshots)

	eligiblePersonIDs := make(map[uint]bool)
	for _, s := range snapshots {
		if s.EntryDate != nil {
			entryMonth := s.EntryDate.Time().Month()
			if entryMonth == nextMonthStart.Month() {
				eligiblePersonIDs[s.PersonID] = true
			}
		}
	}

	if len(eligiblePersonIDs) == 0 {
		return nil, fmt.Errorf("当月无符合条件的在职人员")
	}

	batchNo := "ALC-" + monthStart.Format("20060102") + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	batch := model.SysBatch{
		BatchNo:        batchNo,
		BusinessType:   "annual_leave_carryover",
		BusinessPeriod: nextMonthStart.Format("2006-01"),
		OperatorID:     operatorID,
		Status:         1,
		TotalCount:     len(eligiblePersonIDs),
	}
	if err := dao.DB.Create(&batch).Error; err != nil {
		return nil, err
	}

	yearlyHours := getYearlyAnnualLeaveHours()

	success := 0
	fail := 0
	now := time.Now()

	for personID := range eligiblePersonIDs {
		err := utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
			tx.Where("person_id = ? AND source_type = ? AND batch_id = ?", personID, "system_period", batch.ID).Delete(&model.AnnualLeaveAccountEvent{})

			balance := calculatePersonAnnualBalance(tx, personID)

			if balance > 0 {
				monthEnd := monthStart.AddDate(0, 1, -1)
				deduct := model.AnnualLeaveAccountEvent{
					PersonID:      personID,
					EventType:     "carryover_deduct",
					SourceType:    "system_period",
					BatchID:       &batch.ID,
					Hours:         balance,
					EffectiveDate: utils.DateOnlyFromTime(monthEnd),
					Remark:        fmt.Sprintf("周年结转扣除 %s", nextMonthStart.Format("2006-01")),
				}
				if err := createSystemLeaveEventInTx(tx, &deduct); err != nil {
					return err
				}
			}

			var curSnap model.PositionSnapshot
			tx.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
				personID, utils.DateOnlyFromTime(nextMonthEnd), utils.DateOnlyFromTime(nextMonthStart)).First(&curSnap)

			if curSnap.HasAnnualLeave {
				grant := model.AnnualLeaveAccountEvent{
					PersonID:      personID,
					EventType:     "grant",
					SourceType:    "system_period",
					BatchID:       &batch.ID,
					Hours:         yearlyHours,
					EffectiveDate: utils.DateOnlyFromTime(nextMonthStart),
					Remark:        fmt.Sprintf("周年配发 %s", nextMonthStart.Format("2006-01")),
				}
				if err := createSystemLeaveEventInTx(tx, &grant); err != nil {
					return err
				}
			}

			return RebuildAnnualLeaveBalance(tx, personID)
		})
		if err != nil {
			fail++
		} else {
			success++
		}
	}

	updates := map[string]interface{}{
		"status":      2,
		"total_count": success,
		"executed_at": now,
	}
	if fail > 0 {
		updates["status"] = 4
	}
	dao.DB.Model(&batch).Updates(updates)

	return map[string]interface{}{
		"batch_no": batchNo,
		"success":  success,
		"fail":     fail,
		"total":    len(eligiblePersonIDs),
	}, nil
}

func createSystemLeaveEventInTx(tx *gorm.DB, event *model.AnnualLeaveAccountEvent) error {
	var maxSeq int
	tx.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ?", event.PersonID).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
	event.Seq = maxSeq + 1
	return tx.Create(event).Error
}

func calculatePersonAnnualBalance(tx *gorm.DB, personID uint) float64 {
	var accountEvents []model.AnnualLeaveAccountEvent
	tx.Where("person_id = ?", personID).Find(&accountEvents)

	var attendEvents []model.AttendanceEventDetail
	tx.Table("attendance_event_details").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id").
		Where("attendance_daily.person_id = ? AND attendance_event_details.event_type = ? AND attendance_event_details.sub_type = ?", personID, "休假", "年假").
		Select("attendance_event_details.hours").
		Scan(&attendEvents)

	var balance float64
	for _, e := range accountEvents {
		switch e.EventType {
		case "grant", "adjust":
			balance += e.Hours
		case "carryover_deduct":
			balance -= e.Hours
		}
	}
	for _, e := range attendEvents {
		balance -= e.Hours
	}
	return balance
}

func CancelCarryover(batchID uint) error {
	var batch model.SysBatch
	if err := dao.DB.First(&batch, batchID).Error; err != nil {
		return fmt.Errorf("批次不存在")
	}
	if batch.Status != 2 {
		return fmt.Errorf("仅可反结账已生效的批次")
	}

	var events []model.AnnualLeaveAccountEvent
	dao.DB.Where("batch_id = ?", batchID).Find(&events)

	for _, e := range events {
		dao.DB.Delete(&e)
		if err := RebuildAnnualLeaveBalance(dao.DB, e.PersonID); err != nil {
			return err
		}
	}

	now := time.Now()
	dao.DB.Model(&batch).Updates(map[string]interface{}{
		"status":      3,
		"canceled_at": now,
	})
	return nil
}

func GetCarryoverBatches() ([]model.SysBatch, error) {
	var batches []model.SysBatch
	err := dao.DB.Where("business_type = ?", "annual_leave_carryover").Order("id DESC").Find(&batches).Error
	return batches, err
}

func getYearlyAnnualLeaveHours() float64 {
	v := GetConfigValueOrDefault("annual_leave.yearly_hours", "40")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func GetBatchEvents(batchID uint) ([]map[string]interface{}, error) {
	var events []model.AnnualLeaveAccountEvent
	dao.DB.Where("batch_id = ?", batchID).Find(&events)
	var result []map[string]interface{}
	for _, e := range events {
		item := map[string]interface{}{
			"id":             e.ID,
			"person_id":      e.PersonID,
			"event_type":     e.EventType,
			"hours":          e.Hours,
			"effective_date": e.EffectiveDate,
		}
		var personName string
		dao.DB.Table("persons").Select("name").Where("id = ?", e.PersonID).Scan(&personName)
		item["person_name"] = personName
		result = append(result, item)
	}
	return result, nil
}
