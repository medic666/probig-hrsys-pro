package middleware

import (
	"strings"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/service"
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
		if service.IsTokenBlacklisted(token) {
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

		info := dao.AuditInfo{OperatorID: user.ID, OperatorName: user.Username, IP: c.ClientIP()}
		c.Request = c.Request.WithContext(dao.WithAuditContext(c.Request.Context(), info))
		// 数据范围上下文：人员维度权限（own 时按关联人员过滤/校验）
		scope := dao.ScopeInfo{UserID: user.ID, DataScope: user.DataScope, PersonID: user.PersonID}
		c.Request = c.Request.WithContext(dao.WithScopeInfo(c.Request.Context(), scope))
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("user", &user)

		// 滑动续期：剩余有效期不足一半时签发新 token 经响应头返回，前端静默更新
		if claims.ExpiresAt != nil && utils.ShouldRenew(claims.ExpiresAt.Time) {
			if newToken, err := utils.GenerateToken(user.ID, user.Username); err == nil {
				c.Header("X-New-Token", newToken)
			}
		}
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

		for _, key := range service.GetUserPermissionKeys(userID.(uint)) {
			if key == permKey {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "无操作权限")
		c.Abort()
	}
}
