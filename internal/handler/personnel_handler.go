package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/service"
	"probig/internal/utils"
)

type PersonnelHandler struct {
	svc *service.PersonnelService
}

func NewPersonnelHandler() *PersonnelHandler {
	return &PersonnelHandler{svc: service.NewPersonnelService()}
}

func (h *PersonnelHandler) List(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	snapshots, total, err := h.svc.ListSnapshots(search, page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询人员列表失败")
		return
	}
	utils.SuccessPage(c, snapshots, total, page, pageSize)
}

func (h *PersonnelHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	snapshot, err := h.svc.GetSnapshot(uint(id))
	if err != nil {
		utils.NotFound(c, "人员不存在")
		return
	}
	utils.Success(c, snapshot)
}

func (h *PersonnelHandler) History(c *gin.Context) {
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

func (h *PersonnelHandler) ListEvents(c *gin.Context) {
	entityIDStr := c.Query("entity_id")
	var entityID uint
	if entityIDStr != "" {
		id, err := strconv.ParseUint(entityIDStr, 10, 64)
		if err != nil {
			utils.BadRequest(c, "无效的人员ID")
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

func (h *PersonnelHandler) CreateEvent(c *gin.Context) {
	var input service.PersonnelEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
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

func (h *PersonnelHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	var input service.PersonnelEventInput
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

func (h *PersonnelHandler) DeleteEvent(c *gin.Context) {
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

func getEventContext(c *gin.Context) service.EventContext {
	uid, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	var userID uint
	var username string
	if v, ok := uid.(uint); ok {
		userID = v
	}
	if v, ok := uname.(string); ok {
		username = v
	}
	return service.EventContext{UserID: userID, Username: username}
}
