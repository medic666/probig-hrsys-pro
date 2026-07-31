package model

import (
	"probig/server/internal/utils"
	"time"
)

type AnnualLeaveBalanceSnapshot struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	PersonID           uint           `gorm:"not null;index" json:"person_id"`
	EffectiveStartDate utils.DateOnly `gorm:"type:date;not null" json:"effective_start_date"`
	EffectiveEndDate   utils.DateOnly `gorm:"type:date;not null" json:"effective_end_date"`
	BalanceHours       float64        `gorm:"type:decimal(5,1);default:0" json:"balance_hours"`
	LastCalcAt         utils.DateOnly `json:"last_calc_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (AnnualLeaveBalanceSnapshot) TableName() string { return "annual_leave_balance_snapshots" }
