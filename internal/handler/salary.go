package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetSalaryEventList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	belongMonth := c.Query("belong_month")
	eventType := c.Query("event_type")
	list, total, err := service.GetSalaryEventList(page, pageSize, uint(personID), belongMonth, eventType)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func GetSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	e, err := service.GetSalaryEvent(uint(id))
	if err != nil {
		response.Error(c, "事件不存在")
		return
	}
	response.Success(c, e)
}

func CreateSalaryEvent(c *gin.Context) {
	var e models.SalaryEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.CreateSalaryEvent(c, &e); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, e)
}

func UpdateSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e models.SalaryEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		response.Error(c, "参数错误")
		return
	}
	e.ID = uint(id)
	if err := service.UpdateSalaryEvent(c, &e); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, e)
}

func DeleteSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteSalaryEvent(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func CalculateAttendance(c *gin.Context) {
	var req struct {
		PersonID    uint   `json:"person_id"`
		BelongMonth string `json:"belong_month"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	summary, err := service.CalculateAttendance(c, req.PersonID, req.BelongMonth)
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, summary)
}

func GetAttendanceSummaryList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	belongMonth := c.Query("belong_month")
	list, total, err := service.GetAttendanceSummaryList(page, pageSize, uint(personID), belongMonth)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func LockAttendanceSummary(c *gin.Context) {
	var req struct {
		PersonID    uint   `json:"person_id"`
		BelongMonth string `json:"belong_month"`
		IsLocked    bool   `json:"is_locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.LockAttendanceSummary(c, req.PersonID, req.BelongMonth, req.IsLocked); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func CalculateSalary(c *gin.Context) {
	var req struct {
		PersonID    uint   `json:"person_id"`
		BelongMonth string `json:"belong_month"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	summary, err := service.CalculateSalary(c, req.PersonID, req.BelongMonth)
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, summary)
}

func GetSalarySummaryList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	belongMonth := c.Query("belong_month")
	list, total, err := service.GetSalarySummaryList(page, pageSize, uint(personID), belongMonth)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func LockSalarySummary(c *gin.Context) {
	var req struct {
		PersonID    uint   `json:"person_id"`
		BelongMonth string `json:"belong_month"`
		IsLocked    bool   `json:"is_locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.LockSalarySummary(c, req.PersonID, req.BelongMonth, req.IsLocked); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}
