package models

import (
	"time"
)

type AuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index" json:"user_id"`
	Username      string    `gorm:"size:64" json:"username"`
	Action        string    `gorm:"size:16;not null" json:"action"`
	TargetType    string    `gorm:"size:32;not null" json:"target_type"`
	TargetID      uint      `json:"target_id"`
	TargetName    string    `gorm:"size:128" json:"target_name"`
	TargetSummary string    `gorm:"size:256" json:"target_summary"`
	Payload       JSONMap   `gorm:"type:text" json:"payload"`
	CreatedAt     time.Time `json:"created_at"`
}
