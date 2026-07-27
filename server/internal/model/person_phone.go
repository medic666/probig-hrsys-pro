package model

import "gorm.io/gorm"

type PersonPhone struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"person_id"`
	PhoneType string         `gorm:"type:varchar(16);default:mobile" json:"phone_type"`
	Phone     string         `gorm:"type:varchar(32);not null" json:"phone"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PersonPhone) TableName() string { return "person_phones" }
