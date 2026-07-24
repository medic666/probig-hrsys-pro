package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type PositionEvent struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	PersonID             uint           `gorm:"not null;index" json:"person_id"`
	EventName            string         `gorm:"type:varchar(128)" json:"event_name"`
	EffectiveDate        *time.Time     `gorm:"type:date" json:"effective_date"`
	AttendanceGroup      *string        `gorm:"type:varchar(64)" json:"attendance_group"`
	EntryDate            *time.Time     `gorm:"type:date" json:"entry_date"`
	LeaveDate            *time.Time     `gorm:"type:date" json:"leave_date"`
	HasAnnualLeave       *bool          `gorm:"type:bool" json:"has_annual_leave"`
	HasAttendanceBonus   *bool          `gorm:"type:bool" json:"has_attendance_bonus"`
	BaseSalary           sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	PerformanceSalary    sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	SalaryDays           *int           `json:"salary_days"`
	PostAllowance        sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	MealAllowance        sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	HousingAllowance     sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	TransportAllowance   sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	HighTempAllowance    sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	InsuranceCompensation sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	FundCompensation     sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	SocialSecurityDeduct sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	HousingFundDeduct    sql.NullFloat64 `gorm:"type:decimal(10,2)" json:"-"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Person *Person `gorm:"foreignKey:PersonID" json:"person"`
}

func (PositionEvent) TableName() string {
	return "position_events"
}
