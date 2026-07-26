package rbac

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/pkg/config"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"
)

type Handler struct {
	svc *Service
}

func NewHandler() *Handler {
	return &Handler{svc: DefaultService}
}

func getPageParams(c *gin.Context) (int, int) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	if pageNum < 1 {
		pageNum = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", config.Get("system.page_size")))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageNum, pageSize
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	ip := c.ClientIP()
	res, err := h.svc.Login(req.Username, req.Password, ip)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, res)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	userID := utils.GetUserID(c)
	if err := h.svc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "密码修改成功", nil)
}

func (h *Handler) GetUserInfo(c *gin.Context) {
	userID := utils.GetUserID(c)
	info, err := h.svc.GetUserInfo(userID)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}
	response.Success(c, info)
}

func (h *Handler) ListUsers(c *gin.Context) {
	pageNum, pageSize := getPageParams(c)
	username := c.Query("username")

	var status *int8
	if statusStr := c.Query("status"); statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		st := int8(s)
		status = &st
	}

	users, total, err := h.svc.ListUsers(pageNum, pageSize, username, status)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}

	response.Success(c, &response.PageResult{
		List:  users,
		Total: total,
	})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.CreateUser(req.Username, req.Password, req.PersonID)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "创建成功", user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "用户ID无效")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateUser(id, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "用户ID无效")
		return
	}

	if err := h.svc.DeleteUser(id); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func (h *Handler) GetUserRoles(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "用户ID无效")
		return
	}

	roles, err := h.svc.GetUserRoles(id)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}

	response.Success(c, roles)
}

func (h *Handler) AssignUserRoles(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "用户ID无效")
		return
	}

	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.AssignUserRoles(id, req.RoleIDs); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "分配角色成功", nil)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "用户ID无效")
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.ResetUserPassword(id, req.NewPassword); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "密码重置成功", nil)
}

func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles()
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}
	response.Success(c, roles)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	role, err := h.svc.CreateRole(req.Name, req.Remark)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "创建成功", role)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "角色ID无效")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateRole(id, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "角色ID无效")
		return
	}

	if err := h.svc.DeleteRole(id); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func (h *Handler) GetRolePermissions(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "角色ID无效")
		return
	}

	permissions, err := h.svc.GetRolePermissions(id)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}

	response.Success(c, permissions)
}

func (h *Handler) AssignRolePermissions(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "角色ID无效")
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.AssignRolePermissions(id, req.PermissionIDs); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "分配权限成功", nil)
}

func (h *Handler) ListPermissions(c *gin.Context) {
	permissions, err := h.svc.ListAllPermissions()
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}
	response.Success(c, permissions)
}
