package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	RoleID   uint     `json:"role_id"`
	RoleName string   `json:"role_name"`
	Perms    []string `json:"perms"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret       []byte
	expireHours  int
}

func NewJWTManager(secret string, expireHours int) *JWTManager {
	return &JWTManager{
		secret:      []byte(secret),
		expireHours: expireHours,
	}
}

func (m *JWTManager) Generate(userID uint, username, roleName string, roleID uint, perms []string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		Perms:    perms,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
