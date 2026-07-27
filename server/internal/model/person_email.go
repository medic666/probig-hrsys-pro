package model

import "gorm.io/gorm"

type PersonEmail struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"person_id"`
	EmailType string         `gorm:"type:varchar(16);default:personal" json:"email_type"`
	Email     string         `gorm:"type:varchar(128);not null" json:"email"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PersonEmail) TableName() string { return "person_emails" }
