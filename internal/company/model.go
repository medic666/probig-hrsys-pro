package company

import "gorm.io/gorm"

type Company struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(128);not null" json:"name"`
	CreditCode   string         `gorm:"type:varchar(256);uniqueIndex" json:"creditCode"`
	Address      string         `gorm:"type:varchar(256)" json:"address"`
	ContactPhone string         `gorm:"type:varchar(32)" json:"contactPhone"`
	BankName     string         `gorm:"type:varchar(64)" json:"bankName"`
	BankAccount  string         `gorm:"type:varchar(256)" json:"bankAccount"`
	CreatedAt    string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt    string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Company) TableName() string { return "companies" }
