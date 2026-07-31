package model

import (
	"probig/server/internal/utils"
	"time"
)

type AttendanceDailyProjection struct {
	ID                      uint           `gorm:"primarykey" json:"id"`
	PersonID                uint           `gorm:"not null;index" json:"person_id"`
	WorkDate                utils.DateOnly `gorm:"type:date;not null" json:"work_date"`
	PunchTime               string         `gorm:"type:varchar(32)" json:"punch_time"`
	WorkHours               float64        `gorm:"type:decimal(4,1);default:0" json:"work_hours"`
	OvertimeWorkdayHours    float64        `gorm:"type:decimal(4,1);default:0" json:"overtime_workday_hours"`
	OvertimeHolidayHours    float64        `gorm:"type:decimal(4,1);default:0" json:"overtime_holiday_hours"`
	HasPersonalLeave        bool           `gorm:"default:false" json:"has_personal_leave"`
	ViolationCount          int            `gorm:"default:0" json:"violation_count"`
	Remark                  string         `gorm:"type:varchar(256)" json:"remark"`
	Status                  string         `gorm:"type:varchar(16);default:confirmed" json:"status"`
	LastCalcAt              utils.DateOnly `json:"last_calc_at"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

func (AttendanceDailyProjection) TableName() string { return "attendance_daily_projections" }
