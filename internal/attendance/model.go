package attendance

import (
	"time"

	"gorm.io/gorm"
)

type AttendanceEvent struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	PersonID         uint           `gorm:"index;not null" json:"person_id"`
	EventDate        string         `gorm:"size:10;index" json:"event_date"`
	PunchTime        string         `gorm:"size:32" json:"punch_time"`
	EventType        string         `gorm:"size:32" json:"event_type"`
	SubType          string         `gorm:"size:32" json:"sub_type"`
	Hours            float64        `gorm:"type:decimal(4,1)" json:"hours"`
	LateMinutes      int            `json:"late_minutes"`
	IsSpecialApproval bool          `gorm:"default:false" json:"is_special_approval"`
	Remark           string         `gorm:"size:256" json:"remark"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type AttendanceDaily struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	PersonID              uint      `gorm:"index" json:"person_id"`
	WorkDate              string    `gorm:"size:10;index" json:"work_date"`
	PunchTime             string    `gorm:"size:32" json:"punch_time"`
	WorkHours             float64   `gorm:"type:decimal(4,1)" json:"work_hours"`
	OvertimeWorkdayHours  float64   `gorm:"type:decimal(4,1)" json:"overtime_workday_hours"`
	OvertimeHolidayHours  float64   `gorm:"type:decimal(4,1)" json:"overtime_holiday_hours"`
	HasPersonalLeave      bool      `json:"has_personal_leave"`
	ViolationCount        int       `json:"violation_count"`
	Remark                string    `gorm:"size:256" json:"remark"`
	LastCalcAt            time.Time `json:"last_calc_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type AttendanceSalary struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	PersonID                uint      `gorm:"index" json:"person_id"`
	BelongMonth             string    `gorm:"size:7;index" json:"belong_month"`
	SalaryDays              int       `json:"salary_days"`
	WeightedBaseSalary      float64   `gorm:"type:decimal(10,2)" json:"weighted_base_salary"`
	WeightedMealAllowance   float64   `gorm:"type:decimal(10,2)" json:"weighted_meal_allowance"`
	TotalWorkHours          float64   `gorm:"type:decimal(5,1)" json:"total_work_hours"`
	TotalOvertimeWorkdayH   float64   `gorm:"type:decimal(5,1)" json:"total_overtime_workday_hours"`
	TotalOvertimeHolidayH   float64   `gorm:"type:decimal(5,1)" json:"total_overtime_holiday_hours"`
	AttendanceSalary        float64   `gorm:"type:decimal(10,2)" json:"attendance_salary"`
	OvertimeWorkdaySalary   float64   `gorm:"type:decimal(10,2)" json:"overtime_workday_salary"`
	OvertimeHolidaySalary   float64   `gorm:"type:decimal(10,2)" json:"overtime_holiday_salary"`
	HasPersonalLeaveMonth   bool      `json:"has_personal_leave_month"`
	TotalViolationCount     int       `json:"total_violation_count"`
	AttendanceBonus         float64   `gorm:"type:decimal(10,2)" json:"attendance_bonus"`
	LastCalcAt              *time.Time `json:"last_calc_at"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

const (
	EventTypeAttendance = "出勤"
	EventTypeLeave      = "休假"
	EventTypeOvertime   = "加班"
	EventTypeViolation  = "违纪"
)
