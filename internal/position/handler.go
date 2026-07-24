package position

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
	var personID uint
	if personIDStr != "" {
		if parsed, err := strconv.ParseUint(personIDStr, 10, 64); err == nil {
			personID = uint(parsed)
		}
	}

	events, total, err := h.service.ListEvents(personID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, events, total, page, pageSize)
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var event PositionEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if event.PersonID == 0 {
		response.ParamError(c, "personId is required")
		return
	}
	if event.EffectiveDate == "" {
		response.ParamError(c, "effectiveDate is required")
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

	var event PositionEvent
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

func (h *Handler) GetEventByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	event, err := h.service.GetEventByID(uint(id))
	if err != nil {
		response.NotFound(c, "event not found")
		return
	}

	response.Success(c, event)
}

func (h *Handler) GetSnapshot(c *gin.Context) {
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

	date := c.Query("date")
	if date == "" {
		response.ParamError(c, "date is required")
		return
	}

	snapshot, err := h.service.GetSnapshotByPersonAndDate(uint(personID), date)
	if err != nil {
		response.NotFound(c, "snapshot not found")
		return
	}

	response.Success(c, snapshot)
}

func (h *Handler) RebuildSnapshots(c *gin.Context) {
	personIDStr := c.Param("personId")
	personID, err := strconv.ParseUint(personIDStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid personId")
		return
	}

	if err := h.service.RebuildSnapshotsForPerson(uint(personID)); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
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
