package service

import (
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// 同日多事件：seq 大的覆盖同名字段，无效时段（end<start）跳过
func TestSnapshotBoundarySameDayEvents(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		d, _ := utils.ParseDate("2026-01-01")
		dOnly := utils.DateOnlyFromTime(d)
		if err := db.Create(&model.PositionEvent{
			PersonID: 1, Seq: 1, EventType: "入职", EffectiveDate: dOnly,
			EntryDate: &dOnly, BaseSalary: ptrFloat(8000), MealAllowance: ptrFloat(300),
		}).Error; err != nil {
			t.Fatalf("seed entry: %v", err)
		}
		if err := db.Create(&model.PositionEvent{
			PersonID: 1, Seq: 2, EventType: "调薪调岗", EffectiveDate: dOnly,
			BaseSalary: ptrFloat(9000),
		}).Error; err != nil {
			t.Fatalf("seed adjust: %v", err)
		}
		RebuildPositionSnapshots(db, 1)

		var snaps []model.PositionSnapshot
		db.Where("person_id = ?", 1).Find(&snaps)
		if len(snaps) != 1 {
			t.Fatalf("expected 1 snapshot (invalid same-day segment skipped), got %d", len(snaps))
		}
		if snaps[0].BaseSalary != 9000 {
			t.Errorf("base salary: got %.2f, want 9000 (later seq wins)", snaps[0].BaseSalary)
		}
		if snaps[0].MealAllowance != 300 {
			t.Errorf("meal should inherit from seq1: got %.2f, want 300", snaps[0].MealAllowance)
		}
		if snaps[0].EffectiveEndDate.String() != "9999-12-31" {
			t.Errorf("end date: got %s, want 9999-12-31", snaps[0].EffectiveEndDate.String())
		}
	})
}

// 离职后恢复在职：三段时段，离职段 is_active=false，9999-12-31 收尾
func TestSnapshotBoundaryLeaveAndReentry(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		entry, _ := utils.ParseDate("2026-01-01")
		entryD := utils.DateOnlyFromTime(entry)
		leave, _ := utils.ParseDate("2026-06-30")
		leaveD := utils.DateOnlyFromTime(leave)
		reentry, _ := utils.ParseDate("2026-09-01")
		reentryD := utils.DateOnlyFromTime(reentry)

		db.Create(&model.PositionEvent{PersonID: 2, Seq: 1, EventType: "入职", EffectiveDate: entryD, EntryDate: &entryD, BaseSalary: ptrFloat(8000)})
		db.Create(&model.PositionEvent{PersonID: 2, Seq: 2, EventType: "离职", EffectiveDate: leaveD, LeaveDate: &leaveD})
		db.Create(&model.PositionEvent{PersonID: 2, Seq: 3, EventType: "入职", EffectiveDate: reentryD, EntryDate: &reentryD, BaseSalary: ptrFloat(8500)})
		RebuildPositionSnapshots(db, 2)

		var snaps []model.PositionSnapshot
		db.Where("person_id = ?", 2).Order("effective_start_date ASC").Find(&snaps)
		if len(snaps) != 3 {
			t.Fatalf("expected 3 snapshots, got %d", len(snaps))
		}
		if !snaps[0].IsActive || snaps[1].IsActive || !snaps[2].IsActive {
			t.Errorf("active flags wrong: %v %v %v", snaps[0].IsActive, snaps[1].IsActive, snaps[2].IsActive)
		}
		if snaps[0].EffectiveEndDate.String() != "2026-06-29" {
			t.Errorf("seg1 end: got %s, want 2026-06-29", snaps[0].EffectiveEndDate.String())
		}
		if snaps[1].EffectiveStartDate.String() != "2026-06-30" || snaps[1].EffectiveEndDate.String() != "2026-08-31" {
			t.Errorf("seg2 range wrong: %s ~ %s", snaps[1].EffectiveStartDate.String(), snaps[1].EffectiveEndDate.String())
		}
		if snaps[2].EffectiveStartDate.String() != "2026-09-01" || snaps[2].EffectiveEndDate.String() != "9999-12-31" {
			t.Errorf("seg3 range wrong: %s ~ %s", snaps[2].EffectiveStartDate.String(), snaps[2].EffectiveEndDate.String())
		}
		if snaps[2].BaseSalary != 8500 {
			t.Errorf("reentry base salary: got %.2f, want 8500", snaps[2].BaseSalary)
		}
	})
}

// 无事件：重建后无快照
func TestSnapshotBoundaryEmptyEvents(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		RebuildPositionSnapshots(db, 3)
		var count int64
		db.Model(&model.PositionSnapshot{}).Where("person_id = ?", 3).Count(&count)
		if count != 0 {
			t.Errorf("expected no snapshots for empty events, got %d", count)
		}
	})
}

// 单字段变更：生成两段，未变更字段继承
func TestSnapshotBoundarySingleFieldChange(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		entry, _ := utils.ParseDate("2026-01-01")
		entryD := utils.DateOnlyFromTime(entry)
		change, _ := utils.ParseDate("2026-04-01")
		changeD := utils.DateOnlyFromTime(change)

		db.Create(&model.PositionEvent{PersonID: 4, Seq: 1, EventType: "入职", EffectiveDate: entryD, EntryDate: &entryD, BaseSalary: ptrFloat(8000), MealAllowance: ptrFloat(300)})
		db.Create(&model.PositionEvent{PersonID: 4, Seq: 2, EventType: "调薪调岗", EffectiveDate: changeD, MealAllowance: ptrFloat(400)})
		RebuildPositionSnapshots(db, 4)

		var snaps []model.PositionSnapshot
		db.Where("person_id = ?", 4).Order("effective_start_date ASC").Find(&snaps)
		if len(snaps) != 2 {
			t.Fatalf("expected 2 snapshots, got %d", len(snaps))
		}
		if snaps[0].MealAllowance != 300 || snaps[1].MealAllowance != 400 {
			t.Errorf("meal split wrong: %v %v", snaps[0].MealAllowance, snaps[1].MealAllowance)
		}
		if snaps[0].BaseSalary != 8000 || snaps[1].BaseSalary != 8000 {
			t.Errorf("base salary should inherit: %v %v", snaps[0].BaseSalary, snaps[1].BaseSalary)
		}
		if snaps[0].EffectiveEndDate.String() != "2026-03-31" {
			t.Errorf("seg1 end: got %s, want 2026-03-31", snaps[0].EffectiveEndDate.String())
		}
	})
}
