package model

import (
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

type AttendanceEvent struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"person_id"`
	Seq       int            `gorm:"not null" json:"seq"`
	EventDate utils.DateOnly `gorm:"type:date;not null" json:"event_date"`
	PunchTime string         `gorm:"type:varchar(32)" json:"punch_time"`
	EventType string         `gorm:"type:varchar(32);not null" json:"event_type"`
	SubType   string         `gorm:"type:varchar(32);not null" json:"sub_type"`
	Hours     float64        `gorm:"type:decimal(4,1);not null" json:"hours"`
	Remark    string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt utils.DateOnly `json:"created_at"`
	UpdatedAt utils.DateOnly `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AttendanceEvent) TableName() string { return "attendance_events" }
