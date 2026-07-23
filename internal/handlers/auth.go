package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/auth"
	"github.com/medic666/probig/internal/models"
	"github.com/medic666/probig/internal/response"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db     *sqlx.DB
	jwtMgr *auth.JWTManager
}

func NewAuthHandler(db *sqlx.DB, jwtMgr *auth.JWTManager) *AuthHandler {
	return &AuthHandler{db: db, jwtMgr: jwtMgr}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var user models.User
	if err := h.db.Get(&user, "SELECT * FROM users WHERE username = ? AND status = 'active'", req.Username); err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	var role models.Role
	if err := h.db.Get(&role, "SELECT * FROM roles WHERE id = ?", user.RoleID); err != nil {
		response.InternalError(c, "角色信息异常")
		return
	}

	var perms []models.RolePermission
	if err := h.db.Select(&perms, "SELECT * FROM role_permissions WHERE role_id = ?", user.RoleID); err != nil {
		response.InternalError(c, "权限信息异常")
		return
	}

	permStrings := make([]string, len(perms))
	for i, p := range perms {
		permStrings[i] = p.Module + ":" + p.Action
	}

	token, err := h.jwtMgr.Generate(user.ID, user.Username, role.Name, user.RoleID, permStrings)
	if err != nil {
		response.InternalError(c, "生成令牌失败")
		return
	}

	response.Success(c, models.LoginResponse{
		Token: token,
		User:  user,
		Perms: permStrings,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")
	roleID := c.GetUint("role_id")
	roleName := c.GetString("role_name")
	perms, _ := c.Get("perms")

	response.Success(c, gin.H{
		"user_id":   userID,
		"username":  username,
		"role_id":   roleID,
		"role_name": roleName,
		"perms":     perms,
	})
}
