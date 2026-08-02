package handler

import (
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// salaryEventListQuery 工资事件列表筛选解析（列表与导出共用）
func salaryEventListQuery(c *gin.Context) service.SalaryEventListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	return service.SalaryEventListQuery{
		PageNum:     pageReq.PageNum,
		PageSize:    pageReq.PageSize,
		PersonID:    uint(personID),
		BelongMonth: c.Query("belong_month"),
		EventType:   c.Query("event_type"),
	}
}

func GetSalaryEvents(c *gin.Context) {
	q := salaryEventListQuery(c)
	list, total, err := service.GetSalaryEventList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, utils.PageRequest{PageNum: q.PageNum, PageSize: q.PageSize}))
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
	if err := service.CreateSalaryEvent(c.Request.Context(), &e); err != nil {
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
	if err := service.UpdateSalaryEvent(c.Request.Context(), uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteSalaryEvent(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreSalaryEvent(c.Request.Context(), uint(id)); err != nil {
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

// salaryEventExportFilters 工资事件导出文件名筛选摘要
func salaryEventExportFilters(q service.SalaryEventListQuery) []string {
	var parts []string
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	if q.BelongMonth != "" {
		parts = append(parts, "月份="+q.BelongMonth)
	}
	if q.EventType != "" {
		parts = append(parts, "类型="+q.EventType)
	}
	return parts
}

func ExportSalaryEvents(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := salaryEventListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, _ := service.GetSalaryEventList(q)

	var rows [][]interface{}
	for _, e := range list {
		rows = append(rows, []interface{}{
			e["person_name"], e["belong_month"], e["event_type"], e["amount"], e["remark"],
		})
	}
	writeExcel(c, "工资事件",
		[]string{"人员", "归属月份", "类型", "值", "备注"}, rows,
		salaryEventExportFilters(q)...)
}
