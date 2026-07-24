package models

import "time"

type FileAssociation struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	FileID     uint   `gorm:"index;not null" json:"file_id"`
	TargetType string `gorm:"size:32;not null" json:"target_type"` // entity / event
	TargetID   uint   `gorm:"index;not null" json:"target_id"`

	CreatedAt time.Time `json:"created_at"`
}
