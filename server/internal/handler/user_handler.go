package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	pageReq := utils.BindPage(c)
	username := c.Query("username")

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		val := activeStr == "true" || activeStr == "1"
		isActive = &val
	}

	list, total, err := service.GetUserList(pageReq.PageNum, pageReq.PageSize, username, isActive)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func CreateUser(c *gin.Context) {
	var req service.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if req.Password == "" {
		req.Password = service.DefaultPassword
	}

	user, err := service.CreateUser(c.Request.Context(), req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": user.ID})
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	var req service.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := service.UpdateUser(c.Request.Context(), uint(id), req); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	operatorID := c.GetUint("userID")
	if err := service.DeleteUser(c.Request.Context(), uint(id), operatorID); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := service.RestoreUser(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func ResetUserPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	newPwd, err := service.ResetPassword(c.Request.Context(), uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "密码已重置为: "+newPwd, nil)
}

type assignRolesReq struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

func AssignUserRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	var req assignRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := service.AssignUserRoles(c.Request.Context(), uint(id), req.RoleIDs); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "分配成功", nil)
}

func GetDeletedUsers(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedUserList(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}
