package model

import (
	"time"

	"gorm.io/gorm"
)

type AttendanceEventDetail struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	DailyID   uint           `gorm:"not null;index" json:"daily_id"`
	EventType string         `gorm:"type:varchar(32);not null" json:"event_type"`
	SubType   string         `gorm:"type:varchar(32);not null" json:"sub_type"`
	Hours     float64        `gorm:"type:decimal(4,1);not null" json:"hours"`
	Minutes   int            `gorm:"default:0" json:"minutes"`
	Remark    string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AttendanceEventDetail) TableName() string { return "attendance_event_details" }
