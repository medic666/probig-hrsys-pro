package attendance

import (
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

type AttendanceEvent = database.AttendanceEvent
type AttendanceDailyProjection = database.AttendanceDailyProjection
type AttendanceSalaryMonthly = database.AttendanceSalaryMonthly
type PositionSnapshot = database.PositionSnapshot
type Person = database.Person

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

type AttendanceEventWithName struct {
	AttendanceEvent
	PersonName string `json:"person_name"`
}

type DailyProjectionWithName struct {
	AttendanceDailyProjection
	PersonName string `json:"person_name"`
}

type MonthlySalaryWithName struct {
	AttendanceSalaryMonthly
	PersonName string `json:"person_name"`
}

type MonthlySalaryStatus struct {
	AttendanceSalaryMonthly
	PersonName string `json:"person_name"`
	Status     string `json:"status"`
}

var FarFutureDate = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
