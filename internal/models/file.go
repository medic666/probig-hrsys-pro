package models

import "time"

type File struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Filename     string    `gorm:"size:256;not null" json:"filename"`
	OriginalName string    `gorm:"size:256;not null" json:"original_name"`
	MimeType     string    `gorm:"size:128" json:"mime_type"`
	Size         int64     `json:"size"`
	StoragePath  string    `gorm:"size:512;not null" json:"-"`
	UploadedBy   uint      `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}
