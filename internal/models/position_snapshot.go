package models

import (
	"time"

	"gorm.io/gorm"
)

type PositionSnapshot struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"not null;index" json:"person_id"`
	SnapshotDate          *time.Time     `gorm:"type:date" json:"snapshot_date"`
	AttendanceGroup       string         `gorm:"type:varchar(64)" json:"attendance_group"`
	EntryDate             *time.Time     `gorm:"type:date" json:"entry_date"`
	LeaveDate             *time.Time     `gorm:"type:date" json:"leave_date"`
	HasAnnualLeave        bool           `json:"has_annual_leave"`
	HasAttendanceBonus    bool           `json:"has_attendance_bonus"`
	BaseSalary            float64        `gorm:"type:decimal(10,2)" json:"base_salary"`
	PerformanceSalary     float64        `gorm:"type:decimal(10,2)" json:"performance_salary"`
	SalaryDays            int            `json:"salary_days"`
	PostAllowance         float64        `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance         float64        `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance      float64        `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance    float64        `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance     float64        `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceCompensation float64        `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundCompensation      float64        `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	SocialSecurityDeduct  float64        `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct     float64        `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (PositionSnapshot) TableName() string {
	return "position_snapshots"
}
