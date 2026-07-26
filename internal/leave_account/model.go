package leave_account

import (
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

type LeaveAccountEvent = database.LeaveAccountEvent
type LeaveAccountBalance = database.LeaveAccountBalance
type AttendanceEvent = database.AttendanceEvent
type PositionSnapshot = database.PositionSnapshot
type Person = database.Person
type SysBatch = database.SysBatch

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

type LeaveAccountEventWithName struct {
	LeaveAccountEvent
	PersonName string `json:"person_name"`
}

type LeaveAccountBalanceWithName struct {
	LeaveAccountBalance
	PersonName string `json:"person_name"`
}

type BalanceDetail struct {
	PersonID      uint    `json:"person_id"`
	PersonName    string  `json:"person_name"`
	LeaveType     string  `json:"leave_type"`
	BalanceHours  float64 `json:"balance_hours"`
	TotalAccrued  float64 `json:"total_accrued"`
	TotalTaken    float64 `json:"total_taken"`
	TotalAdjusted float64 `json:"total_adjusted"`
	TotalCarryover float64 `json:"total_carryover"`
}

var FarFutureDate = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
