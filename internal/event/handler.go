package event

import (
	"strconv"

	"probig/internal/auth"
	"probig/internal/common"

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

	var filter EventFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		common.Error(c, common.CodeBadRequest, "参数错误")
		return
	}

	events, total, err := h.service.List(filter, p.Page, p.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessPage(c, events, total, p.Page, p.PageSize)
}

func (h *Handler) GetEntityEvents(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID, err := strconv.ParseInt(c.Param("entityId"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效的实体ID")
		return
	}

	events, err := h.service.GetByEntity(entityType, entityID)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, events)
}

func (h *Handler) SoftDelete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	if err := h.service.SoftDelete(id); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (h *Handler) UpdateRemark(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	if err := common.BindJSON(c, &req); err != nil {
		return
	}

	if err := h.service.UpdateRemark(id, req.Remark); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func GetOperatorID(c *gin.Context) int64 {
	claims := auth.GetUserClaims(c)
	if claims == nil {
		return 0
	}
	return claims.UserID
}
