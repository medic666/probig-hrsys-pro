package attendance

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

func (h *Handler) ListEvents(c *gin.Context) {
	p := common.GetPagination(c)

	personIDStr := c.Query("personId")
	personID, _ := strconv.ParseInt(personIDStr, 10, 64)
	yearMonth := c.Query("yearMonth")
	eventType := c.Query("eventType")

	events, total, err := h.service.ListEvents(personID, yearMonth, eventType, p.Page, p.PageSize)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.SuccessPage(c, events, total, p.Page, p.PageSize)
}

func (h *Handler) GetEvent(c *gin.Context) {
	common.Error(c, common.CodeNotFound, "not implemented")
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var attEvent AttendanceEvent
	if err := common.BindJSON(c, &attEvent); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	attEvent.OperatorID = operatorID

	if err := h.service.CreateEvent(&attEvent, operatorID, "创建考勤事件"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, attEvent)
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	var attEvent AttendanceEvent
	if err := common.BindJSON(c, &attEvent); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.UpdateEvent(id, &attEvent, operatorID, "更新考勤事件"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, attEvent)
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效ID")
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.DeleteEvent(id, operatorID, "删除考勤事件"); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (h *Handler) GetLeaveBalance(c *gin.Context) {
	personID, err := strconv.ParseInt(c.Query("personId"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效人员ID")
		return
	}

	balance, err := h.service.GetLeaveBalance(personID)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, balance)
}

func (h *Handler) GrantAnnualLeave(c *gin.Context) {
	var req struct {
		PersonID int64 `json:"personId"`
	}
	if err := common.BindJSON(c, &req); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.GrantAnnualLeave(req.PersonID, operatorID); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (h *Handler) CloseMonth(c *gin.Context) {
	var req struct {
		PersonID  int64  `json:"personId"`
		YearMonth string `json:"yearMonth"`
	}
	if err := common.BindJSON(c, &req); err != nil {
		return
	}

	operatorID := event.GetOperatorID(c)
	if err := h.service.CloseMonth(req.PersonID, req.YearMonth, operatorID); err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, nil)
}

func (h *Handler) GetMonthlyEvents(c *gin.Context) {
	personID, err := strconv.ParseInt(c.Query("personId"), 10, 64)
	if err != nil {
		common.Error(c, common.CodeBadRequest, "无效人员ID")
		return
	}
	yearMonth := c.Query("yearMonth")

	events, err := h.service.GetEventsByPersonMonth(personID, yearMonth)
	if err != nil {
		common.Error(c, common.CodeInternalError, err.Error())
		return
	}

	common.Success(c, events)
}
