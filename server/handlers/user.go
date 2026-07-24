package handlers

import (
	"strconv"

	"probig/middleware"
	"probig/models"
	"probig/services"
	"probig/utils"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

func ListUsers(c *gin.Context) {
	query := c.Query("query")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize
	users, total, _ := services.ListUsers(query, offset, pageSize)
	utils.SuccessPage(c, users, total, page, pageSize)
}

func GetUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := services.GetUser(uint(id))
	if err != nil {
		utils.ErrBadRequest(c, "用户不存在")
		return
	}
	utils.Success(c, user)
}

func CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		PersonID *uint  `json:"person_id"`
		Status   int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := models.User{
		Username:     input.Username,
		PasswordHash: string(hash),
		PersonID:     input.PersonID,
		IsFirstLogin: true,
		Status:       input.Status,
	}
	if user.Status == 0 {
		user.Status = 1
	}
	if err := services.CreateUser(&user); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "user", user.ID, "新增", "{}", user)
	utils.Success(c, user)
}

func UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		Username string `json:"username"`
		PersonID *uint  `json:"person_id"`
		Status   int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	updates := map[string]interface{}{
		"username":  input.Username,
		"person_id": input.PersonID,
		"status":    input.Status,
	}
	before, _ := services.GetUser(uint(id))
	services.UpdateUser(uint(id), updates)
	after, _ := services.GetUser(uint(id))
	middleware.AuditAction(c, "user", uint(id), "修改", before, after)
	utils.Success(c, nil)
}

func DeleteUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	before, _ := services.GetUser(uint(id))
	services.DeleteUser(uint(id))
	middleware.AuditAction(c, "user", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func ResetUserPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		NewPassword string `json:"new_password"`
	}
	c.ShouldBindJSON(&input)
	if err := services.ResetPassword(uint(id), input.NewPassword); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "user", uint(id), "重置密码", "{}", gin.H{"user_id": id})
	utils.Success(c, nil)
}

func AssignUserRoles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		RoleIDs []uint `json:"role_ids"`
	}
	c.ShouldBindJSON(&input)
	services.AssignUserRoles(uint(id), input.RoleIDs)
	middleware.AuditAction(c, "user", uint(id), "分配角色", "{}", input)
	utils.Success(c, nil)
}

func ListRoles(c *gin.Context) {
	query := c.Query("query")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize
	roles, total, _ := services.ListRoles(query, offset, pageSize)
	utils.SuccessPage(c, roles, total, page, pageSize)
}

func GetAllRoles(c *gin.Context) {
	roles, _ := services.GetAllRoles()
	utils.Success(c, roles)
}

func GetRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, err := services.GetRole(uint(id))
	if err != nil {
		utils.ErrBadRequest(c, "角色不存在")
		return
	}
	utils.Success(c, role)
}

func CreateRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	services.CreateRole(&role)
	middleware.AuditAction(c, "role", role.ID, "新增", "{}", role)
	utils.Success(c, role)
}

func UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	before, _ := services.GetRole(uint(id))
	services.UpdateRole(uint(id), updates)
	after, _ := services.GetRole(uint(id))
	middleware.AuditAction(c, "role", uint(id), "修改", before, after)
	utils.Success(c, nil)
}

func DeleteRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	before, _ := services.GetRole(uint(id))
	services.DeleteRole(uint(id))
	middleware.AuditAction(c, "role", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func AssignRolePermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	c.ShouldBindJSON(&input)
	services.AssignRolePermissions(uint(id), input.PermissionIDs)
	middleware.AuditAction(c, "role", uint(id), "分配权限", "{}", input)
	utils.Success(c, nil)
}

func ListPermissions(c *gin.Context) {
	perms, _ := services.ListPermissions()
	utils.Success(c, perms)
}

func ListUsersSimple(c *gin.Context) {
	users, _ := services.ListUsersSimple()
	utils.Success(c, users)
}
