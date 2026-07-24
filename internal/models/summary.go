package models

import "time"

type AttendanceSummary struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	EntityID          uint      `gorm:"index;not null" json:"entity_id"`
	EntityName        string    `gorm:"-" json:"entity_name"`
	PeriodStart       string    `gorm:"size:16;not null;index" json:"period_start"`
	PeriodEnd         string    `gorm:"size:16;not null" json:"period_end"`

	NormalDays        float64 `json:"normal_days"`
	MakeupDays        float64 `json:"makeup_days"`
	LieuDays          float64 `json:"lieu_days"`
	PersonalDays      float64 `json:"personal_days"`
	SickDays          float64 `json:"sick_days"`
	AnnualDays        float64 `json:"annual_days"`
	StatutoryDays     float64 `json:"statutory_days"`
	WelfareDays       float64 `json:"welfare_days"`
	WorkdayOvertime   float64 `json:"workday_overtime"`
	HolidayOvertime   float64 `json:"holiday_overtime"`
	MissingCardCount  float64 `json:"missing_card_count"`
	LateCount         float64 `json:"late_count"`
	EarlyCount        float64 `json:"early_count"`
	AnnualAllocated   float64 `json:"annual_allocated"`
	AnnualCarriedOver  float64 `json:"annual_carried_over"`

	CalculatedAt time.Time `json:"calculated_at"`
}

type SalarySummary struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	EntityID    uint   `gorm:"index;not null" json:"entity_id"`
	EntityName  string `gorm:"-" json:"entity_name"`
	PeriodStart string `gorm:"size:16;not null;index" json:"period_start"`
	PeriodEnd   string `gorm:"size:16;not null" json:"period_end"`

	BaseSalary              float64 `json:"base_salary"`
	DailySalary             float64 `json:"daily_salary"`
	AttendanceWage          float64 `json:"attendance_wage"`
	FullAttendanceBonus     float64 `json:"full_attendance_bonus"`
	OvertimeWage            float64 `json:"overtime_wage"`
	PerformanceSalary       float64 `json:"performance_salary"`
	PositionAllowance       float64 `json:"position_allowance"`
	MealSubsidy             float64 `json:"meal_subsidy"`
	HousingSubsidy          float64 `json:"housing_subsidy"`
	TransportSubsidy        float64 `json:"transport_subsidy"`
	HeatSubsidy             float64 `json:"heat_subsidy"`
	InsuranceCompensation   float64 `json:"insurance_compensation"`
	HousingFundCompensation float64 `json:"housing_fund_compensation"`
	PerformanceAdjustment   float64 `json:"performance_adjustment"`
	RewardPunishment        float64 `json:"reward_punishment"`
	LoanDeduction           float64 `json:"loan_deduction"`
	SocialInsuranceDeduct   float64 `json:"social_insurance_deduct"`
	HousingFundDeduct       float64 `json:"housing_fund_deduct"`
	TaxDeduction            float64 `json:"tax_deduction"`
	GrossPay                float64 `json:"gross_pay"`
	NetPay                  float64 `json:"net_pay"`

	CalculatedAt time.Time `json:"calculated_at"`
}
