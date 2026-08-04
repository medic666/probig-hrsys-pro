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

// TestCarryoverIdempotentOrphanCleanup 幂等重跑防孤儿：旧批次人员不在新候选名单时，
// 其旧事件也必须被删除（事件不得残留为 batch_id 指向已删批次的孤儿）。
func TestCarryoverIdempotentOrphanCleanup(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateCarryoverTables(t, db)
		// A/B 均 2023-08 入职 → 周年月 8 月，结算 2024-07 命中首周年
		seedEmployee(db, 60, "2023-08-01", 8000, 2000, 300, 500, 26)
		seedEmployee(db, 61, "2023-08-01", 8000, 2000, 300, 500, 26)

		if _, err := ExecuteCarryover(context.Background(), "2024-07", 1, "admin"); err != nil {
			t.Fatalf("first carryover: %v", err)
		}
		var first []model.SysBatch
		db.Find(&first)
		if len(first) != 1 {
			t.Fatalf("expected 1 batch after first run, got %d", len(first))
		}
		firstBatchID := first[0].ID
		// 两人均产生周年配发事件
		var liveEvents int64
		db.Model(&model.AnnualLeaveAccountEvent{}).
			Where("source_type = ? AND batch_id = ?", "system_period", firstBatchID).Count(&liveEvents)
		if liveEvents != 2 {
			t.Fatalf("expected 2 events after first run, got %d", liveEvents)
		}

		// 模拟快照变更：A 入职日期改到 2024-08 → 不再满足 2024-08 周年（diff=0）
		if err := db.Exec("UPDATE position_snapshots SET entry_date = '2024-08-01' WHERE person_id = ?", 60).Error; err != nil {
			t.Fatalf("update snapshot: %v", err)
		}

		// 重跑同结算月：仅 B 仍是候选；A 的旧事件必须一并删除，不得成为孤儿
		if _, err := ExecuteCarryover(context.Background(), "2024-07", 1, "admin"); err != nil {
			t.Fatalf("re-run carryover: %v", err)
		}

		var batches []model.SysBatch
		db.Find(&batches)
		if len(batches) != 1 {
			t.Fatalf("expected 1 batch after re-run, got %d", len(batches))
		}
		newBatchID := batches[0].ID
		if newBatchID == firstBatchID {
			t.Fatalf("new batch should replace old one")
		}

		// A：旧事件已删除（无存活系统事件）；无事件即无余额快照（空态，Rebuild 不生成）
		var aEvents int64
		db.Model(&model.AnnualLeaveAccountEvent{}).
			Where("person_id = ? AND source_type = ?", 60, "system_period").Count(&aEvents)
		if aEvents != 0 {
			t.Errorf("person A old events should be removed, got %d", aEvents)
		}
		var aSnaps int64
		db.Model(&model.AnnualLeaveBalanceSnapshot{}).Where("person_id = ?", 60).Count(&aSnaps)
		if aSnaps != 0 {
			t.Errorf("person A should have no balance snapshot (no events), got %d", aSnaps)
		}

		// B：新批次唯一配发事件（余额 0 无扣减，仅配发）
		var bEvents []model.AnnualLeaveAccountEvent
		db.Where("person_id = ? AND source_type = ?", 61, "system_period").Find(&bEvents)
		if len(bEvents) != 1 || bEvents[0].BatchID == nil || *bEvents[0].BatchID != newBatchID {
			t.Errorf("person B should have exactly 1 event of new batch, got %d events", len(bEvents))
		}
		assertALBalance(t, db, 61, 40)

		// 全局孤儿检查：不存在 batch_id 指向已删批次且未删除的系统事件
		var orphans int64
		db.Model(&model.AnnualLeaveAccountEvent{}).
			Where("source_type = ? AND batch_id IS NOT NULL AND batch_id NOT IN (SELECT id FROM sys_batches)",
				"system_period").Count(&orphans)
		if orphans != 0 {
			t.Errorf("orphan events should be 0, got %d", orphans)
		}
	})
}
