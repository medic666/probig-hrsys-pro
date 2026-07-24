package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"probig/internal/config"
	"probig/internal/database"
	"probig/internal/middleware"
	"probig/internal/models"
)

type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户名或密码错误")
		}
		return "", nil, err
	}

	if user.Status != "active" {
		return "", nil, errors.New("账户已被禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return "", nil, err
	}
	return token, &user, nil
}

func (s *AuthService) GetUserWithRoles(userID uint) (*models.User, error) {
	var user models.User
	if err := database.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) GetPermissions(userID uint) ([]models.Permission, error) {
	var user models.User
	if err := database.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		return nil, err
	}
	var perms []models.Permission
	seen := make(map[uint]bool)
	for _, role := range user.Roles {
		for _, p := range role.Permissions {
			if !seen[p.ID] {
				perms = append(perms, p)
				seen[p.ID] = true
			}
		}
	}
	return perms, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := &middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
