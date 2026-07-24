package dao

import (
	"probig/internal/models"
)

func GetAllSysConfigs() ([]models.SysConfig, error) {
	var list []models.SysConfig
	if err := DB().Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetSysConfigByKey(key string) (*models.SysConfig, error) {
	var c models.SysConfig
	if err := DB().Where("config_key = ?", key).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func UpdateSysConfig(c *models.SysConfig) error {
	return DB().Save(c).Error
}

func UpsertSysConfig(c *models.SysConfig) error {
	var existing models.SysConfig
	err := DB().Where("config_key = ?", c.ConfigKey).First(&existing).Error
	if err != nil {
		return DB().Create(c).Error
	}
	c.ID = existing.ID
	return DB().Save(c).Error
}
