package models

import "time"

type AuditLog struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	OperatorID     uint   `json:"operator_id"`
	OperatorName   string `gorm:"type:varchar(64)" json:"operator_name"`
	TargetType     string `gorm:"type:varchar(64)" json:"target_type"`
	TargetID       uint   `json:"target_id"`
	Action         string `gorm:"type:varchar(32)" json:"action"`
	BeforeSnapshot string `gorm:"type:text" json:"before_snapshot"`
	AfterSnapshot  string `gorm:"type:text" json:"after_snapshot"`
	BatchID        string `gorm:"type:varchar(64)" json:"batch_id"`
	IP             string `gorm:"type:varchar(64)" json:"ip"`
	CreatedAt      time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
