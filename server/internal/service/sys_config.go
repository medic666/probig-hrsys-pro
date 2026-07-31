package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"probig/server/internal/dao"
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

func SetConfig(ctx context.Context, key, value string) error {
	db := dao.DBFromContext(ctx)
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

type SystemConfigItem struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	Desc      string   `json:"desc"`
	ValueType string   `json:"value_type"`
	Options   []string `json:"options,omitempty"`
	Value     string   `json:"value"`
	Group     string   `json:"group"`
}

func configGroup(key string) string {
	switch {
	case strings.HasPrefix(key, "system."):
		return "系统"
	case strings.HasPrefix(key, "attendance."):
		return "考勤"
	case strings.HasPrefix(key, "annual_leave."):
		return "年假"
	}
	return "其他"
}

// GetSystemConfigItems 返回带完整元数据的配置列表，前端按元数据驱动渲染
func GetSystemConfigItems() ([]SystemConfigItem, error) {
	var configs []model.SysConfig
	if err := dao.DB.Order("id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	result := make([]SystemConfigItem, 0, len(configs))
	for _, cfg := range configs {
		item := SystemConfigItem{
			Key:       cfg.ConfigKey,
			Name:      cfg.ConfigName,
			Desc:      cfg.ConfigDesc,
			ValueType: cfg.ValueType,
			Value:     cfg.ConfigValue,
			Group:     configGroup(cfg.ConfigKey),
		}
		if cfg.OptionValues != "" {
			var opts []string
			if err := json.Unmarshal([]byte(cfg.OptionValues), &opts); err == nil {
				item.Options = opts
			}
		}
		result = append(result, item)
	}
	return result, nil
}
