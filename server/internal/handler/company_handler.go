package handler

import (
	"strconv"
	"time"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetCompanies(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetCompanyList(pageReq.PageNum, pageReq.PageSize, c.Query("name"), c.Query("credit_code"))
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
	if err := service.CreateCompany(&co); err != nil {
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
	if err := service.UpdateCompany(uint(id), &co); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteCompany(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreCompany(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreCompany(uint(id)); err != nil {
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
	list, _, err := service.GetCompanyList(1, 10000, c.Query("name"), c.Query("credit_code"))
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "公司列表"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"公司名称", "统一社会信用代码", "地址", "联系电话", "开户行", "银行账号"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, co := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), co.Name)
		f.SetCellValue(sheet, cellName(2, row), co.CreditCode)
		f.SetCellValue(sheet, cellName(3, row), co.Address)
		f.SetCellValue(sheet, cellName(4, row), co.ContactPhone)
		f.SetCellValue(sheet, cellName(5, row), co.BankName)
		f.SetCellValue(sheet, cellName(6, row), co.BankAccount)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=companies_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
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
