package service

import (
	"context"
	"path/filepath"
	"strings"
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

func TestDingTalkExtractPunchTime(t *testing.T) {
	cases := []struct {
		cell string
		want string
	}{
		{"出勤(08:30,18:30)", "08:30,18:30"},
		{"正常(08:47)", "08:47"},
		{"上班外勤,下班缺卡(09:15,-)", "09:15,-"},
		{"下班缺卡(08:30,-)", "08:30,-"},
		{"上班外勤,上班迟到24分钟,下班缺卡(09:54,-)", "09:54,-"},
		{"休息并打卡(-)", ""},
		{"正常(-)", ""},
		{"正常(08:30,-)", "08:30,-"},
		{"正常(-,18:07)", "-,18:07"},
		{"正常", ""},
		{"休息,病假05-22 08:30到05-29 18:30 56小时", ""},
	}
	for _, tc := range cases {
		if got := extractPunchTime(tc.cell); got != tc.want {
			t.Errorf("extractPunchTime(%q): got %q, want %q", tc.cell, got, tc.want)
		}
	}
}

// TestDingTalkParseMultiLeave 同类型多段请假求和 + 休息+请假不再被跳过
func TestDingTalkParseMultiLeave(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-05-18")
		date := utils.DateOnlyFromTime(d)

		cell := "病假05-18 08:30到05-18 12:00 3.5小时,病假05-18 13:30到05-18 18:00 4.5小时"
		c, _ := parseDailyCell(ctx, cell, 200, date)
		if c == 0 {
			t.Fatalf("multi-leave cell: created=0, want >0")
		}
		var daily model.AttendanceDaily
		if err := db.Where("person_id = 200 AND event_date = ?", date).First(&daily).Error; err != nil {
			t.Fatalf("daily not written: %v", err)
		}
		var details []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Find(&details)
		var total float64
		for _, e := range details {
			if e.SubType == "病假" {
				total += e.Hours
			}
		}
		if total != 8 {
			t.Errorf("multi-leave total hours: got %v, want 8", total)
		}

		// 休息+病假：不得被休息日分支丢弃
		cell2 := "休息,病假05-22 08:30到05-29 18:30 56小时"
		c2, _ := parseDailyCell(ctx, cell2, 201, date)
		if c2 == 0 {
			t.Fatalf("rest+leave cell: created=0, want >0 (病假不应被丢弃)")
		}
		var daily2 model.AttendanceDaily
		if err := db.Where("person_id = 201 AND event_date = ?", date).First(&daily2).Error; err != nil {
			t.Fatalf("daily2 not written: %v", err)
		}
		var details2 []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily2.ID).Find(&details2)
		var total2 float64
		for _, e := range details2 {
			if e.SubType == "病假" {
				total2 += e.Hours
			}
		}
		if total2 != 56 {
			t.Errorf("rest+leave hours: got %v, want 56", total2)
		}
	})
}

// TestDingTalkImportIdempotent 重复导入同一日：明细不重复
func TestDingTalkImportIdempotent(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-06-10")
		date := utils.DateOnlyFromTime(d)

		cell := "正常(08:30,18:30)"
		for i := 0; i < 2; i++ {
			c, _ := parseDailyCell(ctx, cell, 300, date)
			if c == 0 {
				t.Fatalf("round %d: created=0", i+1)
			}
		}
		var daily model.AttendanceDaily
		if err := db.Where("person_id = 300 AND event_date = ?", date).First(&daily).Error; err != nil {
			t.Fatalf("daily not written: %v", err)
		}
		var details []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Find(&details)
		if len(details) != 1 {
			t.Errorf("idempotent: details count = %d, want 1", len(details))
		}
	})
}

// TestDingTalkSampleFilesPunchCoverage 真实样本文件：非休息单元格打卡时间覆盖率
func TestDingTalkSampleFilesPunchCoverage(t *testing.T) {
	matches, err := filepath.Glob("../../../examples/广州*月度汇总*.xlsx")
	if err != nil || len(matches) == 0 {
		t.Skip("无样本文件，跳过")
	}
	totalCells, withPunch, restSkip := 0, 0, 0
	missingPunch := map[string]int{}
	for _, fp := range matches {
		_, persons, err := ParseDingTalkExcel(fp)
		if err != nil {
			t.Fatalf("parse %s: %v", fp, err)
		}
		for _, pr := range persons {
			for _, cell := range pr.DailyCells {
				cell = trimSpace(cell)
				if cell == "" {
					continue
				}
				totalCells++
				hasLeave := containsAny(cell, []string{"事假", "年假", "病假", "调休"})
				isRest := strings.HasPrefix(cell, "休息") && !containsAny(cell, []string{"加班", "打卡", "外勤"}) && !hasLeave
				if isRest {
					restSkip++
					continue
				}
				if extractPunchTime(cell) != "" {
					withPunch++
					continue
				}
				// 无打卡时间的单元格必须是"(-)"无打卡标记（换行分隔的括号在末尾）
				if !strings.Contains(cell, "\n(-)") && !strings.Contains(cell, "(-)") {
					missingPunch[cell]++
				}
			}
		}
	}
	t.Logf("样本单元格: %d, 休息跳过: %d, 有打卡时间: %d, 异常无打卡: %d",
		totalCells, restSkip, withPunch, len(missingPunch))
	for cell, cnt := range missingPunch {
		t.Errorf("异常无打卡时间(%d次): %q", cnt, cell)
	}
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
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
