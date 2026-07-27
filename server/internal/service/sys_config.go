package service

import (
	"fmt"
	"sync"

	"probig/server/internal/model"

	"gorm.io/gorm"
)

var (
	configCache sync.Map
	configMu    sync.RWMutex
)

func InitSysConfig(db *gorm.DB) error {
	db.AutoMigrate(&model.SysConfig{})

	if err := LoadAllConfigs(db); err != nil {
		return err
	}

	if err := seedDefaultConfigs(db); err != nil {
		return err
	}

	return nil
}

func LoadAllConfigs(db *gorm.DB) error {
	var configs []model.SysConfig
	if err := db.Find(&configs).Error; err != nil {
		return err
	}

	configMu.Lock()
	defer configMu.Unlock()

	configCache = sync.Map{}
	for _, cfg := range configs {
		configCache.Store(cfg.ConfigKey, cfg.ConfigValue)
	}

	return nil
}

func GetConfigValue(key string) string {
	if val, ok := configCache.Load(key); ok {
		return val.(string)
	}
	return ""
}

func GetConfigValueOrDefault(key, defaultVal string) string {
	if val, ok := configCache.Load(key); ok {
		return val.(string)
	}
	return defaultVal
}

func RefreshConfig(db *gorm.DB, key string) error {
	var cfg model.SysConfig
	if err := db.Where("config_key = ?", key).First(&cfg).Error; err != nil {
		return err
	}

	configMu.Lock()
	defer configMu.Unlock()
	configCache.Store(cfg.ConfigKey, cfg.ConfigValue)

	return nil
}

func RefreshAllConfigs(db *gorm.DB) error {
	return LoadAllConfigs(db)
}

func SetConfig(db *gorm.DB, key, value string) error {
	result := db.Model(&model.SysConfig{}).Where("config_key = ?", key).Update("config_value", value)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("config key not found: %s", key)
	}
	return RefreshConfig(db, key)
}

func GetAllConfigs() map[string]string {
	result := make(map[string]string)
	configMu.RLock()
	defer configMu.RUnlock()

	configCache.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(string)
		return true
	})
	return result
}
