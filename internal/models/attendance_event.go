package models

import (
	"time"

	"gorm.io/gorm"
)

type AttendanceEvent struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	PersonID           uint           `gorm:"not null;index" json:"person_id"`
	EventDate          *time.Time     `gorm:"type:date" json:"event_date"`
	EventType          string         `gorm:"type:varchar(32)" json:"event_type"`
	SubType            string         `gorm:"type:varchar(32)" json:"sub_type"`
	Hours              *float64       `gorm:"type:decimal(4,1)" json:"hours"`
	LateMinutes        *int           `json:"late_minutes"`
	LeaveAdjustAmount  *float64       `gorm:"type:decimal(4,1)" json:"leave_adjust_amount"`
	IsSpecialApproval  bool           `json:"is_special_approval"`
	Remark             string         `gorm:"type:varchar(256)" json:"remark"`
	BatchID            string         `gorm:"type:varchar(64)" json:"batch_id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Person *Person `gorm:"foreignKey:PersonID" json:"person"`
}

func (AttendanceEvent) TableName() string {
	return "attendance_events"
}
