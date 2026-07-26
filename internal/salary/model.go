package salary

import (
	"time"

	"gorm.io/gorm"
)

type SalaryEvent struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	PersonID     uint           `gorm:"index;not null" json:"person_id"`
	BelongMonth  string         `gorm:"size:7;index" json:"belong_month"`
	EventType    string         `gorm:"size:32" json:"event_type"`
	Amount       float64        `gorm:"type:decimal(10,2)" json:"amount"`
	EventName    string         `gorm:"size:128" json:"event_name"`
	Remark       string         `gorm:"size:256" json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type SalarySummary struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	PersonID                uint      `gorm:"index" json:"person_id"`
	BelongMonth             string    `gorm:"size:7;index" json:"belong_month"`
	SalaryDays              int       `json:"salary_days"`
	WeightedBaseSalary      float64   `gorm:"type:decimal(10,2)" json:"weighted_base_salary"`
	TotalWorkHours          float64   `gorm:"type:decimal(5,1)" json:"total_work_hours"`
	TotalOvertimeWorkdayH   float64   `gorm:"type:decimal(5,1)" json:"total_overtime_workday_hours"`
	TotalOvertimeHolidayH   float64   `gorm:"type:decimal(5,1)" json:"total_overtime_holiday_hours"`
	AttendanceSalary        float64   `gorm:"type:decimal(10,2)" json:"attendance_salary"`
	OvertimeWorkdaySalary   float64   `gorm:"type:decimal(10,2)" json:"overtime_workday_salary"`
	OvertimeHolidaySalary   float64   `gorm:"type:decimal(10,2)" json:"overtime_holiday_salary"`
	AnnualLeaveCarryoverSal float64   `gorm:"type:decimal(10,2)" json:"annual_leave_carryover_salary"`
	AttendanceBonus         float64   `gorm:"type:decimal(10,2)" json:"attendance_bonus"`
	PerformanceSalary       float64   `gorm:"type:decimal(10,2)" json:"performance_salary"`
	PostAllowance           float64   `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance           float64   `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance        float64   `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance      float64   `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance       float64   `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceComp           float64   `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundComp                float64   `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	TotalAdjustment         float64   `gorm:"type:decimal(10,2)" json:"total_adjustment"`
	SocialSecurityDeduct    float64   `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct       float64   `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	TaxDeduct               float64   `gorm:"type:decimal(10,2)" json:"tax_deduct"`
	FinalSalary             float64   `gorm:"type:decimal(10,2)" json:"final_salary"`
	LastCalcAt              *time.Time `json:"last_calc_at"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}
