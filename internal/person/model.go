package person

import "gorm.io/gorm"

type Person struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	Name            string           `gorm:"type:varchar(64);not null" json:"name"`
	IDCard          string           `gorm:"type:varchar(256);uniqueIndex" json:"idCard"`
	Gender          *int             `json:"gender"`
	Birthday        *string          `gorm:"type:date" json:"birthday"`
	Nation          string           `gorm:"type:varchar(32)" json:"nation"`
	NativePlace     string           `gorm:"type:varchar(128)" json:"nativePlace"`
	Address         string           `gorm:"type:varchar(256)" json:"address"`
	PoliticalStatus string           `gorm:"type:varchar(32)" json:"politicalStatus"`
	MaritalStatus   *int             `json:"maritalStatus"`
	Alias           string           `gorm:"type:varchar(64)" json:"alias"`
	Phones          []PersonPhone    `gorm:"foreignKey:PersonID" json:"phones,omitempty"`
	Emails          []PersonEmail    `gorm:"foreignKey:PersonID" json:"emails,omitempty"`
	BankCards       []PersonBankCard `gorm:"foreignKey:PersonID" json:"bankCards,omitempty"`
	CreatedAt       string           `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt       string           `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (Person) TableName() string { return "persons" }

type PersonPhone struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"personId"`
	Phone     string         `gorm:"type:varchar(256)" json:"phone"`
	Remark    string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PersonPhone) TableName() string { return "person_phones" }

type PersonEmail struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"personId"`
	Email     string         `gorm:"type:varchar(256)" json:"email"`
	Remark    string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PersonEmail) TableName() string { return "person_emails" }

type PersonBankCard struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"personId"`
	BankCard  string         `gorm:"type:varchar(256)" json:"bankCard"`
	BankName  string         `gorm:"type:varchar(64)" json:"bankName"`
	Remark    string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PersonBankCard) TableName() string { return "person_bank_cards" }
