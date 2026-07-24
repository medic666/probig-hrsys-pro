package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetCompanyList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total, err := service.GetCompanyList(page, pageSize, keyword)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func GetCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	company, err := service.GetCompany(uint(id))
	if err != nil {
		response.Error(c, "公司不存在")
		return
	}
	response.Success(c, company)
}

func CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.CreateCompany(c, &company); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, company)
}

func UpdateCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		response.Error(c, "参数错误")
		return
	}
	company.ID = uint(id)
	if err := service.UpdateCompany(c, &company); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, company)
}

func DeleteCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteCompany(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func RestoreCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreCompany(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}
