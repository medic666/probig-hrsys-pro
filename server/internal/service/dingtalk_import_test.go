package service

import (
	"context"
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func TestDingTalkExtractHelpers(t *testing.T) {
	if h := extractOTHours("休息日加班 3.5小时"); h != 3.5 {
		t.Errorf("extractOTHours: got %v, want 3.5", h)
	}
	if h := extractOTHours("加班"); h != 0 {
		t.Errorf("extractOTHours no-hours: got %v, want 0", h)
	}
	if typ, h := parseLeaveFromCell("事假2.5小时"); typ != "事假" || h != 2.5 {
		t.Errorf("parseLeaveFromCell: got %q/%v, want 事假/2.5", typ, h)
	}
	if typ, _ := parseLeaveFromCell("加班2小时"); typ != "" {
		t.Errorf("parseLeaveFromCell no-leave: got %q, want empty", typ)
	}
	if m := extractLateMinutes("迟到15分钟 早退10分钟", "迟到"); m != 15 {
		t.Errorf("extractLateMinutes: got %d, want 15", m)
	}
	if m := extractLateMinutes("迟到", "迟到"); m != 0 {
		t.Errorf("extractLateMinutes no-minutes: got %d, want 0", m)
	}
}

func TestDingTalkParseDailyCell(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-06-10")
		date := utils.DateOnlyFromTime(d)
		cases := []struct {
			cell    string
			created int
			pending int
		}{
			{"出勤(08:30,18:30)", 1, 0},
			{"外勤", 1, 0},
			{"迟到31分钟", 1, 1},
			{"早退5分钟", 1, 0},
			{"缺卡", 1, 1},
			{"旷工", 1, 1},
			{"休息", 0, 0},
			{"休息并打卡(08:00,12:00)", 1, 1},
			{"休息加班3小时", 1, 0},
			{"事假4小时", 1, 1},
			{"年假8小时", 1, 0},
			{"加班2小时", 1, 0},
		}
		for i, tc := range cases {
			pid := uint(100 + i)
			c, p := parseDailyCell(ctx, tc.cell, pid, date)
			if c != tc.created || p != tc.pending {
				t.Errorf("cell %q: created=%d pending=%d, want %d/%d", tc.cell, c, p, tc.created, tc.pending)
			}
			if c > 0 {
				var daily model.AttendanceDaily
				if err := db.Where("person_id = ? AND event_date = ?", pid, date).First(&daily).Error; err != nil {
					t.Errorf("cell %q: daily not written: %v", tc.cell, err)
					continue
				}
				var details []model.AttendanceEventDetail
				db.Where("daily_id = ?", daily.ID).Find(&details)
				if len(details) == 0 {
					t.Errorf("cell %q: no details written", tc.cell)
				}
			}
		}
	})
}
