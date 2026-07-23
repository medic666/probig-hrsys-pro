package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/medic666/probig/internal/models"
	"github.com/medic666/probig/internal/response"
	"github.com/medic666/probig/internal/services"
)

type OrganizationHandler struct {
	svc *services.OrganizationService
}

func NewOrganizationHandler(svc *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{svc: svc}
}

func (h *OrganizationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	list, total, err := h.svc.List(page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, list, total, page, pageSize)
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	var req models.OrgEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.EventType == "" {
		req.EventType = "establish"
	}
	entity, err := h.svc.Create(req, c.GetUint("user_id"), c.ClientIP())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, entity)
}

func (h *OrganizationHandler) GetDetail(c *gin.Context) {
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

func (h *OrganizationHandler) CreateEvent(c *gin.Context) {
	var req models.OrgEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.OrgID == 0 {
		response.BadRequest(c, "请指定组织ID")
		return
	}
	event, err := h.svc.CreateEvent(req, c.GetUint("user_id"), c.ClientIP())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, event)
}

func (h *OrganizationHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	var req models.OrgEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if req.OrgID == 0 {
		response.BadRequest(c, "请指定组织ID")
		return
	}
	if err := h.svc.UpdateEvent(uint(id), req, c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *OrganizationHandler) DeleteEvent(c *gin.Context) {
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
