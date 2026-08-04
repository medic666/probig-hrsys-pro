package handler

import (
	"errors"
	"time"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"probig/server/internal/config"
	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// attendanceDailyListQuery 考勤日记录列表筛选解析（列表/待确认/导出共用）
func attendanceDailyListQuery(c *gin.Context) service.AttendanceDailyListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	return service.AttendanceDailyListQuery{
		PageNum:   pageReq.PageNum,
		PageSize:  pageReq.PageSize,
		PersonID:  uint(personID),
		DateStart: c.Query("date_start"),
		DateEnd:   c.Query("date_end"),
		Status:    c.Query("status"),
	}
}

func GetAttendanceEvents(c *gin.Context) {
	q := attendanceDailyListQuery(c)
	list, total, err := service.GetAttendanceDailyList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, q.PageNum, q.PageSize)
}

type createDailyReq struct {
	PersonID  uint                          `json:"person_id"`
	EventDate string                        `json:"event_date"`
	Status    string                        `json:"status"`
	PunchTime string                        `json:"punch_time"`
	Remark    string                        `json:"remark"`
	Details   []model.AttendanceEventDetail `json:"details"`
}

// normalizeDailyStatus 状态归一：空值默认已确认（向后兼容，原子卡片一键确认不传状态），
// 非法值报错。新增/编辑/批量三入口共用，保证状态取值唯一来源。
func normalizeDailyStatus(s string) (string, error) {
	if s == "" {
		return "confirmed", nil
	}
	if s != "pending" && s != "confirmed" {
		return "", errors.New("状态取值仅支持 pending/confirmed")
	}
	return s, nil
}

func CreateAttendanceEvent(c *gin.Context) {
	var req createDailyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误:"+err.Error())
		return
	}
	if req.PersonID == 0 || req.EventDate == "" {
		utils.BadRequest(c, "人员和日期为必填项")
		return
	}
	status, err := normalizeDailyStatus(req.Status)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	d, _ := utils.ParseDate(req.EventDate)
	dateOnly := utils.DateOnlyFromTime(d)

	err = utils.WithTransaction(dao.DBFromContext(c.Request.Context()), func(tx *gorm.DB) error {
		// 颗粒化 upsert（提供即覆盖）：明细/打卡时间/备注/状态按本次提交值写入，当天已有记录被替换
		return service.UpsertAttendanceDaily(tx, service.AttendanceDailyUpsert{
			PersonID:  req.PersonID,
			Date:      dateOnly,
			Status:    &status,
			PunchTime: &req.PunchTime,
			Remark:    &req.Remark,
			Details:   req.Details,
		})
	})
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "录入成功", nil)
}

// GetAttendanceEventByID 考勤日记录完整详情（页面化"编辑=查看"取数）
func GetAttendanceEventByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	daily, err := service.GetAttendanceDailyByID(uint(id))
	if err != nil {
		utils.Error(c, "记录不存在")
		return
	}
	details, err := service.GetDetailsByDailyID(dao.DB, daily.ID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"id": daily.ID, "person_id": daily.PersonID, "person_name": service.PersonName(daily.PersonID),
		"event_date": daily.EventDate, "status": daily.Status,
		"punch_time": daily.PunchTime, "remark": daily.Remark, "details": details,
	})
}

// DeleteAttendanceEvent 删除考勤日记录（软删除 + 投影重建）
func DeleteAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAttendanceDaily(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreAttendanceDaily(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedAttendanceEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedAttendanceDailies(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
}

func CreateBatchAttendanceEvents(c *gin.Context) {
	var req service.BatchAttendanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	status, err := normalizeDailyStatus(req.Status)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	req.Status = status
	success, fail, err := service.CreateBatchAttendanceDailies(c.Request.Context(), req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "批量创建完成", gin.H{"success": success, "fail": fail})
}

func GetPendingDailyList(c *gin.Context) {
	q := attendanceDailyListQuery(c)
	list, total, err := service.GetPendingDailyList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, q.PageNum, q.PageSize)
}

type confirmDailyReq struct {
	Details   []model.AttendanceEventDetail `json:"details"`
	PunchTime string                        `json:"punch_time"`
	Remark    string                        `json:"remark"`
	Status    string                        `json:"status"`
}

func ConfirmAttendanceDaily(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req confirmDailyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	status, err := normalizeDailyStatus(req.Status)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	err = utils.WithTransaction(dao.DBFromContext(c.Request.Context()), func(tx *gorm.DB) error {
		// 确认时同步应用打卡时间/备注（编辑弹窗"确定"一次提交）
		if err := tx.Model(&model.AttendanceDaily{}).Where("id = ?", id).
			Updates(map[string]interface{}{"punch_time": req.PunchTime, "remark": req.Remark}).Error; err != nil {
			return err
		}
		return service.ConfirmDaily(c.Request.Context(), tx, uint(id), req.Details, status)
	})
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "确认成功", nil)
}

// ---- DingTalk import ----

func DingTalkPreview(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请上传文件")
		return
	}
	defer file.Close()
	tmpDir := config.ResolvePath(config.AppConfig.FileStorage.Path)
	os.MkdirAll(tmpDir, 0755)
	// 懒清理：仅删除超过 24h 的陈旧导入临时文件，保留当前预览文件供执行步骤使用
	service.CleanupStaleDingTalkFiles(tmpDir, 24*time.Hour)
	savePath := filepath.Join(tmpDir, fmt.Sprintf("dingtalk_%d.xlsx", time.Now().UnixNano()))
	out, _ := os.Create(savePath)
	io.Copy(out, file)
	out.Close()

	result, err := service.DingTalkPreview(savePath)
	if err != nil {
		os.Remove(savePath)
		utils.Error(c, "解析失败:"+err.Error())
		return
	}
	utils.Success(c, gin.H{"preview": result, "file_path": savePath, "file_name": header.Filename})
}

type dingTalkExecuteReq struct {
	Month    string                           `json:"month" binding:"required"`
	Mappings []service.DingTalkImportMapping  `json:"mappings" binding:"required"`
	FilePath string                           `json:"file_path" binding:"required"`
}

func DingTalkExecute(c *gin.Context) {
	var req dingTalkExecuteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if !isDingTalkTempPath(req.FilePath) {
		utils.BadRequest(c, "文件路径不合法")
		return
	}
	if _, err := os.Stat(req.FilePath); err != nil {
		utils.Error(c, "上传文件已过期，请重新上传")
		return
	}
	created, pending, fail, err := service.DingTalkExecute(c.Request.Context(), req.FilePath, req.Month, req.Mappings)
	if err != nil {
		utils.Error(c, "导入失败:"+err.Error())
		return
	}
	// 执行成功后清理本次导入临时文件
	os.Remove(req.FilePath)
	utils.Success(c, gin.H{"created": created, "pending": pending, "fail": fail})
}

// isDingTalkTempPath 校验导入文件位于上传目录内且为 dingtalk_*.xlsx 临时文件
func isDingTalkTempPath(p string) bool {
	tmpDir := config.ResolvePath(config.AppConfig.FileStorage.Path)
	abs, err := filepath.Abs(p)
	if err != nil || !strings.HasPrefix(abs, filepath.Clean(tmpDir)+string(filepath.Separator)) {
		return false
	}
	name := filepath.Base(abs)
	return strings.HasPrefix(name, "dingtalk_") && strings.HasSuffix(name, ".xlsx")
}

// attendanceDailyExportFilters 考勤事件导出文件名筛选摘要
func attendanceDailyExportFilters(q service.AttendanceDailyListQuery) []string {
	var parts []string
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	if q.Status != "" {
		parts = append(parts, "状态="+map[string]string{"pending": "待确认", "confirmed": "已确认"}[q.Status])
	}
	if p := dateRangePiece("日期", q.DateStart, q.DateEnd); p != "" {
		parts = append(parts, p)
	}
	return parts
}

func ExportAttendanceEvents(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := attendanceDailyListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, _ := service.GetAttendanceDailyList(q)

	var rows [][]interface{}
	for _, e := range list {
		summary := ""
		if details, ok := e["details"].([]map[string]interface{}); ok {
			for _, d := range details {
				summary += fmt.Sprintf("%s-%s(%.1fh);", d["event_type"], d["sub_type"], d["hours"])
			}
		}
		rows = append(rows, []interface{}{
			e["person_name"], e["event_date"], e["status"], e["punch_time"], summary,
		})
	}
	writeExcel(c, "考勤事件",
		[]string{"人员", "日期", "状态", "打卡时间", "事件摘要"}, rows,
		attendanceDailyExportFilters(q)...)
}

// dailyProjectionListQuery 日记工时投影列表筛选解析（列表与导出共用）
func dailyProjectionListQuery(c *gin.Context) service.DailyProjectionListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	return service.DailyProjectionListQuery{
		PageNum:   pageReq.PageNum,
		PageSize:  pageReq.PageSize,
		PersonID:  uint(personID),
		DateStart: c.Query("date_start"),
		DateEnd:   c.Query("date_end"),
	}
}

func GetDailyProjections(c *gin.Context) {
	q := dailyProjectionListQuery(c)
	list, total, err := service.GetDailyProjections(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, q.PageNum, q.PageSize)
}

// dailyProjectionExportFilters 日记工时导出文件名筛选摘要
func dailyProjectionExportFilters(q service.DailyProjectionListQuery) []string {
	var parts []string
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	if p := dateRangePiece("日期", q.DateStart, q.DateEnd); p != "" {
		parts = append(parts, p)
	}
	return parts
}

// ExportDailyProjections 日记工时导出（严格关联列表视图的当前筛选）
func ExportDailyProjections(c *gin.Context) {
	q := dailyProjectionListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, err := service.GetDailyProjections(q)
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	var rows [][]interface{}
	for _, p := range list {
		rows = append(rows, []interface{}{
			p["person_name"], p["work_date"], p["punch_time"], p["work_hours"],
			p["overtime_workday_hours"], p["overtime_holiday_hours"],
			p["violation_count"], exportBool(p["has_personal_leave"].(bool)),
			map[string]string{"pending": "待确认", "confirmed": "已确认"}[p["status"].(string)],
			p["remark"],
		})
	}
	writeExcel(c, "日记工时",
		[]string{"人员", "日期", "打卡时间", "记出勤工时", "工作日加班", "节假日加班", "违纪次数", "有事假", "状态", "备注"}, rows,
		dailyProjectionExportFilters(q)...)
}

func GetEventsByPersonDate(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Param("personId"), 10, 64)
	date := c.Param("date")
	list, _, _ := service.GetAttendanceDailyList(service.AttendanceDailyListQuery{
		PageNum: 1, PageSize: 100, PersonID: uint(personID), DateStart: date, DateEnd: date,
	})
	if len(list) > 0 {
		utils.Success(c, gin.H{
			"daily_id": list[0]["id"],
			"details":  list[0]["details"],
		})
		return
	}
	utils.Success(c, gin.H{"daily_id": 0, "details": []interface{}{}})
}
