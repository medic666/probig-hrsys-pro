package system

import (
	"probig/internal/pkg/config"

	"gorm.io/gorm"
)

type Service struct {
	dao *DAO
}

var globalService *Service

func NewService(db *gorm.DB) *Service {
	svc := &Service{dao: NewDAO(db)}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) GetAllConfigs() ([]SysConfig, error) {
	return s.dao.GetAll()
}

func (s *Service) UpdateConfig(key string, value string) error {
	if err := s.dao.Update(key, value); err != nil {
		return err
	}
	config.SetSysConfig(key, value)
	return nil
}

func (s *Service) LoadToCache() error {
	configs, err := s.dao.GetAll()
	if err != nil {
		return err
	}
	items := make([]config.SysConfigItem, len(configs))
	for i, cfg := range configs {
		items[i] = config.SysConfigItem{ConfigKey: cfg.ConfigKey, ConfigValue: cfg.ConfigValue}
	}
	config.LoadSysConfigs(items)
	return nil
}

func (s *Service) GetConfigValue(key string) string {
	return config.GetSysConfig(key)
}
