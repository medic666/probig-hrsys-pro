package audit

import (
	"probig/internal/audit_log"

	"gorm.io/gorm"
)

var GlobalAuditService *AuditService

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Log(operatorID uint, operatorName, targetType string, targetID uint, action, beforeJSON, afterJSON, batchID, ip string) error {
	log := audit_log.AuditLog{
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		TargetType:     targetType,
		TargetID:       targetID,
		Action:         action,
		BeforeSnapshot: beforeJSON,
		AfterSnapshot:  afterJSON,
		BatchID:        batchID,
		IP:             ip,
	}
	return s.db.Create(&log).Error
}
