package service

import (
	"context"
	"testing"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func TestLatestTime(t *testing.T) {
	if got := LatestTime(nil); !got.IsZero() {
		t.Errorf("empty slice should return zero time, got %v", got)
	}
	base := time.Now()
	before, after := base.Add(-time.Hour), base.Add(time.Hour)
	// nil 忽略 + 取最大
	if got := LatestTime([]*time.Time{nil, &before, &after}); !got.Equal(after) {
		t.Errorf("LatestTime = %v, want %v", got, after)
	}
	// 全 nil → 零值
	if got := LatestTime([]*time.Time{nil, nil}); !got.IsZero() {
		t.Errorf("all-nil should return zero time, got %v", got)
	}
}

func TestIsStaleAfter(t *testing.T) {
	now := time.Now()
	newer := now.Add(time.Hour)
	older := now.Add(-time.Hour)

	if !IsStaleAfter(now, []*time.Time{&older}, []*time.Time{&newer}) {
		t.Error("any source newer than result should be stale")
	}
	if IsStaleAfter(now, []*time.Time{&older}, []*time.Time{&older}) {
		t.Error("all sources older should not be stale")
	}
	if IsStaleAfter(now) {
		t.Error("no sources should not be stale")
	}
	if IsStaleAfter(now, []*time.Time{nil}) {
		t.Error("nil-only source should not affect staleness")
	}
}

func TestStaleDetectionPreciseWithinSameDay(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 60, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 60, "2026-06", 26, 8)

		if _, err := CalculateMonthlyAttendance(context.Background(), 60, "2026-06"); err != nil {
			t.Fatalf("calc: %v", err)
		}
		var calc model.AttendanceCalculationMonthly
		if err := db.Where("person_id = ? AND belong_month = ?", 60, "2026-06").First(&calc).Error; err != nil {
			t.Fatalf("load calc: %v", err)
		}
		if st := IsAttendanceMonthlyStale(&calc); st != "calculated" {
			t.Fatalf("expected calculated, got %s", st)
		}

		// 同日变动：重建日记工时投影（时间戳推进）
		time.Sleep(5 * time.Millisecond)
		d, _ := utils.ParseDate("2026-06-01")
		if err := RebuildDailyProjection(db, 60, utils.DateOnlyFromTime(d)); err != nil {
			t.Fatalf("rebuild projection: %v", err)
		}
		if err := db.Where("person_id = ? AND belong_month = ?", 60, "2026-06").First(&calc).Error; err != nil {
			t.Fatalf("reload calc: %v", err)
		}
		if st := IsAttendanceMonthlyStale(&calc); st != "data_changed" {
			t.Fatalf("expected data_changed after same-day change, got %s", st)
		}

		// 重算后恢复 calculated
		if _, err := CalculateMonthlyAttendance(context.Background(), 60, "2026-06"); err != nil {
			t.Fatalf("recalc: %v", err)
		}
		var calc2 model.AttendanceCalculationMonthly
		if err := db.Where("person_id = ? AND belong_month = ?", 60, "2026-06").First(&calc2).Error; err != nil {
			t.Fatalf("reload calc: %v", err)
		}
		if st := IsAttendanceMonthlyStale(&calc2); st != "calculated" {
			t.Fatalf("expected calculated after recalc, got %s", st)
		}
	})
}

func TestSalaryStaleDetectionOnEventChange(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 61, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 61, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 61, "2026-06"); err != nil {
			t.Fatalf("calc: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 61, "2026-06", 1, "admin"); err != nil {
			t.Fatalf("calc salary: %v", err)
		}

		var s model.SalarySummary
		if err := db.Where("person_id = ? AND belong_month = ?", 61, "2026-06").First(&s).Error; err != nil {
			t.Fatalf("load summary: %v", err)
		}
		if st := IsSalarySummaryStale(&s); st != "calculated" {
			t.Fatalf("expected calculated, got %s", st)
		}

		// 同日新增工资事件 → data_changed
		time.Sleep(5 * time.Millisecond)
		seedSalaryEvent(db, 61, "2026-06", "提成", 100)
		if err := db.Where("person_id = ? AND belong_month = ?", 61, "2026-06").First(&s).Error; err != nil {
			t.Fatalf("reload summary: %v", err)
		}
		if st := IsSalarySummaryStale(&s); st != "data_changed" {
			t.Fatalf("expected data_changed after salary event, got %s", st)
		}

		// 同日删除工资事件 → data_changed（软删除时间纳入检测）
		time.Sleep(5 * time.Millisecond)
		var ev model.SalaryEvent
		if err := db.Where("person_id = ? AND belong_month = ?", 61, "2026-06").First(&ev).Error; err != nil {
			t.Fatalf("load event: %v", err)
		}
		if err := db.Delete(&ev).Error; err != nil {
			t.Fatalf("delete event: %v", err)
		}
		if err := db.Where("person_id = ? AND belong_month = ?", 61, "2026-06").First(&s).Error; err != nil {
			t.Fatalf("reload summary: %v", err)
		}
		if st := IsSalarySummaryStale(&s); st != "data_changed" {
			t.Fatalf("expected data_changed after event deletion, got %s", st)
		}
	})
}
