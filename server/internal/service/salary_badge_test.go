package service

import (
	"context"
	"testing"
	"time"

	"probig/server/internal/model"

	"gorm.io/gorm"
)

// TestSalaryAdvanceBalances 工资预支余额：SUM(工资预支) 正累加，预支还款不参与
func TestSalaryAdvanceBalances(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.Create(&model.Person{ID: 55, Name: "预支测试"})
		seedSalaryEvent(db, 55, "2026-05", "工资预支", 3000)
		seedSalaryEvent(db, 55, "2026-06", "工资预支", 2000)
		seedSalaryEvent(db, 55, "2026-06", "预支还款", -1000)

		balances, err := GetSalaryAdvanceBalances()
		if err != nil {
			t.Fatalf("advance balances: %v", err)
		}
		found := false
		for _, b := range balances {
			if b.PersonID == 55 {
				found = true
				if b.Balance != 5000 {
					t.Errorf("advance balance = %v, want 5000 (只累加工资预支)", b.Balance)
				}
			}
		}
		if !found {
			t.Errorf("person 55 missing in advance balances")
		}
	})
}

// TestSalarySummariesBadges 工资汇总徽章三态：无汇总 gray / 正常 green / 事件变动后 orange
func TestSalarySummariesBadges(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 56, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 56, "2026-06", 5, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 56, "2026-06"); err != nil {
			t.Fatalf("calc attendance: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 56, "2026-06", 1, "admin"); err != nil {
			t.Fatalf("calc salary: %v", err)
		}

		// 无汇总记录人员 → gray
		db.Create(&model.Person{ID: 57, Name: "无汇总"})
		seedEmployee(db, 57, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 57, "2026-06", 5, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 57, "2026-06"); err != nil {
			t.Fatalf("calc attendance 57: %v", err)
		}

		badges, err := GetSalarySummariesBadges("2026-06")
		if err != nil {
			t.Fatalf("badges: %v", err)
		}
		if badgeLevelOf(badges, 56) != "green" {
			t.Errorf("person 56 want green, got %s", badgeLevelOf(badges, 56))
		}
		if badgeLevelOf(badges, 57) != "gray" {
			t.Errorf("person 57 (无汇总) want gray, got %s", badgeLevelOf(badges, 57))
		}

		// 工资事件变动（updated_at 推进）→ orange
		time.Sleep(5 * time.Millisecond)
		seedSalaryEvent(db, 56, "2026-06", "预支还款", -500)
		badges2, err := GetSalarySummariesBadges("2026-06")
		if err != nil {
			t.Fatalf("badges2: %v", err)
		}
		if badgeLevelOf(badges2, 56) != "orange" {
			t.Errorf("person 56 after event change want orange, got %s", badgeLevelOf(badges2, 56))
		}
	})
}

// TestSalaryAdvanceEventsNotInCalc 工资预支不参与核算；预支还款参与（替代借款还款位置）
func TestSalaryAdvanceEventsNotInCalc(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 58, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 58, "2026-06", 5, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 58, "2026-06"); err != nil {
			t.Fatalf("calc attendance: %v", err)
		}
		seedSalaryEvent(db, 58, "2026-06", "工资预支", 3000)  // 不参与核算
		seedSalaryEvent(db, 58, "2026-06", "预支还款", -800)  // 参与核算（BorrowingRepayment = -800）
		s, err := CalculateSalary(context.Background(), 58, "2026-06", 1, "admin")
		if err != nil || s == nil {
			t.Fatalf("calc salary: %v %v", s, err)
		}
		if s.BorrowingRepayment != -800 {
			t.Errorf("BorrowingRepayment = %v, want -800（预支还款参与、工资预支不参与）", s.BorrowingRepayment)
		}
	})
}
