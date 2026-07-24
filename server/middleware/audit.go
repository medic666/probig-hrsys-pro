package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"probig/database"
	"probig/models"

	"github.com/gin-gonic/gin"
)

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		c.Next()
	}
}

func WriteAuditLog(c *gin.Context, targetType string, targetID uint, action string, before, after interface{}, batchID string) {
	beforeJSON := toJSON(before)
	afterJSON := toJSON(after)

	operatorID := uint(0)
	operatorName := "系统"
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint); ok {
			operatorID = id
		}
	}
	if uname, exists := c.Get("username"); exists {
		if name, ok := uname.(string); ok {
			operatorName = name
		}
	}

	if operatorID == 0 {
		operatorID = 1
		operatorName = "系统"
	}

	log := models.AuditLog{
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		TargetType:     targetType,
		TargetID:       targetID,
		Action:         action,
		BeforeSnapshot: beforeJSON,
		AfterSnapshot:  afterJSON,
		BatchID:        batchID,
		IP:             c.ClientIP(),
		CreatedAt:      time.Now(),
	}

	if err := database.DB.Create(&log).Error; err != nil {
		fmt.Printf("audit log write failed: %v\n", err)
	}
}

func AuditAction(c *gin.Context, targetType string, targetID uint, action string, before, after interface{}) {
	WriteAuditLog(c, targetType, targetID, action, before, after, "")
}

func AuditBatch(c *gin.Context, targetType string, targetID uint, action string, before, after interface{}, batchID string) {
	WriteAuditLog(c, targetType, targetID, action, before, after, batchID)
}

func toJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func GetUserID(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint); ok {
			return id
		}
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if uname, exists := c.Get("username"); exists {
		if name, ok := uname.(string); ok {
			return name
		}
	}
	return ""
}

func CheckConfigLocked(month, module string) error {
	switch module {
	case "attendance":
		var summary models.AttendanceSummary
		err := database.DB.Where("belong_month = ? AND is_locked = ?", month, true).First(&summary).Error
		if err == nil {
			return fmt.Errorf("月份 %s 的考勤已锁定，禁止修改假勤事件", month)
		}
	case "salary":
		var summary models.SalarySummary
		err := database.DB.Where("belong_month = ? AND is_locked = ?", month, true).First(&summary).Error
		if err == nil {
			return fmt.Errorf("月份 %s 的工资已锁定，禁止修改工资相关数据", month)
		}
	}
	return nil
}

func CheckConfigLockedBoth(month string) error {
	if err := CheckConfigLocked(month, "attendance"); err != nil {
		return err
	}
	if err := CheckConfigLocked(month, "salary"); err != nil {
		return err
	}
	return nil
}
