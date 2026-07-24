package audit_log

import (
	"gorm.io/gorm"
)

type AuditListParams struct {
	Page         int
	PageSize     int
	OperatorName string
	TargetType   string
	Action       string
	BatchID      string
	StartTime    string
	EndTime      string
}

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) Create(log *AuditLog) error {
	return d.db.Create(log).Error
}

func (d *DAO) List(params AuditListParams) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := d.db.Model(&AuditLog{})

	if params.OperatorName != "" {
		query = query.Where("operator_name LIKE ?", "%"+params.OperatorName+"%")
	}
	if params.TargetType != "" {
		query = query.Where("target_type = ?", params.TargetType)
	}
	if params.Action != "" {
		query = query.Where("action = ?", params.Action)
	}
	if params.BatchID != "" {
		query = query.Where("batch_id = ?", params.BatchID)
	}
	if params.StartTime != "" {
		query = query.Where("created_at >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("created_at <= ?", params.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
