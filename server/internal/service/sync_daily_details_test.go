package service

import (
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func registerCreateCounter(db *gorm.DB, table string, count *int) func() {
	db.Callback().Create().After("gorm:after_create").Register("test:count-create-"+table, func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			*count++
		}
	})
	return func() { db.Callback().Create().Remove("test:count-create-" + table) }
}

func mustDaily(t *testing.T, db *gorm.DB, personID uint, date string) model.AttendanceDaily {
	t.Helper()
	d, _ := utils.ParseDate(date)
	dOnly := utils.DateOnlyFromTime(d)
	daily := model.AttendanceDaily{PersonID: personID, EventDate: dOnly, Status: "confirmed"}
	if err := db.Create(&daily).Error; err != nil {
		t.Fatalf("create daily: %v", err)
	}
	return daily
}

// TestSyncDailyDetails_RepeatUpsertKeepsRows 重复写入相同明细：零操作零审计，行 id 保持不变
func TestSyncDailyDetails_RepeatUpsertKeepsRows(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		daily := mustDaily(t, db, 1, "2026-06-01")
		rows := []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤", Hours: 8},
			{EventType: "加班", SubType: "工作日加班", Hours: 2},
		}
		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, rows) }); err != nil {
			t.Fatalf("first upsert: %v", err)
		}
		var first []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&first)
		if len(first) != 2 {
			t.Fatalf("first: want 2 rows, got %d", len(first))
		}

		var creates, deletes int
		rmC := registerCreateCounter(db, "attendance_event_details", &creates)
		defer rmC()
		rmD := registerDeleteCounter(db, "attendance_event_details", &deletes)
		defer rmD()

		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, rows) }); err != nil {
			t.Fatalf("repeat upsert: %v", err)
		}
		if creates != 0 || deletes != 0 {
			t.Errorf("repeat upsert must be zero-op: creates=%d deletes=%d", creates, deletes)
		}
		var after []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&after)
		if len(after) != 2 || after[0].ID != first[0].ID || after[1].ID != first[1].ID {
			t.Errorf("row ids changed: before=%v after=%v", first, after)
		}
	})
}

// TestSyncDailyDetails_UpdateOnlyChangedRow 修改一行：仅该行就地更新（id 不变），其余行零操作
func TestSyncDailyDetails_UpdateOnlyChangedRow(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		daily := mustDaily(t, db, 1, "2026-06-01")
		rows := []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤", Hours: 8},
			{EventType: "加班", SubType: "工作日加班", Hours: 2},
		}
		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, rows) }); err != nil {
			t.Fatalf("first upsert: %v", err)
		}
		var before []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&before)

		changed := []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤", Hours: 7},
			{EventType: "加班", SubType: "工作日加班", Hours: 2},
		}
		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, changed) }); err != nil {
			t.Fatalf("update upsert: %v", err)
		}
		var after []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&after)
		if len(after) != 2 {
			t.Fatalf("want 2 rows, got %d", len(after))
		}
		if after[0].ID != before[0].ID || after[0].Hours != 7 {
			t.Errorf("row0 not updated in place: before=%v after=%v", before[0], after[0])
		}
		if after[1].ID != before[1].ID || after[1].Hours != 2 {
			t.Errorf("row1 should be untouched: before=%v after=%v", before[1], after[1])
		}
	})
}

// TestSyncDailyDetails_RemoveSoftDeleteAndAdd 移除的行软删除（可恢复），新类型行新增
func TestSyncDailyDetails_RemoveSoftDeleteAndAdd(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		daily := mustDaily(t, db, 1, "2026-06-01")
		rows := []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤", Hours: 8},
			{EventType: "加班", SubType: "工作日加班", Hours: 2},
		}
		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, rows) }); err != nil {
			t.Fatalf("first upsert: %v", err)
		}
		var before []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&before)

		next := []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤", Hours: 8},
			{EventType: "休假", SubType: "年假", Hours: 4},
		}
		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, next) }); err != nil {
			t.Fatalf("next upsert: %v", err)
		}
		var after []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&after)
		if len(after) != 2 {
			t.Fatalf("want 2 active rows, got %d", len(after))
		}
		if after[0].ID != before[0].ID || after[0].SubType != "普通出勤" {
			t.Errorf("kept row not in place: before=%v after=%v", before[0], after[0])
		}
		if after[1].SubType != "年假" || after[1].ID == before[1].ID {
			t.Errorf("new row should be inserted with fresh id: after=%v", after[1])
		}
		var removed model.AttendanceEventDetail
		if err := db.Unscoped().First(&removed, before[1].ID).Error; err != nil {
			t.Fatalf("removed row should be soft-deletable, not gone: %v", err)
		}
		if removed.DeletedAt.Time.IsZero() {
			t.Errorf("removed row should have deleted_at set")
		}
	})
}

// TestSyncDailyDetails_EditWithIDsUsesPrimaryKey 编辑回显（带 id）场景：按主键精确对齐
func TestSyncDailyDetails_EditWithIDsUsesPrimaryKey(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		daily := mustDaily(t, db, 1, "2026-06-01")
		rows := []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤", Hours: 8},
			{EventType: "加班", SubType: "工作日加班", Hours: 2},
		}
		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, rows) }); err != nil {
			t.Fatalf("first upsert: %v", err)
		}
		var before []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&before)

		// 带 id 回显：仅改第一条的备注，第二条原样带回
		edited := []model.AttendanceEventDetail{
			{ID: before[0].ID, EventType: "出勤", SubType: "普通出勤", Hours: 8, Remark: "调休半天"},
			{ID: before[1].ID, EventType: "加班", SubType: "工作日加班", Hours: 2},
		}
		if err := db.Transaction(func(tx *gorm.DB) error { return SyncDailyDetails(tx, daily.ID, edited) }); err != nil {
			t.Fatalf("edit upsert: %v", err)
		}
		var after []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Order("id").Find(&after)
		if len(after) != 2 {
			t.Fatalf("want 2 rows, got %d", len(after))
		}
		if after[0].ID != before[0].ID || after[0].Remark != "调休半天" {
			t.Errorf("row0 remark not updated: after=%v", after[0])
		}
		if after[1].ID != before[1].ID || after[1].Remark != "" {
			t.Errorf("row1 should stay untouched: after=%v", after[1])
		}
	})
}

// TestUpsertAttendanceDaily_GranularUpsert 统一入口颗粒化：重复 upsert 相同内容不产生新行
func TestUpsertAttendanceDaily_GranularUpsert(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		d, _ := utils.ParseDate("2026-06-02")
		dOnly := utils.DateOnlyFromTime(d)
		status, punch, remark := "confirmed", "08:30,18:00", "正常"
		details := []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤", Hours: 8},
		}
		upsert := func() error {
			return db.Transaction(func(tx *gorm.DB) error {
				return UpsertAttendanceDaily(tx, AttendanceDailyUpsert{
					PersonID: 1, Date: dOnly,
					Status: &status, PunchTime: &punch, Remark: &remark, Details: details,
				})
			})
		}
		if err := upsert(); err != nil {
			t.Fatalf("first upsert: %v", err)
		}
		var first model.AttendanceDaily
		if err := db.Where("person_id = 1 AND event_date = ?", dOnly).First(&first).Error; err != nil {
			t.Fatalf("load daily: %v", err)
		}
		var firstDetails []model.AttendanceEventDetail
		db.Where("daily_id = ?", first.ID).Find(&firstDetails)

		var creates, deletes int
		rmC := registerCreateCounter(db, "attendance_event_details", &creates)
		defer rmC()
		rmD := registerDeleteCounter(db, "attendance_event_details", &deletes)
		defer rmD()

		if err := upsert(); err != nil {
			t.Fatalf("repeat upsert: %v", err)
		}
		if creates != 0 || deletes != 0 {
			t.Errorf("repeat upsert must not recreate details: creates=%d deletes=%d", creates, deletes)
		}
		var second []model.AttendanceEventDetail
		db.Where("daily_id = ?", first.ID).Find(&second)
		if len(second) != 1 || second[0].ID != firstDetails[0].ID {
			t.Errorf("detail row changed: before=%v after=%v", firstDetails, second)
		}
	})
}
