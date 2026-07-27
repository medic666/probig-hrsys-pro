package model

import "time"

type AuditLog struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	OperatorID     uint      `gorm:"not null" json:"operator_id"`
	OperatorName   string    `gorm:"type:varchar(64)" json:"operator_name"`
	TargetType     string    `gorm:"type:varchar(64);not null" json:"target_type"`
	TargetID       uint      `gorm:"not null" json:"target_id"`
	TargetName     string    `gorm:"type:varchar(128)" json:"target_name"`
	Action         string    `gorm:"type:varchar(32);not null" json:"action"`
	BeforeSnapshot string    `gorm:"type:text" json:"before_snapshot"`
	AfterSnapshot  string    `gorm:"type:text" json:"after_snapshot"`
	IP             string    `gorm:"type:varchar(64)" json:"ip"`
	CreatedAt      time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
