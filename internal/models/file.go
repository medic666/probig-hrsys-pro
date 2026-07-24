package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"type:varchar(128)" json:"name"`
	Size       int64          `gorm:"type:bigint" json:"size"`
	MimeType   string         `gorm:"type:varchar(64)" json:"mime_type"`
	Content    []byte         `gorm:"type:blob" json:"-"`
	UploaderID uint           `json:"uploader_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (File) TableName() string {
	return "files"
}

type FileRelation struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileID     uint           `gorm:"not null;index" json:"file_id"`
	TargetType string         `gorm:"type:varchar(32);not null" json:"target_type"`
	TargetID   uint           `gorm:"not null;index" json:"target_id"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (FileRelation) TableName() string {
	return "file_relations"
}
