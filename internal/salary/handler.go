package salary

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
	personIDStr := c.Query("personId")
	personID, _ := strconv.ParseInt(personIDStr, 10, 64)
	yearMonth := c.Query("yearMonth")

	records, total, err := h.service.ListRecords(personID, yearMonth, p.Page, p.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessPage(c, records, total, p.Page, p.PageSize)
}

func (h *Handler) Calculate(c *gin.Context) {
	var req struct {
		PersonID  int64  `json:"personId" binding:"required"`
		YearMonth string `json:"yearMonth" binding:"required"`
	}
	if err := common.BindJSON(c, &req); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	record, err := h.service.CalculateSalary(req.PersonID, req.YearMonth, operatorID)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, record)
}

func (h *Handler) GetRecord(c *gin.Context) {
	personID, err := strconv.ParseInt(c.Query("personId"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效人员ID")
		return
	}
	yearMonth := c.Query("yearMonth")

	record, err := h.service.GetRecord(personID, yearMonth)
	if err != nil {
		common.Error(c, common.CodeNotFound, "工资记录不存在")
		return
	}

	common.Success(c, record)
}

func (h *Handler) AddAdjustment(c *gin.Context) {
	var adj SalaryAdjustment
	if err := common.BindJSON(c, &adj); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.AddAdjustment(&adj, operatorID, "添加工资调整"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, adj)
}

func (h *Handler) DeleteAdjustment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.DeleteAdjustment(id, operatorID, "删除工资调整"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (h *Handler) GetAdjustments(c *gin.Context) {
	personID, err := strconv.ParseInt(c.Query("personId"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效人员ID")
		return
	}
	yearMonth := c.Query("yearMonth")

	adjs, err := h.service.GetAdjustments(personID, yearMonth)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, adjs)
}

func (h *Handler) ListByMonth(c *gin.Context) {
	yearMonth := c.Query("yearMonth")
	if yearMonth == "" {
		common.Error(c, common.CodeBadRequest, "缺少月份参数")
		return
	}

	records, err := h.service.ListByMonth(yearMonth)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, records)
}
