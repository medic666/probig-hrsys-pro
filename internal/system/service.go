package system

import (
	"gorm.io/gorm"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
)

func ListAllConfigs() ([]SysConfig, error) {
	configs, err := ListAllConfigsFromDB()
	if err != nil {
		return nil, err
	}
	for i := range configs {
		if configs[i].ConfigKey == "system.encrypt_key" && len(configs[i].ConfigValue) > 8 {
			configs[i].ConfigValue = configs[i].ConfigValue[:4] + "****" + configs[i].ConfigValue[len(configs[i].ConfigValue)-4:]
		}
	}
	return configs, nil
}

func UpdateConfig(id uint, value string, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var cfg SysConfig
		if err := tx.First(&cfg, id).Error; err != nil {
			return err
		}

		if cfg.ConfigKey == "system.encrypt_key" {
			return gorm.ErrInvalidData
		}

		oldValue := cfg.ConfigValue

		cfg.ConfigValue = value
		if err := tx.Save(&cfg).Error; err != nil {
			return err
		}

		if err := config.Set(cfg.ConfigKey, cfg.ConfigValue); err != nil {
			return err
		}

		audit.CreateAuditLog(tx, operatorID, operatorName, "sys_config", cfg.ID, cfg.ConfigName, "配置修改", oldValue, cfg.ConfigValue, clientIP)

		return nil
	})
}
