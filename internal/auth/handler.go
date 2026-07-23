package auth

import (
	"probig/internal/common"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := common.BindJSON(c, &req); err != nil {
		return
	}

	resp, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		common.Error(c, common.CodeUnauthorized, err.Error())
		return
	}

	common.Success(c, resp)
}

func (h *Handler) Me(c *gin.Context) {
	claims := GetUserClaims(c)
	if claims == nil {
		common.Error(c, common.CodeUnauthorized, "未登录")
		return
	}

	perms := RolePermissions[claims.Role]
	if perms == nil {
		perms = []string{}
	}

	common.Success(c, gin.H{
		"user": UserInfo{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		},
		"permissions": perms,
	})
}

func (h *Handler) GetMenus(c *gin.Context) {
	claims := GetUserClaims(c)
	if claims == nil {
		common.Error(c, common.CodeUnauthorized, "未登录")
		return
	}

	perms := RolePermissions[claims.Role]
	permSet := make(map[string]bool)
	for _, p := range perms {
		permSet[p] = true
	}

	allMenus := []MenuItem{
		{Key: "dashboard", Label: "仪表盘", Icon: "DashboardOutlined", Path: "/", Permission: "dashboard:read"},
		{Key: "person", Label: "人员管理", Icon: "UserOutlined", Path: "/persons", Permission: "person:read"},
		{Key: "policy", Label: "制度管理", Icon: "FileTextOutlined", Path: "/policies", Permission: "policy:read"},
		{Key: "attendance", Label: "考勤管理", Icon: "ClockCircleOutlined", Path: "/attendance", Permission: "attendance:read"},
		{Key: "salary", Label: "工资管理", Icon: "DollarOutlined", Path: "/salary", Permission: "salary:read"},
		{Key: "asset", Label: "资产管理", Icon: "DatabaseOutlined", Path: "/assets", Permission: "asset:read"},
		{Key: "event", Label: "审计日志", Icon: "AuditOutlined", Path: "/events", Permission: "event:read"},
	}

	var filtered []MenuItem
	for _, m := range allMenus {
		if permSet[m.Permission] {
			filtered = append(filtered, m)
		}
	}

	common.Success(c, filtered)
}

func (h *Handler) GetPermissions(c *gin.Context) {
	claims := GetUserClaims(c)
	if claims == nil {
		common.Error(c, common.CodeUnauthorized, "未登录")
		return
	}

	perms := RolePermissions[claims.Role]
	if perms == nil {
		perms = []string{}
	}
	common.Success(c, perms)
}
