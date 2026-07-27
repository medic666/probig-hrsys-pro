package model

import "gorm.io/gorm"

type PersonBankCard struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	PersonID      uint           `gorm:"not null;index" json:"person_id"`
	BankName      string         `gorm:"type:varchar(64)" json:"bank_name"`
	AccountNumber string         `gorm:"type:varchar(64);not null" json:"account_number"`
	AccountHolder string         `gorm:"type:varchar(64)" json:"account_holder"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PersonBankCard) TableName() string { return "person_bank_cards" }
