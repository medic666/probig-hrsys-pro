package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/medic666/probig/internal/models"
	"github.com/medic666/probig/internal/response"
	"github.com/medic666/probig/internal/services"
)

type PersonHandler struct {
	svc *services.PersonService
}

func NewPersonHandler(svc *services.PersonService) *PersonHandler {
	return &PersonHandler{svc: svc}
}

func (h *PersonHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	status := c.Query("status")

	list, total, err := h.svc.List(page, pageSize, keyword, status)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, list, total, page, pageSize)
}

func (h *PersonHandler) Create(c *gin.Context) {
	var req models.PersonEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.EventType == "" {
		req.EventType = "onboard"
	}
	entity, err := h.svc.Create(req, c.GetUint("user_id"), c.ClientIP())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, entity)
}

func (h *PersonHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	entity, events, snapshots, err := h.svc.GetDetail(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"entity":    entity,
		"events":    events,
		"snapshots": snapshots,
	})
}

func (h *PersonHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.svc.UpdateStatus(uint(id), req.Status, c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *PersonHandler) CreateEvent(c *gin.Context) {
	var req models.PersonEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.PersonID == 0 {
		response.BadRequest(c, "请指定人员ID")
		return
	}
	event, err := h.svc.CreateEvent(req, c.GetUint("user_id"), c.ClientIP())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, event)
}

func (h *PersonHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	var req models.PersonEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.PersonID == 0 {
		response.BadRequest(c, "请指定人员ID")
		return
	}
	if err := h.svc.UpdateEvent(uint(id), req, c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *PersonHandler) DeleteEvent(c *gin.Context) {
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

func (h *PersonHandler) GetSnapshots(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	_, _, snapshots, err := h.svc.GetDetail(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, snapshots)
}
