package models

import "time"

type PersonnelSnapshot struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	EntityID      uint      `gorm:"index;not null" json:"entity_id"`
	EventID       uint      `gorm:"index" json:"event_id"`
	EffectiveDate time.Time `gorm:"not null;index" json:"effective_date"`
	IsLatest      bool      `gorm:"index" json:"is_latest"`

	Name                    string  `json:"name"`
	AttendanceGroup         string  `json:"attendance_group"`
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

	ExtendedInfo JSONMap `gorm:"type:text" json:"extended_info"`

	CreatedAt time.Time `json:"created_at"`
}

type OrganizationSnapshot struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	EntityID      uint      `gorm:"index;not null" json:"entity_id"`
	EventID       uint      `gorm:"index" json:"event_id"`
	EffectiveDate time.Time `gorm:"not null;index" json:"effective_date"`
	IsLatest      bool      `gorm:"index" json:"is_latest"`

	CompanyName           string `json:"company_name"`
	CreditCode            string `json:"credit_code"`
	Address               string `json:"address"`
	Phone                 string `json:"phone"`
	BankName              string `json:"bank_name"`
	BankAccount           string `json:"bank_account"`
	BusinessLicenseFileID *uint  `json:"business_license_file_id"`
	OfficialSealFileID    *uint  `json:"official_seal_file_id"`

	CreatedAt time.Time `json:"created_at"`
}
