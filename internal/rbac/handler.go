package rbac

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/pkg/config"
	"probig/internal/pkg/jwt"
	"probig/internal/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	user, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	cfg := config.Load()
	token, err := jwt.GenerateToken(user.ID, user.Username, cfg.JwtSecret, 24)
	if err != nil {
		response.ServerError(c, "failed to generate token")
		return
	}

	permissions, _ := h.service.GetAllPermissions()
	var permKeys []string
	for _, p := range permissions {
		if h.service.CheckPermission(user.ID, p.PermissionKey) {
			permKeys = append(permKeys, p.PermissionKey)
		}
	}

	response.Success(c, gin.H{
		"user":        user,
		"token":       token,
		"permissions": permKeys,
	})
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := toUint(userIDVal)

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		response.ErrorWithMsg(c, "user not found")
		return
	}

	permissions, _ := h.service.GetAllPermissions()
	var permKeys []string
	for _, p := range permissions {
		if h.service.CheckPermission(user.ID, p.PermissionKey) {
			permKeys = append(permKeys, p.PermissionKey)
		}
	}

	response.Success(c, gin.H{
		"user":        user,
		"permissions": permKeys,
	})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := toUint(userIDVal)

	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	username := c.Query("username")

	users, total, err := h.service.ListUsers(page, pageSize, username)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, users, total, page, pageSize)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if user.Username == "" {
		response.ParamError(c, "username is required")
		return
	}

	if err := h.service.CreateUser(&user); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	user.ID = uint(id)
	if err := h.service.UpdateUser(&user); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	if err := h.service.DeleteUser(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if err := h.service.ResetPassword(uint(id), req.NewPassword); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) AssignRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	var req struct {
		RoleIDs []uint `json:"roleIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if err := h.service.AssignRoles(uint(id), req.RoleIDs); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	name := c.Query("name")

	roles, total, err := h.service.ListRoles(page, pageSize, name)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, roles, total, page, pageSize)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var role Role
	if err := c.ShouldBindJSON(&role); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if role.Name == "" {
		response.ParamError(c, "name is required")
		return
	}

	if err := h.service.CreateRole(&role); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, role)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	var role Role
	if err := c.ShouldBindJSON(&role); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	role.ID = uint(id)
	if err := h.service.UpdateRole(&role); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	if err := h.service.DeleteRole(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) AssignPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permissionIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if err := h.service.AssignPermissions(uint(id), req.PermissionIDs); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.service.GetAllPermissions()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, permissions)
}

func toUint(val interface{}) uint {
	switch v := val.(type) {
	case uint:
		return v
	case float64:
		return uint(v)
	case int:
		return uint(v)
	}
	return 0
}
