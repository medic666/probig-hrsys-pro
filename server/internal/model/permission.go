package model

import "time"

type Permission struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Module    string    `gorm:"type:varchar(64);not null" json:"module"`
	Action    string    `gorm:"type:varchar(32);not null" json:"action"`
	Name      string    `gorm:"type:varchar(128);not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (Permission) TableName() string {
	return "permissions"
}
