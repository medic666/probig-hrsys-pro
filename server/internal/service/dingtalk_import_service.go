package service

import (
	"errors"
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

var reClockTime = regexp.MustCompile(`\((\d{1,2}:\d{2}),(\d{1,2}:\d{2})\)`)

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

func DingTalkExecute(filePath, month string, mappings []DingTalkImportMapping) (created, pending int, err error) {
	_, persons, err := ParseDingTalkExcel(filePath)
	if err != nil {
		return 0, 0, err
	}
	personMap := make(map[string]uint)
	for _, m := range mappings {
		personMap[m.ExcelName] = m.PersonID
	}
	monthStart, _ := time.Parse("2006-01", month)
	monthDays := int(monthStart.AddDate(0, 1, -1).Day())

	for _, pr := range persons {
		pid, ok := personMap[pr.Name]
		if !ok {
			continue
		}
		for dayIdx := 0; dayIdx < monthDays && dayIdx < len(pr.DailyCells); dayIdx++ {
			cell := pr.DailyCells[dayIdx]
			cell = strings.TrimSpace(cell)
			if cell == "" {
				continue
			}
			date := monthStart.AddDate(0, 0, dayIdx)
			dateOnly := utils.DateOnlyFromTime(date)
			c, p := parseDailyCell(cell, pid, dateOnly)
			created += c
			pending += p
		}
	}
	return created, pending, nil
}

func parseDailyCell(cell string, personID uint, date utils.DateOnly) (created, pending int) {
	status := "confirmed"
	var events []model.AttendanceEventDetail

	isRest := strings.HasPrefix(cell, "休息") && !strings.Contains(cell, "加班") && !strings.Contains(cell, "打卡") && !strings.Contains(cell, "外勤")
	if isRest {
		return 0, 0
	}

	hasLeave := containsAny(cell, []string{"事假", "年假", "病假", "调休"})
	hasOT := containsAny(cell, []string{"加班"})
	isRestOT := strings.Contains(cell, "休息并打卡")
	isRestWithOT := strings.Contains(cell, "休息") && hasOT && !isRestOT
	isAbsent := strings.Contains(cell, "旷工")
	isFieldWork := containsAny(cell, []string{"外勤"})
	isLate := strings.Contains(cell, "迟到")
	isEarly := strings.Contains(cell, "早退")
	isMissingCard := strings.Contains(cell, "缺卡")

	punchTime := ""
	if m := reClockTime.FindStringSubmatch(cell); len(m) >= 3 {
		punchTime = m[1] + "," + m[2]
	}

	createEvents := func() (int, int) {
		if len(events) == 0 {
			return 0, 0
		}
		if punchTime != "" {
			events = append(events, model.AttendanceEventDetail{
				EventType: "打卡时间戳", SubType: "", Hours: 0, Remark: punchTime,
			})
		}
		err := utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
			daily, err := GetOrCreateDaily(tx, personID, date, status)
			if err != nil {
				return err
			}
			for _, e := range events {
				if err := CreateDetail(tx, daily.ID, e.EventType, e.SubType, e.Hours, e.Minutes, e.Remark); err != nil {
					return err
				}
			}
			if punchTime != "" {
				tx.Model(daily).Update("punch_time", punchTime)
			}
			if status == "confirmed" {
				return RebuildDailyProjection(tx, personID, date)
			}
			return nil
		})
		if err != nil {
			return 0, 0
		}
		return created, pending
	}

	if isAbsent {
		events = append(events, model.AttendanceEventDetail{
			EventType: "休假", SubType: "事假", Hours: 8, Remark: "钉钉导入:旷工",
		})
		status = "pending"; pending = 1
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
		events = append(events, model.AttendanceEventDetail{
			EventType: "加班", SubType: "节假日加班", Hours: 8, Remark: "钉钉导入:休息并打卡",
		})
		status = "pending"; pending = 1
		return createEvents()
	}

	if hasLeave {
		leaveType, leaveHours := parseLeaveFromCell(cell)
		if leaveHours > 0 {
			events = append(events, model.AttendanceEventDetail{
				EventType: "休假", SubType: leaveType, Hours: leaveHours, Remark: "钉钉导入:"+cell,
			})
			if leaveHours < 8 {
				events = append(events, model.AttendanceEventDetail{
					EventType: "出勤", SubType: "普通出勤", Hours: 8 - leaveHours,
				})
			}
			status = "pending"; pending = 1
			if !isLate && !isEarly && !isMissingCard {
				return createEvents()
			}
		}
	}

	if hasOT && !isRestOT && !isRestWithOT {
		h := extractOTHours(cell)
		if h == 0 { h = 8 }
		events = append(events, model.AttendanceEventDetail{
			EventType: "加班", SubType: "工作日加班", Hours: h, Remark: "钉钉导入:加班",
		})
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
	re := regexp.MustCompile(`加班[^\d]*(\d+\.?\d*)\s*小时`)
	m := re.FindStringSubmatch(cell)
	if len(m) >= 2 {
		h, _ := strconv.ParseFloat(m[1], 64)
		return h
	}
	return 0
}

func parseLeaveFromCell(cell string) (string, float64) {
	leaveTypes := []string{"事假", "年假", "病假", "调休"}
	for _, lt := range leaveTypes {
		re := regexp.MustCompile(lt + `[^\d]*(\d+\.?\d*)\s*小时`)
		m := re.FindStringSubmatch(cell)
		if len(m) >= 2 {
			h, _ := strconv.ParseFloat(m[1], 64)
			return lt, h
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
