package audit

import (
	"encoding/json"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

type AuditLog = database.AuditLog

func CreateAuditLog(tx *gorm.DB, operatorID uint, operatorName string, targetType string, targetID uint, targetName string, action string, beforeData interface{}, afterData interface{}, ip string) error {
	var beforeJSON, afterJSON string

	if beforeData != nil {
		data, err := json.Marshal(beforeData)
		if err != nil {
			return err
		}
		beforeJSON = string(data)
	}

	if afterData != nil {
		data, err := json.Marshal(afterData)
		if err != nil {
			return err
		}
		afterJSON = string(data)
	}

	log := AuditLog{
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		TargetType:     targetType,
		TargetID:       targetID,
		TargetName:     targetName,
		Action:         action,
		BeforeSnapshot: beforeJSON,
		AfterSnapshot:  afterJSON,
		IP:             ip,
	}

	if tx != nil {
		return tx.Create(&log).Error
	}
	return database.DB.Create(&log).Error
}
