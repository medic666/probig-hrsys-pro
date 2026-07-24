package handlers

import (
	"strconv"

	"probig/database"
	"probig/middleware"
	"probig/models"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func ListSalaryEvents(c *gin.Context) {
	personIDStr := c.Query("person_id")
	var personID uint
	if personIDStr != "" {
		pid, _ := strconv.ParseUint(personIDStr, 10, 64)
		personID = uint(pid)
	}
	belongMonth := c.Query("belong_month")
	eventType := c.Query("event_type")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	events, total, err := services.ListSalaryEvents(personID, belongMonth, eventType, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, events, total, page, pageSize)
}

func CreateSalaryEvent(c *gin.Context) {
	var event models.SalaryEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.CreateSalaryEvent(&event); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "salary_event", event.ID, "新增", "{}", event)
	utils.Success(c, event)
}

func UpdateSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	var event models.SalaryEvent
	database.DB.First(&event, uint(id))
	before := event
	if err := services.UpdateSalaryEvent(uint(id), updates); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	database.DB.First(&event, uint(id))
	middleware.AuditAction(c, "salary_event", uint(id), "修改", before, event)
	utils.Success(c, nil)
}

func DeleteSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var event models.SalaryEvent
	database.DB.First(&event, uint(id))
	before := event
	if err := services.DeleteSalaryEvent(uint(id)); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "salary_event", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func CalcSalary(c *gin.Context) {
	var input struct {
		BelongMonth string `json:"belong_month"`
		PersonIDs   []uint `json:"person_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	summaries, err := services.CalcSalary(c, input.BelongMonth, input.PersonIDs)
	if err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.WriteAuditLog(c, "salary_summary", 0, "核算", "{}", input, "")
	utils.Success(c, summaries)
}

func ListSalarySummaries(c *gin.Context) {
	month := c.Query("belong_month")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize
	summaries, total, _ := services.ListSalarySummaries(month, offset, pageSize)
	utils.SuccessPage(c, summaries, total, page, pageSize)
}

func LockSalarySummary(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		IsLocked bool `json:"is_locked"`
	}
	c.ShouldBindJSON(&input)
	services.LockSalarySummary(uint(id), input.IsLocked)
	action := "解锁"
	if input.IsLocked {
		action = "锁定"
	}
	middleware.AuditAction(c, "salary_summary", uint(id), action, "{}", input)
	utils.Success(c, nil)
}
