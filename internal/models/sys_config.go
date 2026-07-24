package models

import "time"

type SysConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ConfigKey    string    `gorm:"type:varchar(64);uniqueIndex" json:"config_key"`
	ConfigValue  string    `gorm:"type:text" json:"config_value"`
	ConfigName   string    `gorm:"type:varchar(128)" json:"config_name"`
	ConfigDesc   string    `gorm:"type:varchar(256)" json:"config_desc"`
	ValueType    string    `gorm:"type:varchar(16)" json:"value_type"`
	OptionValues string    `gorm:"type:text" json:"option_values"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SysConfig) TableName() string {
	return "sys_config"
}
