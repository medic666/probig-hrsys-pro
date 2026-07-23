package person

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

	persons, total, err := h.service.List(p.Search, p.Page, p.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessPage(c, persons, total, p.Page, p.PageSize)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	person, err := h.service.GetByID(id)
	if err != nil {
		common.Error(c, common.CodeNotFound, "人员不存在")
		return
	}

	common.Success(c, person)
}

func (h *Handler) Create(c *gin.Context) {
	var person Person
	if err := common.BindJSON(c, &person); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.Create(&person, operatorID, "创建人员"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, person)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	var person Person
	if err := common.BindJSON(c, &person); err != nil {
		return
	}

	person.ID = id
	operatorID := event.GetOperatorID(c)
	if err := h.service.Update(&person, operatorID, "更新人员"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, person)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.Delete(id, operatorID, "删除人员"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (h *Handler) GetSnapshot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	at := c.Query("at")
	if at == "" {
		common.Error(c, common.CodeBadRequest, "缺少时间参数")
		return
	}

	person, err := h.service.GetSnapshot(id, at)
	if err != nil {
		common.Error(c, common.CodeNotFound, "未找到快照")
		return
	}

	common.Success(c, person)
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

func (h *Handler) All(c *gin.Context) {
	persons, err := h.service.All()
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, persons)
}
