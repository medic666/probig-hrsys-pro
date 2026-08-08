package service

import (
	"context"
	"testing"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// 快照局部重建（以时间为轴）回归测试：
// 过去段原样保留（LastCalcAt 不变），仅重建事件生效日及之后的段——
// 未来职务事件不误标已核算月份（考勤核算/工资汇总/徽章同源），改晚/倒填正确标橙且内容正确。

func loadCalc(t *testing.T, db *gorm.DB, personID uint, month string) model.AttendanceCalculationMonthly {
	t.Helper()
	var calc model.AttendanceCalculationMonthly
	if err := db.Where("person_id = ? AND belong_month = ?", personID, month).First(&calc).Error; err != nil {
		t.Fatalf("load calc %s: %v", month, err)
	}
	return calc
}

func loadSummary(t *testing.T, db *gorm.DB, personID uint, month string) model.SalarySummary {
	t.Helper()
	var s model.SalarySummary
	if err := db.Where("person_id = ? AND belong_month = ?", personID, month).First(&s).Error; err != nil {
		t.Fatalf("load summary %s: %v", month, err)
	}
	return s
}

// T1: 新增未来职务事件不误报已核算月份；删除后同样不误报
func TestFuturePositionEventKeepsPastMonthsFresh(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 70, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 70, "2026-03", 26, 8)
		seedAttendanceDays(db, 70, "2026-05", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 70, "2026-03"); err != nil {
			t.Fatalf("calc 2026-03: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 70, "2026-03", 1, "admin"); err != nil {
			t.Fatalf("calc salary 2026-03: %v", err)
		}
		if _, err := CalculateMonthlyAttendance(context.Background(), 70, "2026-05"); err != nil {
			t.Fatalf("calc 2026-05: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 70, "2026-05", 1, "admin"); err != nil {
			t.Fatalf("calc salary 2026-05: %v", err)
		}

		calc3 := loadCalc(t, db, 70, "2026-03")
		sum3 := loadSummary(t, db, 70, "2026-03")
		calc5 := loadCalc(t, db, 70, "2026-05")
		if IsAttendanceMonthlyStale(&calc3) != "calculated" || IsSalarySummaryStale(&sum3) != "calculated" {
			t.Fatalf("initial 2026-03 should be calculated")
		}
		if IsAttendanceMonthlyStale(&calc5) != "calculated" {
			t.Fatalf("initial 2026-05 should be calculated")
		}

		// 记录覆盖 3 月的保留段原始时间戳
		var seg1 model.PositionSnapshot
		if err := db.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			70, "2026-03-15", "2026-03-15").First(&seg1).Error; err != nil {
			t.Fatalf("load seg1: %v", err)
		}
		seg1OldCalcAt := seg1.LastCalcAt

		time.Sleep(5 * time.Millisecond)
		mayD, _ := utils.ParseDate("2026-05-01")
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 70, EventType: "调薪", EffectiveDate: utils.DateOnlyFromTime(mayD),
			BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("create may event: %v", err)
		}

		// 3 月核算/汇总与徽章均不误报
		calc3 = loadCalc(t, db, 70, "2026-03")
		sum3 = loadSummary(t, db, 70, "2026-03")
		if st := IsAttendanceMonthlyStale(&calc3); st != "calculated" {
			t.Errorf("2026-03 attendance stale = %s, want calculated", st)
		}
		if st := IsSalarySummaryStale(&sum3); st != "calculated" {
			t.Errorf("2026-03 salary stale = %s, want calculated", st)
		}
		if badges, err := GetAttendanceMonthlyBadges(context.Background(), "2026-03"); err != nil || badgeLevelOf(badges, 70) != "green" {
			t.Errorf("2026-03 attendance badge: err=%v level=%s, want green", err, badgeLevelOf(badges, 70))
		}
		if badges, err := GetSalarySummariesBadges(context.Background(), "2026-03"); err != nil || badgeLevelOf(badges, 70) != "green" {
			t.Errorf("2026-03 salary badge: err=%v level=%s, want green", err, badgeLevelOf(badges, 70))
		}

		// 5 月（事件生效月）正确标橙
		calc5 = loadCalc(t, db, 70, "2026-05")
		sum5 := loadSummary(t, db, 70, "2026-05")
		if st := IsAttendanceMonthlyStale(&calc5); st != "data_changed" {
			t.Errorf("2026-05 attendance stale = %s, want data_changed", st)
		}
		if st := IsSalarySummaryStale(&sum5); st != "data_changed" {
			t.Errorf("2026-05 salary stale = %s, want data_changed", st)
		}
		if badges, err := GetAttendanceMonthlyBadges(context.Background(), "2026-05"); err != nil || badgeLevelOf(badges, 70) != "orange" {
			t.Errorf("2026-05 attendance badge: err=%v level=%s, want orange", err, badgeLevelOf(badges, 70))
		}

		// 快照结构：过去段截断保留（时间戳不变），新段 [5/1..9999] 内容为新薪资
		var seg1After model.PositionSnapshot
		if err := db.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			70, "2026-03-15", "2026-03-15").First(&seg1After).Error; err != nil {
			t.Fatalf("load seg1 after: %v", err)
		}
		if seg1After.EffectiveEndDate.String() != "2026-04-30" {
			t.Errorf("preserved segment end = %s, want 2026-04-30", seg1After.EffectiveEndDate.String())
		}
		if !seg1After.LastCalcAt.Equal(seg1OldCalcAt) {
			t.Errorf("preserved segment LastCalcAt changed: %v -> %v", seg1OldCalcAt, seg1After.LastCalcAt)
		}
		var seg2 model.PositionSnapshot
		if err := db.Where("person_id = ? AND effective_start_date = ?", 70, "2026-05-01").First(&seg2).Error; err != nil {
			t.Fatalf("load seg2: %v", err)
		}
		if seg2.BaseSalary != 9000 {
			t.Errorf("rebuilt segment base salary = %v, want 9000", seg2.BaseSalary)
		}

		// 删除 5/1 事件：3 月仍不误报，5 月状态回退为事件前薪资
		var mayEvent model.PositionEvent
		if err := db.Where("person_id = ? AND effective_date = ?", 70, "2026-05-01").First(&mayEvent).Error; err != nil {
			t.Fatalf("load may event: %v", err)
		}
		time.Sleep(5 * time.Millisecond)
		if err := DeletePositionEvent(context.Background(), mayEvent.ID); err != nil {
			t.Fatalf("delete may event: %v", err)
		}
		calc3 = loadCalc(t, db, 70, "2026-03")
		if st := IsAttendanceMonthlyStale(&calc3); st != "calculated" {
			t.Errorf("after delete, 2026-03 attendance stale = %s, want calculated", st)
		}
		var maySeg model.PositionSnapshot
		if err := db.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			70, "2026-05-15", "2026-05-15").First(&maySeg).Error; err != nil {
			t.Fatalf("load may seg after delete: %v", err)
		}
		if maySeg.BaseSalary != 8000 {
			t.Errorf("may segment after delete = %v, want 8000 (reverted)", maySeg.BaseSalary)
		}
	})
}

// T2: 事件生效日改晚（5/1→6/1）：中间区间状态回退 + 标橙（验证切点取 min(旧,新)）
func TestPositionEventMovedLaterRevertsStateAndFlags(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 71, "2026-01-01", 8000, 2000, 300, 500, 26)
		mayD, _ := utils.ParseDate("2026-05-01")
		mayDOnly := utils.DateOnlyFromTime(mayD)
		var ev model.PositionEvent
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 71, EventType: "调薪", EffectiveDate: mayDOnly, BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("create may event: %v", err)
		}
		if err := db.Where("person_id = ? AND effective_date = ?", 71, mayDOnly).First(&ev).Error; err != nil {
			t.Fatalf("load event: %v", err)
		}

		seedAttendanceDays(db, 71, "2026-05", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 71, "2026-05"); err != nil {
			t.Fatalf("calc 2026-05: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 71, "2026-05", 1, "admin"); err != nil {
			t.Fatalf("calc salary 2026-05: %v", err)
		}
		calc5 := loadCalc(t, db, 71, "2026-05")
		if IsAttendanceMonthlyStale(&calc5) != "calculated" {
			t.Fatalf("initial 2026-05 should be calculated")
		}

		time.Sleep(5 * time.Millisecond)
		juneD, _ := utils.ParseDate("2026-06-01")
		if err := UpdatePositionEvent(context.Background(), ev.ID, &model.PositionEvent{
			EventType: "调薪", EffectiveDate: utils.DateOnlyFromTime(juneD), BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("move event to june: %v", err)
		}

		// 5 月标橙（状态回退 = 内容变化）
		calc5 = loadCalc(t, db, 71, "2026-05")
		if st := IsAttendanceMonthlyStale(&calc5); st != "data_changed" {
			t.Errorf("2026-05 attendance stale = %s, want data_changed", st)
		}
		if badges, err := GetSalarySummariesBadges(context.Background(), "2026-05"); err != nil || badgeLevelOf(badges, 71) != "orange" {
			t.Errorf("2026-05 salary badge: err=%v level=%s, want orange", err, badgeLevelOf(badges, 71))
		}

		// 内容正确：5 月已回退为事件前薪资，6 月起仍为事件后薪资
		var maySeg model.PositionSnapshot
		if err := db.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			71, "2026-05-15", "2026-05-15").First(&maySeg).Error; err != nil {
			t.Fatalf("load may segment: %v", err)
		}
		if maySeg.BaseSalary != 8000 {
			t.Errorf("may segment salary = %v, want 8000 (reverted)", maySeg.BaseSalary)
		}
		var juneSeg model.PositionSnapshot
		if err := db.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
			71, "2026-06-15", "2026-06-15").First(&juneSeg).Error; err != nil {
			t.Fatalf("load june segment: %v", err)
		}
		if juneSeg.BaseSalary != 9000 {
			t.Errorf("june segment salary = %v, want 9000", juneSeg.BaseSalary)
		}
	})
}

// T3: 事件倒填（生效于已核算月份内）→ 正确标橙
func TestBackdatedPositionEventFlagsMonth(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 72, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 72, "2026-03", 26, 8)
		if _, err := CalculateMonthlyAttendance(context.Background(), 72, "2026-03"); err != nil {
			t.Fatalf("calc 2026-03: %v", err)
		}
		if _, err := CalculateSalary(context.Background(), 72, "2026-03", 1, "admin"); err != nil {
			t.Fatalf("calc salary 2026-03: %v", err)
		}
		calc3 := loadCalc(t, db, 72, "2026-03")
		if IsAttendanceMonthlyStale(&calc3) != "calculated" {
			t.Fatalf("initial 2026-03 should be calculated")
		}

		time.Sleep(5 * time.Millisecond)
		midD, _ := utils.ParseDate("2026-03-15")
		if err := CreatePositionEvent(context.Background(), &model.PositionEvent{
			PersonID: 72, EventType: "调薪", EffectiveDate: utils.DateOnlyFromTime(midD), BaseSalary: ptrFloat(9000),
		}); err != nil {
			t.Fatalf("create backdated event: %v", err)
		}

		calc3 = loadCalc(t, db, 72, "2026-03")
		if st := IsAttendanceMonthlyStale(&calc3); st != "data_changed" {
			t.Errorf("2026-03 attendance stale = %s, want data_changed", st)
		}
		if badges, err := GetAttendanceMonthlyBadges(context.Background(), "2026-03"); err != nil || badgeLevelOf(badges, 72) != "orange" {
			t.Errorf("2026-03 attendance badge: err=%v level=%s, want orange", err, badgeLevelOf(badges, 72))
		}
	})
}
