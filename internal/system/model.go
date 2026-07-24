package system

import "gorm.io/gorm"

type SysConfig struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ConfigKey    string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"configKey"`
	ConfigValue  string         `gorm:"type:text" json:"configValue"`
	ConfigName   string         `gorm:"type:varchar(128)" json:"configName"`
	ConfigDesc   string         `gorm:"type:varchar(256)" json:"configDesc"`
	ValueType    string         `gorm:"type:varchar(16)" json:"valueType"`
	OptionValues string         `gorm:"type:text" json:"optionValues"`
	UpdatedAt    *string        `gorm:"type:datetime" json:"updatedAt"`
	CreatedAt    string         `gorm:"type:datetime" json:"createdAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SysConfig) TableName() string { return "sys_config" }
