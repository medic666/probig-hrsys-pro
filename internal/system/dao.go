package system

import (
	"time"

	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) GetAll() ([]SysConfig, error) {
	var configs []SysConfig
	err := d.db.Find(&configs).Error
	return configs, err
}

func (d *DAO) GetByKey(key string) (*SysConfig, error) {
	var config SysConfig
	err := d.db.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (d *DAO) Update(key string, value string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	return d.db.Model(&SysConfig{}).Where("config_key = ?", key).Updates(map[string]interface{}{
		"config_value": value,
		"updated_at":   &now,
	}).Error
}
