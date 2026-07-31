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
		if err := CalculateSalary(context.Background(), 20, "2026-06", 1, "admin"); err != nil {
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

		if err := CalculateSalary(context.Background(), 20, "2026-06", 1, "admin"); err == nil {
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

func TestCalculateMonthlyAttendanceKeepsOldResultOnFailure(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 21, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 21, "2026-06", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 21, "2026-06"); err != nil {
			t.Fatalf("calc attendance: %v", err)
		}

		// 新增一条 pending 记录 → 重算必须失败且不破坏旧结果
		d, _ := utils.ParseDate("2026-06-20")
		dOnly := utils.DateOnlyFromTime(d)
		db.Create(&model.AttendanceDailyProjection{
			PersonID:   21,
			WorkDate:   dOnly,
			Status:     "pending",
			LastCalcAt: time.Now(),
		})

		if _, err := CalculateMonthlyAttendance(context.Background(), 21, "2026-06"); err == nil {
			t.Fatalf("expected failure with pending records, got nil")
		}

		var calc model.AttendanceCalculationMonthly
		if err := db.Where("person_id = ? AND belong_month = ?", 21, "2026-06").First(&calc).Error; err != nil {
			t.Fatalf("old calc should be preserved: %v", err)
		}
		if calc.TotalWorkHours == 0 {
			t.Errorf("old calc result lost, got TotalWorkHours=0")
		}
	})
}
