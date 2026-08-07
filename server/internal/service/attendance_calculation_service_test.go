package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

var testMu sync.Mutex

func ptr(s string) *string       { return &s }
func ptrBool(b bool) *bool        { return &b }
func ptrFloat(f float64) *float64 { return &f }

func withTestDB(t *testing.T, fn func(db *gorm.DB)) {
	t.Helper()
	testMu.Lock()
	defer testMu.Unlock()

	db, err := gorm.Open(dao.GetSQLiteDialector(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.PositionEvent{},
		&model.PositionSnapshot{},
		&model.AttendanceDaily{},
		&model.AttendanceEventDetail{},
		&model.AttendanceDailyProjection{},
		&model.AttendanceCalculationMonthly{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	orig := dao.DB
	dao.DB = db
	defer func() { dao.DB = orig }()

	fn(db)
}

func TestCalculateMonthlyAttendance_Basic(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		entryDate, _ := utils.ParseDate("2026-01-01")
		entryD := utils.DateOnlyFromTime(entryDate)

		db.Create(&model.PositionEvent{
			PersonID: 1, Seq: 1, EventType: "入职", EffectiveDate: entryD,
			EntryDate: &entryD, AttendanceGroup: ptr("标准"),
			HasAnnualLeave: ptrBool(true), HasAttendanceBonus: ptrBool(true),
			BaseSalary: ptrFloat(8000), PerformanceSalary: ptrFloat(2000),
			SalaryDays: ptrFloat(26), MealAllowance: ptrFloat(300),
		})
		RebuildPositionSnapshots(db, 1)

		workD, _ := utils.ParseDate("2026-06-15")
		daily := model.AttendanceDaily{PersonID: 1, EventDate: utils.DateOnlyFromTime(workD), Status: "confirmed"}
		db.Create(&daily)
		db.Create(&model.AttendanceEventDetail{
			DailyID: daily.ID, EventType: "出勤", SubType: "普通出勤", Hours: 8,
		})
		RebuildDailyProjection(db, 1, utils.DateOnlyFromTime(workD))

		result, err := CalculateMonthlyAttendance(context.Background(), 1, "2026-06")
		if err != nil {
			t.Fatalf("CalculateMonthlyAttendance: %v", err)
		}
		if result.TotalWorkHours != 8 {
			t.Errorf("TotalWorkHours = %f, want 8", result.TotalWorkHours)
		}
		if result.AttendanceSalary <= 0 {
			t.Errorf("AttendanceSalary = %f, want > 0", result.AttendanceSalary)
		}
		if result.BelongMonth != "2026-06" {
			t.Errorf("BelongMonth = %s, want 2026-06", result.BelongMonth)
		}
	})
}

func TestIsAttendanceMonthlyStale_NotStale(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		now := time.Now()
		calc := model.AttendanceCalculationMonthly{
			PersonID: 1, BelongMonth: "2026-06", LastCalcAt: now,
		}
		db.Create(&calc)

		status := IsAttendanceMonthlyStale(&calc)
		if status == "data_changed" {
			t.Error("should not be stale with no underlying changes")
		}
	})
}
