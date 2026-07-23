package auth

import (
	"net/http"
	"strings"

	"probig/internal/common"

	"github.com/gin-gonic/gin"
)

func GetUserClaims(c *gin.Context) *Claims {
	val, exists := c.Get("userClaims")
	if !exists {
		return nil
	}
	claims, ok := val.(*Claims)
	if !ok {
		return nil
	}
	return claims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.CodeUnauthorized, "未提供认证令牌")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.CodeUnauthorized, "认证格式错误")
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			common.ErrorWithStatus(c, http.StatusUnauthorized, common.CodeUnauthorized, "认证令牌无效或已过期")
			return
		}

		c.Set("userClaims", claims)
		c.Next()
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetUserClaims(c)
		if claims == nil {
			common.ErrorWithStatus(c, http.StatusForbidden, common.CodeForbidden, "无权限")
			return
		}

		perms := RolePermissions[claims.Role]
		hasPermission := false
		for _, p := range perms {
			if p == permission || p == "admin" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			common.ErrorWithStatus(c, http.StatusForbidden, common.CodeForbidden, "无权限: "+permission)
			return
		}

		c.Next()
	}
}
