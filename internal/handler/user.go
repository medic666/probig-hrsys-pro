package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total, err := service.GetUserList(page, pageSize, keyword)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func GetUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	u, err := service.GetUser(uint(id))
	if err != nil {
		response.Error(c, "用户不存在")
		return
	}
	response.Success(c, u)
}

func CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		PersonID *uint  `json:"person_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	u := &models.User{
		Username: req.Username,
		PersonID: req.PersonID,
	}
	if err := service.CreateUser(c, u, req.Password); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, u)
}

func UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		response.Error(c, "参数错误")
		return
	}
	u.ID = uint(id)
	if err := service.UpdateUser(c, &u); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, u)
}

func DeleteUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteUser(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func GetRoleList(c *gin.Context) {
	list, err := service.GetRoleList()
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func GetRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	r, err := service.GetRole(uint(id))
	if err != nil {
		response.Error(c, "角色不存在")
		return
	}
	response.Success(c, r)
}

func CreateRole(c *gin.Context) {
	var r models.Role
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.CreateRole(c, &r); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, r)
}

func UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var r models.Role
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Error(c, "参数错误")
		return
	}
	r.ID = uint(id)
	if err := service.UpdateRole(c, &r); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, r)
}

func DeleteRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteRole(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func GetAllPermissions(c *gin.Context) {
	list, err := service.GetAllPermissions()
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, list)
}
