package middleware

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"probig/internal/dao"
	"probig/internal/models"
)

func AuditLog(action, targetType string, targetID uint, before, after interface{}, batchID string) func(c *gin.Context) {
	return func(c *gin.Context) {
		claims := GetUserClaims(c)
		operatorID := uint(0)
		operatorName := "system"
		if claims != nil {
			operatorID = claims.UserID
			operatorName = claims.Username
		}
		beforeJSON := marshalJSON(before)
		afterJSON := marshalJSON(after)
		log := &models.AuditLog{
			OperatorID:     operatorID,
			OperatorName:   operatorName,
			TargetType:     targetType,
			TargetID:       targetID,
			Action:         action,
			BeforeSnapshot: beforeJSON,
			AfterSnapshot:  afterJSON,
			BatchID:        batchID,
			IP:             c.ClientIP(),
		}
		dao.CreateAuditLog(log)
		c.Next()
	}
}

func RecordAudit(c *gin.Context, action, targetType string, targetID uint, before, after interface{}, batchID string) {
	claims := GetUserClaims(c)
	operatorID := uint(0)
	operatorName := "system"
	if claims != nil {
		operatorID = claims.UserID
		operatorName = claims.Username
	}
	dao.CreateAuditLog(&models.AuditLog{
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		TargetType:     targetType,
		TargetID:       targetID,
		Action:         action,
		BeforeSnapshot: marshalJSON(before),
		AfterSnapshot:  marshalJSON(after),
		BatchID:        batchID,
		IP:             c.ClientIP(),
	})
}

func RecordAuditBatch(c *gin.Context, action, targetType string, before, after interface{}, batchID string) {
	RecordAudit(c, action, targetType, 0, before, after, batchID)
}

func marshalJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
