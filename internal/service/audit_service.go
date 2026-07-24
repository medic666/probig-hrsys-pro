package service

import (
	"gorm.io/gorm"

	"probig/internal/database"
	"probig/internal/models"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

func (s *AuditService) Log(userID uint, username, action, targetType string, targetID uint, targetName, targetSummary string, payload models.JSONMap) {
	log := models.AuditLog{
		UserID:        userID,
		Username:      username,
		Action:        action,
		TargetType:    targetType,
		TargetID:      targetID,
		TargetName:    targetName,
		TargetSummary: targetSummary,
		Payload:       payload,
	}
	database.DB.Create(&log)
}

func (s *AuditService) LogTx(tx *gorm.DB, userID uint, username, action, targetType string, targetID uint, targetName, targetSummary string, payload models.JSONMap) error {
	log := models.AuditLog{
		UserID:        userID,
		Username:      username,
		Action:        action,
		TargetType:    targetType,
		TargetID:      targetID,
		TargetName:    targetName,
		TargetSummary: targetSummary,
		Payload:       payload,
	}
	return tx.Create(&log).Error
}
