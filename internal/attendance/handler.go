package attendance

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

type createCrossDayRequest struct {
	PersonID  uint    `json:"personId"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	EventType string  `json:"eventType"`
	SubType   string  `json:"subType"`
	Hours     float64 `json:"hours"`
	Remark    string  `json:"remark"`
}

type batchCreateRequest struct {
	PersonIDs []uint  `json:"personIds"`
	EventDate string  `json:"eventDate"`
	EventType string  `json:"eventType"`
	SubType   string  `json:"subType"`
	Hours     float64 `json:"hours"`
	Remark    string  `json:"remark"`
}

type calculateRequest struct {
	PersonID    uint   `json:"personId"`
	BelongMonth string `json:"belongMonth"`
}

type batchCalculateRequest struct {
	PersonIDs   []uint `json:"personIds"`
	BelongMonth string `json:"belongMonth"`
}

type lockUnlockRequest struct {
	PersonID    uint   `json:"personId"`
	BelongMonth string `json:"belongMonth"`
}

func (h *Handler) ListEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	var personIDs []uint
	if pidStr := c.Query("personId"); pidStr != "" {
		if pid, err := strconv.ParseUint(pidStr, 10, 64); err == nil {
			personIDs = []uint{uint(pid)}
		}
	}

	params := AttendanceListParams{
		Page:           page,
		PageSize:       pageSize,
		PersonIDs:      personIDs,
		EventDateStart: c.Query("eventDateStart"),
		EventDateEnd:   c.Query("eventDateEnd"),
		EventType:      c.Query("eventType"),
		SubType:        c.Query("subType"),
	}

	events, total, err := h.service.ListEvents(params)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, events, total, page, pageSize)
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var event AttendanceEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if event.PersonID == 0 {
		response.ParamError(c, "personId is required")
		return
	}
	if event.EventDate == "" {
		response.ParamError(c, "eventDate is required")
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

	var event AttendanceEvent
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

func (h *Handler) CreateCrossDayEvents(c *gin.Context) {
	var req createCrossDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if req.PersonID == 0 {
		response.ParamError(c, "personId is required")
		return
	}
	if req.StartDate == "" || req.EndDate == "" {
		response.ParamError(c, "startDate and endDate are required")
		return
	}
	if req.EventType == "" {
		response.ParamError(c, "eventType is required")
		return
	}

	operatorID, operatorName := h.getOperator(c)
	if err := h.service.CreateCrossDayEvents(req.PersonID, req.StartDate, req.EndDate, req.EventType, req.SubType, req.Hours, req.Remark, operatorID, operatorName); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) BatchCreateEvents(c *gin.Context) {
	var req batchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if len(req.PersonIDs) == 0 {
		response.ParamError(c, "personIds is required")
		return
	}
	if req.EventDate == "" {
		response.ParamError(c, "eventDate is required")
		return
	}
	if req.EventType == "" {
		response.ParamError(c, "eventType is required")
		return
	}

	operatorID, operatorName := h.getOperator(c)
	if err := h.service.BatchCreateEvents(req.PersonIDs, req.EventDate, req.EventType, req.SubType, req.Hours, req.Remark, operatorID, operatorName); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) CalculateSummary(c *gin.Context) {
	var req calculateRequest
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

	summary, err := h.service.CalculateSummary(req.PersonID, req.BelongMonth)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, summary)
}

func (h *Handler) CalculateSummaries(c *gin.Context) {
	var req batchCalculateRequest
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

	summaries, err := h.service.CalculateSummaries(req.PersonIDs, req.BelongMonth)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, summaries)
}

func (h *Handler) LockSummary(c *gin.Context) {
	var req lockUnlockRequest
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
	var req lockUnlockRequest
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
	yearMonth := c.Query("belongMonth")

	var personIDs []uint
	if pidStr := c.Query("personId"); pidStr != "" {
		if pid, err := strconv.ParseUint(pidStr, 10, 64); err == nil {
			personIDs = []uint{uint(pid)}
		}
	}

	var isLocked *bool
	if lockedStr := c.Query("isLocked"); lockedStr != "" {
		parsed := lockedStr == "true" || lockedStr == "1"
		isLocked = &parsed
	}

	summaries, total, err := h.service.ListSummaries(personIDs, yearMonth, isLocked, page, pageSize)
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

	summary, err := h.service.GetSummary(uint(personID), belongMonth)
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
