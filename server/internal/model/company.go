package model

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `gorm:"type:varchar(128);not null" json:"name"`
	CreditCode   string         `gorm:"type:varchar(64)" json:"credit_code"`
	Address      string         `gorm:"type:varchar(256)" json:"address"`
	ContactPhone string         `gorm:"type:varchar(32)" json:"contact_phone"`
	BankName     string         `gorm:"type:varchar(64)" json:"bank_name"`
	BankAccount  string         `gorm:"type:varchar(64)" json:"bank_account"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Company) TableName() string { return "companies" }
