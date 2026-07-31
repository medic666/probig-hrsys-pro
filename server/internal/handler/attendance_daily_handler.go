package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"probig/server/internal/config"
	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func GetAttendanceEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetAttendanceDailyList(uint(personID),
		c.Query("date_start"), c.Query("date_end"), c.Query("status"), pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
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

	err := utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		daily, err := service.GetOrCreateDaily(tx, req.PersonID, dateOnly, "confirmed")
		if err != nil {
			return err
		}
		if req.PunchTime != "" {
			tx.Model(daily).Updates(map[string]interface{}{"punch_time": req.PunchTime, "remark": req.Remark})
		}
		for _, dt := range req.Details {
			if err := service.CreateDetail(tx, daily.ID, dt.EventType, dt.SubType, dt.Hours, dt.Minutes, dt.Remark); err != nil {
				return err
			}
		}
		return service.RebuildDailyProjection(tx, req.PersonID, dateOnly)
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
	var daily model.AttendanceDaily
	if err := dao.DB.First(&daily, uint(id)).Error; err != nil {
		utils.Error(c, "记录不存在")
		return
	}
	err := utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
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
		tx.Model(&daily).Updates(map[string]interface{}{"punch_time": req.PunchTime, "remark": req.Remark})
		return service.RebuildDailyProjection(tx, daily.PersonID, daily.EventDate)
	})
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAttendanceDaily(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreAttendanceDaily(uint(id)); err != nil {
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
	success, fail, err := service.CreateBatchAttendanceDailies(req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "批量创建完成", gin.H{"success": success, "fail": fail})
}

func GetPendingDailyList(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetPendingDailyList(pageReq.PageNum, pageReq.PageSize, uint(personID))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

type confirmDailyReq struct {
	Details []model.AttendanceEventDetail `json:"details"`
}

func ConfirmPendingDaily(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req confirmDailyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	err := utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		return service.ConfirmDaily(tx, uint(id), req.Details)
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
	savePath := filepath.Join(tmpDir, fmt.Sprintf("dingtalk_%d.xlsx", time.Now().UnixNano()))
	out, _ := os.Create(savePath)
	io.Copy(out, file)
	out.Close()
	defer os.Remove(savePath)

	result, err := service.DingTalkPreview(savePath)
	if err != nil {
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
	created, pending, err := service.DingTalkExecute(req.FilePath, req.Month, req.Mappings)
	if err != nil {
		utils.Error(c, "导入失败:"+err.Error())
		return
	}
	utils.Success(c, gin.H{"created": created, "pending": pending})
}

func ExportAttendanceEvents(c *gin.Context) {
	list, _, _ := service.GetAttendanceDailyList(0, "", "", "", 1, 10000)
	f := excelize.NewFile()
	defer f.Close()
	sheet := "考勤事件"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"人员", "日期", "状态", "打卡时间", "事件摘要"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, e := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), e["person_name"])
		f.SetCellValue(sheet, cellName(2, row), e["event_date"])
		f.SetCellValue(sheet, cellName(3, row), e["status"])
		f.SetCellValue(sheet, cellName(4, row), e["punch_time"])
		summary := ""
		if details, ok := e["details"].([]map[string]interface{}); ok {
			for _, d := range details {
				summary += fmt.Sprintf("%s-%s(%.1fh);", d["event_type"], d["sub_type"], d["hours"])
			}
		}
		f.SetCellValue(sheet, cellName(5, row), summary)
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=attendance_events_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
}

func GetDailyProjections(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetDailyProjections(uint(personID), c.Query("date_start"), c.Query("date_end"), pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func GetEventsByPersonDate(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Param("personId"), 10, 64)
	date := c.Param("date")
	list, _, _ := service.GetAttendanceDailyList(uint(personID), date, date, "", 1, 100)
	if len(list) > 0 && list[0]["details"] != nil {
		utils.Success(c, list[0]["details"])
	} else {
		utils.Success(c, []interface{}{})
	}
}
