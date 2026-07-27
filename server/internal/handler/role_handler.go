package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetRoles(c *gin.Context) {
	pageReq := utils.BindPage(c)
	name := c.Query("name")

	list, total, err := service.GetRoleList(pageReq.PageNum, pageReq.PageSize, name)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func GetAllRolesList(c *gin.Context) {
	roles, err := service.GetAllRoles()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, roles)
}

type createRoleReq struct {
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
}

func CreateRole(c *gin.Context) {
	var req createRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	role, err := service.CreateRole(req.Name, req.Remark)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": role.ID})
}

type updateRoleReq struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

func UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	var req updateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := service.UpdateRole(uint(id), req.Name, req.Remark); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := service.DeleteRole(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := service.RestoreRole(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedRoles(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedRoleList(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

type assignPermsReq struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

func AssignRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	var req assignPermsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if err := service.AssignRolePermissions(uint(id), req.PermissionIDs); err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "分配成功", nil)
}

func GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	permIDs := service.GetRolePermissionIDs(uint(id))
	utils.Success(c, permIDs)
}
