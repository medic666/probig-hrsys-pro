package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"probig/pkg/jwt"
	"probig/pkg/response"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			response.Unauthorized(c, "未登录或登录已过期")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := jwt.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			return
		}
		SetUserClaims(c, &UserClaims{
			UserID:   claims.UserID,
			Username: claims.Username,
			PersonID: claims.PersonID,
		})
		c.Next()
	}
}
