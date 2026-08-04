package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	monthEnd := monthStart.AddDate(0, 1, -1)
	monthStartD := utils.DateOnlyFromTime(monthStart)
	monthEndD := utils.DateOnlyFromTime(monthEnd)

	var snapshots []model.PositionSnapshot
	dao.DB.Where("effective_start_date <= ? AND effective_end_date >= ? AND is_active = true",
		utils.DateOnlyFromTime(nextMonthEnd), monthStartD).Find(&snapshots)

	// 周年结算资格：所选月+1 == 员工入职周年月（入职次年同月），且当月在职。
	// 离职结算边界：员工离职月 == 所选月 → 离职结算（仅结算不配发）。
	type settlementKind int
	const (
		settleAnniversary settlementKind = iota
		settleResignation
	)
	type candidate struct {
		personID uint
		kind     settlementKind
	}
	candidateMap := make(map[uint]settlementKind)
	nextYear, nextMonthNum := nextMonthStart.Year(), int(nextMonthStart.Month())
	nextMonths := nextYear*12 + nextMonthNum - 1
	for _, s := range snapshots {
		if s.EntryDate != nil {
			entryMonths := s.EntryDate.Time().Year()*12 + int(s.EntryDate.Time().Month()) - 1
			diff := nextMonths - entryMonths
			// 周年月判定：所选月+1 与入职月同月（月份差为 12 的倍数）且已满一周年（差 ≥ 12）
			if diff >= 12 && diff%12 == 0 {
				candidateMap[s.PersonID] = settleAnniversary
			}
		}
	}

	// 离职月结算：离职发生在所选月内（含所选月末离岗）的离职人员
	var leftSnaps []model.PositionSnapshot
	dao.DB.Where("effective_start_date <= ? AND effective_end_date >= ? AND is_active = false",
		monthEndD, monthStartD).
		Where("leave_date IS NOT NULL AND strftime('%Y-%m', leave_date) = ?", month).
		Find(&leftSnaps)
	for _, s := range leftSnaps {
		if _, exists := candidateMap[s.PersonID]; !exists {
			candidateMap[s.PersonID] = settleResignation
		}
	}

	if len(candidateMap) == 0 {
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
		TotalCount:     len(candidateMap),
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

	success := 0
	fail := 0
	now := time.Now()

	for personID, kind := range candidateMap {
		err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
			// 冲销历史同周期批次的系统事件，事件源删除由 GORM 审计自动留痕；
			// 冲销后立即重建余额快照，保证结算判定读到删除后（无旧 deduct/grant）的真实余额
			if len(oldBatchIDs) > 0 {
				if err := tx.Where("person_id = ? AND source_type = ? AND batch_id IN ?",
					personID, "system_period", oldBatchIDs).Delete(&model.AnnualLeaveAccountEvent{}).Error; err != nil {
					return err
				}
				if err := RebuildAnnualLeaveBalance(tx, personID); err != nil {
					return err
				}
			}

			// 结算：所选月（结算月）最后一日末的余额快照 > 0 → 结转扣除
			balance, ok := GetAnnualLeaveBalanceAt(tx, personID, monthEndD)
			if ok && balance > 0 {
				remark := fmt.Sprintf("周年结算扣除 %s", nextMonthStart.Format("2006-01"))
				if kind == settleResignation {
					remark = fmt.Sprintf("离职结算扣除 %s", month)
				}
				deduct := model.AnnualLeaveAccountEvent{
					PersonID:      personID,
					EventType:     "carryover_deduct",
					SourceType:    "system_period",
					BatchID:       &batch.ID,
					Hours:         balance,
					EffectiveDate: monthEndD,
					Remark:        remark,
				}
				if err := createSystemLeaveEventInTx(tx, &deduct); err != nil {
					return err
				}
			}

			// 配发：仅周年结算；入职周年月第一日末仍在职且享有年假 → 按司龄阶梯配发
			if kind == settleAnniversary {
				var curSnap model.PositionSnapshot
				tx.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ? AND is_active = true",
					personID, utils.DateOnlyFromTime(nextMonthEnd), utils.DateOnlyFromTime(nextMonthStart)).First(&curSnap)
				if curSnap.HasAnnualLeave {
					grant := model.AnnualLeaveAccountEvent{
						PersonID:      personID,
						EventType:     "grant",
						SourceType:    "system_period",
						BatchID:       &batch.ID,
						Hours:         getYearlyAnnualLeaveHours(personID),
						EffectiveDate: utils.DateOnlyFromTime(nextMonthStart),
						Remark:        fmt.Sprintf("周年配发 %s", nextMonthStart.Format("2006-01")),
					}
					if err := createSystemLeaveEventInTx(tx, &grant); err != nil {
						return err
					}
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
			"success": success, "total": len(candidateMap),
		})
		dao.WriteBusinessAudit(ctx, "结转", "annual_leave_carryover", batch.ID, batchNo, "", string(summaryJSON))
	}

	return map[string]interface{}{
		"batch_no": batchNo,
		"success":  success,
		"fail":     fail,
		"total":    len(candidateMap),
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

// AnnualLeaveTier 年假阶梯档位：司龄满 years 年后按 hours 配发
type AnnualLeaveTier struct {
	Years int     `json:"years"`
	Hours float64 `json:"hours"`
}

// GetAnnualLeaveTiers 解析年假配发阶梯配置：
// 兼容旧单值配置（纯数字 = 单档，所有司龄同一小时数）；新配置为 JSON 数组 [{years,hours}...]。
func GetAnnualLeaveTiers() ([]AnnualLeaveTier, bool) {
	v := GetConfigValueOrDefault("annual_leave.yearly_hours", "")
	if v == "" {
		return nil, false
	}
	if n, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
		return []AnnualLeaveTier{{Years: 0, Hours: n}}, true
	}
	var tiers []AnnualLeaveTier
	if err := json.Unmarshal([]byte(v), &tiers); err != nil || len(tiers) == 0 {
		return nil, false
	}
	return tiers, true
}

// getYearlyAnnualLeaveHours 按员工司龄（入职年数）取年假配发小时数。
func getYearlyAnnualLeaveHours(personID uint) float64 {
	tiers, ok := GetAnnualLeaveTiers()
	if !ok {
		return 40
	}
	seniority := 0
	var entry *utils.DateOnly
	dao.DB.Table("position_snapshots").
		Select("entry_date").
		Where("person_id = ? AND effective_end_date = ?", personID, realFarFuture).
		Scan(&entry)
	if entry != nil {
		seniority = time.Now().Year() - entry.Time().Year()
	}
	// 取第一个"司龄未达门槛"的档位（门槛从小到大排序）
	for _, t := range tiers {
		if seniority < t.Years || t.Years == 0 {
			return t.Hours
		}
	}
	return tiers[len(tiers)-1].Hours
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
