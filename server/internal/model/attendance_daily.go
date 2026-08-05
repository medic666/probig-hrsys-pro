package model

import (
	"probig/server/internal/utils"
	"time"

	"gorm.io/gorm"
)

type AttendanceDaily struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	PersonID  uint           `gorm:"not null;index:idx_daily_person_date,priority:1" json:"person_id"`
	EventDate utils.DateOnly `gorm:"type:date;not null;index:idx_daily_person_date,priority:2" json:"event_date"`
	// Seq 当日版本序号：同人同日允许多条记录，seq 最大者为当日有效记录（唯一 confirmed 必为其）；
	// 新录入/导入追加 MAX+1，编辑/确认就地提升为 MAX+1（已是最大则不变），其它组一律降级 pending
	Seq       int            `gorm:"not null;default:1;index:idx_daily_person_date,priority:3" json:"seq"`
	Status    string         `gorm:"type:varchar(16);default:pending" json:"status"`
	PunchTime string         `gorm:"type:varchar(32)" json:"punch_time"`
	Remark    string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Details []AttendanceEventDetail `gorm:"foreignKey:DailyID" json:"details,omitempty"`
}

func (AttendanceDaily) TableName() string { return "attendance_daily" }
