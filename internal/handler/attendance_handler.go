package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/service"
	"probig/internal/utils"
)

type AttendanceHandler struct {
	svc *service.AttendanceService
}

func NewAttendanceHandler() *AttendanceHandler {
	return &AttendanceHandler{svc: service.NewAttendanceService()}
}

func (h *AttendanceHandler) ListEvents(c *gin.Context) {
	entityIDStr := c.Query("entity_id")
	var entityID uint
	if entityIDStr != "" {
		id, _ := strconv.ParseUint(entityIDStr, 10, 64)
		entityID = uint(id)
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	events, total, err := h.svc.ListEvents(entityID, startDate, endDate, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询假勤事件失败")
		return
	}
	utils.SuccessPage(c, events, total, page, pageSize)
}

func (h *AttendanceHandler) CreateEvent(c *gin.Context) {
	var input service.AttendanceEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)
	event, err := h.svc.CreateEvent(auditSvc, ctx, &input)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, event)
}

func (h *AttendanceHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var input service.AttendanceEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)
	event, err := h.svc.UpdateEvent(auditSvc, ctx, uint(id), &input)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, event)
}

func (h *AttendanceHandler) DeleteEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)
	if err := h.svc.DeleteEvent(auditSvc, ctx, uint(id)); err != nil {
		utils.InternalError(c, "删除事件失败")
		return
	}
	utils.Success(c, nil)
}

func (h *AttendanceHandler) ListSummaries(c *gin.Context) {
	entityIDStr := c.Query("entity_id")
	var entityID uint
	if entityIDStr != "" {
		id, _ := strconv.ParseUint(entityIDStr, 10, 64)
		entityID = uint(id)
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	summaries, total, err := h.svc.ListSummaries(entityID, startDate, endDate, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询汇总失败")
		return
	}
	utils.SuccessPage(c, summaries, total, page, pageSize)
}

func (h *AttendanceHandler) Calculate(c *gin.Context) {
	var req struct {
		PeriodStart string `json:"period_start" binding:"required"`
		PeriodEnd   string `json:"period_end" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请输入核算时段")
		return
	}

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)
	summaries, err := h.svc.Calculate(auditSvc, ctx, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		utils.InternalError(c, "考勤核算失败")
		return
	}
	utils.Success(c, summaries)
}
