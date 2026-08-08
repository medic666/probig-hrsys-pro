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

func TestRowDataChanged(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.Create(&model.Person{ID: 62, Name: "stale"})
		sources := []StaleSource{
			{Model: &model.SalaryEvent{}, Column: "updated_at", Unscoped: true,
				Where: "belong_month = ?", Args: []interface{}{"2026-06"}},
		}

		// 无源记录 → 不 stale
		changed, err := RowDataChanged(time.Now(), 62, sources)
		if err != nil {
			t.Fatalf("row data changed: %v", err)
		}
		if changed {
			t.Error("no sources should not be stale")
		}

		// 任一源时间晚于结果时间 → stale
		time.Sleep(5 * time.Millisecond)
		before := time.Now()
		seedSalaryEvent(db, 62, "2026-06", "提成", 100)
		changed, err = RowDataChanged(before, 62, sources)
		if err != nil {
			t.Fatalf("row data changed: %v", err)
		}
		if !changed {
			t.Error("source newer than result should be stale")
		}

		// 所有源早于结果时间 → 不 stale
		after := time.Now().Add(time.Hour)
		changed, err = RowDataChanged(after, 62, sources)
		if err != nil {
			t.Fatalf("row data changed: %v", err)
		}
		if changed {
			t.Error("all sources older should not be stale")
		}
	})
}

// TestAttendanceStaleOnEventDeletion 事件全删 → 投影清空：事件表 deleted_at 信号兜底，
// 核算仍须判 data_changed（防止"投影消失即无信号"的假绿）
func TestAttendanceStaleOnEventDeletion(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 63, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 63, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 63, "2026-06"); err != nil {
			t.Fatalf("calc: %v", err)
		}
		var calc model.AttendanceCalculationMonthly
		if err := db.Where("person_id = ? AND belong_month = ?", 63, "2026-06").First(&calc).Error; err != nil {
			t.Fatalf("load calc: %v", err)
		}
		if st := IsAttendanceMonthlyStale(&calc); st != "calculated" {
			t.Fatalf("expected calculated before deletion, got %s", st)
		}

		// 删除当月全部考勤事件 + 投影（模拟 DeleteAttendanceDaily 后的重建结果）
		time.Sleep(5 * time.Millisecond)
		if err := db.Where("person_id = ?", 63).Delete(&model.AttendanceDaily{}).Error; err != nil {
			t.Fatalf("delete dailies: %v", err)
		}
		if err := db.Where("person_id = ?", 63).Delete(&model.AttendanceDailyProjection{}).Error; err != nil {
			t.Fatalf("delete projections: %v", err)
		}
		var calc2 model.AttendanceCalculationMonthly
		if err := db.Where("person_id = ? AND belong_month = ?", 63, "2026-06").First(&calc2).Error; err != nil {
			t.Fatalf("reload calc: %v", err)
		}
		if st := IsAttendanceMonthlyStale(&calc2); st != "data_changed" {
			t.Fatalf("expected data_changed after event deletion, got %s", st)
		}
	})
}

// TestSalaryStaleOnAttendanceDeletion 考勤事件全删致投影/核算清空：工资汇总仍须判 data_changed
func TestSalaryStaleOnAttendanceDeletion(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 64, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 64, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 64, "2026-06"); err != nil {
			t.Fatalf("calc attendance: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 64, "2026-06", 1, "admin"); err != nil {
			t.Fatalf("calc salary: %v", err)
		}
		var s model.SalarySummary
		if err := db.Where("person_id = ? AND belong_month = ?", 64, "2026-06").First(&s).Error; err != nil {
			t.Fatalf("load summary: %v", err)
		}
		if st := IsSalarySummaryStale(&s); st != "calculated" {
			t.Fatalf("expected calculated before deletion, got %s", st)
		}

		// 删除当月全部考勤事件 + 投影 + 重算考勤（核算被清空）→ 汇总仍须 data_changed
		time.Sleep(5 * time.Millisecond)
		if err := db.Where("person_id = ?", 64).Delete(&model.AttendanceDaily{}).Error; err != nil {
			t.Fatalf("delete dailies: %v", err)
		}
		if err := db.Where("person_id = ?", 64).Delete(&model.AttendanceDailyProjection{}).Error; err != nil {
			t.Fatalf("delete projections: %v", err)
		}
		if _, err := CalculateMonthlyAttendance(context.Background(), 64, "2026-06"); err != nil {
			t.Fatalf("recalc attendance: %v", err)
		}
		var s2 model.SalarySummary
		if err := db.Where("person_id = ? AND belong_month = ?", 64, "2026-06").First(&s2).Error; err != nil {
			t.Fatalf("reload summary: %v", err)
		}
		if st := IsSalarySummaryStale(&s2); st != "data_changed" {
			t.Fatalf("expected data_changed after attendance deletion, got %s", st)
		}
	})
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
