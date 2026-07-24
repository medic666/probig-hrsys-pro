package middleware

import "github.com/gin-gonic/gin"

type UserClaims struct {
	UserID     uint
	Username   string
	PersonID   uint
	PersonName string
}

const ClaimsKey = "userClaims"

func SetUserClaims(c *gin.Context, claims *UserClaims) {
	c.Set(ClaimsKey, claims)
}

func GetUserClaims(c *gin.Context) *UserClaims {
	v, exists := c.Get(ClaimsKey)
	if !exists {
		return nil
	}
	claims, ok := v.(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}
