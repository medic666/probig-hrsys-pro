package service

import (
	"context"
	"testing"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// 列表筛选回归：月份多选（months IN）与状态筛选（派生值过滤 + 分页总数正确）

func TestMonthlyListMonthsAndStatusFilter(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 90, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 90, "2026-03", 26, 8)
		seedAttendanceDays(db, 90, "2026-05", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 90, "2026-03"); err != nil {
			t.Fatalf("calc 2026-03: %v", err)
		}
		if _, err := CalculateMonthlyAttendance(context.Background(), 90, "2026-05"); err != nil {
			t.Fatalf("calc 2026-05: %v", err)
		}
		// 重建 5/10 投影 → 5 月核算过期（data_changed），3 月保持 calculated
		time.Sleep(5 * time.Millisecond)
		d, _ := utils.ParseDate("2026-05-10")
		if err := RebuildDailyProjection(db, 90, utils.DateOnlyFromTime(d)); err != nil {
			t.Fatalf("rebuild projection: %v", err)
		}

		// 状态筛选：多月中仅 5 月 data_changed，total=1
		rows, total, err := GetMonthlyList(MonthlyListQuery{
			PageNum: 1, PageSize: 10, Months: []string{"2026-03", "2026-05"}, Status: "data_changed",
		})
		if err != nil {
			t.Fatalf("status filter: %v", err)
		}
		if total != 1 || len(rows) != 1 {
			t.Fatalf("status filter: total=%d len=%d, want 1/1", total, len(rows))
		}
		if rows[0]["belong_month"] != "2026-05" || rows[0]["status"] != "data_changed" {
			t.Errorf("status filter row = %v/%v, want 2026-05/data_changed", rows[0]["belong_month"], rows[0]["status"])
		}

		// 月份多选（无状态）：两月共 2 行
		rows, total, _ = GetMonthlyList(MonthlyListQuery{PageNum: 1, PageSize: 10, Months: []string{"2026-03", "2026-05"}})
		if total != 2 || len(rows) != 2 {
			t.Errorf("months filter: total=%d len=%d, want 2/2", total, len(rows))
		}

		// 单月兼容（详情页取数路径）
		rows, total, _ = GetMonthlyList(MonthlyListQuery{PageNum: 1, PageSize: 10, Month: "2026-03"})
		if total != 1 || len(rows) != 1 || rows[0]["belong_month"] != "2026-03" {
			t.Errorf("single month: total=%d len=%d row=%v, want 1/1/2026-03", total, len(rows), rows[0]["belong_month"])
		}

		// 状态筛选分页：pageSize=1 时第二页返回剩余 1 行
		rows, total, _ = GetMonthlyList(MonthlyListQuery{PageNum: 2, PageSize: 1, Months: []string{"2026-03", "2026-05"}})
		if total != 2 || len(rows) != 1 {
			t.Errorf("status pagination: total=%d len=%d, want 2/1", total, len(rows))
		}
	})
}

func TestSalaryEventMonthsFilter(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.Create(&model.Person{ID: 91, Name: "测试员工"})
		seedSalaryEvent(db, 91, "2026-03", "提成", 100)
		seedSalaryEvent(db, 91, "2026-04", "提成", 200)
		seedSalaryEvent(db, 91, "2026-04", "奖惩", 50)

		// 月份多选
		rows, total, err := GetSalaryEventList(SalaryEventListQuery{PageNum: 1, PageSize: 10, Months: []string{"2026-03"}})
		if err != nil {
			t.Fatalf("months filter: %v", err)
		}
		if total != 1 || len(rows) != 1 {
			t.Errorf("single month: total=%d len=%d, want 1/1", total, len(rows))
		}
		rows, total, _ = GetSalaryEventList(SalaryEventListQuery{
			PageNum: 1, PageSize: 10, PersonID: 91, Months: []string{"2026-03", "2026-04"},
		})
		if total != 3 || len(rows) != 3 {
			t.Errorf("multi months: total=%d len=%d, want 3/3", total, len(rows))
		}
	})
}
