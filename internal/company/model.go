package company

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	CreditCode   string         `gorm:"size:64;uniqueIndex" json:"credit_code"`
	Address      string         `gorm:"size:256" json:"address"`
	ContactPhone string         `gorm:"size:32" json:"contact_phone"`
	BankName     string         `gorm:"size:64" json:"bank_name"`
	BankAccount  string         `gorm:"size:256" json:"bank_account"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
