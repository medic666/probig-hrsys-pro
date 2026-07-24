package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/database"
	"probig/internal/models"
	"probig/internal/utils"
)

type AuditHandler struct{}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{}
}

func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	targetType := c.Query("target_type")
	userIDStr := c.Query("user_id")

	var logs []models.AuditLog
	var total int64
	query := database.DB.Model(&models.AuditLog{})
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if userIDStr != "" {
		query = query.Where("user_id = ?", userIDStr)
	}
	query.Count(&total)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		utils.InternalError(c, "查询审计日志失败")
		return
	}
	utils.SuccessPage(c, logs, total, page, pageSize)
}
