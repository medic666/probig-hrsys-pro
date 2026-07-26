package middleware

import (
	"probig/internal/pkg/config"
	"probig/internal/pkg/jwt"
	"probig/internal/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.ErrorWithCode(c, response.CodeAuthErr, "未登录或登录已过期")
		c.Abort()
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.ErrorWithCode(c, response.CodeAuthErr, "认证格式错误")
		c.Abort()
		return
	}
	claims, err := jwt.ParseToken(parts[1])
	if err != nil {
		response.ErrorWithCode(c, response.CodeAuthErr, "登录已过期，请重新登录")
		c.Abort()
		return
	}
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Next()
}

func Permission(permKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		uid := userID.(uint)

		type UserRole struct {
			RoleID uint
		}
		type RolePerm struct {
			PermKey string
		}

		var roleIDs []struct{ RoleID uint }
		config.DB.Table("user_roles").Select("role_id").Where("user_id = ?", uid).Find(&roleIDs)

		var rids []uint
		for _, r := range roleIDs {
			rids = append(rids, r.RoleID)
		}

		if len(rids) == 0 {
			response.ErrorWithCode(c, response.CodeForbid, "无操作权限")
			c.Abort()
			return
		}

		var permissions []string
		config.DB.Table("role_permissions").Select("permission.perm_key").
			Joins("left join permission on permission.id = role_permissions.permission_id").
			Where("role_permissions.role_id in ?", rids).
			Pluck("permission.perm_key", &permissions)

		for _, p := range permissions {
			if p == permKey || p == "all" {
				c.Next()
				return
			}
		}

		var isAdminCount int64
		config.DB.Table("user_roles").
			Joins("left join role on role.id = user_roles.role_id").
			Where("user_roles.user_id = ? AND role.is_admin = ?", uid, true).
			Count(&isAdminCount)
		if isAdminCount > 0 {
			c.Next()
			return
		}

		response.ErrorWithCode(c, response.CodeForbid, "无操作权限")
		c.Abort()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.Error(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
