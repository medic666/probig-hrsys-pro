package service

import (
	"probig/server/internal/dao"
	"probig/server/internal/model"

	"gorm.io/gorm"
)

// AuditLogListQuery 审计日志列表查询（列表与导出共用）
type AuditLogListQuery struct {
	PageNum      int
	PageSize     int
	OperatorName string
	Action       string
	TargetType   string
	DateStart    string
	DateEnd      string
}

func GetAuditLogList(q AuditLogListQuery) ([]model.AuditLog, int64, error) {
	tx := dao.DB.Model(&model.AuditLog{})
	if q.OperatorName != "" {
		tx = tx.Where("operator_name LIKE ?", "%"+q.OperatorName+"%")
	}
	if q.Action != "" {
		tx = tx.Where("action = ?", q.Action)
	}
	if q.TargetType != "" {
		tx = tx.Where("target_type = ?", q.TargetType)
	}
	if q.DateStart != "" {
		tx = tx.Where("created_at >= ?", q.DateStart)
	}
	if q.DateEnd != "" {
		tx = tx.Where("created_at <= ?", q.DateEnd+" 23:59:59")
	}
	var total int64
	tx.Count(&total)
	var list []model.AuditLog
	offset := (q.PageNum - 1) * q.PageSize
	tx.Offset(offset).Limit(q.PageSize).Order("id DESC").Find(&list)
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
