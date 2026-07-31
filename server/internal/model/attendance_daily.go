package model

import (
	"probig/server/internal/utils"
	"time"

	"gorm.io/gorm"
)

type AttendanceDaily struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	PersonID  uint           `gorm:"not null;index:idx_daily_person_date" json:"person_id"`
	EventDate utils.DateOnly `gorm:"type:date;not null;index:idx_daily_person_date" json:"event_date"`
	Status    string         `gorm:"type:varchar(16);default:pending" json:"status"`
	PunchTime string         `gorm:"type:varchar(32)" json:"punch_time"`
	Remark    string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Details []AttendanceEventDetail `gorm:"foreignKey:DailyID" json:"details,omitempty"`
}

func (AttendanceDaily) TableName() string { return "attendance_daily" }
