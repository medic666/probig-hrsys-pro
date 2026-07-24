package models

import (
	"time"

	"gorm.io/gorm"
)

type SalarySummary struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	PersonID           uint           `gorm:"not null;index" json:"person_id"`
	BelongMonth        string         `gorm:"type:varchar(7)" json:"belong_month"`
	AttendanceSalary   float64        `gorm:"type:decimal(10,2)" json:"attendance_salary"`
	OvertimeSalary     float64        `gorm:"type:decimal(10,2)" json:"overtime_salary"`
	AttendanceBonus    float64        `gorm:"type:decimal(10,2)" json:"attendance_bonus"`
	PerformanceSalary  float64        `gorm:"type:decimal(10,2)" json:"performance_salary"`
	TotalAllowance     float64        `gorm:"type:decimal(10,2)" json:"total_allowance"`
	TotalAdjustment    float64        `gorm:"type:decimal(10,2)" json:"total_adjustment"`
	TotalDeduction     float64        `gorm:"type:decimal(10,2)" json:"total_deduction"`
	FinalSalary        float64        `gorm:"type:decimal(10,2)" json:"final_salary"`
	LastCalcAt         *time.Time     `json:"last_calc_at"`
	IsLocked           bool           `json:"is_locked"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (SalarySummary) TableName() string {
	return "salary_summaries"
}
