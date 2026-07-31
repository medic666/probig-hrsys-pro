package service

import (
	"context"
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// 入职月且有请假：补贴/绩效按实际出勤比例折算（企划 3.4.2 规则 3/4）
func TestSalaryMidMonthEntryWithLeave(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		// 6/10 入职，在职 21 天（6/10-6/30），其中最后 3 天事假
		d, _ := utils.ParseDate("2026-06-10")
		dOnly := utils.DateOnlyFromTime(d)
		if err := db.Create(&model.PositionEvent{
			PersonID: 99, Seq: 1, EventType: "入职", EffectiveDate: dOnly,
			EntryDate: &dOnly, AttendanceGroup: ptr("标准"),
			HasAnnualLeave:     ptrBool(true),
			HasAttendanceBonus: ptrBool(true),
			BaseSalary:         ptrFloat(8000),
			PerformanceSalary:  ptrFloat(2000),
			SalaryDays:         ptrInt(26),
			MealAllowance:      ptrFloat(300),
		}).Error; err != nil {
			t.Fatalf("seed entry: %v", err)
		}
		RebuildPositionSnapshots(db, 99)

		for i := 0; i < 21; i++ {
			date := dOnly.AddDate(0, 0, i)
			daily := model.AttendanceDaily{PersonID: 99, EventDate: date, Status: "confirmed"}
			if err := db.Create(&daily).Error; err != nil {
				t.Fatalf("seed daily: %v", err)
			}
			evType, subType := "出勤", "普通出勤"
			if i >= 18 {
				evType, subType = "休假", "事假"
			}
			if err := db.Create(&model.AttendanceEventDetail{DailyID: daily.ID, EventType: evType, SubType: subType, Hours: 8}).Error; err != nil {
				t.Fatalf("seed detail: %v", err)
			}
			if err := RebuildDailyProjection(db, 99, date); err != nil {
				t.Fatalf("rebuild: %v", err)
			}
		}

		if _, err := CalculateMonthlyAttendance(context.Background(), 99, "2026-06"); err != nil {
			t.Fatalf("calc: %v", err)
		}
		if err := CalculateSalary(context.Background(), 99, "2026-06", 1, "admin"); err != nil {
			t.Fatalf("salary: %v", err)
		}
		var s model.SalarySummary
		if err := db.Where("person_id = ? AND belong_month = ?", 99, "2026-06").First(&s).Error; err != nil {
			t.Fatalf("load summary: %v", err)
		}

		// 18 天出勤 × 8h → attendanceDays=18；在职 21 天全勤即 21 天，请假 3 天
		// 餐补 = (300×21/21) × (18/26) = 207.69
		expect := 300.0 * 18.0 / 26.0
		if absDiff(s.MealAllowance, expect) > 0.03 {
			t.Errorf("MealAllowance: got %.2f, want %.2f (请假应折算)", s.MealAllowance, expect)
		}
		// 绩效同构折算
		expectPerf := 2000.0 * 18.0 / 26.0
		if absDiff(s.PerformanceSalary, expectPerf) > 0.03 {
			t.Errorf("PerfSalary: got %.2f, want %.2f", s.PerformanceSalary, expectPerf)
		}
		// 出勤工资按工时
		expectAtt := 144.0 * (8000.0 / 26.0 / 8.0)
		if absDiff(s.AttendanceSalary, expectAtt) > 0.03 {
			t.Errorf("AttendanceSalary: got %.2f, want %.2f", s.AttendanceSalary, expectAtt)
		}
		t.Logf("entry-with-leave: Meal=%.2f Perf=%.2f Att=%.2f Final=%.2f",
			s.MealAllowance, s.PerformanceSalary, s.AttendanceSalary, s.FinalSalary)
	})
}
