package handler

import (
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
	utils.Success(c, utils.NewPageResult(list, total, utils.PageRequest{PageNum: q.PageNum, PageSize: q.PageSize}))
}

type createDailyReq struct {
	PersonID  uint                           `json:"person_id"`
	EventDate string                         `json:"event_date"`
	PunchTime string                         `json:"punch_time"`
	Remark    string                         `json:"remark"`
	Details   []model.AttendanceEventDetail  `json:"details"`
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
	d, _ := utils.ParseDate(req.EventDate)
	dateOnly := utils.DateOnlyFromTime(d)

	err := utils.WithTransaction(dao.DBFromContext(c.Request.Context()), func(tx *gorm.DB) error {
		daily, err := service.GetOrCreateDaily(tx, req.PersonID, dateOnly, "confirmed")
		if err != nil {
			return err
		}
		if req.PunchTime != "" {
			if err := tx.Model(daily).Updates(map[string]interface{}{"punch_time": req.PunchTime, "remark": req.Remark}).Error; err != nil {
				return err
			}
		}
		for _, dt := range req.Details {
			if err := service.CreateDetail(tx, daily.ID, dt.EventType, dt.SubType, dt.Hours, dt.Minutes, dt.Remark); err != nil {
				return err
			}
		}
		return service.RebuildProjectionsAfterAttendanceChange(tx, req.PersonID, dateOnly, req.Details)
	})
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", nil)
}

type updateDailyReq struct {
	PunchTime string                         `json:"punch_time"`
	Remark    string                         `json:"remark"`
	Details   []model.AttendanceEventDetail  `json:"details"`
}

func UpdateAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req updateDailyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	daily, err := service.GetAttendanceDailyByID(uint(id))
	if err != nil {
		utils.Error(c, "记录不存在")
		return
	}
	err = utils.WithTransaction(dao.DBFromContext(c.Request.Context()), func(tx *gorm.DB) error {
		oldDetails, err := service.GetDetailsByDailyID(tx, uint(id))
		if err != nil {
			return err
		}
		if err := tx.Where("daily_id = ?", uint(id)).Delete(&model.AttendanceEventDetail{}).Error; err != nil {
			return err
		}
		for _, dt := range req.Details {
			dt.ID = 0
			dt.DailyID = uint(id)
			if err := tx.Create(&dt).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&daily).Updates(map[string]interface{}{"punch_time": req.PunchTime, "remark": req.Remark}).Error; err != nil {
			return err
		}
		involved := append(oldDetails, req.Details...)
		return service.RebuildProjectionsAfterAttendanceChange(tx, daily.PersonID, daily.EventDate, involved)
	})
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

// SaveAttendanceDetails 暂存保存当日事件明细（保持原状态不变），供每日方块"编辑-保存"使用
func SaveAttendanceDetails(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req updateDailyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	daily, err := service.GetAttendanceDailyByID(uint(id))
	if err != nil {
		utils.Error(c, "记录不存在")
		return
	}
	err = utils.WithTransaction(dao.DBFromContext(c.Request.Context()), func(tx *gorm.DB) error {
		if err := service.SaveDailyDetailsKeepStatus(tx, uint(id), req.Details); err != nil {
			return err
		}
		if err := tx.Model(&daily).Updates(map[string]interface{}{"punch_time": req.PunchTime, "remark": req.Remark}).Error; err != nil {
			return err
		}
		return service.RebuildProjectionsAfterAttendanceChange(tx, daily.PersonID, daily.EventDate, req.Details)
	})
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "已保存", nil)
}

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
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func CreateBatchAttendanceEvents(c *gin.Context) {
	var req service.BatchAttendanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
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
	utils.Success(c, utils.NewPageResult(list, total, utils.PageRequest{PageNum: q.PageNum, PageSize: q.PageSize}))
}

type confirmDailyReq struct {
	Details   []model.AttendanceEventDetail `json:"details"`
	PunchTime string                        `json:"punch_time"`
	Remark    string                        `json:"remark"`
}

func ConfirmPendingDaily(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req confirmDailyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	err := utils.WithTransaction(dao.DBFromContext(c.Request.Context()), func(tx *gorm.DB) error {
		// 确认时同步应用打卡时间/备注（编辑弹窗"确定"一次提交）
		if err := tx.Model(&model.AttendanceDaily{}).Where("id = ?", id).
			Updates(map[string]interface{}{"punch_time": req.PunchTime, "remark": req.Remark}).Error; err != nil {
			return err
		}
		return service.ConfirmDaily(c.Request.Context(), tx, uint(id), req.Details)
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
	created, pending, err := service.DingTalkExecute(c.Request.Context(), req.FilePath, req.Month, req.Mappings)
	if err != nil {
		utils.Error(c, "导入失败:"+err.Error())
		return
	}
	// 执行成功后清理本次导入临时文件
	os.Remove(req.FilePath)
	utils.Success(c, gin.H{"created": created, "pending": pending})
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
	utils.Success(c, utils.NewPageResult(list, total, utils.PageRequest{PageNum: q.PageNum, PageSize: q.PageSize}))
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
	if len(list) > 0 && list[0]["details"] != nil {
		utils.Success(c, list[0]["details"])
	} else {
		utils.Success(c, []interface{}{})
	}
}
