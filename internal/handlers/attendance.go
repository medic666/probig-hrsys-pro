package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/medic666/probig/internal/models"
	"github.com/medic666/probig/internal/response"
	"github.com/medic666/probig/internal/services"
)

type AttendanceHandler struct {
	svc *services.AttendanceService
}

func NewAttendanceHandler(svc *services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

func (h *AttendanceHandler) ListEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	period := c.Query("period")
	eventType := c.Query("event_type")

	list, total, err := h.svc.ListEvents(page, pageSize, uint(personID), period, eventType)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, list, total, page, pageSize)
}

func (h *AttendanceHandler) CreateEvent(c *gin.Context) {
	var req models.AttendanceEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	event, err := h.svc.CreateEvent(req, c.GetUint("user_id"), c.ClientIP())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, event)
}

func (h *AttendanceHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	var req models.AttendanceEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateEvent(uint(id), req, c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AttendanceHandler) DeleteEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	if err := h.svc.DeleteEvent(uint(id), c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AttendanceHandler) Calculate(c *gin.Context) {
	var req models.CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Calculate(req, c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AttendanceHandler) ListSummaries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	period := c.Query("period")

	list, total, err := h.svc.ListSummaries(page, pageSize, uint(personID), period)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, list, total, page, pageSize)
}
