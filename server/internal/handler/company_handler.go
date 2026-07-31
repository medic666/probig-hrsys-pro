package handler

import (
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetCompanies(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetCompanyList(pageReq.PageNum, pageReq.PageSize, c.Query("name"), c.Query("credit_code"), c.Query("id"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
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
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func ExportCompanies(c *gin.Context) {
	list, _, err := service.GetCompanyList(1, 10000, c.Query("name"), c.Query("credit_code"), "")
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
	writeExcel(c, "公司列表", "companies",
		[]string{"公司名称", "统一社会信用代码", "地址", "联系电话", "开户行", "银行账号"}, rows)
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
