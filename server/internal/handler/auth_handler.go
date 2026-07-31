package handler

import (
	"strings"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func AuthLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写用户名和密码")
		return
	}

	token, user, err := service.Login(req.Username, req.Password)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	perms, menus, _ := service.GetUserPermissions(user.ID)

	utils.Success(c, gin.H{
		"token":          token,
		"user":           gin.H{"id": user.ID, "username": user.Username, "name": user.Username},
		"is_first_login": user.IsFirstLogin,
		"permissions":    perms,
		"menus":          menus,
	})
}

func AuthLogout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if claims, err := utils.ParseToken(token); err == nil {
		service.BlacklistToken(token, claims.ExpiresAt.Time)
	}
	utils.SuccessWithMsg(c, "已退出登录", nil)
}

type changePwdReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func AuthChangePassword(c *gin.Context) {
	var req changePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请填写密码")
		return
	}

	userID := c.GetUint("userID")
	if err := service.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "密码修改成功", nil)
}

func AuthMe(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := service.GetUserByID(userID)
	if err != nil {
		utils.Error(c, "用户不存在")
		return
	}

	perms, menus, _ := service.GetUserPermissions(userID)

	utils.Success(c, gin.H{
		"user":           gin.H{"id": user.ID, "username": user.Username, "name": user.Username},
		"is_first_login": user.IsFirstLogin,
		"permissions":    perms,
		"menus":          menus,
	})
}
