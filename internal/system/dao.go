package system

import (
	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

func db() *gorm.DB {
	return database.DB
}

func ListAllConfigsFromDB() ([]SysConfig, error) {
	var configs []SysConfig
	err := db().Order("id ASC").Find(&configs).Error
	return configs, err
}

func GetConfigByID(id uint) (*SysConfig, error) {
	var cfg SysConfig
	err := db().Where("id = ?", id).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func GetConfigByKey(key string) (*SysConfig, error) {
	var cfg SysConfig
	err := db().Where("config_key = ?", key).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func UpdateConfigValue(id uint, value string) error {
	return db().Model(&SysConfig{}).Where("id = ?", id).Updates(map[string]interface{}{
		"config_value": value,
	}).Error
}

func UpdateConfigValueByKey(key string, value string) error {
	return db().Model(&SysConfig{}).Where("config_key = ?", key).Update("config_value", value).Error
}
