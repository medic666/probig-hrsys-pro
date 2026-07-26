package leave_account

import (
	"time"

	"gorm.io/gorm"
)

type LeaveAccountEvent struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	PersonID      uint           `gorm:"index;not null" json:"person_id"`
	LeaveType     string         `gorm:"size:32" json:"leave_type"`
	EventType     string         `gorm:"size:32" json:"event_type"`
	SourceType    string         `gorm:"size:16" json:"source_type"`
	BatchID       *uint          `json:"batch_id"`
	Hours         float64        `gorm:"type:decimal(5,1)" json:"hours"`
	EffectiveDate string         `gorm:"size:10" json:"effective_date"`
	Remark        string         `gorm:"size:256" json:"remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type LeaveAccountBalance struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PersonID    uint      `gorm:"index" json:"person_id"`
	LeaveType   string    `gorm:"size:32" json:"leave_type"`
	BalanceHours float64  `gorm:"type:decimal(5,1)" json:"balance_hours"`
	LastCalcAt  time.Time `json:"last_calc_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	LeaveTypeAnnual   = "annual_leave"
	LeaveTypeTimeOff  = "time_off"

	EventTypeGrant           = "grant"
	EventTypeAdjust          = "adjust"
	EventTypeCarryoverDeduct = "carryover_deduct"
	EventTypeTimeOffAccrue   = "time_off_accrue"

	SourceManual       = "manual"
	SourceSystemPeriod = "system_period"
)
