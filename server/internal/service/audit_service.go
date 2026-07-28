package service

import (
	"probig/server/internal/dao"
	"probig/server/internal/model"

	"gorm.io/gorm"
)

func GetAuditLogList(pageNum, pageSize int, operatorName, action, targetType, dateStart, dateEnd string) ([]model.AuditLog, int64, error) {
	tx := dao.DB.Model(&model.AuditLog{})
	if operatorName != "" {
		tx = tx.Where("operator_name LIKE ?", "%"+operatorName+"%")
	}
	if action != "" {
		tx = tx.Where("action = ?", action)
	}
	if targetType != "" {
		tx = tx.Where("target_type = ?", targetType)
	}
	if dateStart != "" {
		tx = tx.Where("created_at >= ?", dateStart)
	}
	if dateEnd != "" {
		tx = tx.Where("created_at <= ?", dateEnd+" 23:59:59")
	}
	var total int64
	tx.Count(&total)
	var list []model.AuditLog
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)
	return list, total, nil
}

func GetAuditLogByID(id uint) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := dao.DB.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func GetDB() *gorm.DB {
	return dao.DB
}
