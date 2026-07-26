package person

import (
	"time"

	"gorm.io/gorm"
)

type Person struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:64;not null" json:"name"`
	IDCard          string         `gorm:"size:256;uniqueIndex" json:"id_card"`
	Gender          int            `gorm:"default:0" json:"gender"`
	Birthday        *time.Time     `json:"birthday"`
	Nation          string         `gorm:"size:32" json:"nation"`
	NativePlace     string         `gorm:"size:128" json:"native_place"`
	Address         string         `gorm:"size:256" json:"address"`
	PoliticalStatus string         `gorm:"size:32" json:"political_status"`
	MaritalStatus   int            `gorm:"default:0" json:"marital_status"`
	Alias           string         `gorm:"size:64" json:"alias"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type PersonPhone struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"index;not null" json:"person_id"`
	Phone     string         `gorm:"size:256" json:"phone"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PersonEmail struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"index;not null" json:"person_id"`
	Email     string         `gorm:"size:256" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PersonBankCard struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"index;not null" json:"person_id"`
	CardNo    string         `gorm:"size:256" json:"card_no"`
	BankName  string         `gorm:"size:64" json:"bank_name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
