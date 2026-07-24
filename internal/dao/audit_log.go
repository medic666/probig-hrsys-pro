package dao

import (
	"probig/internal/models"
)

func CreateAuditLog(log *models.AuditLog) error {
	return DB().Create(log).Error
}

func GetAuditLogList(page, pageSize int, operatorID uint, targetType, action, startDate, endDate, batchID string) ([]models.AuditLog, int64, error) {
	var list []models.AuditLog
	var total int64
	q := DB().Model(&models.AuditLog{})
	if operatorID > 0 {
		q = q.Where("operator_id = ?", operatorID)
	}
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if startDate != "" {
		q = q.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		q = q.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if batchID != "" {
		q = q.Where("batch_id = ?", batchID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func CreateAuditLogs(logs []models.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return DB().Create(&logs).Error
}
