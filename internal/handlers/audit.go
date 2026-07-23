package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/medic666/probig/internal/response"
	"github.com/medic666/probig/internal/services"
)

type AuditHandler struct {
	svc *services.AuditService
}

func NewAuditHandler(svc *services.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	entityType := c.Query("entity_type")
	action := c.Query("action")

	list, total, err := h.svc.List(page, pageSize, uint(userID), entityType, action)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, list, total, page, pageSize)
}
