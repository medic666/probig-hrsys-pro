package service

import (
	"context"
	"testing"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func migrateCarryoverTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&model.SysBatch{}, &model.AnnualLeaveBalanceSnapshot{}, &model.LeaveInLieuBalanceSnapshot{}); err != nil {
		t.Fatalf("migrate carryover tables: %v", err)
	}
}

func seedManualGrant(t *testing.T, db *gorm.DB, personID uint, date string, hours float64) {
	t.Helper()
	d, _ := utils.ParseDate(date)
	dOnly := utils.DateOnlyFromTime(d)
	var maxSeq int
	db.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ?", personID).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
	if err := db.Create(&model.AnnualLeaveAccountEvent{
		PersonID: personID, Seq: maxSeq + 1, EventType: "grant", SourceType: "manual",
		Hours: hours, EffectiveDate: dOnly,
	}).Error; err != nil {
		t.Fatalf("seed grant: %v", err)
	}
	if err := RebuildAnnualLeaveBalance(db, personID); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
}

// TestCarryoverAnniversarySettlement 周年结算：结算月最后一日末余额>0 → deduct + 周年配发
func TestCarryoverAnniversarySettlement(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateCarryoverTables(t, db)
		// 入职 2025-07-15 → 周年月 2026-07，结算月 2026-06
		seedEmployee(db, 70, "2025-07-15", 8000, 2000, 300, 500, 26)
		seedManualGrant(t, db, 70, "2025-08-01", 40)
		seedConfirmedDetail(t, db, 70, "2026-06-10", "年假", 8) // 余额 32

		result, err := ExecuteCarryover(context.Background(), "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("carryover: %v", err)
		}
		if result["success"] != 1 {
			t.Fatalf("expected 1 success, got %v", result["success"])
		}

		// deduct：结算月末快照余额 32
		var deduct model.AnnualLeaveAccountEvent
		if err := db.Where("person_id = ? AND event_type = ?", 70, "carryover_deduct").First(&deduct).Error; err != nil {
			t.Fatalf("deduct missing: %v", err)
		}
		if deduct.Hours != 32 || deduct.EffectiveDate.String() != "2026-06-30" {
			t.Errorf("deduct: hours=%v date=%v, want 32 / 2026-06-30", deduct.Hours, deduct.EffectiveDate.String())
		}
		// grant：周年月首日阶梯配发（司龄 1 年 → 40h）
		var grant model.AnnualLeaveAccountEvent
		if err := db.Where("person_id = ? AND event_type = ? AND source_type = ?", 70, "grant", "system_period").First(&grant).Error; err != nil {
			t.Fatalf("grant missing: %v", err)
		}
		if grant.Hours != 40 || grant.EffectiveDate.String() != "2026-07-01" {
			t.Errorf("grant: hours=%v date=%v, want 40 / 2026-07-01", grant.Hours, grant.EffectiveDate.String())
		}
		assertALBalance(t, db, 70, 40)
	})
}

// TestCarryoverResignationSettlement 离职月结算边界：所选月==离职月 → 仅结算不配发
func TestCarryoverResignationSettlement(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateCarryoverTables(t, db)
		seedEmployee(db, 71, "2024-01-01", 8000, 2000, 300, 500, 26)
		seedManualGrant(t, db, 71, "2024-01-15", 40)
		seedConfirmedDetail(t, db, 71, "2026-05-10", "年假", 8) // 余额 32
		seedLeave(db, 71, "2026-06-15")                        // 离职月 2026-06

		result, err := ExecuteCarryover(context.Background(), "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("carryover: %v", err)
		}
		if result["success"] != 1 {
			t.Fatalf("expected 1 success, got %v", result["success"])
		}

		var deduct model.AnnualLeaveAccountEvent
		if err := db.Where("person_id = ? AND event_type = ?", 71, "carryover_deduct").First(&deduct).Error; err != nil {
			t.Fatalf("resignation deduct missing: %v", err)
		}
		if deduct.Hours != 32 {
			t.Errorf("resignation deduct hours=%v, want 32", deduct.Hours)
		}
		// 离职结算不配发
		var grantCount int64
		db.Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ? AND event_type = ? AND source_type = ?", 71, "grant", "system_period").Count(&grantCount)
		if grantCount != 0 {
			t.Errorf("resignation settlement should not grant, got %d", grantCount)
		}
		assertALBalance(t, db, 71, 0)
	})
}

// TestCarryoverAnyAnniversary 任意周年（第二个周年）均匹配
func TestCarryoverAnyAnniversary(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateCarryoverTables(t, db)
		seedEmployee(db, 72, "2024-07-01", 8000, 2000, 300, 500, 26) // 周年月每年 7 月
		seedManualGrant(t, db, 72, "2024-07-01", 40)
		seedConfirmedDetail(t, db, 72, "2026-06-10", "年假", 40) // 余额 0

		// 第二个周年 2026-07，结算月 2026-06（与入职差 24 个月）
		result, err := ExecuteCarryover(context.Background(), "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("carryover: %v", err)
		}
		if result["success"] != 1 {
			t.Fatalf("second anniversary should settle, got %v", result)
		}
		assertALBalance(t, db, 72, 40) // 余额 0 不 deduct，仅配发 40
	})
}

// TestBalanceRebuildPriorityOrdering 同日多变更次序：配发先于考勤、结转后于考勤
func TestBalanceRebuildPriorityOrdering(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateCarryoverTables(t, db)
		db.Create(&model.Person{ID: 73, Name: "排序测试"})

		// 2026-06-30：先考勤消费 8（priority 2）再系统结转 20（priority 3）→ 当日余额 40-8-20=12
		grantD, _ := utils.ParseDate("2026-01-01")
		deductD, _ := utils.ParseDate("2026-06-30")
		db.Create(&model.AnnualLeaveAccountEvent{PersonID: 73, Seq: 1, EventType: "grant", SourceType: "manual", Hours: 40, EffectiveDate: utils.DateOnlyFromTime(grantD)})
		db.Create(&model.AnnualLeaveAccountEvent{PersonID: 73, Seq: 2, EventType: "carryover_deduct", SourceType: "system_period", Hours: 20, EffectiveDate: utils.DateOnlyFromTime(deductD)})
		seedConfirmedDetail(t, db, 73, "2026-06-30", "年假", 8)

		RebuildAnnualLeaveBalance(db, 73)
		bal, ok := GetAnnualLeaveBalanceAt(db, 73, utils.DateOnlyFromTime(deductD))
		if !ok || bal != 12 {
			t.Errorf("2026-06-30 balance = %v (ok=%v), want 12 (消费优先于结转)", bal, ok)
		}

		// 2026-08-01：先系统配发 40（priority 0）再考勤消费 8（priority 2）→ 当日余额 40-8=32
		grant2D, _ := utils.ParseDate("2026-08-01")
		consume2D, _ := utils.ParseDate("2026-08-01")
		db.Create(&model.AnnualLeaveAccountEvent{PersonID: 73, Seq: 3, EventType: "grant", SourceType: "system_period", Hours: 40, EffectiveDate: utils.DateOnlyFromTime(grant2D)})
		seedConfirmedDetail(t, db, 73, "2026-08-01", "年假", 8)

		RebuildAnnualLeaveBalance(db, 73)
		bal2, ok2 := GetAnnualLeaveBalanceAt(db, 73, utils.DateOnlyFromTime(consume2D))
		if !ok2 || bal2 != 44 {
			t.Errorf("2026-08-01 balance = %v (ok=%v), want 44 (12基底+配发40-消费8)", bal2, ok2)
		}
	})
}

// TestAnnualLeaveEventBadges 周年月未结转橙点
func TestAnnualLeaveEventBadges(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateCarryoverTables(t, db)
		now := time.Now()
		entry := "2019-08-15" // 周年月每年 8 月 == 当前月
		seedEmployee(db, 80, entry, 8000, 2000, 300, 500, 26)
		seedManualGrant(t, db, 80, "2019-08-01", 40) // 余额 40 > 0 → 未结转

		seedEmployee(db, 81, "2020-01-01", 8000, 2000, 300, 500, 26) // 周年月 1 月 ≠ 当前月
		seedManualGrant(t, db, 81, "2020-01-01", 40)

		_ = now
		badges, err := GetAnnualLeaveEventBadges(context.Background())
		if err != nil {
			t.Fatalf("badges: %v", err)
		}
		if badgeLevelOf(badges, 80) != "orange" {
			t.Errorf("person 80 (周年月未结转) want orange, got %s", badgeLevelOf(badges, 80))
		}
		if badgeLevelOf(badges, 81) != "green" {
			t.Errorf("person 81 (非周年月) want green, got %s", badgeLevelOf(badges, 81))
		}
	})
}

// TestAnnualLeaveTiers 阶梯配发解析：单值兼容 + 按司龄取档
func TestAnnualLeaveTiers(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		// 单值兼容（老库配置 "40"）
		db.Model(&model.SysConfig{}).Where("config_key = ?", "annual_leave.yearly_hours").
			Update("config_value", "40")
		RefreshConfig(db, "annual_leave.yearly_hours")
		tiers, ok := GetAnnualLeaveTiers()
		if !ok || len(tiers) != 1 || tiers[0].Hours != 40 {
			t.Fatalf("legacy single value parse: %v %v", tiers, ok)
		}

		// 阶梯配置（下界语义：满 X 司龄年配发；默认 seed 已是新语义阶梯 JSON）
		db.Model(&model.SysConfig{}).Where("config_key = ?", "annual_leave.yearly_hours").
			Update("config_value", `[{"years":1,"hours":40},{"years":10,"hours":80},{"years":20,"hours":120}]`)
		RefreshConfig(db, "annual_leave.yearly_hours")

		// 司龄 1 年 → 40h；司龄 15 年 → 80h；司龄 25 年 → 120h（配发生效年 2026）
		seedEmployee(db, 90, "2025-01-01", 8000, 2000, 300, 500, 26)
		seedEmployee(db, 91, "2011-01-01", 8000, 2000, 300, 500, 26)
		seedEmployee(db, 92, "2001-01-01", 8000, 2000, 300, 500, 26)
		if h := getYearlyAnnualLeaveHours(90, 2026); h != 40 {
			t.Errorf("seniority 1y want 40, got %v", h)
		}
		if h := getYearlyAnnualLeaveHours(91, 2026); h != 80 {
			t.Errorf("seniority 15y want 80, got %v", h)
		}
		if h := getYearlyAnnualLeaveHours(92, 2026); h != 120 {
			t.Errorf("seniority 25y want 120, got %v", h)
		}

		// 边界：恰满 10 / 20 年命中对应档；未达任何门槛回退第一档
		seedEmployee(db, 93, "2016-01-01", 8000, 2000, 300, 500, 26)
		seedEmployee(db, 94, "2006-01-01", 8000, 2000, 300, 500, 26)
		seedEmployee(db, 95, "2026-01-01", 8000, 2000, 300, 500, 26)
		if h := getYearlyAnnualLeaveHours(93, 2026); h != 80 {
			t.Errorf("seniority 10y (boundary) want 80, got %v", h)
		}
		if h := getYearlyAnnualLeaveHours(94, 2026); h != 120 {
			t.Errorf("seniority 20y (boundary) want 120, got %v", h)
		}
		if h := getYearlyAnnualLeaveHours(95, 2026); h != 40 {
			t.Errorf("seniority 0y (fallback first tier) want 40, got %v", h)
		}

		// 执行年无关性：配发额度由配发生效年（结算月+1）决定，与操作时间无关
		seedEmployee(db, 96, "2015-01-01", 8000, 2000, 300, 500, 26)
		if h := getYearlyAnnualLeaveHours(96, 2025); h != 80 {
			t.Errorf("grant year 2025 (seniority 10y) want 80, got %v", h)
		}
		if h := getYearlyAnnualLeaveHours(96, 2024); h != 40 {
			t.Errorf("grant year 2024 (seniority 9y) want 40, got %v", h)
		}
	})
}
