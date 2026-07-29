package model

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"type:varchar(256);not null" json:"name"`
	OriginalName string         `gorm:"type:varchar(256)" json:"original_name"`
	Path         string         `gorm:"type:varchar(512);not null" json:"path"`
	Size         int64          `gorm:"default:0" json:"size"`
	MimeType     string         `gorm:"type:varchar(128)" json:"mime_type"`
	MD5          string         `gorm:"type:varchar(64)" json:"md5"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (File) TableName() string { return "files" }
