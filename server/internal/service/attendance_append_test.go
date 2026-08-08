package service

import (
	"context"
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// appendDaily 追加一条考勤组（默认按传入状态），返回当日最新组
func appendDaily(t *testing.T, db *gorm.DB, personID uint, date, status string, details ...model.AttendanceEventDetail) model.AttendanceDaily {
	t.Helper()
	d, _ := utils.ParseDate(date)
	dOnly := utils.DateOnlyFromTime(d)
	if err := db.Transaction(func(tx *gorm.DB) error {
		return AppendAttendanceDaily(tx, AttendanceDailyUpsert{
			PersonID: personID, Date: dOnly, Status: &status, Details: details,
		})
	}); err != nil {
		t.Fatalf("append daily: %v", err)
	}
	var daily model.AttendanceDaily
	if err := db.Where("person_id = ? AND event_date = ?", personID, dOnly).Order("seq DESC").First(&daily).Error; err != nil {
		t.Fatalf("load appended daily: %v", err)
	}
	return daily
}

func detailRow(evType, subType string, hours float64) model.AttendanceEventDetail {
	return model.AttendanceEventDetail{EventType: evType, SubType: subType, Hours: hours}
}

func dailyOf(t *testing.T, db *gorm.DB, id uint) model.AttendanceDaily {
	t.Helper()
	var daily model.AttendanceDaily
	if err := db.First(&daily, id).Error; err != nil {
		t.Fatalf("load daily %d: %v", id, err)
	}
	return daily
}

func dailyDetails(t *testing.T, db *gorm.DB, dailyID uint) []model.AttendanceEventDetail {
	t.Helper()
	var details []model.AttendanceEventDetail
	db.Where("daily_id = ?", dailyID).Find(&details)
	return details
}

func confirmedCountOfDay(t *testing.T, db *gorm.DB, personID uint, date string) int {
	t.Helper()
	d, _ := utils.ParseDate(date)
	var count int64
	db.Model(&model.AttendanceDaily{}).
		Where("person_id = ? AND event_date = ? AND deleted_at IS NULL AND status = ?", personID, utils.DateOnlyFromTime(d), "confirmed").
		Count(&count)
	return int(count)
}

func maxSeqOfDay(t *testing.T, db *gorm.DB, personID uint, date string) int {
	t.Helper()
	d, _ := utils.ParseDate(date)
	var daily model.AttendanceDaily
	if err := db.Where("person_id = ? AND event_date = ?", personID, utils.DateOnlyFromTime(d)).
		Order("seq DESC").First(&daily).Error; err != nil {
		t.Fatalf("load newest daily: %v", err)
	}
	return daily.Seq
}

func assertDailyProjectionStatus(t *testing.T, db *gorm.DB, personID uint, date, status string) {
	t.Helper()
	d, _ := utils.ParseDate(date)
	var p model.AttendanceDailyProjection
	if err := db.Where("person_id = ? AND work_date = ?", personID, utils.DateOnlyFromTime(d)).First(&p).Error; err != nil {
		t.Fatalf("load projection: %v", err)
	}
	if p.Status != status {
		t.Errorf("projection status = %s, want %s", p.Status, status)
	}
}

// TestPendingDailyListDayLevel 待确认页日级语义：仅当日最新组为 pending 的记录，
// 陈旧 pending（同日有更新的 confirmed）不进入待确认页
func TestPendingDailyListDayLevel(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 9, "2026-01-01", 8000, 2000, 300, 500, 26)
		// 日 A：最新 confirmed + 陈旧 pending → 不进待确认页
		appendDaily(t, db, 9, "2026-06-01", "pending", detailRow("出勤", "普通出勤", 8))
		appendDaily(t, db, 9, "2026-06-01", "confirmed", detailRow("出勤", "普通出勤", 8))
		// 日 B：最新 pending → 进待确认页
		appendDaily(t, db, 9, "2026-06-02", "pending", detailRow("出勤", "普通出勤", 8))

		list, total, err := GetPendingDailyList(context.Background(), AttendanceDailyListQuery{PageNum: 1, PageSize: 20})
		if err != nil {
			t.Fatalf("pending list: %v", err)
		}
		if total != 1 {
			t.Fatalf("pending total = %d, want 1 (仅日B)", total)
		}
		if len(list) != 1 {
			t.Fatalf("pending list len = %d, want 1", len(list))
		}
		date, ok := list[0]["event_date"].(utils.DateOnly)
		if !ok || date.String() != "2026-06-02" {
			t.Errorf("pending row should be 2026-06-02, got %v", list[0]["event_date"])
		}
	})
}

// TestProjectionMakeupCountsButLILNot 补班出勤计出勤工时，调休不计（避免调休 8h 与补班重复记工时）
func TestProjectionMakeupCountsButLILNot(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		seedEmployee(db, 10, "2026-01-01", 8000, 2000, 300, 500, 26)
		// 补班出勤 8h → work_hours 8
		appendDaily(t, db, 10, "2026-06-01", "confirmed", detailRow("出勤", "补班出勤", 8))
		// 调休 8h → work_hours 0（仅为调休余额消费）
		appendDaily(t, db, 10, "2026-06-02", "confirmed", detailRow("休假", "调休", 8))
		// 出勤 4h + 调休 4h → work_hours 4
		appendDaily(t, db, 10, "2026-06-03", "confirmed", detailRow("出勤", "普通出勤", 4), detailRow("休假", "调休", 4))

		assertProjectionWorkHours(t, db, 10, "2026-06-01", 8)
		assertProjectionWorkHours(t, db, 10, "2026-06-02", 0)
		assertProjectionWorkHours(t, db, 10, "2026-06-03", 4)
	})
}

func assertProjectionWorkHours(t *testing.T, db *gorm.DB, personID uint, date string, want float64) {
	t.Helper()
	d, _ := utils.ParseDate(date)
	var p model.AttendanceDailyProjection
	if err := db.Where("person_id = ? AND work_date = ?", personID, utils.DateOnlyFromTime(d)).First(&p).Error; err != nil {
		t.Fatalf("load projection: %v", err)
	}
	if p.WorkHours != want {
		t.Errorf("work_hours = %v, want %v", p.WorkHours, want)
	}
}

// TestGetLILEventList 调休事件列表：仅已确认组的补班/调休明细，先过滤后分页
func TestGetLILEventList(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 20, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedLILRow := func(date, status, evType, subType string, hours float64) {
			d, _ := utils.ParseDate(date)
			daily := model.AttendanceDaily{PersonID: 20, EventDate: utils.DateOnlyFromTime(d), Status: status, Seq: 1}
			if err := db.Create(&daily).Error; err != nil {
				t.Fatalf("create daily: %v", err)
			}
			if err := db.Create(&model.AttendanceEventDetail{DailyID: daily.ID, EventType: evType, SubType: subType, Hours: hours}).Error; err != nil {
				t.Fatalf("create detail: %v", err)
			}
		}
		seedLILRow("2026-06-01", "confirmed", "出勤", "补班出勤", 8)
		seedLILRow("2026-06-02", "confirmed", "休假", "调休", 4)
		seedLILRow("2026-06-03", "pending", "出勤", "补班出勤", 8) // 未确认不入列

		list, total, err := GetLILEventList(context.Background(), AttendanceDailyListQuery{PageNum: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("lil list: %v", err)
		}
		if total != 2 {
			t.Fatalf("total = %d, want 2（pending 不入列）", total)
		}
		if len(list) != 2 || list[0].SubType != "调休" || list[1].SubType != "补班出勤" {
			t.Errorf("order/fields wrong: %+v", list)
		}
		if list[0].PersonName != "测试员工" || list[0].EventDate.String() != "2026-06-02" {
			t.Errorf("person/date wrong: %+v", list[0])
		}

		list2, total2, _ := GetLILEventList(context.Background(), AttendanceDailyListQuery{PageNum: 1, PageSize: 1})
		if total2 != 2 || len(list2) != 1 {
			t.Errorf("pagination: total=%d len=%d, want 2/1", total2, len(list2))
		}
		_, total3, _ := GetLILEventList(context.Background(), AttendanceDailyListQuery{PageNum: 1, PageSize: 10, PersonID: 999})
		if total3 != 0 {
			t.Errorf("person filter: total=%d, want 0", total3)
		}
	})
}

// TestAppendCreatesNewVersionAndDemotes 追加式写入：新组 seq 递增、旧组降级 pending、投影取最新
func TestAppendCreatesNewVersionAndDemotes(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 1, "2026-01-01", 8000, 2000, 300, 500, 26)

		v1 := appendDaily(t, db, 1, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8))
		if v1.Seq != 1 {
			t.Fatalf("first append seq = %d, want 1", v1.Seq)
		}
		v2 := appendDaily(t, db, 1, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8))
		if v2.Seq != 2 {
			t.Fatalf("second append seq = %d, want 2", v2.Seq)
		}

		if s := dailyOf(t, db, v1.ID).Status; s != "pending" {
			t.Errorf("old version should be demoted to pending, got %s", s)
		}
		if s := dailyOf(t, db, v2.ID).Status; s != "confirmed" {
			t.Errorf("newest should stay confirmed, got %s", s)
		}
		if n := confirmedCountOfDay(t, db, 1, "2026-06-10"); n != 1 {
			t.Errorf("at most one confirmed per day, got %d", n)
		}
		assertDailyProjectionStatus(t, db, 1, "2026-06-10", "confirmed")
	})
}

// TestAppendPendingKeepsDayPending 最新组为 pending → 当日投影 pending
func TestAppendPendingKeepsDayPending(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 2, "2026-01-01", 8000, 2000, 300, 500, 26)
		appendDaily(t, db, 2, "2026-06-10", "pending", detailRow("出勤", "普通出勤", 8))
		assertDailyProjectionStatus(t, db, 2, "2026-06-10", "pending")
	})
}

// TestEditPromotesStaleGroup 编辑陈旧组：就地提升为最新（seq=MAX+1），其它组全部降级，投影/余额切换
func TestEditPromotesStaleGroup(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		seedEmployee(db, 3, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedGrant(t, db, 3, 40)

		v1 := appendDaily(t, db, 3, "2026-06-10", "confirmed", detailRow("休假", "年假", 8)) // 余额 32
		assertALBalance(t, db, 3, 32)
		v2 := appendDaily(t, db, 3, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8)) // v1 降级，消费撤销 → 40
		assertALBalance(t, db, 3, 40)
		_ = v2

		// 编辑陈旧组 v1（含 4h 年假）→ v1 成为最新（seq 3），v2 降级 pending，余额 36
		if err := db.Transaction(func(tx *gorm.DB) error {
			return ConfirmDaily(context.Background(), tx, v1.ID, []model.AttendanceEventDetail{detailRow("休假", "年假", 4)}, "confirmed", "", "")
		}); err != nil {
			t.Fatalf("edit stale group: %v", err)
		}
		if d := dailyOf(t, db, v1.ID); d.Seq != 3 || d.Status != "confirmed" {
			t.Errorf("edited group should be seq=3 confirmed, got seq=%d status=%s", d.Seq, d.Status)
		}
		if d := dailyOf(t, db, v2.ID); d.Status != "pending" {
			t.Errorf("previous newest should be demoted to pending, got %s", d.Status)
		}
		assertALBalance(t, db, 3, 36)
		assertDailyProjectionStatus(t, db, 3, "2026-06-10", "confirmed")
	})
}

// TestEditNewestKeepsSeq 编辑最新组：seq 保持不变，仍是唯一 confirmed
func TestEditNewestKeepsSeq(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 4, "2026-01-01", 8000, 2000, 300, 500, 26)
		v1 := appendDaily(t, db, 4, "2026-06-10", "pending", detailRow("出勤", "普通出勤", 8))
		if err := db.Transaction(func(tx *gorm.DB) error {
			return ConfirmDaily(context.Background(), tx, v1.ID, []model.AttendanceEventDetail{detailRow("出勤", "普通出勤", 8)}, "confirmed", "08:30", "正常")
		}); err != nil {
			t.Fatalf("confirm newest: %v", err)
		}
		d := dailyOf(t, db, v1.ID)
		if d.Seq != 1 {
			t.Errorf("editing the newest group should keep seq, got %d", d.Seq)
		}
		if d.Status != "confirmed" || d.PunchTime != "08:30" || d.Remark != "正常" {
			t.Errorf("confirm should apply status/punch/remark in place, got %+v", d)
		}
		assertDailyProjectionStatus(t, db, 4, "2026-06-10", "confirmed")
	})
}

// TestConfirmPromotesStalePending 确认陈旧 pending 组（转正）：原 confirmed 被降级，同日至多一条 confirmed
func TestConfirmPromotesStalePending(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 5, "2026-01-01", 8000, 2000, 300, 500, 26)
		v1 := appendDaily(t, db, 5, "2026-06-10", "pending", detailRow("出勤", "普通出勤", 8))
		v2 := appendDaily(t, db, 5, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8))
		_ = v2

		if err := db.Transaction(func(tx *gorm.DB) error {
			return ConfirmDaily(context.Background(), tx, v1.ID, []model.AttendanceEventDetail{detailRow("出勤", "普通出勤", 8)}, "confirmed", "", "")
		}); err != nil {
			t.Fatalf("confirm stale pending: %v", err)
		}
		if d := dailyOf(t, db, v1.ID); d.Seq != 3 || d.Status != "confirmed" {
			t.Errorf("promoted group should be seq=3 confirmed, got seq=%d status=%s", d.Seq, d.Status)
		}
		if d := dailyOf(t, db, v2.ID); d.Status != "pending" {
			t.Errorf("old confirmed should be demoted, got %s", d.Status)
		}
		if n := confirmedCountOfDay(t, db, 5, "2026-06-10"); n != 1 {
			t.Errorf("at most one confirmed per day, got %d", n)
		}
		if maxSeqOfDay(t, db, 5, "2026-06-10") != 3 {
			t.Errorf("confirmed must be the max seq")
		}
	})
}

// TestDeleteOperativeFallsBackToPending 删除最新（有效）组：当日剩余全为 pending → 投影回落 pending
func TestDeleteOperativeFallsBackToPending(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 6, "2026-01-01", 8000, 2000, 300, 500, 26)
		v1 := appendDaily(t, db, 6, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8))
		v2 := appendDaily(t, db, 6, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8))
		_ = v1
		if err := DeleteAttendanceDaily(context.Background(), v2.ID); err != nil {
			t.Fatalf("delete operative: %v", err)
		}
		assertDailyProjectionStatus(t, db, 6, "2026-06-10", "pending")
	})
}

// TestRestoreSeqReassignAndDemote 恢复 = 复活为新版：seq 重分配为当日最大+1，其它组降级 pending
func TestRestoreSeqReassignAndDemote(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 7, "2026-01-01", 8000, 2000, 300, 500, 26)
		v1 := appendDaily(t, db, 7, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8))
		if err := DeleteAttendanceDaily(context.Background(), v1.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
		// 删除后追加：剩余组 seq 从 1 重新计数
		v2 := appendDaily(t, db, 7, "2026-06-10", "confirmed", detailRow("出勤", "普通出勤", 8))
		if v2.Seq != 1 {
			t.Fatalf("after delete, append seq = %d, want 1", v2.Seq)
		}

		if err := RestoreAttendanceDaily(context.Background(), v1.ID); err != nil {
			t.Fatalf("restore: %v", err)
		}
		d1 := dailyOf(t, db, v1.ID)
		if d1.Seq != 2 {
			t.Errorf("restored group seq = %d, want 2 (reassigned to newest)", d1.Seq)
		}
		if d1.Status != "confirmed" {
			t.Errorf("restored confirmed group should stay confirmed, got %s", d1.Status)
		}
		if d := dailyOf(t, db, v2.ID); d.Status != "pending" {
			t.Errorf("other group should be demoted after restore, got %s", d.Status)
		}
		if n := confirmedCountOfDay(t, db, 7, "2026-06-10"); n != 1 {
			t.Errorf("at most one confirmed per day, got %d", n)
		}
	})
}

// TestConfirmReplacesDetailsWholesale 编辑/确认时详情整体替换（旧明细软删 + 新明细新建）
func TestConfirmReplacesDetailsWholesale(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 8, "2026-01-01", 8000, 2000, 300, 500, 26)
		v1 := appendDaily(t, db, 8, "2026-06-10", "pending", detailRow("出勤", "普通出勤", 8), detailRow("加班", "工作日加班", 2))
		if err := db.Transaction(func(tx *gorm.DB) error {
			return ConfirmDaily(context.Background(), tx, v1.ID, []model.AttendanceEventDetail{detailRow("休假", "事假", 8)}, "confirmed", "", "")
		}); err != nil {
			t.Fatalf("confirm: %v", err)
		}
		details := dailyDetails(t, db, v1.ID)
		if len(details) != 1 || details[0].SubType != "事假" {
			t.Fatalf("details should be wholly replaced, got %+v", details)
		}
		var deleted int64
		db.Unscoped().Model(&model.AttendanceEventDetail{}).Where("daily_id = ? AND deleted_at IS NOT NULL", v1.ID).Count(&deleted)
		if deleted != 2 {
			t.Errorf("old details should be soft-deleted, got %d", deleted)
		}
	})
}
