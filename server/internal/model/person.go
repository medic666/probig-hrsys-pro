package model

import (
	"time"

	"gorm.io/gorm"
)

type Person struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	Name            string         `gorm:"type:varchar(64);not null" json:"name"`
	IDCard          string         `gorm:"type:varchar(32)" json:"id_card"`
	Gender          int8           `gorm:"default:0" json:"gender"`
	Birthday        *time.Time     `json:"birthday"`
	Nation          string         `gorm:"type:varchar(32)" json:"nation"`
	NativePlace     string         `gorm:"type:varchar(128)" json:"native_place"`
	Address         string         `gorm:"type:varchar(256)" json:"address"`
	PoliticalStatus string         `gorm:"type:varchar(32)" json:"political_status"`
	MaritalStatus   int8           `gorm:"default:0" json:"marital_status"`
	Alias           string         `gorm:"type:varchar(64)" json:"alias"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Phones    []PersonPhone    `gorm:"foreignKey:PersonID" json:"phones,omitempty"`
	Emails    []PersonEmail    `gorm:"foreignKey:PersonID" json:"emails,omitempty"`
	BankCards []PersonBankCard `gorm:"foreignKey:PersonID" json:"bank_cards,omitempty"`
}

func (Person) TableName() string { return "persons" }
