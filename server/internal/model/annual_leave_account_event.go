package model

import (
	"probig/server/internal/utils"
	"time"

	"gorm.io/gorm"
)

type AnnualLeaveAccountEvent struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	PersonID      uint           `gorm:"not null;index" json:"person_id"`
	Seq           int            `gorm:"not null" json:"seq"`
	EventType     string         `gorm:"type:varchar(32);not null" json:"event_type"`
	SourceType    string         `gorm:"type:varchar(16);default:manual" json:"source_type"`
	BatchID       *uint          `json:"batch_id"`
	Hours         float64        `gorm:"type:decimal(5,1);not null" json:"hours"`
	EffectiveDate utils.DateOnly `gorm:"type:date;not null" json:"effective_date"`
	Remark        string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AnnualLeaveAccountEvent) TableName() string { return "annual_leave_account_events" }
