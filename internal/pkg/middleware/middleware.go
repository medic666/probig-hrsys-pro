package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"probig/internal/pkg/database"
	jwtPkg "probig/internal/pkg/jwt"
	"probig/internal/pkg/response"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/auth/login") || strings.HasPrefix(path, "/api/login") {
			c.Next()
			return
		}
		if strings.HasPrefix(path, "/api/health") {
			c.Next()
			return
		}
		if !strings.HasPrefix(path, "/api") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, response.Unauthorized, "未登录，请先登录")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.Error(c, response.Unauthorized, "认证格式错误")
			c.Abort()
			return
		}

		claims, err := jwtPkg.ParseToken(tokenString)
		if err != nil {
			response.Error(c, response.Unauthorized, "认证已过期，请重新登录")
			c.Abort()
			return
		}

		var user database.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			response.Error(c, response.Unauthorized, "用户不存在")
			c.Abort()
			return
		}

		if user.Status != 1 {
			response.Error(c, response.Forbidden, "账号已被禁用")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func PermissionMiddleware(permKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		if userID == 0 {
			c.Next()
			return
		}

		var permission database.Permission
		if err := database.DB.Where("perm_key = ?", permKey).First(&permission).Error; err != nil {
			c.Next()
			return
		}

		var count int64
		database.DB.Table("user_roles").
			Joins("JOIN role_permissions ON role_permissions.role_id = user_roles.role_id").
			Where("user_roles.user_id = ? AND role_permissions.permission_id = ?", userID, permission.ID).
			Count(&count)

		if count == 0 {
			response.Error(c, response.Forbidden, "无操作权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				c.AbortWithStatusJSON(http.StatusOK, response.Response{
					Code: response.InternalError,
					Msg:  "服务器内部错误",
					Data: nil,
				})
			}
		}()
		c.Next()
	}
}
