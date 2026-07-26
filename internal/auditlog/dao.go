package auditlog

import (
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

func db() *gorm.DB {
	return database.DB
}

func ListLogs(filter AuditLogFilter) ([]AuditLog, int64, error) {
	query := db().Model(&AuditLog{})

	if filter.OperatorID != nil {
		query = query.Where("operator_id = ?", *filter.OperatorID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.TargetType != "" {
		query = query.Where("target_type = ?", filter.TargetType)
	}
	if filter.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", filter.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if filter.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", filter.EndDate)
		if err == nil {
			query = query.Where("created_at <= ?", endTime.Add(24*time.Hour))
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []AuditLog
	pageNum := filter.PageNum
	pageSize := filter.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func GetAllLogs(filter AuditLogFilter) ([]AuditLog, error) {
	query := db().Model(&AuditLog{})

	if filter.OperatorID != nil {
		query = query.Where("operator_id = ?", *filter.OperatorID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.TargetType != "" {
		query = query.Where("target_type = ?", filter.TargetType)
	}
	if filter.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", filter.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if filter.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", filter.EndDate)
		if err == nil {
			query = query.Where("created_at <= ?", endTime.Add(24*time.Hour))
		}
	}

	var logs []AuditLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func GetLogByID(id uint) (*AuditLog, error) {
	var log AuditLog
	err := db().Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}
