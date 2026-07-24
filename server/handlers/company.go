package handlers

import (
	"strconv"

	"probig/middleware"
	"probig/models"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func ListCompanies(c *gin.Context) {
	query := c.Query("query")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	companies, total, err := services.ListCompanies(query, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, companies, total, page, pageSize)
}

func ListDeletedCompanies(c *gin.Context) {
	query := c.Query("query")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	companies, total, err := services.ListDeletedCompanies(query, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, companies, total, page, pageSize)
}

func GetCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	company, err := services.GetCompany(uint(id))
	if err != nil {
		utils.ErrBadRequest(c, "公司不存在")
		return
	}
	utils.Success(c, company)
}

func CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.CreateCompany(&company); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "company", company.ID, "新增", "{}", company)
	utils.Success(c, company)
}

func UpdateCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	before, _ := services.GetCompany(uint(id))
	if err := services.UpdateCompany(uint(id), updates); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	after, _ := services.GetCompany(uint(id))
	middleware.AuditAction(c, "company", uint(id), "修改", before, after)
	utils.Success(c, nil)
}

func DeleteCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	before, _ := services.GetCompany(uint(id))
	services.DeleteCompany(uint(id))
	middleware.AuditAction(c, "company", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func RestoreCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	services.RestoreCompany(uint(id))
	middleware.AuditAction(c, "company", uint(id), "恢复", "{}", "restored")
	utils.Success(c, nil)
}
