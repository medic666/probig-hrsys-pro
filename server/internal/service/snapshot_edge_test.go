package service

import (
	"context"
	"testing"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// 快照局部重建（以时间为轴）边界测试：
// 改早/改晚跨段、倒填至入职前、同日多事件、删除唯一未来事件、
// 快照表缺失兜底（G1）、月边界移动与 remark-only 编辑（混合方案内容比对）。

func mustDate(t *testing.T, s string) utils.DateOnly {
	t.Helper()
	d, err := utils.ParseDate(s)
	if err != nil {
		t.Fatalf("parse date %s: %v", s, err)
	}
	return utils.DateOnlyFromTime(d)
}

func findSegment(t *testing.T, db *gorm.DB, personID uint, date string) model.PositionSnapshot {
	t.Helper()
	var s model.PositionSnapshot
	if err := db.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
		personID, date, date).First(&s).Error; err != nil {
		t.Fatalf("load segment covering %s: %v", date, err)
	}
	return s
}

func findEvent(t *testing.T, db *gorm.DB, personID uint, date string, eventType string) model.PositionEvent {
	t.Helper()
	var e model.PositionEvent
	if err := db.Where("person_id = ? AND effective_date = ? AND event_type = ?",
		personID, date, eventType).First(&e).Error; err != nil {
		t.Fatalf("load event %s@%s: %v", eventType, date, err)
	}
	return e
}

func countSnapshots(t *testing.T, db *gorm.DB, personID uint) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&model.PositionSnapshot{}).Where("person_id = ?", personID).Count(&n).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	return n
}

// staleOf / salaryStaleOf 断言辅助：加载核算结果并返回 stale 状态
func staleOf(t *testing.T, db *gorm.DB, personID uint, month string) string {
	t.Helper()
	calc := loadCalc(t, db, personID, month)
	return IsAttendanceMonthlyStale(&calc)
}

func salaryStaleOf(t *testing.T, db *gorm.DB, personID uint, month string) string {
	t.Helper()
	s := loadSummary(t, db, personID, month)
	return IsSalarySummaryStale(&s)
}

// B1: 事件生效日改早且跨多段（7/1→5/15，中间有 3/1 段）：中间段截断保留，4 月不误报
func TestEventMovedEarlierAcrossSegments(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 81, "2026-01-01", 8000, 2000, 300, 500, 26)
		marchD := mustDate(t, "2026-03-01")
		julyD := mustDate(t, "2026-07-01")
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 81, EventType: "调薪", EffectiveDate: marchD, BaseSalary: ptrFloat(8500),
		}); err != nil {
			t.Fatalf("create march event: %v", err)
		}
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 81, EventType: "调薪", EffectiveDate: julyD, BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("create july event: %v", err)
		}

		seedAttendanceDays(db, 81, "2026-04", 26, 8)
		seedAttendanceDays(db, 81, "2026-05", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 81, "2026-04"); err != nil {
			t.Fatalf("calc 2026-04: %v", err)
		}
		if _, err := CalculateMonthlyAttendance(context.Background(), 81, "2026-05"); err != nil {
			t.Fatalf("calc 2026-05: %v", err)
		}
		if staleOf(t, db, 81, "2026-04") != "calculated" || staleOf(t, db, 81, "2026-05") != "calculated" {
			t.Fatalf("initial should be calculated")
		}
		segMarchOld := findSegment(t, db, 81, "2026-03-01")

		time.Sleep(5 * time.Millisecond)
		ev := findEvent(t, db, 81, "2026-07-01", "调薪")
		may15 := mustDate(t, "2026-05-15")
		if err := UpdatePositionEvent(context.Background(), ev.ID, &model.PositionEvent{
			EventType: "调薪", EffectiveDate: may15,
		}); err != nil {
			t.Fatalf("move event to 5/15: %v", err)
		}

		// 4 月不误报（覆盖段为截断保留段 [3/1..5/14]），5 月正确标橙
		if st := staleOf(t, db, 81, "2026-04"); st != "calculated" {
			t.Errorf("2026-04 stale = %s, want calculated", st)
		}
		if st := staleOf(t, db, 81, "2026-05"); st != "data_changed" {
			t.Errorf("2026-05 stale = %s, want data_changed", st)
		}

		// 段结构：[1/1..2/28]8000 / [3/1..5/14]8500（时间戳不变）/ [5/15..9999]9000
		segMarchNew := findSegment(t, db, 81, "2026-03-01")
		if segMarchNew.EffectiveEndDate.String() != "2026-05-14" {
			t.Errorf("middle segment end = %s, want 2026-05-14", segMarchNew.EffectiveEndDate.String())
		}
		if !segMarchNew.LastCalcAt.Equal(segMarchOld.LastCalcAt) {
			t.Errorf("middle segment LastCalcAt changed: %v -> %v", segMarchOld.LastCalcAt, segMarchNew.LastCalcAt)
		}
		if segMarchNew.BaseSalary != 8500 {
			t.Errorf("middle segment salary = %v, want 8500", segMarchNew.BaseSalary)
		}
		if seg := findSegment(t, db, 81, "2026-05-20"); seg.BaseSalary != 9000 {
			t.Errorf("rebuilt segment salary = %v, want 9000", seg.BaseSalary)
		}
	})
}

// B2: 事件生效日改晚且跨段（5/1→8/1，中间 6/15 调岗）：中间区间完整回退重建
func TestEventMovedLaterAcrossSegmentsReverts(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 82, "2026-01-01", 8000, 2000, 300, 500, 26)
		mayD := mustDate(t, "2026-05-01")
		june15 := mustDate(t, "2026-06-15")
		dept := "业务部"
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 82, EventType: "调薪", EffectiveDate: mayD, BaseSalary: ptrFloat(8500),
		}); err != nil {
			t.Fatalf("create may event: %v", err)
		}
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 82, EventType: "调岗", EffectiveDate: june15, Department: &dept,
		}); err != nil {
			t.Fatalf("create june event: %v", err)
		}

		seedAttendanceDays(db, 82, "2026-05", 26, 8)
		seedAttendanceDays(db, 82, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 82, "2026-05"); err != nil {
			t.Fatalf("calc 2026-05: %v", err)
		}
		if _, err := CalculateMonthlyAttendance(context.Background(), 82, "2026-06"); err != nil {
			t.Fatalf("calc 2026-06: %v", err)
		}
		if staleOf(t, db, 82, "2026-05") != "calculated" || staleOf(t, db, 82, "2026-06") != "calculated" {
			t.Fatalf("initial should be calculated")
		}
		segJanOld := findSegment(t, db, 82, "2026-01-01")

		time.Sleep(5 * time.Millisecond)
		ev := findEvent(t, db, 82, "2026-05-01", "调薪")
		augD := mustDate(t, "2026-08-01")
		if err := UpdatePositionEvent(context.Background(), ev.ID, &model.PositionEvent{
			EventType: "调薪", EffectiveDate: augD, BaseSalary: ptrFloat(8500),
		}); err != nil {
			t.Fatalf("move event to 8/1: %v", err)
		}

		// 5/6 月标橙（中间区间回退），[1/1..4/30] 保留段时间戳不变
		if st := staleOf(t, db, 82, "2026-05"); st != "data_changed" {
			t.Errorf("2026-05 stale = %s, want data_changed", st)
		}
		if st := staleOf(t, db, 82, "2026-06"); st != "data_changed" {
			t.Errorf("2026-06 stale = %s, want data_changed", st)
		}
		if seg := findSegment(t, db, 82, "2026-01-01"); !seg.LastCalcAt.Equal(segJanOld.LastCalcAt) {
			t.Errorf("preserved segment LastCalcAt changed")
		}

		// 内容正确：5 月回退 8000 无部门；6/15..7/31 = 8000+deptX；8/1+ = 8500+deptX
		segMay := findSegment(t, db, 82, "2026-05-15")
		if segMay.BaseSalary != 8000 || segMay.Department != "" {
			t.Errorf("may segment = salary %v dept %q, want 8000/''", segMay.BaseSalary, segMay.Department)
		}
		segJun := findSegment(t, db, 82, "2026-06-20")
		if segJun.BaseSalary != 8000 || segJun.Department != "业务部" {
			t.Errorf("june segment = salary %v dept %q, want 8000/业务部", segJun.BaseSalary, segJun.Department)
		}
		segAug := findSegment(t, db, 82, "2026-08-15")
		if segAug.BaseSalary != 8500 || segAug.Department != "业务部" {
			t.Errorf("aug segment = salary %v dept %q, want 8500/业务部", segAug.BaseSalary, segAug.Department)
		}
	})
}

// B3: 事件倒填至入职之前（6/1 事件，入职 7/1）：零状态起点 + 入职效果完整
func TestBackdatedEventBeforeFirstSegment(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 83, "2026-07-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 83, "2026-08", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 83, "2026-08"); err != nil {
			t.Fatalf("calc 2026-08: %v", err)
		}
		if staleOf(t, db, 83, "2026-08") != "calculated" {
			t.Fatalf("initial should be calculated")
		}

		time.Sleep(5 * time.Millisecond)
		juneD := mustDate(t, "2026-06-01")
		dept := "业务部"
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 83, EventType: "调岗", EffectiveDate: juneD, Department: &dept,
		}); err != nil {
			t.Fatalf("create backdated event: %v", err)
		}

		if st := staleOf(t, db, 83, "2026-08"); st != "data_changed" {
			t.Errorf("2026-08 stale = %s, want data_changed", st)
		}
		// [6/1..6/30]：零状态 + 新事件效果（无入职继承）
		segJun := findSegment(t, db, 83, "2026-06-15")
		if segJun.Department != "业务部" || segJun.BaseSalary != 0 || segJun.IsActive {
			t.Errorf("pre-entry segment = dept %q salary %v active %v, want 业务部/0/false",
				segJun.Department, segJun.BaseSalary, segJun.IsActive)
		}
		// [7/1..9999]：入职效果完整
		segJul := findSegment(t, db, 83, "2026-07-15")
		if segJul.BaseSalary != 8000 || !segJul.IsActive {
			t.Errorf("entry segment = salary %v active %v, want 8000/true", segJul.BaseSalary, segJul.IsActive)
		}
	})
}

// B4: 同日多事件编辑其一：折叠段内容正确、后 seq 覆盖、无空段
func TestSameDayEventEditViaUpdate(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 84, "2026-01-01", 8000, 2000, 300, 500, 26)
		mayD := mustDate(t, "2026-05-01")
		dept := "业务部"
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 84, EventType: "调薪", EffectiveDate: mayD, BaseSalary: ptrFloat(8500),
		}); err != nil {
			t.Fatalf("create salary event: %v", err)
		}
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 84, EventType: "调岗", EffectiveDate: mayD, Department: &dept,
		}); err != nil {
			t.Fatalf("create dept event: %v", err)
		}

		seedAttendanceDays(db, 84, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 84, "2026-06"); err != nil {
			t.Fatalf("calc 2026-06: %v", err)
		}
		if staleOf(t, db, 84, "2026-06") != "calculated" {
			t.Fatalf("initial should be calculated")
		}

		time.Sleep(5 * time.Millisecond)
		ev := findEvent(t, db, 84, "2026-05-01", "调薪")
		if err := UpdatePositionEvent(context.Background(), ev.ID, &model.PositionEvent{
			EventType: "调薪", EffectiveDate: mayD, BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("update salary event: %v", err)
		}

		if st := staleOf(t, db, 84, "2026-06"); st != "data_changed" {
			t.Errorf("2026-06 stale = %s, want data_changed", st)
		}
		// 同日双事件折叠为单段，后 seq（调岗）不覆盖薪资
		var snaps []model.PositionSnapshot
		db.Where("person_id = ?", 84).Order("effective_start_date ASC").Find(&snaps)
		if len(snaps) != 2 {
			t.Fatalf("expected 2 segments, got %d", len(snaps))
		}
		seg := findSegment(t, db, 84, "2026-06-15")
		if seg.BaseSalary != 9000 || seg.Department != "业务部" {
			t.Errorf("folded segment = salary %v dept %q, want 9000/业务部", seg.BaseSalary, seg.Department)
		}
	})
}

// B5: 删除唯一未来事件：过去段不误报、未来段回退标橙；再删唯一事件 → 快照清空
func TestDeleteOnlyFutureEvent(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 85, "2026-01-01", 8000, 2000, 300, 500, 26)
		mayD := mustDate(t, "2026-05-01")
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 85, EventType: "调薪", EffectiveDate: mayD, BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("create may event: %v", err)
		}

		seedAttendanceDays(db, 85, "2026-04", 26, 8)
		seedAttendanceDays(db, 85, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 85, "2026-04"); err != nil {
			t.Fatalf("calc 2026-04: %v", err)
		}
		if _, err := CalculateMonthlyAttendance(context.Background(), 85, "2026-06"); err != nil {
			t.Fatalf("calc 2026-06: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 85, "2026-06", 1, "admin"); err != nil {
			t.Fatalf("calc salary 2026-06: %v", err)
		}
		if badges, err := GetSalarySummariesBadges("2026-06"); err != nil || badgeLevelOf(badges, 85) != "green" {
			t.Fatalf("initial salary badge: err=%v level=%s, want green", err, badgeLevelOf(badges, 85))
		}

		time.Sleep(5 * time.Millisecond)
		ev := findEvent(t, db, 85, "2026-05-01", "调薪")
		if err := DeletePositionEvent(context.Background(), ev.ID); err != nil {
			t.Fatalf("delete may event: %v", err)
		}

		// 4 月不误报；6 月回退标橙（内容真变）
		if st := staleOf(t, db, 85, "2026-04"); st != "calculated" {
			t.Errorf("2026-04 stale = %s, want calculated", st)
		}
		if st := staleOf(t, db, 85, "2026-06"); st != "data_changed" {
			t.Errorf("2026-06 stale = %s, want data_changed", st)
		}
		if badges, err := GetSalarySummariesBadges("2026-06"); err != nil || badgeLevelOf(badges, 85) != "orange" {
			t.Errorf("salary badge: err=%v level=%s, want orange", err, badgeLevelOf(badges, 85))
		}
		if seg := findSegment(t, db, 85, "2026-06-15"); seg.BaseSalary != 8000 {
			t.Errorf("reverted segment salary = %v, want 8000", seg.BaseSalary)
		}

		// 删除唯一剩余事件 → 快照清空
		entryEv := findEvent(t, db, 85, "2026-01-01", "入职")
		if err := DeletePositionEvent(context.Background(), entryEv.ID); err != nil {
			t.Fatalf("delete entry event: %v", err)
		}
		if n := countSnapshots(t, db, 85); n != 0 {
			t.Errorf("snapshots after deleting only event = %d, want 0", n)
		}
	})
}

// B6: 快照表物理清空后重建（G1 兜底）：起始状态由 cut 前事件回放得出，入职效果不丢失
func TestRebuildAfterSnapshotLoss(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 86, "2026-01-01", 8000, 2000, 300, 500, 26)
		mayD := mustDate(t, "2026-05-01")
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 86, EventType: "调薪", EffectiveDate: mayD, BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("create may event: %v", err)
		}
		// 模拟快照表缺失/损坏
		db.Where("person_id = ?", 86).Delete(&model.PositionSnapshot{})

		time.Sleep(5 * time.Millisecond)
		ev := findEvent(t, db, 86, "2026-05-01", "调薪")
		if err := UpdatePositionEvent(context.Background(), ev.ID, &model.PositionEvent{
			EventType: "调薪", EffectiveDate: mayD, BaseSalary: ptrFloat(9500),
		}); err != nil {
			t.Fatalf("update after snapshot loss: %v", err)
		}

		// G1 兜底回放入职事件：段仍含入职效果
		seg := findSegment(t, db, 86, "2026-06-15")
		if seg.BaseSalary != 9500 {
			t.Errorf("segment salary = %v, want 9500", seg.BaseSalary)
		}
		if seg.EntryDate == nil || seg.EntryDate.String() != "2026-01-01" || !seg.IsActive {
			t.Errorf("entry effects lost: entry=%v active=%v", seg.EntryDate, seg.IsActive)
		}
	})
}

// B7: 事件跨月边界移动（3/31→4/1，混合方案核心）：3 月标橙（当天真变）、4 月不误报（内容未变）
func TestEventMoveAcrossMonthBoundary(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 87, "2026-01-01", 8000, 2000, 300, 500, 26)
		marchEnd := mustDate(t, "2026-03-31")
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 87, EventType: "调薪", EffectiveDate: marchEnd, BaseSalary: ptrFloat(8500),
		}); err != nil {
			t.Fatalf("create 3/31 event: %v", err)
		}

		seedAttendanceDays(db, 87, "2026-03", 26, 8)
		seedAttendanceDays(db, 87, "2026-04", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 87, "2026-03"); err != nil {
			t.Fatalf("calc 2026-03: %v", err)
		}
		if _, err := CalculateMonthlyAttendance(context.Background(), 87, "2026-04"); err != nil {
			t.Fatalf("calc 2026-04: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 87, "2026-04", 1, "admin"); err != nil {
			t.Fatalf("calc salary 2026-04: %v", err)
		}
		if staleOf(t, db, 87, "2026-03") != "calculated" || staleOf(t, db, 87, "2026-04") != "calculated" {
			t.Fatalf("initial should be calculated")
		}

		time.Sleep(5 * time.Millisecond)
		ev := findEvent(t, db, 87, "2026-03-31", "调薪")
		aprilD := mustDate(t, "2026-04-01")
		if err := UpdatePositionEvent(context.Background(), ev.ID, &model.PositionEvent{
			EventType: "调薪", EffectiveDate: aprilD, BaseSalary: ptrFloat(8500),
		}); err != nil {
			t.Fatalf("move event to 4/1: %v", err)
		}

		// 3 月标橙：3/31 当天 8500→8000 内容真变
		if st := staleOf(t, db, 87, "2026-03"); st != "data_changed" {
			t.Errorf("2026-03 stale = %s, want data_changed", st)
		}
		if seg := findSegment(t, db, 87, "2026-03-31"); seg.BaseSalary != 8000 {
			t.Errorf("3/31 segment salary = %v, want 8000 (reverted)", seg.BaseSalary)
		}
		// 4 月内容未变（8500→8500）：混合方案沿用旧时间戳，不标橙
		if st := staleOf(t, db, 87, "2026-04"); st != "calculated" {
			t.Errorf("2026-04 stale = %s, want calculated (content unchanged)", st)
		}
		if st := salaryStaleOf(t, db, 87, "2026-04"); st != "calculated" {
			t.Errorf("2026-04 salary stale = %s, want calculated", st)
		}
		if badges, err := GetSalarySummariesBadges("2026-04"); err != nil || badgeLevelOf(badges, 87) != "green" {
			t.Errorf("2026-04 salary badge: err=%v level=%s, want green", err, badgeLevelOf(badges, 87))
		}
	})
}

// B8: 仅改 remark（no-op 编辑，混合方案）：内容未变月份不标橙
func TestRemarkOnlyEditDoesNotFlag(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 88, "2026-01-01", 8000, 2000, 300, 500, 26)
		mayD := mustDate(t, "2026-05-01")
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 88, EventType: "调薪", EffectiveDate: mayD, BaseSalary: ptrFloat(9000), Remark: "初始",
		}); err != nil {
			t.Fatalf("create may event: %v", err)
		}

		seedAttendanceDays(db, 88, "2026-05", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 88, "2026-05"); err != nil {
			t.Fatalf("calc 2026-05: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 88, "2026-05", 1, "admin"); err != nil {
			t.Fatalf("calc salary 2026-05: %v", err)
		}
		if staleOf(t, db, 88, "2026-05") != "calculated" {
			t.Fatalf("initial should be calculated")
		}

		time.Sleep(5 * time.Millisecond)
		ev := findEvent(t, db, 88, "2026-05-01", "调薪")
		if err := UpdatePositionEvent(context.Background(), ev.ID, &model.PositionEvent{
			EventType: "调薪", EffectiveDate: mayD, Remark: "修正",
		}); err != nil {
			t.Fatalf("update remark only: %v", err)
		}

		// 内容字段未变 → 沿用旧时间戳 → 不标橙
		if st := staleOf(t, db, 88, "2026-05"); st != "calculated" {
			t.Errorf("2026-05 stale = %s, want calculated (remark-only edit)", st)
		}
		if st := salaryStaleOf(t, db, 88, "2026-05"); st != "calculated" {
			t.Errorf("2026-05 salary stale = %s, want calculated", st)
		}
		if badges, err := GetAttendanceMonthlyBadges("2026-05"); err != nil || badgeLevelOf(badges, 88) != "green" {
			t.Errorf("attendance badge: err=%v level=%s, want green", err, badgeLevelOf(badges, 88))
		}
	})
}
