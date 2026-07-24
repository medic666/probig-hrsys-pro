package services

import (
	"errors"

	"probig/database"
	"probig/middleware"
	"probig/models"

	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token       string   `json:"token"`
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	IsFirstLogin bool    `json:"is_first_login"`
	Permissions []string `json:"permissions"`
	PersonID    *uint    `json:"person_id"`
}

func Login(input LoginInput) (*LoginResult, error) {
	var user models.User
	if err := database.DB.Where("username = ? AND status = 1", input.Username).First(&user).Error; err != nil {
		return nil, errors.New("账号或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("账号或密码错误")
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	perms := middleware.GetUserPermissions(user.ID)

	return &LoginResult{
		Token:       token,
		UserID:      user.ID,
		Username:    user.Username,
		IsFirstLogin: user.IsFirstLogin,
		Permissions: perms,
		PersonID:    user.PersonID,
	}, nil
}

func ChangePassword(userID uint, oldPwd, newPwd string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return errors.New("原密码错误")
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	return database.DB.Model(&user).Updates(map[string]interface{}{
		"password_hash":  string(hash),
		"is_first_login": false,
	}).Error
}

func ResetPassword(userID uint, newPwd string) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password_hash":  string(hash),
		"is_first_login": true,
	}).Error
}
