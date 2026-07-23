package attendance

import "time"

type AttendanceEvent struct {
	ID            int64     `json:"id" db:"id"`
	PersonID      int64     `json:"personId" db:"person_id"`
	EventDate     string    `json:"eventDate" db:"event_date"`
	EventType     string    `json:"eventType" db:"event_type"`
	StartTime     string    `json:"startTime" db:"start_time"`
	EndTime       string    `json:"endTime" db:"end_time"`
	DurationHours float64   `json:"durationHours" db:"duration_hours"`
	Description   string    `json:"description" db:"description"`
	OperatorID    int64     `json:"operatorId" db:"operator_id"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

type AnnualLeaveGrant struct {
	ID            int64     `json:"id" db:"id"`
	PersonID      int64     `json:"personId" db:"person_id"`
	GrantDate     string    `json:"grantDate" db:"grant_date"`
	DaysGranted   float64   `json:"daysGranted" db:"days_granted"`
	DaysRemaining float64   `json:"daysRemaining" db:"days_remaining"`
	YearMonth     string    `json:"yearMonth" db:"year_month"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

type MonthlyLeaveBalance struct {
	TotalDays float64 `json:"totalDays"`
	UsedDays  float64 `json:"usedDays"`
	Remaining float64 `json:"remaining"`
}
