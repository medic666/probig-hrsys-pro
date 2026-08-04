package service

import (
	"context"
	"testing"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func TestCalculateSalaryRollbackOnFailure(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 20, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 20, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 20, "2026-06"); err != nil {
			t.Fatalf("calc attendance: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 20, "2026-06", 1, "admin"); err != nil {
			t.Fatalf("calc salary: %v", err)
		}

		var s model.SalarySummary
		if err := db.Where("person_id = ? AND belong_month = ?", 20, "2026-06").First(&s).Error; err != nil {
			t.Fatalf("load summary: %v", err)
		}
		oldFinal := s.FinalSalary

		// 注入：工资汇总插入失败（SQLite trigger 模拟真实 SQL 错误）
		if err := db.Exec("CREATE TRIGGER test_fail_summary BEFORE INSERT ON salary_summaries BEGIN SELECT RAISE(ABORT, 'injected failure'); END").Error; err != nil {
			t.Fatalf("create trigger: %v", err)
		}
		defer db.Exec("DROP TRIGGER IF EXISTS test_fail_summary")

		if _, err := CalculateSalary(context.Background(), 20, "2026-06", 1, "admin"); err == nil {
			t.Fatalf("expected failure, got nil")
		}

		// 汇总应保持旧值
		var after model.SalarySummary
		if err := db.Where("person_id = ? AND belong_month = ?", 20, "2026-06").First(&after).Error; err != nil {
			t.Fatalf("summary should still exist: %v", err)
		}
		if after.FinalSalary != oldFinal {
			t.Errorf("summary should be rolled back, got %.2f want %.2f", after.FinalSalary, oldFinal)
		}

		// 无孤儿版本记录
		var verCount int64
		db.Model(&model.SalarySummaryVersion{}).Where("person_id = ? AND belong_month = ?", 20, "2026-06").Count(&verCount)
		if verCount != 1 {
			t.Errorf("expected 1 version, got %d (orphan version created)", verCount)
		}
	})
}

// TestCalculateMonthlyAttendanceClearsOnPending 存在 pending 日记工时 → 核算结果置空（无值语义）：
// 不报错、不保留旧结果，旧核算记录被物理删除，等待确认后重新核算
func TestCalculateMonthlyAttendanceClearsOnPending(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 21, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 21, "2026-06", 26, 8)
		r, err := CalculateMonthlyAttendance(context.Background(), 21, "2026-06")
		if err != nil || r == nil {
			t.Fatalf("calc attendance: result=%v err=%v", r, err)
		}

		// 新增一条 pending 记录 → 重算：结果置空（无值）
		d, _ := utils.ParseDate("2026-06-20")
		dOnly := utils.DateOnlyFromTime(d)
		db.Create(&model.AttendanceDailyProjection{
			PersonID:   21,
			WorkDate:   dOnly,
			Status:     "pending",
			LastCalcAt: time.Now(),
		})

		r2, err2 := CalculateMonthlyAttendance(context.Background(), 21, "2026-06")
		if err2 != nil {
			t.Fatalf("pending should be empty result, got err: %v", err2)
		}
		if r2 != nil {
			t.Fatalf("pending should yield nil result, got %v", r2)
		}

		var count int64
		db.Model(&model.AttendanceCalculationMonthly{}).Where("person_id = ? AND belong_month = ?", 21, "2026-06").Count(&count)
		if count != 0 {
			t.Errorf("old calc should be cleared (empty), got %d records", count)
		}
	})
}
