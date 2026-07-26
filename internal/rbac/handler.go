package rbac

import (
	"probig/internal/pkg/audit"
	"probig/internal/pkg/middleware"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/login", Login)
	r.POST("/register", Register)

	auth := r.Group("").Use(middleware.Auth)
	{
		auth.GET("/current-user", GetCurrentUser)
		auth.POST("/change-password", ChangePassword)
		auth.GET("/permissions", GetPermissions)
	}

	users := r.Group("/users").Use(middleware.Auth, middleware.Permission("rbac:read"))
	{
		users.GET("", GetUserList)
		users.POST("", middleware.Permission("rbac:write"), CreateUser)
		users.PUT("/:id", middleware.Permission("rbac:write"), UpdateUser)
		users.DELETE("/:id", middleware.Permission("rbac:write"), DeleteUser)
		users.PUT("/:id/status", middleware.Permission("rbac:write"), ToggleUserStatus)
		users.POST("/:id/reset-password", middleware.Permission("rbac:write"), ResetPassword)
		users.PUT("/:id/bind", middleware.Permission("rbac:write"), BindPerson)
	}

	roles := r.Group("/roles").Use(middleware.Auth, middleware.Permission("rbac:read"))
	{
		roles.GET("", GetRoles)
		roles.POST("", middleware.Permission("rbac:write"), CreateRole)
		roles.PUT("/:id", middleware.Permission("rbac:write"), UpdateRole)
		roles.DELETE("/:id", middleware.Permission("rbac:write"), DeleteRole)
		roles.GET("/:id/permissions", GetRolePermissions)
		roles.PUT("/:id/permissions", middleware.Permission("rbac:write"), SetRolePermissions)
		roles.GET("/permissions", GetAllPermissions)
	}
}

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请输入用户名和密码")
		return
	}
	svc := NewService()
	user, token, err := svc.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, "用户名或密码错误")
		return
	}
	audit.Write(c, user.ID, user.Username, "系统", 0, "", "登录", nil, nil)
	response.Success(c, gin.H{
		"token":          token,
		"is_first_login": user.IsFirstLogin,
	})
}

func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.CreateUser(req.Username, req.Password, nil); err != nil {
		response.Error(c, "创建用户失败")
		return
	}
	response.SuccessMsg(c, "注册成功")
}

func GetCurrentUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	svc := NewService()
	user, err := svc.GetCurrentUser(userID)
	if err != nil {
		response.Error(c, "获取用户信息失败")
		return
	}
	response.Success(c, user)
}

func ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	userID := c.GetUint("user_id")
	svc := NewService()
	if err := svc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, "原密码错误")
		return
	}
	username, _ := c.Get("username")
	audit.Write(c, userID, username.(string), "用户", userID, username.(string), "修改密码", nil, nil)
	response.SuccessMsg(c, "密码修改成功")
}

func GetPermissions(c *gin.Context) {
	userID := c.GetUint("user_id")
	svc := NewService()
	perms, err := svc.GetUserPermissions(userID)
	if err != nil {
		response.Error(c, "获取权限失败")
		return
	}
	response.Success(c, perms)
}

func GetUserList(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	username := c.Query("username")
	svc := NewService()
	list, total, err := svc.GetUserList(pageNum, pageSize, username)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		RoleIDs  []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.CreateUser(req.Username, req.Password, req.RoleIDs); err != nil {
		response.Error(c, "创建失败，用户名可能已存在")
		return
	}
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	audit.Write(c, userID.(uint), username.(string), "用户", 0, req.Username, "新增", nil, nil)
	response.SuccessMsg(c, "创建成功")
}

func UpdateUser(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		Username string `json:"username"`
		RoleIDs  []uint `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.UpdateUser(id, req.Username, req.RoleIDs); err != nil {
		response.Error(c, "更新失败")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "用户", id, req.Username, "修改", nil, nil)
	response.SuccessMsg(c, "更新成功")
}

func DeleteUser(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeleteUser(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "用户", id, "", "删除", nil, nil)
	response.SuccessMsg(c, "删除成功")
}

func ToggleUserStatus(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.ToggleUserStatus(id); err != nil {
		response.Error(c, "操作失败")
		return
	}
	response.SuccessMsg(c, "操作成功")
}

func ResetPassword(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	newPW, err := svc.ResetPassword(id)
	if err != nil {
		response.Error(c, "重置失败")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "用户", id, "", "重置密码", nil, nil)
	response.Success(c, gin.H{"new_password": newPW})
}

func BindPerson(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		PersonID uint `json:"person_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.BindPerson(id, req.PersonID); err != nil {
		response.Error(c, "绑定失败")
		return
	}
	response.SuccessMsg(c, "绑定成功")
}

func GetRoles(c *gin.Context) {
	svc := NewService()
	roles, err := svc.GetRoles()
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, roles)
}

func CreateRole(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.CreateRole(req.Name, req.Remark); err != nil {
		response.Error(c, "创建失败")
		return
	}
	response.SuccessMsg(c, "创建成功")
}

func UpdateRole(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		Name   string `json:"name"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.UpdateRole(id, req.Name, req.Remark); err != nil {
		response.Error(c, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func DeleteRole(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeleteRole(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func GetRolePermissions(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	perms, err := svc.GetRolePermissions(id)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, perms)
}

func SetRolePermissions(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		PermIDs []uint `json:"perm_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.SetRolePermissions(id, req.PermIDs); err != nil {
		response.Error(c, "设置失败")
		return
	}
	response.SuccessMsg(c, "设置成功")
}

func GetAllPermissions(c *gin.Context) {
	svc := NewService()
	perms, err := svc.GetAllPermissions()
	if err != nil {
		response.Error(c, "查询失败")
		return
	}

	grouped := make(map[string][]Permission)
	for _, p := range perms {
		grouped[p.Module] = append(grouped[p.Module], p)
	}
	response.Success(c, grouped)
}
