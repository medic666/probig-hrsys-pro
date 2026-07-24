package service

import (
	"errors"

	"probig/internal/dao"
	"probig/internal/models"
	"probig/pkg/crypto"
	"probig/pkg/jwt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	IsFirstLogin bool   `json:"is_first_login"`
	Username     string `json:"username"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := dao.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}
	personID := uint(0)
	if user.PersonID != nil {
		personID = *user.PersonID
	}
	token, err := jwt.GenerateToken(user.ID, user.Username, personID)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}
	return &LoginResponse{
		Token:        token,
		IsFirstLogin: user.IsFirstLogin,
		Username:     user.Username,
	}, nil
}

func GetUserInfo(userID uint) (*models.User, error) {
	return dao.GetUserByID(userID)
}

func ChangePassword(userID uint, req *ChangePasswordRequest) error {
	user, err := dao.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if !crypto.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errors.New("原密码错误")
	}
	hash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.IsFirstLogin = false
	return dao.UpdateUser(user)
}

func ResetUserPassword(userID uint, newPassword string) error {
	user, err := dao.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.IsFirstLogin = true
	return dao.UpdateUser(user)
}
