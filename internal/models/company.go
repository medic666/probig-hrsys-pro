package models

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(128);not null" json:"name"`
	CreditCode      string         `gorm:"type:varchar(64);uniqueIndex" json:"credit_code"`
	Address         string         `gorm:"type:varchar(256)" json:"address"`
	ContactPhone    string         `gorm:"type:varchar(32)" json:"contact_phone"`
	BankName        string         `gorm:"type:varchar(64)" json:"bank_name"`
	BankAccount     string         `gorm:"type:varchar(128)" json:"-"`
	BankAccountPlain string        `gorm:"-" json:"bank_account"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Company) TableName() string {
	return "companies"
}
