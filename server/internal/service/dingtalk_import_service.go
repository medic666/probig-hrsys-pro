package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var reClockTime = regexp.MustCompile(`\(([^()]*)\)`)

// rePunchTime 打卡时间括号内容里的时间点/缺卡标记，如 08:30 / 09:15,- / -,18:07
var rePunchTime = regexp.MustCompile(`\d{1,2}:\d{2}|-`)

var reDingTalkTemp = regexp.MustCompile(`^dingtalk_\d+\.xlsx$`)

// extractPunchTime 提取打卡时间：取首个括号内容；内容全为 "-" 或空则视为无打卡记录
func extractPunchTime(cell string) string {
	m := reClockTime.FindStringSubmatch(cell)
	if len(m) < 2 {
		return ""
	}
	content := strings.TrimSpace(m[1])
	if content == "" || content == "-" {
		return ""
	}
	has := false
	for _, p := range rePunchTime.FindAllString(content, -1) {
		if p != "-" {
			has = true
			break
		}
	}
	if !has {
		return ""
	}
	return content
}

// CleanupStaleDingTalkFiles 清理 uploads 目录中超过 maxAge 的钉钉导入临时文件
func CleanupStaleDingTalkFiles(uploadDir string, maxAge time.Duration) {
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return
	}
	deadline := time.Now().Add(-maxAge)
	for _, e := range entries {
		if e.IsDir() || !reDingTalkTemp.MatchString(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().Before(deadline) {
			os.Remove(filepath.Join(uploadDir, e.Name()))
		}
	}
}

type DingTalkPersonRow struct {
	Name         string
	LeaveColumns map[int]string
	DailyCells   []string
}

type DingTalkPreviewResult struct {
	ExcelName   string `json:"excel_name"`
	MatchedName string `json:"matched_name"`
	MatchedID   uint   `json:"matched_id"`
	Confidence  string `json:"confidence"`
	DimPerson   bool   `json:"dim_person"`
}

type DingTalkImportMapping struct {
	ExcelName string `json:"excel_name"`
	PersonID  uint   `json:"person_id"`
}

func ParseDingTalkExcel(filePath string) (sheetDate string, persons []DingTalkPersonRow, err error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	rows, err := f.GetRows("月度汇总")
	if err != nil {
		return "", nil, err
	}
	if len(rows) < 5 {
		return "", nil, errors.New("excel 格式错误: 不足5行")
	}
	r0 := rows[0][0]
	sheetDate = strings.TrimPrefix(r0, "月度汇总 统计日期：")
	sheetDate = strings.TrimSuffix(sheetDate, strings.TrimPrefix(sheetDate, strings.SplitN(sheetDate, " 至 ", 2)[0]+" 至 "))
	for i := strings.Index(sheetDate, "20"); i >= 0; i = strings.Index(sheetDate, "20") {
		sheetDate = sheetDate[i:]
		break
	}

	for i := 4; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || row[0] == "" {
			continue
		}
		pr := DingTalkPersonRow{Name: row[0], LeaveColumns: make(map[int]string)}
		for j := 21; j <= 24 && j < len(row); j++ {
			if row[j] != "" {
				pr.LeaveColumns[j] = row[j]
			}
		}
		for j := 37; j < len(row); j++ {
			pr.DailyCells = append(pr.DailyCells, row[j])
		}
		persons = append(persons, pr)
	}
	return sheetDate, persons, nil
}

func DingTalkPreview(filePath string) ([]DingTalkPreviewResult, error) {
	_, persons, err := ParseDingTalkExcel(filePath)
	if err != nil {
		return nil, err
	}
	var allNames []model.Person
	dao.DB.Select("id, name").Find(&allNames)
	nameMap := make(map[string]uint)
	for _, p := range allNames {
		nameMap[p.Name] = p.ID
	}
	var result []DingTalkPreviewResult
	for _, pr := range persons {
		preview := DingTalkPreviewResult{ExcelName: pr.Name}
		if id, ok := nameMap[pr.Name]; ok {
			preview.MatchedID = id
			preview.MatchedName = pr.Name
			preview.Confidence = "exact"
		} else {
			for name, id := range nameMap {
				if strings.Contains(name, pr.Name) || strings.Contains(pr.Name, name) {
					preview.MatchedID = id
					preview.MatchedName = name
					preview.Confidence = "fuzzy"
					break
				}
			}
			if preview.Confidence == "" {
				preview.DimPerson = true
			}
		}
		result = append(result, preview)
	}
	return result, nil
}

func DingTalkExecute(ctx context.Context, filePath, month string, mappings []DingTalkImportMapping) (created, pending, fail int, err error) {
	_, persons, err := ParseDingTalkExcel(filePath)
	if err != nil {
		return 0, 0, 0, err
	}
	personMap := make(map[string]uint)
	for _, m := range mappings {
		personMap[m.ExcelName] = m.PersonID
	}
	monthStart, _ := time.Parse("2006-01", month)
	monthDays := int(monthStart.AddDate(0, 1, -1).Day())

	for _, pr := range persons {
		pid, ok := personMap[pr.Name]
		if !ok || pid == 0 {
			continue // 未匹配或留空跳过（部分导入）
		}
		for dayIdx := 0; dayIdx < monthDays && dayIdx < len(pr.DailyCells); dayIdx++ {
			cell := pr.DailyCells[dayIdx]
			cell = strings.TrimSpace(cell)
			if cell == "" {
				continue
			}
			date := monthStart.AddDate(0, 0, dayIdx)
			dateOnly := utils.DateOnlyFromTime(date)
			c, p, f := parseDailyCell(ctx, cell, pid, dateOnly)
			created += c
			pending += p
			fail += f
		}
	}
	return created, pending, fail, nil
}

func parseDailyCell(ctx context.Context, cell string, personID uint, date utils.DateOnly) (created, pending, fail int) {
	status := "confirmed"
	var events []model.AttendanceEventDetail

	hasLeave := containsAny(cell, []string{"事假", "年假", "病假", "调休"})
	hasOT := containsAny(cell, []string{"加班"})
	// 纯休息日：以"休息"开头且不含加班/打卡/外勤/请假，才视为休息日跳过
	isRest := strings.HasPrefix(cell, "休息") && !hasOT && !strings.Contains(cell, "打卡") && !strings.Contains(cell, "外勤") && !hasLeave
	if isRest {
		return 0, 0, 0
	}

	isRestOT := strings.Contains(cell, "休息并打卡")
	isRestWithOT := strings.Contains(cell, "休息") && hasOT && !isRestOT
	isAbsent := strings.Contains(cell, "旷工")
	isFieldWork := containsAny(cell, []string{"外勤"})
	isLate := strings.Contains(cell, "迟到")
	isEarly := strings.Contains(cell, "早退")
	isMissingCard := strings.Contains(cell, "缺卡")

	punchTime := extractPunchTime(cell)

	createEvents := func() (int, int, int) {
		err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
			// 颗粒化 upsert（提供即覆盖，与单条/批量录入同一规则）：
			// 明细/打卡时间/状态按解析结果写入；导入数据不含备注（nil 保持原值）。
			// pending 事件同样进入投影（状态 pending），使日记工时/月度核算/工资
			// 逐层感知待确认状态，形成 L0→L1→L2→L3 完整控制链
			return AppendAttendanceDaily(tx, AttendanceDailyUpsert{
				PersonID:  personID,
				Date:      date,
				Status:    &status,
				PunchTime: &punchTime,
				Remark:    nil,
				Details:   events,
			})
		})
		if err != nil {
			// 失败兜底：覆盖为 pending 空日记录（无事件明细），管理员可在待确认页分辨导入失败项；
			// 覆盖当天已有记录（与导入"覆盖当天"语义一致，强制人工处理）
			pStatus, pPunch, pRemark := "pending", "", ""
			_ = utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
				return AppendAttendanceDaily(tx, AttendanceDailyUpsert{
					PersonID:  personID,
					Date:      date,
					Status:    &pStatus,
					PunchTime: &pPunch,
					Remark:    &pRemark,
					Details:   []model.AttendanceEventDetail{},
				})
			})
			return 0, 0, 1
		}
		return created, pending, 0
	}

	if isAbsent {
		events = append(events, model.AttendanceEventDetail{
			EventType: "休假", SubType: "事假", Hours: 8, Remark: "钉钉导入:旷工",
		})
		status = "pending"; pending = 1; created = 1
		return createEvents()
	}

	if isRestWithOT {
		h := extractOTHours(cell)
		if h == 0 { h = 8 }
		events = append(events, model.AttendanceEventDetail{
			EventType: "加班", SubType: "节假日加班", Hours: h, Remark: "钉钉导入:休息日加班",
		})
		created = 1
		return createEvents()
	}

	if isRestOT {
		// 休息并打卡：提取实际加班时长（与"休息+加班"分支同构），无时长兜底 8h
		h := extractOTHours(cell)
		if h == 0 {
			h = 8
		}
		events = append(events, model.AttendanceEventDetail{
			EventType: "加班", SubType: "节假日加班", Hours: h, Remark: "钉钉导入:休息并打卡",
		})
		status = "pending"; pending = 1; created = 1
		return createEvents()
	}

	if hasLeave {
		leaveType, leaveHours := parseLeaveFromCell(cell)
		if leaveHours > 0 {
			events = append(events, model.AttendanceEventDetail{
				EventType: "休假", SubType: leaveType, Hours: leaveHours, Remark: "钉钉导入:"+cell,
			})
			created = 1
			// 事假一律待确认（无论是否满 8 小时），由管理员核实
			if leaveType == "事假" {
				status = "pending"; pending = 1
			}
			if leaveHours < 8 {
				// 不满 8 小时：补足出勤并标记待确认（企划 4.3.4）
				events = append(events, model.AttendanceEventDetail{
					EventType: "出勤", SubType: "普通出勤", Hours: 8 - leaveHours,
				})
				status = "pending"; pending = 1
			}
			// 请假与违纪组合自包含处理，避免落入下方普通出勤分支重复记 8h
			lateMin := extractLateMinutes(cell, "迟到")
			earlyMin := extractLateMinutes(cell, "早退")
			if isLate && lateMin > 0 {
				events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "迟到", Minutes: lateMin})
			}
			if isEarly && earlyMin > 0 {
				events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "早退", Minutes: earlyMin})
			}
			// 请假分支与普通出勤分支对齐：迟到+早退合计 > 30 分钟 → 待确认
			if lateMin+earlyMin > 30 {
				status = "pending"; pending = 1
			}
			if isMissingCard {
				events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "缺卡", Hours: 0})
				status = "pending"; pending = 1
			}
			return createEvents()
		}
	}

	if hasOT && !isRestOT && !isRestWithOT {
		// 工作日加班：当日正常出勤 8h + 加班工时（样本单元格均为"正常,加班...X小时"形态），
		// 漏记出勤会导致当日记工时缺失正常出勤时间
		events = append(events, model.AttendanceEventDetail{
			EventType: "出勤", SubType: "普通出勤", Hours: 8,
		})
		h := extractOTHours(cell)
		if h == 0 { h = 8 }
		events = append(events, model.AttendanceEventDetail{
			EventType: "加班", SubType: "工作日加班", Hours: h, Remark: "钉钉导入:加班",
		})
		// 与普通出勤分支对齐：迟到/早退/缺卡违纪 + 迟到早退合计 > 30 分钟标待确认
		lateMin := extractLateMinutes(cell, "迟到")
		earlyMin := extractLateMinutes(cell, "早退")
		if isLate && lateMin > 0 {
			events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "迟到", Minutes: lateMin})
		}
		if isEarly && earlyMin > 0 {
			events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "早退", Minutes: earlyMin})
		}
		if lateMin+earlyMin > 30 {
			status = "pending"; pending = 1
		}
		if isMissingCard {
			events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "缺卡", Hours: 0})
			status = "pending"; pending = 1
		}
		created = 1
		return createEvents()
	}

	if isFieldWork {
		events = append(events, model.AttendanceEventDetail{EventType: "出勤", SubType: "外勤出勤", Hours: 8})
	} else {
		events = append(events, model.AttendanceEventDetail{EventType: "出勤", SubType: "普通出勤", Hours: 8})
	}
	created = 1

	lateMin := extractLateMinutes(cell, "迟到")
	earlyMin := extractLateMinutes(cell, "早退")
	if isLate && lateMin > 0 {
		events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "迟到", Minutes: lateMin})
	}
	if isEarly && earlyMin > 0 {
		events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "早退", Minutes: earlyMin})
	}
	if lateMin+earlyMin > 30 {
		status = "pending"; pending = 1
	}
	if isMissingCard {
		events = append(events, model.AttendanceEventDetail{EventType: "违纪", SubType: "缺卡", Hours: 0})
		status = "pending"; pending = 1
	}
	return createEvents()
}

func extractOTHours(cell string) float64 {
	// 支持 "加班04-17 12:00到04-17 13:30 1.5小时" 等带时间段的表述：关键词后跨任意字符到"数字+小时"
	re := regexp.MustCompile(`加班.*?(\d+(?:\.\d+)?)\s*小时`)
	var total float64
	for _, m := range re.FindAllStringSubmatch(cell, -1) {
		if len(m) >= 2 {
			h, _ := strconv.ParseFloat(m[1], 64)
			total += h
		}
	}
	return total
}

// parseLeaveFromCell 提取请假类型与总时长：支持带时间段与同类型多段请假（如 病假3.5小时,病假4.5小时）求和
func parseLeaveFromCell(cell string) (string, float64) {
	leaveTypes := []string{"事假", "年假", "病假", "调休"}
	for _, lt := range leaveTypes {
		re := regexp.MustCompile(lt + `.*?(\d+(?:\.\d+)?)\s*小时`)
		var total float64
		found := false
		for _, m := range re.FindAllStringSubmatch(cell, -1) {
			if len(m) >= 2 {
				h, _ := strconv.ParseFloat(m[1], 64)
				total += h
				found = true
			}
		}
		if found {
			return lt, total
		}
	}
	return "", 0
}

func extractLateMinutes(cell, typ string) int {
	re := regexp.MustCompile(typ + `(\d+)\s*分钟`)
	m := re.FindStringSubmatch(cell)
	if len(m) >= 2 {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
