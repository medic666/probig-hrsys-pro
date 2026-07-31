package service

import (
	"context"
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func TestCarryoverIdempotentReExecute(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		if err := db.AutoMigrate(&model.SysBatch{}, &model.AnnualLeaveBalanceSnapshot{}); err != nil {
			t.Fatalf("migrate: %v", err)
		}
		seedEmployee(db, 50, "2026-01-01", 8000, 2000, 300, 500, 26)

		// 手动配发 40h + 已休 8h → 余额 32
		d0, _ := utils.ParseDate("2026-01-01")
		d0Only := utils.DateOnlyFromTime(d0)
		if err := db.Create(&model.AnnualLeaveAccountEvent{
			PersonID: 50, Seq: 2, EventType: "grant", SourceType: "manual",
			Hours: 40, EffectiveDate: d0Only,
		}).Error; err != nil {
			t.Fatalf("seed grant: %v", err)
		}
		seedConfirmedDetail(t, db, 50, "2026-06-10", "年假", 8)
		assertALBalance(t, db, 50, 32)

		// 第一次结转（周年处于次月 2027-01）
		if _, err := ExecuteCarryover(context.Background(), "2026-12", 1, "admin"); err != nil {
			t.Fatalf("first carryover: %v", err)
		}
		assertALBalance(t, db, 50, 40)

		var batches []model.SysBatch
		db.Find(&batches)
		if len(batches) != 1 {
			t.Fatalf("expected 1 batch after first run, got %d", len(batches))
		}
		firstBatchID := batches[0].ID
		if batches[0].OperatorName != "admin" {
			t.Errorf("batch operator_name should be admin, got %q", batches[0].OperatorName)
		}

		// 第二次结转（幂等）：余额与单次执行一致，旧批次被清除
		if _, err := ExecuteCarryover(context.Background(), "2026-12", 1, "admin"); err != nil {
			t.Fatalf("second carryover: %v", err)
		}
		assertALBalance(t, db, 50, 40)

		batches = nil
		db.Find(&batches)
		if len(batches) != 1 {
			t.Fatalf("expected 1 batch after re-run, got %d (old batch not cleared)", len(batches))
		}
		if batches[0].ID == firstBatchID {
			t.Fatalf("new batch should replace old one")
		}

		// 系统事件仅保留新一批：扣减 1 + 配发 1
		var sysEvents int64
		db.Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ? AND source_type = ?", 50, "system_period").Count(&sysEvents)
		if sysEvents != 2 {
			t.Errorf("expected 2 system events (deduct+grant), got %d", sysEvents)
		}

		// 反结账：余额回到结转前，批次记录清除
		if err := CancelCarryover(context.Background(), batches[0].ID); err != nil {
			t.Fatalf("cancel: %v", err)
		}
		assertALBalance(t, db, 50, 32)
		var batchCount int64
		db.Model(&model.SysBatch{}).Count(&batchCount)
		if batchCount != 0 {
			t.Errorf("batch record should be removed after cancel, got %d", batchCount)
		}
	})
}
