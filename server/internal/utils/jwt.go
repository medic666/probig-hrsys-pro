package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"probig/server/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// randomJTI 随机 token 标识，保证同一秒内生成的 token 互不相同（登出黑名单依赖此唯一性）
func randomJTI() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}

// JwtTTL 登录态有效期（小时，读配置，缺省 8h）
func JwtTTL() time.Duration {
	hours := 8
	if config.AppConfig != nil && config.AppConfig.Jwt.ExpireHours > 0 {
		hours = config.AppConfig.Jwt.ExpireHours
	}
	return time.Duration(hours) * time.Hour
}

// ShouldRenew 滑动续期判定：剩余有效期不足一半时签发新 token
// （活跃用户不中断，闲置满 TTL 自动失效）
func ShouldRenew(exp time.Time) bool {
	return time.Until(exp) < JwtTTL()/2
}

func GenerateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JwtTTL())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "probig",
			ID:        randomJTI(),
		},
	}

	secret := getJwtSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(tokenString string) (*Claims, error) {
	secret := getJwtSecret()
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func getJwtSecret() []byte {
	secret := "probig-default-secret"
	if config.AppConfig != nil && config.AppConfig.Jwt.Secret != "" {
		secret = config.AppConfig.Jwt.Secret
	}
	return []byte(secret)
}
