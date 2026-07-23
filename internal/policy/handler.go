package policy

import (
	"strconv"

	"probig/internal/common"
	"probig/internal/event"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	p := common.GetPagination(c)
	policyType := c.Query("policyType")

	policies, total, err := h.service.List(policyType, p.Search, p.Page, p.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessPage(c, policies, total, p.Page, p.PageSize)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	policy, err := h.service.GetByID(id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "制度不存在")
		return
	}

	common.Success(c, policy)
}

func (h *Handler) Create(c *gin.Context) {
	var policy Policy
	if err := common.BindJSON(c, &policy); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.Create(&policy, operatorID, "创建制度"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, policy)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	var policy Policy
	if err := common.BindJSON(c, &policy); err != nil {
		return
	}

	policy.ID = id
	operatorID := event.GetOperatorID(c)
	if err := h.service.Update(&policy, operatorID, "更新制度"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, policy)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.Delete(id, operatorID, "删除制度"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (h *Handler) GetVersions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	versions, err := h.service.GetVersions(id)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, versions)
}

func (h *Handler) GetTimeline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	events, err := h.service.GetTimeline(id)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, events)
}
