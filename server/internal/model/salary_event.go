package model

import (
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

type SalaryEvent struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	PersonID     uint           `gorm:"not null;index" json:"person_id"`
	Seq          int            `gorm:"not null" json:"seq"`
	BelongMonth  string         `gorm:"type:varchar(7);not null" json:"belong_month"`
	EventType    string         `gorm:"type:varchar(32);not null" json:"event_type"`
	Amount       float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Remark       string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt    utils.DateOnly `json:"created_at"`
	UpdatedAt    utils.DateOnly `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SalaryEvent) TableName() string { return "salary_events" }
