package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"probig/internal/config"
	"probig/internal/database"
	"probig/internal/models"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "未提供认证令牌", "data": nil})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "认证令牌格式错误", "data": nil})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "message": "认证令牌无效或已过期", "data": nil})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func RBAC(requiredModule, requiredAction string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Abort()
			return
		}

		var user models.User
		if err := database.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": 40300, "message": "无权访问", "data": nil})
			c.Abort()
			return
		}

		for _, role := range user.Roles {
			for _, perm := range role.Permissions {
				if perm.Module == requiredModule && perm.Action == requiredAction {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"code": 40300, "message": "无权执行此操作", "data": nil})
		c.Abort()
	}
}
