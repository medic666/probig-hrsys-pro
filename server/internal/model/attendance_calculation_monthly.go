package model

import (
	"time"
)

type AttendanceCalculationMonthly struct {
	ID                        uint      `gorm:"primarykey" json:"id"`
	PersonID                  uint      `gorm:"not null;index" json:"person_id"`
	PersonName                string    `gorm:"-" json:"person_name"`
	BelongMonth               string    `gorm:"type:varchar(7);not null" json:"belong_month"`
	SalaryDays                float64   `gorm:"type:decimal(4,1);default:0" json:"salary_days"`
	WeightedBaseSalary        float64   `gorm:"type:decimal(10,2);default:0" json:"weighted_base_salary"`
	WeightedMealAllowance     float64   `gorm:"type:decimal(10,2);default:0" json:"weighted_meal_allowance"`
	TotalWorkHours            float64   `gorm:"type:decimal(5,1);default:0" json:"total_work_hours"`
	TotalOvertimeWorkdayHours float64   `gorm:"type:decimal(5,1);default:0" json:"total_overtime_workday_hours"`
	TotalOvertimeHolidayHours float64   `gorm:"type:decimal(5,1);default:0" json:"total_overtime_holiday_hours"`
	AttendanceSalary          float64   `gorm:"type:decimal(10,2);default:0" json:"attendance_salary"`
	OvertimeWorkdaySalary     float64   `gorm:"type:decimal(10,2);default:0" json:"overtime_workday_salary"`
	OvertimeHolidaySalary     float64   `gorm:"type:decimal(10,2);default:0" json:"overtime_holiday_salary"`
	HasPersonalLeaveMonth     bool      `gorm:"default:false" json:"has_personal_leave_month"`
	TotalViolationCount       int       `gorm:"default:0" json:"total_violation_count"`
	AttendanceBonus           float64   `gorm:"type:decimal(10,2);default:0" json:"attendance_bonus"`
	LastCalcAt                time.Time `json:"last_calc_at"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

func (AttendanceCalculationMonthly) TableName() string { return "attendance_calculation_monthly" }
