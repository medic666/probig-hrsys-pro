package services

import (
	"sync"

	"probig/database"
	"probig/models"
)

var ConfigCache sync.Map

func LoadConfigs() {
	var configs []models.SysConfig
	database.DB.Find(&configs)
	for _, c := range configs {
		ConfigCache.Store(c.ConfigKey, c.ConfigValue)
	}
}

func ListConfigs() []models.SysConfig {
	var configs []models.SysConfig
	database.DB.Order("id ASC").Find(&configs)
	return configs
}

func GetConfig(key string) (string, bool) {
	v, ok := ConfigCache.Load(key)
	if ok {
		return v.(string), true
	}
	return "", false
}

func UpdateConfig(configKey, configValue string) error {
	err := database.DB.Model(&models.SysConfig{}).Where("config_key = ?", configKey).Updates(map[string]interface{}{
		"config_value": configValue,
	}).Error
	if err != nil {
		return err
	}
	ConfigCache.Store(configKey, configValue)
	return nil
}

func ListAuditLogs(operatorID uint, targetType, action, batchID, startDate, endDate string, offset, limit int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64
	db := database.DB.Model(&models.AuditLog{})
	if operatorID > 0 {
		db = db.Where("operator_id = ?", operatorID)
	}
	if targetType != "" {
		db = db.Where("target_type = ?", targetType)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if batchID != "" {
		db = db.Where("batch_id = ?", batchID)
	}
	if startDate != "" {
		db = db.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("created_at <= ?", endDate+" 23:59:59")
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&logs).Error
	return logs, total, err
}
