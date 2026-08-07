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

// TestDingTalkWorkdayOTIncludesAttendance 工作日加班：当日正常出勤 8h + 加班工时均需记录
// （样本为"正常,加班MM-DD HH:MM到MM-DD HH:MM X小时(打卡)"形态，如熊圣平 2026-03-06）
func TestDingTalkWorkdayOTIncludesAttendance(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-03-06")
		date := utils.DateOnlyFromTime(d)

		cell := "正常,加班03-06 18:00到03-06 20:30 2.5小时\n(08:01,20:22)"
		c, p, f := parseDailyCell(ctx, cell, 500, date)
		if c != 1 || p != 0 || f != 0 {
			t.Fatalf("created=%d pending=%d fail=%d, want 1/0/0", c, p, f)
		}

		var daily model.AttendanceDaily
		if err := db.Where("person_id = 500 AND event_date = ?", date).First(&daily).Error; err != nil {
			t.Fatalf("daily not written: %v", err)
		}
		var details []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Find(&details)
		if len(details) != 2 {
			t.Fatalf("details count = %d, want 2（出勤+加班）", len(details))
		}
		foundAtt, foundOT := false, false
		for _, e := range details {
			if e.EventType == "出勤" && e.SubType == "普通出勤" && e.Hours == 8 {
				foundAtt = true
			}
			if e.EventType == "加班" && e.SubType == "工作日加班" && e.Hours == 2.5 {
				foundOT = true
			}
		}
		if !foundAtt {
			t.Errorf("missing 出勤-普通出勤 8h, got %+v", details)
		}
		if !foundOT {
			t.Errorf("missing 加班-工作日加班 2.5h, got %+v", details)
		}
		// 投影：出勤 8h + 工作日加班 2.5h
		var proj model.AttendanceDailyProjection
		if err := db.Where("person_id = 500 AND work_date = ?", date).First(&proj).Error; err != nil {
			t.Fatalf("load projection: %v", err)
		}
		if proj.WorkHours != 8 || proj.OvertimeWorkdayHours != 2.5 {
			t.Errorf("projection work_hours=%v ot_workday=%v, want 8/2.5", proj.WorkHours, proj.OvertimeWorkdayHours)
		}
	})
}

// TestDingTalkRestPunchOTIncludesHours 休息并打卡带加班时长：按实际加班小时记录并标待确认
// （如万丹 2026-04-19 加班 12.5h，此前被硬编码为 8h）
func TestDingTalkRestPunchOTIncludesHours(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-04-19")
		date := utils.DateOnlyFromTime(d)

		cell := "休息并打卡,加班04-19 08:30到04-19 22:30 12.5小时\n(08:14,22:37)"
		c, p, f := parseDailyCell(ctx, cell, 600, date)
		if c != 1 || p != 1 || f != 0 {
			t.Fatalf("created=%d pending=%d fail=%d, want 1/1/0", c, p, f)
		}

		var daily model.AttendanceDaily
		if err := db.Where("person_id = 600 AND event_date = ?", date).First(&daily).Error; err != nil {
			t.Fatalf("daily not written: %v", err)
		}
		if daily.Status != "pending" {
			t.Errorf("daily status = %s, want pending", daily.Status)
		}
		var details []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Find(&details)
		if len(details) != 1 {
			t.Fatalf("details count = %d, want 1", len(details))
		}
		if details[0].EventType != "加班" || details[0].SubType != "节假日加班" || details[0].Hours != 12.5 {
			t.Errorf("detail = %+v, want 加班-节假日加班 12.5h", details[0])
		}
	})
}

// TestDingTalkParseMultiLeave 同类型多段请假求和 + 休息+请假不再被跳过
func TestDingTalkParseMultiLeave(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-05-18")
		date := utils.DateOnlyFromTime(d)

		cell := "病假05-18 08:30到05-18 12:00 3.5小时,病假05-18 13:30到05-18 18:00 4.5小时"
		c, _, _ := parseDailyCell(ctx, cell, 200, date)
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

		// 跨天病假（整段描述重复于每个日单元格，无法单日拆分）：忠实透传备注 + pending，不解析时长
		cell2 := "休息,病假05-22 08:30到05-29 18:30 56小时"
		c2, p2, f2 := parseDailyCell(ctx, cell2, 201, date)
		if c2 != 1 || p2 != 1 || f2 != 0 {
			t.Fatalf("cross-day leave: created=%d pending=%d fail=%d, want 1/1/0", c2, p2, f2)
		}
		var daily2 model.AttendanceDaily
		if err := db.Where("person_id = 201 AND event_date = ?", date).First(&daily2).Error; err != nil {
			t.Fatalf("daily2 not written: %v", err)
		}
		if daily2.Status != "pending" {
			t.Errorf("cross-day daily status = %s, want pending", daily2.Status)
		}
		if daily2.Remark != "钉钉导入:"+cell2 {
			t.Errorf("cross-day remark should pass through raw cell, got %q", daily2.Remark)
		}
		var details2 []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily2.ID).Find(&details2)
		if len(details2) != 0 {
			t.Errorf("cross-day passthrough should have no details, got %d", len(details2))
		}
	})
}

// TestDingTalkImportIdempotent 重复导入同一日：追加新版本——旧组降级 pending、新组 confirmed，
// 各 1 条明细不重复；年假消费只按最新确认组计一次
func TestDingTalkImportIdempotent(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		seedGrant(t, db, 300, 40)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-06-10")
		date := utils.DateOnlyFromTime(d)

		cell := "正常(08:30,18:30)"
		for i := 0; i < 2; i++ {
			c, _, _ := parseDailyCell(ctx, cell, 300, date)
			if c == 0 {
				t.Fatalf("round %d: created=0", i+1)
			}
		}
		var dailies []model.AttendanceDaily
		if err := db.Where("person_id = 300 AND event_date = ?", date).Order("seq ASC").Find(&dailies).Error; err != nil {
			t.Fatalf("load dailies: %v", err)
		}
		if len(dailies) != 2 {
			t.Fatalf("second import should append a new version, got %d dailies", len(dailies))
		}
		if dailies[0].Seq != 1 || dailies[0].Status != "pending" {
			t.Errorf("old version should be seq=1 pending, got seq=%d status=%s", dailies[0].Seq, dailies[0].Status)
		}
		if dailies[1].Seq != 2 || dailies[1].Status != "confirmed" {
			t.Errorf("newest should be seq=2 confirmed, got seq=%d status=%s", dailies[1].Seq, dailies[1].Status)
		}
		for i, dl := range dailies {
			var details []model.AttendanceEventDetail
			db.Where("daily_id = ?", dl.ID).Find(&details)
			if len(details) != 1 {
				t.Errorf("daily[%d] details count = %d, want 1", i, len(details))
			}
		}
		// 该日无年假消费，余额保持 40（重复导入不产生额外消费）
		assertALBalance(t, db, 300, 40)
	})
}

// TestDingTalkImportFailFallback 单元格写入失败：fail 计数 + 兜底生成 pending 空日记录（覆盖当天）
func TestDingTalkImportFailFallback(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-06-10")
		date := utils.DateOnlyFromTime(d)

		// 注入：考勤明细插入失败（模拟真实 SQL 错误）
		if err := db.Exec("CREATE TRIGGER test_fail_detail BEFORE INSERT ON attendance_event_details BEGIN SELECT RAISE(ABORT, 'injected failure'); END").Error; err != nil {
			t.Fatalf("create trigger: %v", err)
		}
		defer db.Exec("DROP TRIGGER IF EXISTS test_fail_detail")

		cell := "正常(08:30,18:30)"
		c, p, f := parseDailyCell(ctx, cell, 400, date)
		if c != 0 || p != 0 || f != 1 {
			t.Fatalf("expected fail=1 only, got created=%d pending=%d fail=%d", c, p, f)
		}

		// 兜底 pending 空日记录已生成（无事件明细）
		var daily model.AttendanceDaily
		if err := db.Where("person_id = 400 AND event_date = ?", date).First(&daily).Error; err != nil {
			t.Fatalf("fallback pending daily should exist: %v", err)
		}
		if daily.Status != "pending" {
			t.Errorf("fallback daily status = %s, want pending", daily.Status)
		}
		var details []model.AttendanceEventDetail
		db.Where("daily_id = ?", daily.ID).Find(&details)
		if len(details) != 0 {
			t.Errorf("fallback daily should have no details, got %d", len(details))
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

// TestDingTalkCrossDayPassthrough 跨天事件（整段描述重复于每个日单元格）：
// 忠实透传备注 + pending + 无明细，不做单日解析；无日期区间变体走时长护栏
func TestDingTalkCrossDayPassthrough(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateBalanceSnapshots(t, db)
		ctx := context.Background()
		d, _ := utils.ParseDate("2026-05-22")
		date := utils.DateOnlyFromTime(d)

		cases := []string{
			"病假05-22 08:30到05-29 18:30 56小时\n(-)",
			"休息,病假05-22 08:30到05-29 18:30 56小时\n(-)",
			"病假05-22 08:30到05-29 18:30 56小时\n(-,19:40)",
			"正常,加班05-20 22:00到05-21 01:00 5小时",
			"病假56小时", // 无日期区间：时长护栏（56 > 日标准 8h）
		}
		for i, cell := range cases {
			pid := uint(700 + i)
			c, p, f := parseDailyCell(ctx, cell, pid, date)
			if c != 1 || p != 1 || f != 0 {
				t.Fatalf("cell %q: created=%d pending=%d fail=%d, want 1/1/0", cell, c, p, f)
			}
			var daily model.AttendanceDaily
			if err := db.Where("person_id = ? AND event_date = ?", pid, date).First(&daily).Error; err != nil {
				t.Fatalf("cell %q: daily not written: %v", cell, err)
			}
			if daily.Status != "pending" {
				t.Errorf("cell %q: status = %s, want pending", cell, daily.Status)
			}
			if daily.Remark != "钉钉导入:"+cell {
				t.Errorf("cell %q: remark = %q, want raw cell passthrough", cell, daily.Remark)
			}
			var details []model.AttendanceEventDetail
			db.Where("daily_id = ?", daily.ID).Find(&details)
			if len(details) != 0 {
				t.Errorf("cell %q: details = %d, want 0", cell, len(details))
			}
		}
	})
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
		{"休息并打卡,加班04-19 08:30到04-19 22:30 12.5小时", 1, 1},
		{"休息加班3小时", 1, 0},
		{"事假4小时", 1, 1},
		{"事假8小时", 1, 1},
		{"病假8小时", 1, 0},
		{"年假8小时", 1, 0},
		{"年假8小时,迟到10分钟", 1, 0},
		{"年假8小时,迟到20分钟,早退15分钟", 1, 1},
		{"加班2小时", 1, 0},
		{"加班2小时,迟到15分钟,早退20分钟", 1, 1},
		}
		for i, tc := range cases {
			pid := uint(100 + i)
			c, p, _ := parseDailyCell(ctx, tc.cell, pid, date)
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
