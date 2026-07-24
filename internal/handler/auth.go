package handler

import (
	"github.com/gin-gonic/gin"
	"probig/internal/middleware"
	"probig/internal/service"
	"probig/pkg/response"
)

func Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请输入用户名和密码")
		return
	}
	result, err := service.Login(&req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, result)
}

func GetUserInfo(c *gin.Context) {
	claims := middleware.GetUserClaims(c)
	user, err := service.GetUserInfo(claims.UserID)
	if err != nil {
		response.Error(c, "获取用户信息失败")
		return
	}
	perms, _ := service.GetUserPermissions(claims.UserID)
	response.Success(c, gin.H{
		"user":        user,
		"permissions": perms,
	})
}

func ChangePassword(c *gin.Context) {
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	claims := middleware.GetUserClaims(c)
	if err := service.ChangePassword(claims.UserID, &req); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func ResetUserPassword(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.ResetUserPassword(req.UserID, req.NewPassword); err != nil {
		response.Error(c, err.Error())
		return
	}
	middleware.RecordAudit(c, "重置密码", "user", req.UserID, nil, nil, "")
	response.Success(c, nil)
}
