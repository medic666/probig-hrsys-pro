package position

import (
	"time"

	"gorm.io/gorm"
)

type PositionEvent struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	PersonID            uint           `gorm:"index;not null" json:"person_id"`
	EventName           string         `gorm:"size:128" json:"event_name"`
	EffectiveDate       string         `gorm:"size:10" json:"effective_date"`
	AttendanceGroup     *string        `gorm:"size:64" json:"attendance_group"`
	HasAnnualLeave      *bool          `json:"has_annual_leave"`
	HasAttendanceBonus  *bool          `json:"has_attendance_bonus"`
	BaseSalary          *float64       `gorm:"type:decimal(10,2)" json:"base_salary"`
	PerformanceSalary   *float64       `gorm:"type:decimal(10,2)" json:"performance_salary"`
	SalaryDays          *int           `json:"salary_days"`
	PostAllowance       *float64       `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance       *float64       `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance    *float64       `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance  *float64       `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance   *float64       `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceComp       *float64       `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundComp            *float64       `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	SocialSecurityDeduct *float64      `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct   *float64       `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

type PositionSnapshot struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	PersonID            uint      `gorm:"index" json:"person_id"`
	EffectiveStartDate  string    `gorm:"size:10" json:"effective_start_date"`
	EffectiveEndDate    string    `gorm:"size:10" json:"effective_end_date"`
	EntryDate           string    `gorm:"size:10" json:"entry_date"`
	LeaveDate           *string   `gorm:"size:10" json:"leave_date"`
	AttendanceGroup     string    `gorm:"size:64" json:"attendance_group"`
	HasAnnualLeave      bool      `json:"has_annual_leave"`
	HasAttendanceBonus  bool      `json:"has_attendance_bonus"`
	BaseSalary          float64   `gorm:"type:decimal(10,2)" json:"base_salary"`
	PerformanceSalary   float64   `gorm:"type:decimal(10,2)" json:"performance_salary"`
	SalaryDays          int       `json:"salary_days"`
	PostAllowance       float64   `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance       float64   `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance    float64   `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance  float64   `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance   float64   `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceComp       float64   `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundComp            float64   `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	SocialSecurityDeduct float64  `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct   float64   `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	LastCalcAt          time.Time `json:"last_calc_at"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

const EventTypeEntry = "入职"
const EventTypeLeave = "离职"
