package models

import (
	"time"

	"gorm.io/gorm"
)

type SalaryEvent struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PersonID    uint           `gorm:"not null;index" json:"person_id"`
	BelongMonth string         `gorm:"type:varchar(7)" json:"belong_month"`
	EventType   string         `gorm:"type:varchar(32)" json:"event_type"`
	Amount      float64        `gorm:"type:decimal(10,2)" json:"amount"`
	EventName   string         `gorm:"type:varchar(128)" json:"event_name"`
	Remark      string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Person *Person `gorm:"foreignKey:PersonID" json:"person"`
}

func (SalaryEvent) TableName() string {
	return "salary_events"
}
