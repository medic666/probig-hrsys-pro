package audit

import (
	"encoding/json"
	"time"

	"probig/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

type AuditLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OperatorID     uint      `gorm:"index" json:"operator_id"`
	OperatorName   string    `gorm:"size:64" json:"operator_name"`
	TargetType     string    `gorm:"size:64;index" json:"target_type"`
	TargetID       uint      `gorm:"index" json:"target_id"`
	TargetName     string    `gorm:"size:128" json:"target_name"`
	Action         string    `gorm:"size:32;index" json:"action"`
	BeforeSnapshot string    `gorm:"type:text" json:"before_snapshot"`
	AfterSnapshot  string    `gorm:"type:text" json:"after_snapshot"`
	IP             string    `gorm:"size:64" json:"ip"`
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

func Write(c *gin.Context, operatorID uint, operatorName, targetType string, targetID uint, targetName, action string, before, after interface{}) {
	var beforeStr, afterStr string
	if before != nil {
		b, _ := json.Marshal(before)
		beforeStr = string(b)
	}
	if after != nil {
		b, _ := json.Marshal(after)
		afterStr = string(b)
	}
	ip := c.ClientIP()
	go func() {
		config.DB.Create(&AuditLog{
			OperatorID:     operatorID,
			OperatorName:   operatorName,
			TargetType:     targetType,
			TargetID:       targetID,
			TargetName:     targetName,
			Action:         action,
			BeforeSnapshot: beforeStr,
			AfterSnapshot:  afterStr,
			IP:             ip,
		})
	}()
}

func WriteBatch(c *gin.Context, operatorID uint, operatorName string, logs []AuditLog) {
	ip := c.ClientIP()
	for i := range logs {
		logs[i].IP = ip
		logs[i].OperatorID = operatorID
		logs[i].OperatorName = operatorName
	}
	go func() {
		config.DB.Create(&logs)
	}()
}
