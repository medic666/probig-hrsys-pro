package auditlog

import (
	"time"

	"probig/internal/pkg/database"
)

type AuditLog = database.AuditLog

type AuditLogFilter struct {
	OperatorID *uint   `form:"operator_id"`
	Action     string  `form:"action"`
	TargetType string  `form:"target_type"`
	StartDate  string  `form:"start_date"`
	EndDate    string  `form:"end_date"`
	PageNum    int     `form:"page_num"`
	PageSize   int     `form:"page_size"`
}

type AuditLogVO struct {
	ID           uint      `json:"id"`
	OperatorID   uint      `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	TargetType   string    `json:"target_type"`
	TargetID     uint      `json:"target_id"`
	TargetName   string    `json:"target_name"`
	Action       string    `json:"action"`
	IP           string    `json:"ip"`
	CreatedAt    time.Time `json:"created_at"`
}
