package handler

import (
	"strconv"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetAttendanceEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetAttendanceEventList(pageReq.PageNum, pageReq.PageSize,
		uint(personID), c.Query("date_start"), c.Query("date_end"), c.Query("event_type"), c.Query("sub_type"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func GetAttendanceEventByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	event, err := service.GetAttendanceEvent(uint(id))
	if err != nil {
		utils.Error(c, "考勤事件不存在")
		return
	}
	utils.Success(c, event)
}

func CreateAttendanceEvent(c *gin.Context) {
	var e model.AttendanceEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if e.PersonID == 0 || e.EventType == "" {
		utils.BadRequest(c, "人员、事件类型为必填项")
		return
	}
	if e.EventType != "打卡时间戳" && e.SubType == "" {
		utils.BadRequest(c, "子类型为必填项")
		return
	}
	if !isValidSubType(e.EventType, e.SubType) {
		utils.BadRequest(c, "事件类型与子类型不匹配")
		return
	}
	if err := service.CreateAttendanceEvent(&e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": e.ID})
}

func UpdateAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e model.AttendanceEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if e.EventType != "" && e.SubType != "" && !isValidSubType(e.EventType, e.SubType) {
		utils.BadRequest(c, "事件类型与子类型不匹配")
		return
	}
	if err := service.UpdateAttendanceEvent(uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAttendanceEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreAttendanceEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedAttendanceEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedAttendanceEvents(pageReq.PageNum, pageReq.PageSize)
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
	success, fail, err := service.CreateBatchAttendanceEvents(req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "批量创建完成", gin.H{"success": success, "fail": fail})
}

func GetAttendanceEventsByDate(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Param("personId"), 10, 64)
	date := c.Param("date")
	events, err := service.GetAttendanceEventsByPersonDate(uint(personID), date)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, events)
}

var validSubTypes = map[string][]string{
	"出勤":   {"普通出勤", "补班出勤", "外勤出勤"},
	"休假":   {"调休", "事假", "病假", "年假", "法定假", "福利假"},
	"加班":   {"工作日加班", "节假日加班"},
	"违纪":   {"缺卡", "迟到", "早退"},
	"打卡时间戳": {},
}

func isValidSubType(eventType, subType string) bool {
	if eventType == "打卡时间戳" {
		return true
	}
	list := validSubTypes[eventType]
	for _, v := range list {
		if v == subType {
			return true
		}
	}
	return false
}

func ExportAttendanceEvents(c *gin.Context) {
	list, _, _ := service.GetAttendanceEventList(1, 10000, 0, c.Query("date_start"), c.Query("date_end"), c.Query("event_type"), c.Query("sub_type"))
	f := excelize.NewFile()
	defer f.Close()
	sheet := "考勤事件"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"人员", "日期", "事件类型", "子类型", "时长", "打卡时间", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, e := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), e["person_name"])
		f.SetCellValue(sheet, cellName(2, row), e["event_date"])
		f.SetCellValue(sheet, cellName(3, row), e["event_type"])
		f.SetCellValue(sheet, cellName(4, row), e["sub_type"])
		f.SetCellValue(sheet, cellName(5, row), e["hours"])
		f.SetCellValue(sheet, cellName(6, row), e["punch_time"])
		f.SetCellValue(sheet, cellName(7, row), e["remark"])
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=attendance_events_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
}
