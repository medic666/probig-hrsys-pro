package handler

import (
	"fmt"
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// companyListQuery 公司列表筛选解析（列表与导出共用）
func companyListQuery(c *gin.Context) service.CompanyListQuery {
	pageReq := utils.BindPage(c)
	return service.CompanyListQuery{
		PageNum:    pageReq.PageNum,
		PageSize:   pageReq.PageSize,
		Name:       c.Query("name"),
		CreditCode: c.Query("credit_code"),
		ID:         c.Query("id"),
	}
}

func GetCompanies(c *gin.Context) {
	q := companyListQuery(c)
	list, total, err := service.GetCompanyList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, q.PageNum, q.PageSize)
}

func GetCompanyByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	co, err := service.GetCompanyByID(uint(id))
	if err != nil {
		utils.Error(c, "公司不存在")
		return
	}
	utils.Success(c, co)
}

func CreateCompany(c *gin.Context) {
	var co model.Company
	if err := c.ShouldBindJSON(&co); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.CreateCompany(c.Request.Context(), &co); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": co.ID})
}

func UpdateCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var co model.Company
	if err := c.ShouldBindJSON(&co); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdateCompany(c.Request.Context(), uint(id), &co); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteCompany(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreCompany(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedCompanies(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedCompanies(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
}

// companyExportFilters 公司导出文件名筛选摘要
func companyExportFilters(q service.CompanyListQuery) []string {
	var parts []string
	if q.ID != "" {
		var id uint
		fmt.Sscanf(q.ID, "%d", &id)
		if id > 0 {
			parts = append(parts, "公司="+service.CompanyNameMap([]uint{id})[id])
		}
	}
	if q.Name != "" {
		parts = append(parts, "名称="+q.Name)
	}
	if q.CreditCode != "" {
		parts = append(parts, "信用代码="+q.CreditCode)
	}
	return parts
}

func ExportCompanies(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := companyListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, err := service.GetCompanyList(q)
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	var rows [][]interface{}
	for _, co := range list {
		rows = append(rows, []interface{}{
			co.Name, co.CreditCode, co.Address, co.ContactPhone, co.BankName, co.BankAccount,
		})
	}
	writeExcel(c, "公司列表",
		[]string{"公司名称", "统一社会信用代码", "地址", "联系电话", "开户行", "银行账号"}, rows,
		companyExportFilters(q)...)
}

func GetAllCompaniesList(c *gin.Context) {
	list, err := service.GetAllCompanies()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	var opts []gin.H
	for _, co := range list {
		opts = append(opts, gin.H{"id": co.ID, "name": co.Name})
	}
	utils.Success(c, opts)
}
