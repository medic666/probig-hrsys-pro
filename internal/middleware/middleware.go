package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/medic666/probig/internal/auth"
	"github.com/medic666/probig/internal/response"
)

func AuthRequired(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}
		claims, err := jwtMgr.Parse(parts[1])
		if err != nil {
			response.Unauthorized(c, "认证令牌无效或已过期")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_id", claims.RoleID)
		c.Set("role_name", claims.RoleName)
		c.Set("perms", claims.Perms)
		c.Next()
	}
}

func RBAC(module, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms, exists := c.Get("perms")
		if !exists {
			c.JSON(http.StatusForbidden, response.Response{Code: -1, Message: "无权限"})
			c.Abort()
			return
		}
		permList, ok := perms.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, response.Response{Code: -1, Message: "无权限"})
			c.Abort()
			return
		}
		required := module + ":" + action
		for _, p := range permList {
			if p == required {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, response.Response{Code: -1, Message: "无权限: " + required})
		c.Abort()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Disposition")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
