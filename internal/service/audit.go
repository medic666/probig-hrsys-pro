package service

import (
	"probig/internal/dao"
	"probig/internal/models"
)

func GetAuditLogList(page, pageSize int, operatorID uint, targetType, action, startDate, endDate, batchID string) ([]models.AuditLog, int64, error) {
	return dao.GetAuditLogList(page, pageSize, operatorID, targetType, action, startDate, endDate, batchID)
}
