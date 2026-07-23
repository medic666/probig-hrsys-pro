package asset

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
	assetType := c.Query("assetType")

	assets, total, err := h.service.List(assetType, p.Search, p.Page, p.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessPage(c, assets, total, p.Page, p.PageSize)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	asset, err := h.service.GetByID(id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "资产不存在")
		return
	}

	common.Success(c, asset)
}

func (h *Handler) Create(c *gin.Context) {
	var asset Asset
	if err := common.BindJSON(c, &asset); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.Create(&asset, operatorID, "创建资产"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, asset)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	var asset Asset
	if err := common.BindJSON(c, &asset); err != nil {
		return
	}

	asset.ID = id
	operatorID := event.GetOperatorID(c)
	if err := h.service.Update(&asset, operatorID, "更新资产"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, asset)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.Delete(id, operatorID, "删除资产"); err != nil {
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
