package service

import (
	"context"
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func TestUpdatePositionEventKeepsUnchangedFields(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		entry, _ := utils.ParseDate("2026-01-01")
		entryD := utils.DateOnlyFromTime(entry)
		event := model.PositionEvent{
			PersonID: 1, Seq: 1, EventType: "入职", EffectiveDate: entryD,
			EntryDate: &entryD, AttendanceGroup: ptr("标准"),
			HasAnnualLeave:     ptrBool(true),
			HasAttendanceBonus: ptrBool(true),
			BaseSalary:         ptrFloat(8000),
			MealAllowance:      ptrFloat(300),
		}
		if err := CreatePositionEvent(context.Background(), &event); err != nil {
			t.Fatalf("create: %v", err)
		}

		// 仅更新基本工资，其余字段未提交（模拟前端只勾选基本工资）
		if err := UpdatePositionEvent(context.Background(), event.ID, &model.PositionEvent{
			EventType:     "入职",
			EffectiveDate: entryD,
			BaseSalary:    ptrFloat(9000),
		}); err != nil {
			t.Fatalf("update: %v", err)
		}

		var updated model.PositionEvent
		if err := db.First(&updated, event.ID).Error; err != nil {
			t.Fatalf("load updated: %v", err)
		}
		if updated.EntryDate == nil || updated.EntryDate.String() != "2026-01-01" {
			t.Errorf("entry_date should be preserved, got %v", updated.EntryDate)
		}
		if updated.MealAllowance == nil || *updated.MealAllowance != 300 {
			t.Errorf("meal_allowance should be preserved, got %v", updated.MealAllowance)
		}
		if updated.HasAnnualLeave == nil || !*updated.HasAnnualLeave {
			t.Errorf("has_annual_leave should be preserved, got %v", updated.HasAnnualLeave)
		}
		if updated.BaseSalary == nil || *updated.BaseSalary != 9000 {
			t.Errorf("base_salary should be updated to 9000, got %v", updated.BaseSalary)
		}

		var snap model.PositionSnapshot
		if err := db.Where("person_id = ?", 1).First(&snap).Error; err != nil {
			t.Fatalf("load snapshot: %v", err)
		}
		if snap.BaseSalary != 9000 {
			t.Errorf("snapshot base_salary: got %.2f, want 9000", snap.BaseSalary)
		}
		if snap.MealAllowance != 300 {
			t.Errorf("snapshot meal_allowance: got %.2f, want 300", snap.MealAllowance)
		}
	})
}

func TestUpdatePositionEventKeepsLeaveDateWhenNotSubmitted(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		entry, _ := utils.ParseDate("2026-01-01")
		entryD := utils.DateOnlyFromTime(entry)
		leave, _ := utils.ParseDate("2026-06-30")
		leaveD := utils.DateOnlyFromTime(leave)

		event := model.PositionEvent{
			PersonID: 2, Seq: 1, EventType: "入职", EffectiveDate: entryD,
			EntryDate: &entryD, BaseSalary: ptrFloat(8000),
		}
		if err := CreatePositionEvent(context.Background(), &event); err != nil {
			t.Fatalf("create entry: %v", err)
		}
		leaveEvent := model.PositionEvent{
			PersonID: 2, Seq: 2, EventType: "离职", EffectiveDate: leaveD,
			LeaveDate: &leaveD,
		}
		if err := CreatePositionEvent(context.Background(), &leaveEvent); err != nil {
			t.Fatalf("create leave: %v", err)
		}

		// 编辑离职事件的备注（不带 leave_date 之外的字段，模拟前端只改备注）
		if err := UpdatePositionEvent(context.Background(), leaveEvent.ID, &model.PositionEvent{
			EventType:     "离职",
			Remark:        "主动离职",
			EffectiveDate: leaveD,
		}); err != nil {
			t.Fatalf("update leave event: %v", err)
		}

		var updated model.PositionEvent
		if err := db.First(&updated, leaveEvent.ID).Error; err != nil {
			t.Fatalf("load updated: %v", err)
		}
		if updated.LeaveDate == nil || updated.LeaveDate.String() != "2026-06-30" {
			t.Errorf("leave_date should be preserved, got %v", updated.LeaveDate)
		}
		if updated.Remark != "主动离职" {
			t.Errorf("remark should be updated, got %q", updated.Remark)
		}
	})
}
