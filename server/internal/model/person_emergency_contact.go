package model

import "gorm.io/gorm"

type PersonEmergencyContact struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	PersonID     uint           `gorm:"not null;index" json:"person_id"`
	ContactName  string         `gorm:"type:varchar(64)" json:"contact_name"`
	ContactPhone string         `gorm:"type:varchar(32)" json:"contact_phone"`
	Sort         int            `gorm:"default:1" json:"sort"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PersonEmergencyContact) TableName() string { return "person_emergency_contacts" }
