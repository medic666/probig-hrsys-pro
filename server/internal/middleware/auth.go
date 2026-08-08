package middleware

import (
	"strings"
	"sync"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// 权限键注册表：RequirePermission 构造时登记，供一致性测试断言
// （路由使用的权限键必须全部定义于 model.ModuleActions）
var registeredPermKeys sync.Map

// RegisteredPermissionKeys 返回全部已登记权限键（一致性测试用）
func RegisteredPermissionKeys() []string {
	var keys []string
	registeredPermKeys.Range(func(k, _ any) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

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
		// 数据范围上下文：由角色范围并集派生（用户级不再配置）
		scope := dao.ScopeInfo{UserID: user.ID, DataScope: service.GetUserEffectiveScope(user.ID), PersonID: user.PersonID}
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

// StructureOnly 结构授权点：跨模块「人员×时间」主轴底座数据（人员选项/卡片、公司选项等），
// 仅需认证即可访问，不参与模块×动作授权（与 RequirePermission 并列构成端点授权二元分类）。
// 业务端点必须使用 RequirePermission；结构授权点清单见 model.ReferenceEndpoints。
func StructureOnly() gin.HandlerFunc {
	return AuthRequired()
}

func RequirePermission(permKey string) gin.HandlerFunc {
	registeredPermKeys.Store(permKey, true)
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
