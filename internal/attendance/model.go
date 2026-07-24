package attendance

import "gorm.io/gorm"

type AttendanceEvent struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	PersonID          uint           `gorm:"not null;index" json:"personId"`
	EventDate         string         `gorm:"type:date;not null" json:"eventDate"`
	EventType         string         `gorm:"type:varchar(32);not null" json:"eventType"`
	SubType           string         `gorm:"type:varchar(32)" json:"subType"`
	Hours             *float64       `gorm:"type:decimal(4,1)" json:"hours"`
	LateMinutes       int            `json:"lateMinutes"`
	LeaveAdjustAmount *float64       `gorm:"type:decimal(4,1)" json:"leaveAdjustAmount"`
	IsSpecialApproval bool           `json:"isSpecialApproval"`
	Remark            string         `gorm:"type:varchar(256)" json:"remark"`
	BatchID           string         `gorm:"type:varchar(64)" json:"batchId"`
	CreatedAt         string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt         string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AttendanceEvent) TableName() string { return "attendance_events" }

type AttendanceSummary struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"not null;index" json:"personId"`
	BelongMonth           string         `gorm:"type:varchar(7);not null" json:"belongMonth"`
	WorkDays              float64        `gorm:"type:decimal(4,1)" json:"workDays"`
	MakeUpDays            float64        `gorm:"type:decimal(4,1)" json:"makeUpDays"`
	SickLeaveDays         float64        `gorm:"type:decimal(4,1)" json:"sickLeaveDays"`
	PersonalLeaveDays     float64        `gorm:"type:decimal(4,1)" json:"personalLeaveDays"`
	AnnualLeaveDays       float64        `gorm:"type:decimal(4,1)" json:"annualLeaveDays"`
	StatutoryLeaveDays    float64        `gorm:"type:decimal(4,1)" json:"statutoryLeaveDays"`
	WelfareLeaveDays      float64        `gorm:"type:decimal(4,1)" json:"welfareLeaveDays"`
	OvertimeWorkdayHours  float64        `gorm:"type:decimal(5,1)" json:"overtimeWorkdayHours"`
	OvertimeHolidayHours  float64        `gorm:"type:decimal(5,1)" json:"overtimeHolidayHours"`
	ViolationCount        int            `json:"violationCount"`
	LastCalcAt            *string        `gorm:"type:datetime" json:"lastCalcAt"`
	IsLocked              bool           `gorm:"default:false" json:"isLocked"`
	CreatedAt             string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt             string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AttendanceSummary) TableName() string { return "attendance_summaries" }
