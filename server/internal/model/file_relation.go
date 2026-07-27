package model

import "gorm.io/gorm"

type FileRelation struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	FileID     uint           `gorm:"not null;index" json:"file_id"`
	TargetType string         `gorm:"type:varchar(32);not null;index" json:"target_type"`
	TargetID   uint           `gorm:"not null;index" json:"target_id"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (FileRelation) TableName() string { return "file_relations" }
