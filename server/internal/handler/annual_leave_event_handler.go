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

func GetAnnualLeaveEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	dateStart := c.Query("date_start")
	dateEnd := c.Query("date_end")
	eventType := c.Query("event_type")

	list, _, err := service.GetAnnualLeaveEventList(pageReq.PageNum, pageReq.PageSize, uint(personID),
		dateStart, dateEnd, eventType)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if len(list) > 0 {
		for i := range list {
			list[i]["source"] = "account"
		}
	}

	if eventType == "" || eventType == "休假" {
		attList, _, _ := service.GetAttendanceEventList(pageReq.PageNum, pageReq.PageSize, uint(personID), dateStart, dateEnd, "休假", "年假")
		for i := range attList {
			attList[i]["event_type"] = "休假(年假)"
			attList[i]["source_type"] = "attendance"
			attList[i]["source"] = "attendance"
			attList[i]["hours"] = -(attList[i]["hours"].(float64))
			list = append(list, attList[i])
		}
	}

	utils.Success(c, utils.NewPageResult(list, int64(len(list)), pageReq))
}

func CreateAnnualLeaveEvent(c *gin.Context) {
	var e model.AnnualLeaveAccountEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if e.PersonID == 0 || e.EventType == "" {
		utils.BadRequest(c, "人员和事件类型为必填项")
		return
	}
	e.SourceType = "manual"
	if err := service.CreateAnnualLeaveEvent(&e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": e.ID})
}

func UpdateAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e model.AnnualLeaveAccountEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdateAnnualLeaveEvent(uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAnnualLeaveEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreAnnualLeaveEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedAnnualLeaveEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, _, err := service.GetDeletedAnnualLeaveEvents(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	var filtered []model.AnnualLeaveAccountEvent
	for _, e := range list {
		if e.SourceType != "system_period" {
			filtered = append(filtered, e)
		}
	}
	utils.Success(c, utils.NewPageResult(filtered, int64(len(filtered)), pageReq))
}

func ExportAnnualLeaveEvents(c *gin.Context) {
	list, _, _ := service.GetAnnualLeaveEventList(1, 10000, 0, c.Query("date_start"), c.Query("date_end"), c.Query("event_type"))
	f := excelize.NewFile()
	defer f.Close()
	sheet := "年假事件"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"人员", "事件类型", "来源", "变动时长", "生效日期", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, e := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), e["person_name"])
		f.SetCellValue(sheet, cellName(2, row), e["event_type"])
		f.SetCellValue(sheet, cellName(3, row), e["source_type"])
		f.SetCellValue(sheet, cellName(4, row), e["hours"])
		f.SetCellValue(sheet, cellName(5, row), e["effective_date"])
		f.SetCellValue(sheet, cellName(6, row), e["remark"])
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=annual_leave_events_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
}
