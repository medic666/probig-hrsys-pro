package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetPersonList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total, err := service.GetPersonList(page, pageSize, keyword)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func GetPerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := service.GetPerson(uint(id))
	if err != nil {
		response.Error(c, "人员不存在")
		return
	}
	response.Success(c, p)
}

func CreatePerson(c *gin.Context) {
	var p models.Person
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.CreatePerson(c, &p); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, p)
}

func UpdatePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var p models.Person
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Error(c, "参数错误")
		return
	}
	p.ID = uint(id)
	if err := service.UpdatePerson(c, &p); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, p)
}

func DeletePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePerson(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func RestorePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestorePerson(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}
