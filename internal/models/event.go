package models

import (
	"time"
)

type PersonnelEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EntityID  uint      `gorm:"index;not null" json:"entity_id"`
	EventType string    `gorm:"size:16;not null" json:"event_type"` // create / update / delete
	EventName string    `gorm:"size:64" json:"event_name"`          // human-readable event name

	EffectiveDate time.Time `gorm:"not null;index" json:"effective_date"`

	Name                    string  `gorm:"size:64;not null" json:"name"`
	AttendanceGroup         string  `gorm:"size:64" json:"attendance_group"`
	HireDate                *string `gorm:"size:16" json:"hire_date"`
	BaseSalary              float64 `json:"base_salary"`
	PerformanceSalary       float64 `json:"performance_salary"`
	PayDays                 float64 `json:"pay_days"`
	PositionAllowance       float64 `json:"position_allowance"`
	MealSubsidy             float64 `json:"meal_subsidy"`
	HousingSubsidy          float64 `json:"housing_subsidy"`
	TransportSubsidy        float64 `json:"transport_subsidy"`
	HeatSubsidy             float64 `json:"heat_subsidy"`
	InsuranceCompensation   float64 `json:"insurance_compensation"`
	HousingFundCompensation float64 `json:"housing_fund_compensation"`
	SocialInsuranceDeduct   float64 `json:"social_insurance_deduct"`
	HousingFundDeduct       float64 `json:"housing_fund_deduct"`

	ExtendedInfo  JSONMap `gorm:"type:text" json:"extended_info"`
	ChangedFields JSONMap `gorm:"type:text" json:"changed_fields"`

	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EntityID  uint      `gorm:"index;not null" json:"entity_id"`
	EventType string    `gorm:"size:16;not null" json:"event_type"`
	EventName string    `gorm:"size:64" json:"event_name"`

	EffectiveDate time.Time `gorm:"not null;index" json:"effective_date"`

	CompanyName           string  `gorm:"size:128;not null" json:"company_name"`
	CreditCode            string  `gorm:"size:64" json:"credit_code"`
	Address               string  `gorm:"size:256" json:"address"`
	Phone                 string  `gorm:"size:32" json:"phone"`
	BankName              string  `gorm:"size:128" json:"bank_name"`
	BankAccount           string  `gorm:"size:64" json:"bank_account"`
	BusinessLicenseFileID *uint   `json:"business_license_file_id"`
	OfficialSealFileID    *uint   `json:"official_seal_file_id"`

	ChangedFields JSONMap `gorm:"type:text" json:"changed_fields"`

	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AttendanceEvent struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	EntityID      uint   `gorm:"index;not null" json:"entity_id"`
	EntityName    string `gorm:"-" json:"entity_name"`
	EventCategory string `gorm:"size:32;not null" json:"event_category"`
	EventSubtype  string `gorm:"size:32;not null" json:"event_subtype"`
	EventDate     string `gorm:"size:16;not null;index" json:"event_date"`
	DurationDays  float64 `gorm:"not null" json:"duration_days"`
	Description   string `gorm:"size:512" json:"description"`

	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SalaryEvent struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	EntityID    uint   `gorm:"index;not null" json:"entity_id"`
	EntityName  string `gorm:"-" json:"entity_name"`
	EventType   string `gorm:"size:32;not null" json:"event_type"`
	Amount      float64 `json:"amount"`
	Description string `gorm:"size:512" json:"description"`
	PeriodStart string `gorm:"size:16;not null;index" json:"period_start"`
	PeriodEnd   string `gorm:"size:16;not null" json:"period_end"`

	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
