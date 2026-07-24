package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/service"
	"probig/internal/utils"
)

type OrganizationHandler struct {
	svc *service.OrganizationService
}

func NewOrganizationHandler() *OrganizationHandler {
	return &OrganizationHandler{svc: service.NewOrganizationService()}
}

func (h *OrganizationHandler) List(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	snapshots, total, err := h.svc.ListSnapshots(search, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询组织列表失败")
		return
	}
	utils.SuccessPage(c, snapshots, total, page, pageSize)
}

func (h *OrganizationHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	snapshot, err := h.svc.GetSnapshot(uint(id))
	if err != nil {
		utils.NotFound(c, "组织不存在")
		return
	}
	utils.Success(c, snapshot)
}

func (h *OrganizationHandler) History(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	snapshots, err := h.svc.GetHistory(uint(id))
	if err != nil {
		utils.InternalError(c, "查询历史记录失败")
		return
	}
	utils.Success(c, snapshots)
}

func (h *OrganizationHandler) ListEvents(c *gin.Context) {
	entityIDStr := c.Query("entity_id")
	var entityID uint
	if entityIDStr != "" {
		id, err := strconv.ParseUint(entityIDStr, 10, 64)
		if err != nil {
			utils.BadRequest(c, "无效的组织ID")
			return
		}
		entityID = uint(id)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	events, total, err := h.svc.ListEvents(entityID, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询事件列表失败")
		return
	}
	utils.SuccessPage(c, events, total, page, pageSize)
}

func (h *OrganizationHandler) CreateEvent(c *gin.Context) {
	var input service.OrganizationEventInput
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

func (h *OrganizationHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var input service.OrganizationEventInput
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

func (h *OrganizationHandler) DeleteEvent(c *gin.Context) {
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
