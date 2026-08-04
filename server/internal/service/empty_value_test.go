package service

import (
	"context"
	"testing"

	"probig/server/internal/model"

	"gorm.io/gorm"
)

// TestAttendanceCalcClearsWhenNoProjection 无投影 → 置空（旧核算记录被物理删除）
func TestAttendanceCalcClearsWhenNoProjection(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 30, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 30, "2026-06", 5, 8)

		r, err := CalculateMonthlyAttendance(context.Background(), 30, "2026-06")
		if err != nil || r == nil {
			t.Fatalf("first calc: result=%v err=%v", r, err)
		}
		// 删除当月全部投影（模拟考勤记录被删除 → 投影自动清空）
		db.Where("person_id = ?", 30).Delete(&model.AttendanceDailyProjection{})

		r2, err2 := CalculateMonthlyAttendance(context.Background(), 30, "2026-06")
		if err2 != nil || r2 != nil {
			t.Fatalf("no projection should be empty: result=%v err=%v", r2, err2)
		}
		var count int64
		db.Model(&model.AttendanceCalculationMonthly{}).Where("person_id = ? AND belong_month = ?", 30, "2026-06").Count(&count)
		if count != 0 {
			t.Errorf("calc should be cleared, got %d records", count)
		}
	})
}

// TestAttendanceCalcClearsWhenNoSnapshot 无在职快照 → 置空（无值）
func TestAttendanceCalcClearsWhenNoSnapshot(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.Create(&model.Person{ID: 32, Name: "无快照"})
		r, err := CalculateMonthlyAttendance(context.Background(), 32, "2026-06")
		if err != nil || r != nil {
			t.Fatalf("no snapshot should be empty: result=%v err=%v", r, err)
		}
	})
}

// TestAttendanceCalcFailsOnZeroSalaryDays 计薪天数未配置 → 失败（需人工干预），表不动
func TestAttendanceCalcFailsOnZeroSalaryDays(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 31, "2026-01-01", 8000, 2000, 300, 500, 0)
		seedAttendanceDays(db, 31, "2026-06", 5, 8)

		r, err := CalculateMonthlyAttendance(context.Background(), 31, "2026-06")
		if err == nil || r != nil {
			t.Fatalf("zero salary days should fail: result=%v err=%v", r, err)
		}
		var count int64
		db.Model(&model.AttendanceCalculationMonthly{}).Where("person_id = ? AND belong_month = ?", 31, "2026-06").Count(&count)
		if count != 0 {
			t.Errorf("failed calc should not write, got %d", count)
		}
	})
}

// TestSalaryClearsWhenAttendanceEmpty 空值传递：考勤空 → 工资表物理删除置空
func TestSalaryClearsWhenAttendanceEmpty(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 33, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 33, "2026-06", 5, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 33, "2026-06"); err != nil {
			t.Fatalf("calc attendance: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 33, "2026-06", 1, "admin"); err != nil {
			t.Fatalf("calc salary: %v", err)
		}
		var s model.SalarySummary
		if err := db.Where("person_id = ? AND belong_month = ?", 33, "2026-06").First(&s).Error; err != nil {
			t.Fatalf("summary should exist: %v", err)
		}

		// 考勤投影被清空 → 重新考勤核算置空 → 重新工资核算：空值传递
		db.Where("person_id = ?", 33).Delete(&model.AttendanceDailyProjection{})
		if _, err := CalculateMonthlyAttendance(context.Background(), 33, "2026-06"); err != nil {
			t.Fatalf("recalc attendance: %v", err)
		}
		r, err := CalculateSalary(context.Background(), 33, "2026-06", 1, "admin")
		if err != nil || r != nil {
			t.Fatalf("salary should be empty: result=%v err=%v", r, err)
		}
		var count int64
		db.Model(&model.SalarySummary{}).Where("person_id = ? AND belong_month = ?", 33, "2026-06").Count(&count)
		if count != 0 {
			t.Errorf("salary summary should be cleared, got %d", count)
		}
	})
}

// TestSalaryCalculatesWhenAttendanceZero 考勤 0（有投影无出勤工时）→ 工资正常核算（有值，补贴照发）
func TestSalaryCalculatesWhenAttendanceZero(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 34, "2026-01-01", 8000, 2000, 300, 500, 26)
		// 仅一天事假：投影存在但出勤工时为 0
		seedAttendanceEvent(db, 34, "2026-06-05", "休假", "事假", 8)

		r, err := CalculateMonthlyAttendance(context.Background(), 34, "2026-06")
		if err != nil || r == nil {
			t.Fatalf("attendance with projection should have value: result=%v err=%v", r, err)
		}
		if r.TotalWorkHours != 0 {
			t.Errorf("expected zero work hours, got %v", r.TotalWorkHours)
		}
		s, err := CalculateSalary(context.Background(), 34, "2026-06", 1, "admin")
		if err != nil || s == nil {
			t.Fatalf("salary should have value: result=%v err=%v", s, err)
		}
		if s.AttendanceSalary != 0 {
			t.Errorf("zero attendance salary expected, got %v", s.AttendanceSalary)
		}
		if s.FinalSalary <= 0 {
			t.Errorf("subsidies should still pay on full month, got FinalSalary=%v", s.FinalSalary)
		}
	})
}

// TestBatchCalculateThreeStates 批量核算三态计数
func TestBatchCalculateThreeStates(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 40, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 40, "2026-06", 5, 8) // 有值
		seedEmployee(db, 41, "2026-01-01", 8000, 2000, 300, 500, 0)
		seedAttendanceDays(db, 41, "2026-06", 5, 8) // 计薪天数 0 → 失败
		db.Create(&model.Person{ID: 42, Name: "无快照"}) // 无快照 → 空

		hasValue, empty, fail, err := CalculateMonthlyBatch(context.Background(), "2026-06", []uint{40, 41, 42})
		if err != nil {
			t.Fatalf("batch: %v", err)
		}
		if hasValue != 1 || empty != 1 || fail != 1 {
			t.Errorf("three-state: hasValue=%d empty=%d fail=%d, want 1/1/1", hasValue, empty, fail)
		}
	})
}
