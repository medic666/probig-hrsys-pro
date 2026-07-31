package service

import (
	"testing"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newSalaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.PositionEvent{}, &model.PositionSnapshot{},
		&model.AttendanceEvent{}, &model.AttendanceDailyProjection{},
		&model.AttendanceCalculationMonthly{},
		&model.SalaryEvent{}, &model.SalarySummary{}, &model.SalarySummaryVersion{},
		&model.AnnualLeaveAccountEvent{},
		&model.SysConfig{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	seedDefaultConfigs(db)
	_ = LoadAllConfigs(db)
	return db
}

func withSalaryDB(t *testing.T, fn func(db *gorm.DB)) {
	t.Helper()
	testMu.Lock()
	defer testMu.Unlock()
	db := newSalaryTestDB(t)
	orig := dao.DB
	dao.DB = db
	defer func() { dao.DB = orig }()
	fn(db)
}

func seedEmployee(db *gorm.DB, personID uint, entryDate string, baseSalary, perfSalary, meal, post, salaryDays float64) {
	d, _ := utils.ParseDate(entryDate)
	dOnly := utils.DateOnlyFromTime(d)
	db.Create(&model.PositionEvent{
		PersonID:          personID, Seq: 1, EventType: "入职", EffectiveDate: dOnly,
		EntryDate: &dOnly, AttendanceGroup: ptr("标准"),
		HasAnnualLeave: ptrBool(true), HasAttendanceBonus: ptrBool(true),
		BaseSalary: ptrFloat(baseSalary), PerformanceSalary: ptrFloat(perfSalary),
		SalaryDays: ptrInt(int(salaryDays)), MealAllowance: ptrFloat(meal),
		PostAllowance: ptrFloat(post),
	})
	RebuildPositionSnapshots(db, personID)
}

func seedLeave(db *gorm.DB, personID uint, leaveDate string) {
	d, _ := utils.ParseDate(leaveDate)
	dOnly := utils.DateOnlyFromTime(d)
	db.Create(&model.PositionEvent{
		PersonID: personID, Seq: 2, EventType: "离职", EffectiveDate: dOnly,
		LeaveDate: &dOnly,
	})
	RebuildPositionSnapshots(db, personID)
}

func seedAttendanceDays(db *gorm.DB, personID uint, monthStr string, days int, hoursPerDay float64) {
	m, _ := utils.MonthStart(monthStr)
	for i := 0; i < days; i++ {
		d := utils.DateOnlyFromTime(m.AddDate(0, 0, i))
		db.Create(&model.AttendanceEvent{
			PersonID: personID, Seq: i + 1, EventDate: d,
			EventType: "出勤", SubType: "普通出勤", Hours: hoursPerDay,
		})
		RebuildDailyProjection(db, personID, d)
	}
}

func seedAttendanceEvent(db *gorm.DB, personID uint, date, evType, subType string, hours float64) {
	d, _ := utils.ParseDate(date)
	dOnly := utils.DateOnlyFromTime(d)
	maxSeq := 0
	db.Unscoped().Model(&model.AttendanceEvent{}).Where("person_id = ?", personID).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
	db.Create(&model.AttendanceEvent{
		PersonID: personID, Seq: maxSeq + 1, EventDate: dOnly,
		EventType: evType, SubType: subType, Hours: hours,
	})
	RebuildDailyProjection(db, personID, dOnly)
}

func seedAnnualLeaveDeduct(db *gorm.DB, personID uint, date string, hours float64) {
	d, _ := utils.ParseDate(date)
	dOnly := utils.DateOnlyFromTime(d)
	maxSeq := 0
	db.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ?", personID).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
	db.Create(&model.AnnualLeaveAccountEvent{
		PersonID: personID, Seq: maxSeq + 1, EventType: "carryover_deduct",
		SourceType: "system_period", Hours: hours, EffectiveDate: dOnly,
	})
	RebuildAnnualLeaveBalance(db, personID)
}

// T1: full month normal
func TestSalaryFullMonthNormal(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 1, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 1, "2026-06", 26, 8)
		_, err := CalculateMonthlyAttendance(1, "2026-06")
		if err != nil {
			t.Fatalf("calc: %v", err)
		}
		err = CalculateSalary(1, "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("salary: %v", err)
		}

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 1, "2026-06").First(&s)

		tolerance := 0.02

		expectPost := 500.0
		if absDiff(s.PostAllowance, expectPost) > tolerance {
			t.Errorf("PostAllowance: got %.2f, want %.2f", s.PostAllowance, expectPost)
		}

		expectMeal := 300.0
		if absDiff(s.MealAllowance, expectMeal) > tolerance {
			t.Errorf("MealAllowance: got %.2f, want %.2f", s.MealAllowance, expectMeal)
		}

		expectPerf := 2000.0
		if absDiff(s.PerformanceSalary, expectPerf) > tolerance {
			t.Errorf("PerformanceSalary: got %.2f, want %.2f", s.PerformanceSalary, expectPerf)
		}

		expectAttBonus := 10.0 * 26.0
		if absDiff(s.AttendanceBonus, expectAttBonus) > tolerance*10 {
			t.Errorf("AttendanceBonus: got %.2f, want ~%.2f", s.AttendanceBonus, expectAttBonus)
		}

		if s.FinalSalary <= 0 {
			t.Errorf("FinalSalary should be > 0, got %.2f", s.FinalSalary)
		}
		t.Logf("T1 full-month: PostAllowance=%.2f, FinalSalary=%.2f", s.PostAllowance, s.FinalSalary)
	})
}

// T2: mid-month leave (15th)
func TestSalaryMidMonthLeave(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 2, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedLeave(db, 2, "2026-06-15")
		seedAttendanceDays(db, 2, "2026-06", 14, 8)
		_, err := CalculateMonthlyAttendance(2, "2026-06")
		if err != nil {
			t.Fatalf("calc: %v", err)
		}
		err = CalculateSalary(2, "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("salary: %v", err)
		}

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 2, "2026-06").First(&s)

		tolerance := 0.03
		ratio := 14.0 / 26.0

		expectPost := 500.0 * ratio
		if absDiff(s.PostAllowance, expectPost) > tolerance {
			t.Errorf("PostAllowance: got %.2f, want %.2f", s.PostAllowance, expectPost)
		}

		expectMeal := 300.0 * ratio
		if absDiff(s.MealAllowance, expectMeal) > tolerance {
			t.Errorf("MealAllowance: got %.2f, want %.2f", s.MealAllowance, expectMeal)
		}

		t.Logf("T2 mid-leave: PostAllowance=%.2f Meal=%.2f ratio=%.4f", s.PostAllowance, s.MealAllowance, ratio)
	})
}

// T3: personal leave → bonus = 0
func TestSalaryPersonalLeaveZeroBonus(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 3, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 3, "2026-06", 25, 8)
		seedAttendanceEvent(db, 3, "2026-06-15", "休假", "事假", 8)

		_, err := CalculateMonthlyAttendance(3, "2026-06")
		if err != nil {
			t.Fatalf("calc: %v", err)
		}
		err = CalculateSalary(3, "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("salary: %v", err)
		}

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 3, "2026-06").First(&s)

		if s.AttendanceBonus != 0 {
			t.Errorf("AttendanceBonus should be 0 with personal leave, got %.2f", s.AttendanceBonus)
		}

		t.Logf("T3 personal-leave: Bonus=%.2f", s.AttendanceBonus)
	})
}

// T4: violations but no personal leave → bonus = (workingDays - violations) * dailyBonus
func TestSalaryViolationsBonus(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 4, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 4, "2026-06", 26, 8)
		seedAttendanceEvent(db, 4, "2026-06-03", "违纪", "迟到", 0)
		seedAttendanceEvent(db, 4, "2026-06-05", "违纪", "缺卡", 0)
		seedAttendanceEvent(db, 4, "2026-06-07", "违纪", "迟到", 0)

		_, err := CalculateMonthlyAttendance(4, "2026-06")
		if err != nil {
			t.Fatalf("calc: %v", err)
		}
		err = CalculateSalary(4, "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("salary: %v", err)
		}

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 4, "2026-06").First(&s)

		expectedBonus := (26.0 - 3.0) * 10.0
		tolerance := 0.5
		if absDiff(s.AttendanceBonus, expectedBonus) > tolerance {
			t.Errorf("AttendanceBonus: got %.2f, want %.2f", s.AttendanceBonus, expectedBonus)
		}

		t.Logf("T4 violations: Bonus=%.2f expected=%.2f", s.AttendanceBonus, expectedBonus)
	})
}

// T5: annual leave carryover salary
func TestSalaryCarryover(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 5, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 5, "2026-06", 26, 8)
		seedAnnualLeaveDeduct(db, 5, "2026-06-15", 16)

		_, err := CalculateMonthlyAttendance(5, "2026-06")
		if err != nil {
			t.Fatalf("calc: %v", err)
		}
		err = CalculateSalary(5, "2026-06", 1, "admin")
		if err != nil {
			t.Fatalf("salary: %v", err)
		}

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 5, "2026-06").First(&s)

		expected := 16.0 * (8000.0 + 300.0) / 26.0 / 8.0 * 2.0
		tolerance := 0.03
		if absDiff(s.AnnualLeaveCarryoverSalary, expected) > tolerance {
			t.Errorf("CarryoverSalary: got %.2f, want %.2f", s.AnnualLeaveCarryoverSalary, expected)
		}
		if absDiff(s.AnnualLeaveCarryoverDeduct, 16.0) > 0.01 {
			t.Errorf("CarryoverDeduct: got %.2f, want 16.0", s.AnnualLeaveCarryoverDeduct)
		}

		t.Logf("T5 carryover: Deduct=%.2f Salary=%.2f", s.AnnualLeaveCarryoverDeduct, s.AnnualLeaveCarryoverSalary)
	})
}

func seedSalaryEvent(db *gorm.DB, personID uint, month, evType string, amount float64) {
	maxSeq := 0
	db.Unscoped().Model(&model.SalaryEvent{}).Where("person_id = ?", personID).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
	db.Create(&model.SalaryEvent{
		PersonID: personID, Seq: maxSeq + 1, BelongMonth: month,
		EventType: evType, Amount: amount,
	})
}

// T6: performance coefficient 1.2
func TestSalaryPerfCoeff(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 6, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 6, "2026-06", 26, 8)
		seedSalaryEvent(db, 6, "2026-06", "绩效系数", 1.2)
		_, _ = CalculateMonthlyAttendance(6, "2026-06")
		_ = CalculateSalary(6, "2026-06", 1, "admin")

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 6, "2026-06").First(&s)

		expect := 2000.0 * 1.2
		if absDiff(s.PerformanceSalary, expect) > 0.02 {
			t.Errorf("PerfSalary: got %.2f, want %.2f", s.PerformanceSalary, expect)
		}
		t.Logf("T6 coeff 1.2: PerfSalary=%.2f (expected %.2f)", s.PerformanceSalary, expect)
	})
}

// T7: commission + reward + loan
func TestSalaryAdjustments(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 7, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 7, "2026-06", 26, 8)
		seedSalaryEvent(db, 7, "2026-06", "提成", 3000)
		seedSalaryEvent(db, 7, "2026-06", "奖惩", -200)
		seedSalaryEvent(db, 7, "2026-06", "借款还款", 1500)
		seedSalaryEvent(db, 7, "2026-06", "个税扣除", 800)
		_, _ = CalculateMonthlyAttendance(7, "2026-06")
		_ = CalculateSalary(7, "2026-06", 1, "admin")

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 7, "2026-06").First(&s)

		if absDiff(s.SalesCommission, 3000) > 0.01 {
			t.Errorf("SalesCommission: got %.2f, want 3000", s.SalesCommission)
		}
		if absDiff(s.RewardPunishment, -200) > 0.01 {
			t.Errorf("RewardPunishment: got %.2f, want -200", s.RewardPunishment)
		}
		if absDiff(s.BorrowingRepayment, 1500) > 0.01 {
			t.Errorf("BorrowingRepayment: got %.2f, want 1500", s.BorrowingRepayment)
		}
		if absDiff(s.TaxDeduct, 800) > 0.01 {
			t.Errorf("TaxDeduct: got %.2f, want 800", s.TaxDeduct)
		}
		t.Logf("T7 adjustments: Commission=%.2f Reward=%.2f Loan=%.2f Tax=%.2f Final=%.2f",
			s.SalesCommission, s.RewardPunishment, s.BorrowingRepayment, s.TaxDeduct, s.FinalSalary)
	})
}

// T8: high temp months (June = yes, January = no)
func TestSalaryHighTempMonth(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		d1, _ := utils.ParseDate("2026-01-01")
		db.Create(&model.PositionEvent{
			PersonID: 8, Seq: 1, EventType: "入职", EffectiveDate: utils.DateOnlyFromTime(d1),
			EntryDate: ptrD(d1), AttendanceGroup: ptr("标准"),
			HasAnnualLeave: ptrBool(true), HasAttendanceBonus: ptrBool(true),
			BaseSalary: ptrFloat(8000), PerformanceSalary: ptrFloat(2000),
			SalaryDays: ptrInt(26), MealAllowance: ptrFloat(300),
			PostAllowance: ptrFloat(500), HighTempAllowance: ptrFloat(300),
		})
		RebuildPositionSnapshots(db, 8)

		// June = high-temp month (config: ["06","07","08","09"])
		seedAttendanceDays(db, 8, "2026-06", 26, 8)
		_, _ = CalculateMonthlyAttendance(8, "2026-06")
		_ = CalculateSalary(8, "2026-06", 1, "admin")
		var sJune model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 8, "2026-06").First(&sJune)

		if sJune.HighTempAllowance <= 0 {
			t.Errorf("June should have high-temp allowance, got %.2f", sJune.HighTempAllowance)
		}

		// January = not high-temp
		seedAttendanceDays(db, 8, "2026-01", 26, 8)
		_, _ = CalculateMonthlyAttendance(8, "2026-01")
		_ = CalculateSalary(8, "2026-01", 1, "admin")
		var sJan model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 8, "2026-01").First(&sJan)

		if sJan.HighTempAllowance != 0 {
			t.Errorf("January should NOT have high-temp allowance, got %.2f", sJan.HighTempAllowance)
		}
		t.Logf("T8 high-temp: June=%.2f Jan=%.2f", sJune.HighTempAllowance, sJan.HighTempAllowance)
	})
}

// T9: mid-month salary change (base 5000→6000 on 11th)
func TestSalaryMidMonthAdjust(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		// 1-10: base=5000, perf=1500, post=400
		d1, _ := utils.ParseDate("2026-01-01")
		db.Create(&model.PositionEvent{
			PersonID: 9, Seq: 1, EventType: "入职", EffectiveDate: utils.DateOnlyFromTime(d1),
			EntryDate: ptrD(d1), AttendanceGroup: ptr("标准"),
			HasAnnualLeave: ptrBool(true), HasAttendanceBonus: ptrBool(true),
			BaseSalary: ptrFloat(5000), PerformanceSalary: ptrFloat(1500),
			SalaryDays: ptrInt(26), MealAllowance: ptrFloat(300),
			PostAllowance: ptrFloat(400),
		})
		RebuildPositionSnapshots(db, 9)

		// 6月11日起调薪
		d11, _ := utils.ParseDate("2026-06-11")
		db.Create(&model.PositionEvent{
			PersonID: 9, Seq: 2, EventType: "调薪调岗", EffectiveDate: utils.DateOnlyFromTime(d11),
			BaseSalary: ptrFloat(6000), PerformanceSalary: ptrFloat(1800),
			PostAllowance: ptrFloat(500),
		})
		RebuildPositionSnapshots(db, 9)

		seedAttendanceDays(db, 9, "2026-06", 26, 8)
		_, _ = CalculateMonthlyAttendance(9, "2026-06")
		_ = CalculateSalary(9, "2026-06", 1, "admin")

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 9, "2026-06").First(&s)

		// 加权 base = (5000×10 + 6000×20) / 30 = 170000/30 = 5666.67
		expectBase := (5000.0*10 + 6000.0*20) / 30.0
		if absDiff(s.WeightedBaseSalary, expectBase) > 0.02 {
			t.Errorf("WeightedBase: got %.2f, want %.2f", s.WeightedBaseSalary, expectBase)
		}

		expectPost := (400.0*10 + 500.0*20) / 30.0
		if absDiff(s.PostAllowance, expectPost) > 0.02 {
			t.Errorf("PostAllowance: got %.2f, want %.2f", s.PostAllowance, expectPost)
		}

		expectPerf := (1500.0*10 + 1800.0*20) / 30.0
		if absDiff(s.PerformanceSalary, expectPerf) > 0.02 {
			t.Errorf("PerfSalary: got %.2f, want %.2f", s.PerformanceSalary, expectPerf)
		}

		t.Logf("T9 mid-adjust: Base=%.2f Post=%.2f Perf=%.2f",
			s.WeightedBaseSalary, s.PostAllowance, s.PerformanceSalary)
	})
}

// T10: complex combo: leave + violation + personal leave + carryover
func TestSalaryComplexCombo(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 10, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedLeave(db, 10, "2026-06-20")
		// active June 1-19 = 19 calendar days, work all of them
		seedAttendanceDays(db, 10, "2026-06", 19, 8)
		// add a violation and personal leave
		seedAttendanceEvent(db, 10, "2026-06-05", "违纪", "迟到", 0)
		seedAttendanceEvent(db, 10, "2026-06-10", "休假", "事假", 8)
		seedAnnualLeaveDeduct(db, 10, "2026-06-15", 8)

		_, _ = CalculateMonthlyAttendance(10, "2026-06")
		_ = CalculateSalary(10, "2026-06", 1, "admin")

		var s model.SalarySummary
		db.Where("person_id = ? AND belong_month = ?", 10, "2026-06").First(&s)

		// bonus = 0 (personal leave triggers zero)
		if s.AttendanceBonus != 0 {
			t.Errorf("Bonus should be 0 with personal leave, got %.2f", s.AttendanceBonus)
		}

		// carryover salary: 8h
		var calc model.AttendanceCalculationMonthly
		db.Where("person_id = ? AND belong_month = ?", 10, "2026-06").First(&calc)
		expectCarry := 8.0 * (calc.WeightedBaseSalary + calc.WeightedMealAllowance) / 26.0 / 8.0 * 2.0
		if absDiff(s.AnnualLeaveCarryoverSalary, expectCarry) > 0.03 {
			t.Errorf("CarryoverSalary: got %.2f, want %.2f", s.AnnualLeaveCarryoverSalary, expectCarry)
		}

		// allowances prorated: 19/26
		ratio := 19.0 / 26.0
		expectPost := 500.0 * ratio
		if absDiff(s.PostAllowance, expectPost) > 0.03 {
			t.Errorf("PostAllowance: got %.2f, want %.2f", s.PostAllowance, expectPost)
		}

		t.Logf("T10 complex: Bonus=%.2f Carryover=%.2f Post=%.2f Final=%.2f",
			s.AttendanceBonus, s.AnnualLeaveCarryoverSalary, s.PostAllowance, s.FinalSalary)
	})
}

func ptrD(t time.Time) *utils.DateOnly { d := utils.DateOnlyFromTime(t); return &d }

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
