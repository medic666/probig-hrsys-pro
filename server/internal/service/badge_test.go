package service

import (
	"testing"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func migrateBadgeTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&model.Person{}); err != nil {
		t.Fatalf("migrate persons: %v", err)
	}
}

func seedBadgePersons(t *testing.T, db *gorm.DB, ids ...uint) {
	t.Helper()
	for _, id := range ids {
		if err := db.Create(&model.Person{ID: id, Name: "徽章测试"}).Error; err != nil {
			t.Fatalf("seed person %d: %v", id, err)
		}
	}
}

func badgeLevelOf(badges []PersonBadge, personID uint) string {
	for _, b := range badges {
		if b.PersonID == personID {
			return b.Level
		}
	}
	return ""
}

// TestPositionEventBadges 职务事件徽章：无事件 gray / 超两年未变动 orange / 近期有事件 green
func TestPositionEventBadges(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		migrateBadgeTables(t, db)
		seedBadgePersons(t, db, 1, 2, 3)

		oldEvent := utils.DateOnlyFromTime(time.Now().AddDate(-2, 0, -10))
		recentEvent := utils.DateOnlyFromTime(time.Now().AddDate(0, 0, -30))
		db.Create(&model.PositionEvent{PersonID: 2, Seq: 1, EventType: "调薪调岗", EffectiveDate: oldEvent})
		db.Create(&model.PositionEvent{PersonID: 3, Seq: 1, EventType: "调薪调岗", EffectiveDate: recentEvent})

		badges, err := GetPositionEventBadges()
		if err != nil {
			t.Fatalf("badges: %v", err)
		}
		if badgeLevelOf(badges, 1) != "gray" {
			t.Errorf("person 1 (无事件) want gray, got %s", badgeLevelOf(badges, 1))
		}
		if badgeLevelOf(badges, 2) != "orange" {
			t.Errorf("person 2 (超2年无变动) want orange, got %s", badgeLevelOf(badges, 2))
		}
		if badgeLevelOf(badges, 3) != "green" {
			t.Errorf("person 3 (近期事件) want green, got %s", badgeLevelOf(badges, 3))
		}
	})
}

// TestAttendanceEventBadges 考勤事件徽章（上月）：无记录 gray / 待确认 orange / 全确认 green
func TestAttendanceEventBadges(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		migrateBadgeTables(t, db)
		seedBadgePersons(t, db, 1, 2, 3)

		lastMonth := time.Now().AddDate(0, -1, 0)
		day15 := utils.DateOnlyFromTime(time.Date(lastMonth.Year(), lastMonth.Month(), 15, 0, 0, 0, 0, time.Local))
		db.Create(&model.AttendanceDaily{PersonID: 2, EventDate: day15, Status: "pending"})
		db.Create(&model.AttendanceDaily{PersonID: 3, EventDate: day15, Status: "confirmed"})

		badges, err := GetAttendanceEventBadges(DefaultBadgeMonth())
		if err != nil {
			t.Fatalf("badges: %v", err)
		}
		if badgeLevelOf(badges, 1) != "gray" {
			t.Errorf("person 1 (无记录) want gray, got %s", badgeLevelOf(badges, 1))
		}
		if badgeLevelOf(badges, 2) != "orange" {
			t.Errorf("person 2 (待确认) want orange, got %s", badgeLevelOf(badges, 2))
		}
		if badgeLevelOf(badges, 3) != "green" {
			t.Errorf("person 3 (全确认) want green, got %s", badgeLevelOf(badges, 3))
		}
	})
}

// TestDailyProjectionBadges 日记工时徽章（上月）：无投影 gray / 同月有事假又有加班 orange / 仅加班 green
func TestDailyProjectionBadges(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		migrateBadgeTables(t, db)
		seedBadgePersons(t, db, 1, 2, 3)

		lastMonth := time.Now().AddDate(0, -1, 0)
		day10 := utils.DateOnlyFromTime(time.Date(lastMonth.Year(), lastMonth.Month(), 10, 0, 0, 0, 0, time.Local))
		day20 := utils.DateOnlyFromTime(time.Date(lastMonth.Year(), lastMonth.Month(), 20, 0, 0, 0, 0, time.Local))
		db.Create(&model.AttendanceDailyProjection{PersonID: 2, WorkDate: day10, HasPersonalLeave: true})
		db.Create(&model.AttendanceDailyProjection{PersonID: 2, WorkDate: day20, OvertimeWorkdayHours: 2})
		db.Create(&model.AttendanceDailyProjection{PersonID: 3, WorkDate: day10, OvertimeHolidayHours: 4})

		badges, err := GetDailyProjectionBadges(DefaultBadgeMonth())
		if err != nil {
			t.Fatalf("badges: %v", err)
		}
		if badgeLevelOf(badges, 1) != "gray" {
			t.Errorf("person 1 (无投影) want gray, got %s", badgeLevelOf(badges, 1))
		}
		if badgeLevelOf(badges, 2) != "orange" {
			t.Errorf("person 2 (事假+加班) want orange, got %s", badgeLevelOf(badges, 2))
		}
		if badgeLevelOf(badges, 3) != "green" {
			t.Errorf("person 3 (仅加班) want green, got %s", badgeLevelOf(badges, 3))
		}
	})
}

// TestAttendanceMonthlyBadges 月度核算徽章：无核算 gray / 已核算 green / 核算过期 orange
func TestAttendanceMonthlyBadges(t *testing.T) {
	withTestDB(t, func(db *gorm.DB) {
		migrateBadgeTables(t, db)
		seedBadgePersons(t, db, 1, 2, 3)

		month := DefaultBadgeMonth()
		now := time.Now()
		db.Create(&model.AttendanceCalculationMonthly{PersonID: 2, BelongMonth: month, LastCalcAt: now})
		db.Create(&model.AttendanceCalculationMonthly{PersonID: 3, BelongMonth: month, LastCalcAt: now.Add(-time.Hour)})
		lastMonth := time.Now().AddDate(0, -1, 0)
		day15 := utils.DateOnlyFromTime(time.Date(lastMonth.Year(), lastMonth.Month(), 15, 0, 0, 0, 0, time.Local))
		db.Create(&model.AttendanceDailyProjection{PersonID: 3, WorkDate: day15, LastCalcAt: now})

		badges, err := GetAttendanceMonthlyBadges(month)
		if err != nil {
			t.Fatalf("badges: %v", err)
		}
		if badgeLevelOf(badges, 1) != "gray" {
			t.Errorf("person 1 (无核算) want gray, got %s", badgeLevelOf(badges, 1))
		}
		if badgeLevelOf(badges, 2) != "green" {
			t.Errorf("person 2 (已核算) want green, got %s", badgeLevelOf(badges, 2))
		}
		if badgeLevelOf(badges, 3) != "orange" {
			t.Errorf("person 3 (核算过期) want orange, got %s", badgeLevelOf(badges, 3))
		}
	})
}
