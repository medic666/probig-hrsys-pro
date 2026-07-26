package file

import (
	"time"

	"gorm.io/gorm"
)

type FileModel struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileName   string         `gorm:"size:256;not null" json:"file_name"`
	FilePath   string         `gorm:"size:512" json:"file_path"`
	FileSize   int64          `json:"file_size"`
	FileType   string         `gorm:"size:64" json:"file_type"`
	UploadUser string         `gorm:"size:64" json:"upload_user"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type FileRelation struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileID     uint           `gorm:"index" json:"file_id"`
	TargetType string         `gorm:"size:64" json:"target_type"`
	TargetID   uint           `gorm:"index" json:"target_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
