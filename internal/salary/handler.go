package salary

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) ListEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	personIDStr := c.Query("personId")
	belongMonth := c.Query("belongMonth")
	eventType := c.Query("eventType")

	var personID uint
	if personIDStr != "" {
		if parsed, err := strconv.ParseUint(personIDStr, 10, 64); err == nil {
			personID = uint(parsed)
		}
	}

	events, total, err := h.service.ListEvents(personID, belongMonth, eventType, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, events, total, page, pageSize)
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var event SalaryEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if event.PersonID == 0 {
		response.ParamError(c, "personId is required")
		return
	}
	if event.BelongMonth == "" {
		response.ParamError(c, "belongMonth is required")
		return
	}
	if event.EventType == "" {
		response.ParamError(c, "eventType is required")
		return
	}

	operatorID, operatorName := h.getOperator(c)
	if err := h.service.CreateEvent(&event, operatorID, operatorName); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, event)
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	var event SalaryEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	event.ID = uint(id)
	operatorID, operatorName := h.getOperator(c)
	if err := h.service.UpdateEvent(&event, operatorID, operatorName); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	operatorID, operatorName := h.getOperator(c)
	if err := h.service.DeleteEvent(uint(id), operatorID, operatorName); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) CalculateSalary(c *gin.Context) {
	var req struct {
		PersonID    uint   `json:"personId"`
		BelongMonth string `json:"belongMonth"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if req.PersonID == 0 {
		response.ParamError(c, "personId is required")
		return
	}
	if req.BelongMonth == "" {
		response.ParamError(c, "belongMonth is required")
		return
	}

	summary, err := h.service.CalculateSalary(req.PersonID, req.BelongMonth)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, summary)
}

func (h *Handler) CalculateSalaries(c *gin.Context) {
	var req struct {
		PersonIDs   []uint `json:"personIds"`
		BelongMonth string `json:"belongMonth"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if len(req.PersonIDs) == 0 {
		response.ParamError(c, "personIds is required")
		return
	}
	if req.BelongMonth == "" {
		response.ParamError(c, "belongMonth is required")
		return
	}

	summaries, err := h.service.CalculateSalaries(req.PersonIDs, req.BelongMonth)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, summaries)
}

func (h *Handler) LockSummary(c *gin.Context) {
	var req struct {
		PersonID    uint   `json:"personId"`
		BelongMonth string `json:"belongMonth"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if req.PersonID == 0 {
		response.ParamError(c, "personId is required")
		return
	}
	if req.BelongMonth == "" {
		response.ParamError(c, "belongMonth is required")
		return
	}

	operatorID, operatorName := h.getOperator(c)
	if err := h.service.LockSummary(req.PersonID, req.BelongMonth, operatorID, operatorName); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) UnlockSummary(c *gin.Context) {
	var req struct {
		PersonID    uint   `json:"personId"`
		BelongMonth string `json:"belongMonth"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if req.PersonID == 0 {
		response.ParamError(c, "personId is required")
		return
	}
	if req.BelongMonth == "" {
		response.ParamError(c, "belongMonth is required")
		return
	}

	operatorID, operatorName := h.getOperator(c)
	if err := h.service.UnlockSummary(req.PersonID, req.BelongMonth, operatorID, operatorName); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListSummaries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	belongMonth := c.Query("belongMonth")
	isLockedStr := c.Query("isLocked")

	var personIDs []uint
	personIDsStr := c.QueryArray("personIds")
	for _, s := range personIDsStr {
		if parsed, err := strconv.ParseUint(s, 10, 64); err == nil {
			personIDs = append(personIDs, uint(parsed))
		}
	}

	var isLocked *bool
	if isLockedStr != "" {
		if parsed, err := strconv.ParseBool(isLockedStr); err == nil {
			isLocked = &parsed
		}
	}

	summaries, total, err := h.service.ListSummaries(personIDs, belongMonth, isLocked, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, summaries, total, page, pageSize)
}

func (h *Handler) GetSummary(c *gin.Context) {
	personIDStr := c.Query("personId")
	if personIDStr == "" {
		response.ParamError(c, "personId is required")
		return
	}
	personID, err := strconv.ParseUint(personIDStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid personId")
		return
	}

	belongMonth := c.Query("belongMonth")
	if belongMonth == "" {
		response.ParamError(c, "belongMonth is required")
		return
	}

	summary, err := h.service.GetSummaryByPersonMonth(uint(personID), belongMonth)
	if err != nil {
		response.NotFound(c, "summary not found")
		return
	}

	response.Success(c, summary)
}

func (h *Handler) getOperator(c *gin.Context) (uint, string) {
	var operatorID uint
	var operatorName string
	if val, exists := c.Get("userID"); exists {
		switch v := val.(type) {
		case uint:
			operatorID = v
		case float64:
			operatorID = uint(v)
		case int:
			operatorID = uint(v)
		}
	}
	if val, exists := c.Get("username"); exists {
		if v, ok := val.(string); ok {
			operatorName = v
		}
	}
	return operatorID, operatorName
}
