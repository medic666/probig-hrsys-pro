package service

import (
	"context"
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func registerDeleteCounter(db *gorm.DB, table string, count *int) func() {
	db.Callback().Delete().After("gorm:after_delete").Register("test:count-"+table, func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			*count++
		}
	})
	return func() { db.Callback().Delete().Remove("test:count-" + table) }
}

func migrateBalanceSnapshots(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&model.AnnualLeaveBalanceSnapshot{}, &model.LeaveInLieuBalanceSnapshot{}); err != nil {
		t.Fatalf("migrate balance snapshots: %v", err)
	}
}

func seedGrant(t *testing.T, db *gorm.DB, personID uint, hours float64) {
	t.Helper()
	d, _ := utils.ParseDate("2026-01-01")
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
		t.Fatalf("rebuild grant: %v", err)
	}
}

func seedConfirmedDetail(t *testing.T, db *gorm.DB, personID uint, date, subType string, hours float64) model.AttendanceEventDetail {
	t.Helper()
	d, _ := utils.ParseDate(date)
	dOnly := utils.DateOnlyFromTime(d)
	var daily model.AttendanceDaily
	if err := db.Where("person_id = ? AND event_date = ?", personID, dOnly).First(&daily).Error; err != nil {
		daily = model.AttendanceDaily{PersonID: personID, EventDate: dOnly, Status: "confirmed"}
		if err := db.Create(&daily).Error; err != nil {
			t.Fatalf("create daily: %v", err)
		}
	}
	detail := model.AttendanceEventDetail{DailyID: daily.ID, EventType: "休假", SubType: subType, Hours: hours}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}
	if err := RebuildProjectionsAfterAttendanceChange(db, personID, dOnly, []model.AttendanceEventDetail{detail}); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	return detail
}

func assertALBalance(t *testing.T, db *gorm.DB, personID uint, want float64) {
	t.Helper()
	var snap model.AnnualLeaveBalanceSnapshot
	if err := db.Where("person_id = ?", personID).Order("effective_start_date DESC").First(&snap).Error; err != nil {
		t.Fatalf("load al snapshot: %v", err)
	}
	if snap.BalanceHours != want {
		t.Errorf("al balance: got %.1f want %.1f", snap.BalanceHours, want)
	}
}

func assertLILBalance(t *testing.T, db *gorm.DB, personID uint, want float64) {
	t.Helper()
	var snap model.LeaveInLieuBalanceSnapshot
	if err := db.Where("person_id = ?", personID).Order("effective_start_date DESC").First(&snap).Error; err != nil {
		t.Fatalf("load lil snapshot: %v", err)
	}
	if snap.BalanceHours != want {
		t.Errorf("lil balance: got %.1f want %.1f", snap.BalanceHours, want)
	}
}

func TestProjectionTriggerPreciseByType(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		seedConfirmedDetail(t, db, 40, "2026-06-01", "补班出勤", 8)
		assertLILBalance(t, db, 40, 8)

		alDelete, lilDelete := 0, 0
		unreg1 := registerDeleteCounter(db, "annual_leave_balance_snapshots", &alDelete)
		unreg2 := registerDeleteCounter(db, "leave_in_lieu_balance_snapshots", &lilDelete)
		defer unreg1()
		defer unreg2()

		d, _ := utils.ParseDate("2026-06-10")
		dOnly := utils.DateOnlyFromTime(d)

		// 普通出勤 → 两个余额均不重建
		if err := RebuildProjectionsAfterAttendanceChange(db, 40, dOnly, []model.AttendanceEventDetail{{EventType: "出勤", SubType: "普通出勤"}}); err != nil {
			t.Fatalf("rebuild: %v", err)
		}
		if alDelete != 0 || lilDelete != 0 {
			t.Fatalf("plain attendance should not rebuild balances: al=%d lil=%d", alDelete, lilDelete)
		}

		// 年假休假 → 仅年假余额重建
		if err := RebuildProjectionsAfterAttendanceChange(db, 40, dOnly, []model.AttendanceEventDetail{{EventType: "休假", SubType: "年假"}}); err != nil {
			t.Fatalf("rebuild: %v", err)
		}
		if alDelete != 1 || lilDelete != 0 {
			t.Fatalf("annual leave should rebuild only AL balance: al=%d lil=%d", alDelete, lilDelete)
		}

		// 调休 → 仅调休余额重建
		if err := RebuildProjectionsAfterAttendanceChange(db, 40, dOnly, []model.AttendanceEventDetail{{EventType: "休假", SubType: "调休"}}); err != nil {
			t.Fatalf("rebuild: %v", err)
		}
		if alDelete != 1 || lilDelete != 1 {
			t.Fatalf("leave-in-lieu should rebuild only LIL balance: al=%d lil=%d", alDelete, lilDelete)
		}

		// 编辑场景（旧=普通出勤，新=年假）→ 并集含年假 → 年假重建
		if err := RebuildProjectionsAfterAttendanceChange(db, 40, dOnly, []model.AttendanceEventDetail{
			{EventType: "出勤", SubType: "普通出勤"},
			{EventType: "休假", SubType: "年假"},
		}); err != nil {
			t.Fatalf("rebuild: %v", err)
		}
		if alDelete != 2 || lilDelete != 1 {
			t.Fatalf("mixed change should rebuild AL balance: al=%d lil=%d", alDelete, lilDelete)
		}
	})
}

func TestAnnualLeaveBalanceTracksAttendanceEvent(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		seedGrant(t, db, 41, 40)
		assertALBalance(t, db, 41, 40)

		// 录入 8h 年假休假 → 余额 32
		detail := seedConfirmedDetail(t, db, 41, "2026-06-10", "年假", 8)
		assertALBalance(t, db, 41, 32)

		// 删除该事件 → 余额回升 40
		if err := db.Delete(&detail).Error; err != nil {
			t.Fatalf("delete detail: %v", err)
		}
		var delDaily model.AttendanceDaily
		if err := db.First(&delDaily, detail.DailyID).Error; err != nil {
			t.Fatalf("load daily: %v", err)
		}
		if err := RebuildProjectionsAfterAttendanceChange(db, 41, delDaily.EventDate, []model.AttendanceEventDetail{detail}); err != nil {
			t.Fatalf("rebuild after delete: %v", err)
		}
		assertALBalance(t, db, 41, 40)
	})
}

func TestLILBalanceTracksAttendanceEvent(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		seedConfirmedDetail(t, db, 42, "2026-06-01", "补班出勤", 8)
		assertLILBalance(t, db, 42, 8)
		seedConfirmedDetail(t, db, 42, "2026-06-05", "调休", 4)
		assertLILBalance(t, db, 42, 4)
	})
}

func TestConfirmDailyRebuildsBalances(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		seedGrant(t, db, 43, 40)

		// pending 状态含年假事件 → 不参与投影，余额不变
		d, _ := utils.ParseDate("2026-06-10")
		dOnly := utils.DateOnlyFromTime(d)
		daily := model.AttendanceDaily{PersonID: 43, EventDate: dOnly, Status: "pending"}
		if err := db.Create(&daily).Error; err != nil {
			t.Fatalf("create daily: %v", err)
		}
		if err := db.Create(&model.AttendanceEventDetail{DailyID: daily.ID, EventType: "休假", SubType: "年假", Hours: 8}).Error; err != nil {
			t.Fatalf("create detail: %v", err)
		}
		if err := RebuildProjectionsAfterAttendanceChange(db, 43, dOnly, []model.AttendanceEventDetail{{DailyID: daily.ID, EventType: "休假", SubType: "年假", Hours: 8}}); err != nil {
			t.Fatalf("rebuild pending: %v", err)
		}
		assertALBalance(t, db, 43, 40)

		// 确认后余额生效（确认=就地转正：目标组置为最新并 confirmed，明细整体替换）
		if err := db.Transaction(func(tx *gorm.DB) error {
			return ConfirmDaily(context.Background(), tx, daily.ID, []model.AttendanceEventDetail{{DailyID: daily.ID, EventType: "休假", SubType: "年假", Hours: 8}}, "confirmed", "", "")
		}); err != nil {
			t.Fatalf("confirm: %v", err)
		}
		assertALBalance(t, db, 43, 32)
	})
}
