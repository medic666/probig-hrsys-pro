package models

import "time"

type Entity struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:16;not null;index" json:"type"` // person / organization
	Name      string    `gorm:"size:128;not null" json:"name"`
	Status    string    `gorm:"size:16;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
