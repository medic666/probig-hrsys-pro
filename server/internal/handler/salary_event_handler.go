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

func GetSalaryEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetSalaryEventList(pageReq.PageNum, pageReq.PageSize,
		uint(personID), c.Query("belong_month"), c.Query("event_type"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func CreateSalaryEvent(c *gin.Context) {
	var e model.SalaryEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if e.PersonID == 0 || e.BelongMonth == "" || e.EventType == "" {
		utils.BadRequest(c, "人员、月份、事件类型为必填项")
		return
	}
	if err := service.CreateSalaryEvent(&e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": e.ID})
}

func UpdateSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e model.SalaryEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdateSalaryEvent(uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteSalaryEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreSalaryEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedSalaryEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedSalaryEvents(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func ExportSalaryEvents(c *gin.Context) {
	list, _, _ := service.GetSalaryEventList(1, 10000, 0, c.Query("belong_month"), c.Query("event_type"))
	f := excelize.NewFile()
	defer f.Close()
	sheet := "工资事件"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"人员", "归属月份", "事件类型", "金额", "备注"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, e := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), e["person_name"])
		f.SetCellValue(sheet, cellName(2, row), e["belong_month"])
		f.SetCellValue(sheet, cellName(3, row), e["event_type"])
		f.SetCellValue(sheet, cellName(4, row), e["amount"])
		f.SetCellValue(sheet, cellName(5, row), e["remark"])
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=salary_events_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
}
