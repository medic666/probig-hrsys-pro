package middleware

import (
	"strings"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Unauthorized(c, "未登录或Token已过期")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.Unauthorized(c, "未登录或Token已过期")
			c.Abort()
			return
		}

		var user model.User
		if err := dao.DB.First(&user, claims.UserID).Error; err != nil {
			utils.Unauthorized(c, "用户不存在")
			c.Abort()
			return
		}

		if !user.IsActive {
			utils.Unauthorized(c, "账号已被禁用")
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("user", &user)
		c.Next()
	}
}

func RequirePermission(permKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			utils.Forbidden(c, "无操作权限")
			c.Abort()
			return
		}

		var permissions []model.Permission
		dao.DB.Table("permissions").
			Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
			Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
			Where("user_roles.user_id = ?", userID).
			Find(&permissions)

		for _, p := range permissions {
			key := p.Module + "." + p.Action
			if key == permKey {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "无操作权限")
		c.Abort()
	}
}
