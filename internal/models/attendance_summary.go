package models

import (
	"time"

	"gorm.io/gorm"
)

type AttendanceSummary struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	PersonID            uint           `gorm:"not null;index" json:"person_id"`
	BelongMonth         string         `gorm:"type:varchar(7)" json:"belong_month"`
	WorkDays            float64        `gorm:"type:decimal(4,1)" json:"work_days"`
	MakeUpDays          float64        `gorm:"type:decimal(4,1)" json:"make_up_days"`
	SickLeaveDays       float64        `gorm:"type:decimal(4,1)" json:"sick_leave_days"`
	PersonalLeaveDays   float64        `gorm:"type:decimal(4,1)" json:"personal_leave_days"`
	AnnualLeaveDays     float64        `gorm:"type:decimal(4,1)" json:"annual_leave_days"`
	StatutoryLeaveDays  float64        `gorm:"type:decimal(4,1)" json:"statutory_leave_days"`
	WelfareLeaveDays    float64        `gorm:"type:decimal(4,1)" json:"welfare_leave_days"`
	OvertimeWorkdayHours float64       `gorm:"type:decimal(5,1)" json:"overtime_workday_hours"`
	OvertimeHolidayHours float64       `gorm:"type:decimal(5,1)" json:"overtime_holiday_hours"`
	ViolationCount      int            `json:"violation_count"`
	LastCalcAt          *time.Time     `json:"last_calc_at"`
	IsLocked            bool           `json:"is_locked"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (AttendanceSummary) TableName() string {
	return "attendance_summaries"
}
