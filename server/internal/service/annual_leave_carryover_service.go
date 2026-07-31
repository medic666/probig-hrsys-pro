package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func ExecuteCarryover(ctx context.Context, month string, operatorID uint, operatorName string) (map[string]interface{}, error) {
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

	batchNo := "ALC-" + monthStart.Format("20060102") + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	batch := model.SysBatch{
		BatchNo:        batchNo,
		BusinessType:   "annual_leave_carryover",
		BusinessPeriod: nextMonthStart.Format("2006-01"),
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		Status:         1,
		TotalCount:     len(eligiblePersonIDs),
	}
	if err := dao.DBFromContext(ctx).Create(&batch).Error; err != nil {
		return nil, err
	}

	// 识别历史同周期批次（幂等冲销的载体）：事件将被删除并重建，批次记录本身不保留
	var oldBatchIDs []uint
	dao.DBFromContext(ctx).Model(&model.SysBatch{}).
		Where("business_type = ? AND business_period = ? AND id != ?",
			"annual_leave_carryover", nextMonthStart.Format("2006-01"), batch.ID).
		Pluck("id", &oldBatchIDs)

	yearlyHours := getYearlyAnnualLeaveHours()

	success := 0
	fail := 0
	now := time.Now()

	for personID := range eligiblePersonIDs {
		err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
			// 冲销历史同周期批次的系统事件，事件源删除由 GORM 审计自动留痕
			if len(oldBatchIDs) > 0 {
				if err := tx.Where("person_id = ? AND source_type = ? AND batch_id IN ?",
					personID, "system_period", oldBatchIDs).Delete(&model.AnnualLeaveAccountEvent{}).Error; err != nil {
					return err
				}
			}

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
	dao.DBFromContext(ctx).Model(&batch).Updates(updates)

	// 冲销完成的历史批次记录不再保留
	if len(oldBatchIDs) > 0 {
		if err := dao.DBFromContext(ctx).Where("id IN ?", oldBatchIDs).Delete(&model.SysBatch{}).Error; err != nil {
			return nil, err
		}
	}

	// 结转审计
	if fail == 0 {
		summaryJSON, _ := json.Marshal(map[string]interface{}{
			"batch_no": batchNo, "business_period": nextMonthStart.Format("2006-01"),
			"success": success, "total": len(eligiblePersonIDs),
		})
		dao.WriteBusinessAudit(ctx, "结转", "annual_leave_carryover", batch.ID, batchNo, "", string(summaryJSON))
	}

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
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id AND attendance_daily.deleted_at IS NULL AND attendance_daily.status = 'confirmed'").
		Where("attendance_event_details.deleted_at IS NULL AND attendance_daily.person_id = ? AND attendance_event_details.event_type = ? AND attendance_event_details.sub_type = ?", personID, "休假", "年假").
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

func CancelCarryover(ctx context.Context, batchID uint) error {
	var batch model.SysBatch
	if err := dao.DBFromContext(ctx).First(&batch, batchID).Error; err != nil {
		return fmt.Errorf("批次不存在")
	}
	if batch.Status != 2 {
		return fmt.Errorf("仅可反结账已生效的批次")
	}

	var events []model.AnnualLeaveAccountEvent
	if err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Where("batch_id = ?", batchID).Find(&events).Error; err != nil {
			return err
		}
		for _, e := range events {
			if err := tx.Delete(&e).Error; err != nil {
				return err
			}
			if err := RebuildAnnualLeaveBalance(tx, e.PersonID); err != nil {
				return err
			}
		}
		// 批次记录一并清除，事件源变动已由审计留痕
		if err := tx.Delete(&model.SysBatch{}, batchID).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	// 反结账审计（事务提交后写入）
	dao.WriteBusinessAudit(ctx, "反结账", "annual_leave_carryover", batchID, batch.BatchNo,
		"", fmt.Sprintf("冲销系统事件 %d 条", len(events)))
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

	ids := make([]uint, len(events))
	for i, e := range events {
		ids[i] = e.PersonID
	}
	nameMap := PersonNameMap(ids)

	var result []map[string]interface{}
	for _, e := range events {
		item := map[string]interface{}{
			"id":             e.ID,
			"person_id":      e.PersonID,
			"event_type":     e.EventType,
			"hours":          e.Hours,
			"effective_date": e.EffectiveDate,
		}
		item["person_name"] = nameMap[e.PersonID]
		result = append(result, item)
	}
	return result, nil
}
