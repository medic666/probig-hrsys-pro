package middleware

import (
	"github.com/gin-gonic/gin"
	"probig/internal/dao"
	"probig/pkg/response"
)

func RequirePermission(permKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetUserClaims(c)
		if claims == nil {
			response.Forbidden(c)
			return
		}
		perms, err := dao.GetUserPermissions(claims.UserID)
		if err != nil || !hasPerm(perms, permKey) {
			response.Forbidden(c)
			return
		}
		c.Next()
	}
}

func hasPerm(perms []string, key string) bool {
	for _, p := range perms {
		if p == key || p == "all.all" {
			return true
		}
	}
	return false
}
