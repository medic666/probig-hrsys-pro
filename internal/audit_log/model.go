package audit_log

type AuditLog struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	OperatorID     uint   `gorm:"not null" json:"operatorId"`
	OperatorName   string `gorm:"type:varchar(64)" json:"operatorName"`
	TargetType     string `gorm:"type:varchar(64)" json:"targetType"`
	TargetID       uint   `json:"targetId"`
	Action         string `gorm:"type:varchar(32)" json:"action"`
	BeforeSnapshot string `gorm:"type:text" json:"beforeSnapshot"`
	AfterSnapshot  string `gorm:"type:text" json:"afterSnapshot"`
	BatchID        string `gorm:"type:varchar(64)" json:"batchId"`
	IP             string `gorm:"type:varchar(64)" json:"ip"`
	CreatedAt      string `gorm:"type:datetime" json:"createdAt"`
}

func (AuditLog) TableName() string { return "audit_logs" }
