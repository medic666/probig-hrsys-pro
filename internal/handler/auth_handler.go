package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"probig/internal/config"
	"probig/internal/models"
	"probig/internal/service"
	"probig/internal/utils"
)

type AuthHandler struct {
	cfg  *config.Config
	auth *service.AuthService
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg, auth: service.NewAuthService(cfg)}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请输入用户名和密码")
		return
	}

	token, user, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, 40100, err.Error())
		return
	}

	utils.Success(c, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.auth.GetUserWithRoles(userID.(uint))
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}
	user.Password = ""
	utils.Success(c, user)
}

func (h *AuthHandler) Permissions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	perms, err := h.auth.GetPermissions(userID.(uint))
	if err != nil {
		utils.InternalError(c, "获取权限失败")
		return
	}
	if perms == nil {
		perms = []models.Permission{}
	}
	utils.Success(c, perms)
}
