package audit_log

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

func (h *Handler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	params := AuditListParams{
		Page:         page,
		PageSize:     pageSize,
		OperatorName: c.Query("operatorName"),
		TargetType:   c.Query("targetType"),
		Action:       c.Query("action"),
		BatchID:      c.Query("batchId"),
		StartTime:    c.Query("startTime"),
		EndTime:      c.Query("endTime"),
	}

	logs, total, err := h.service.List(params)
	if err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Page(c, logs, total, page, pageSize)
}
