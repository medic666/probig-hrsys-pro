package audit

import (
	"encoding/json"
	"time"

	"probig/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

type AuditLog struct {
	ID             uint      `gorm:"primaryKey"`
	OperatorID     uint      `gorm:"index"`
	OperatorName   string    `gorm:"size:64"`
	TargetType     string    `gorm:"size:64;index"`
	TargetID       uint      `gorm:"index"`
	TargetName     string    `gorm:"size:128"`
	Action         string    `gorm:"size:32;index"`
	BeforeSnapshot string    `gorm:"type:text"`
	AfterSnapshot  string    `gorm:"type:text"`
	IP             string    `gorm:"size:64"`
	CreatedAt      time.Time `gorm:"index"`
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
