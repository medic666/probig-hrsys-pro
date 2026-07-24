package models

import (
	"time"

	"gorm.io/gorm"
)

type Person struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(64);not null" json:"name"`
	IDCard          string         `gorm:"type:varchar(64);uniqueIndex" json:"-"`
	IDCardPlain     string         `gorm:"-" json:"id_card"`
	Gender          int8           `gorm:"type:tinyint" json:"gender"`
	Birthday        *time.Time     `gorm:"type:date" json:"birthday"`
	Nation          string         `gorm:"type:varchar(32)" json:"nation"`
	NativePlace     string         `gorm:"type:varchar(128)" json:"native_place"`
	Address         string         `gorm:"type:varchar(256)" json:"address"`
	PoliticalStatus string         `gorm:"type:varchar(32)" json:"political_status"`
	MaritalStatus   int8           `gorm:"type:tinyint" json:"marital_status"`
	Alias           string         `gorm:"type:varchar(64)" json:"alias"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Phones    []PersonPhone    `gorm:"foreignKey:PersonID" json:"phones"`
	Emails    []PersonEmail    `gorm:"foreignKey:PersonID" json:"emails"`
	BankCards []PersonBankCard `gorm:"foreignKey:PersonID" json:"bank_cards"`
}

func (Person) TableName() string {
	return "persons"
}

type PersonPhone struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"person_id"`
	Phone     string         `gorm:"type:varchar(64)" json:"-"`
	PhonePlain string        `gorm:"-" json:"phone"`
	Remark    string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (PersonPhone) TableName() string {
	return "person_phones"
}

type PersonEmail struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	PersonID   uint           `gorm:"not null;index" json:"person_id"`
	Email      string         `gorm:"type:varchar(128)" json:"-"`
	EmailPlain string         `gorm:"-" json:"email"`
	Remark     string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (PersonEmail) TableName() string {
	return "person_emails"
}

type PersonBankCard struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	PersonID      uint           `gorm:"not null;index" json:"person_id"`
	BankCard      string         `gorm:"type:varchar(128)" json:"-"`
	BankCardPlain string         `gorm:"-" json:"bank_card"`
	BankName      string         `gorm:"type:varchar(64)" json:"bank_name"`
	Remark        string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (PersonBankCard) TableName() string {
	return "person_bank_cards"
}
