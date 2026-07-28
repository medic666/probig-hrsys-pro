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

func GetPersons(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetPersonList(pageReq.PageNum, pageReq.PageSize, c.Query("name"), c.Query("id_card"), c.Query("person_id"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func GetPersonByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := service.GetPersonByID(uint(id))
	if err != nil {
		utils.Error(c, "人员不存在")
		return
	}
	utils.Success(c, p)
}

func CreatePerson(c *gin.Context) {
	var p model.Person
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.CreatePerson(&p); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": p.ID})
}

func UpdatePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var p model.Person
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdatePerson(uint(id), &p); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePerson(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestorePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestorePerson(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedPersons(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedPersons(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func ExportPersons(c *gin.Context) {
	list, _, err := service.GetPersonList(1, 10000, c.Query("name"), c.Query("id_card"), "")
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "人员列表"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"姓名", "身份证号", "性别", "生日", "民族", "籍贯", "住址", "政治面貌", "婚姻状态", "别名"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, p := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), p.Name)
		f.SetCellValue(sheet, cellName(2, row), p.IDCard)
		f.SetCellValue(sheet, cellName(3, row), genderText(p.Gender))
		if p.Birthday != nil {
			f.SetCellValue(sheet, cellName(4, row), p.Birthday.Format("2006-01-02"))
		}
		f.SetCellValue(sheet, cellName(5, row), p.Nation)
		f.SetCellValue(sheet, cellName(6, row), p.NativePlace)
		f.SetCellValue(sheet, cellName(7, row), p.Address)
		f.SetCellValue(sheet, cellName(8, row), p.PoliticalStatus)
		f.SetCellValue(sheet, cellName(9, row), maritalText(p.MaritalStatus))
		f.SetCellValue(sheet, cellName(10, row), p.Alias)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=persons_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func genderText(g int8) string {
	if g == 1 {
		return "男"
	}
	if g == 2 {
		return "女"
	}
	return ""
}

func maritalText(m int8) string {
	if m == 1 {
		return "已婚"
	}
	if m == 2 {
		return "未婚"
	}
	return ""
}

func AddPersonPhone(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Phone string `json:"phone"`
		Type  string `json:"phone_type"`
	}
	c.ShouldBindJSON(&req)
	if err := service.AddPersonPhone(uint(id), req.Phone, req.Type); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "添加成功", nil)
}

func UpdatePersonPhone(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	var req struct {
		Phone string `json:"phone"`
		Type  string `json:"phone_type"`
	}
	c.ShouldBindJSON(&req)
	if err := service.UpdatePersonPhone(uint(pid), req.Phone, req.Type); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePersonPhone(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err := service.DeletePersonPhone(uint(pid)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func AddPersonEmail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Email string `json:"email"`
		Type  string `json:"email_type"`
	}
	c.ShouldBindJSON(&req)
	if err := service.AddPersonEmail(uint(id), req.Email, req.Type); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "添加成功", nil)
}

func UpdatePersonEmail(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	var req struct {
		Email string `json:"email"`
		Type  string `json:"email_type"`
	}
	c.ShouldBindJSON(&req)
	if err := service.UpdatePersonEmail(uint(pid), req.Email, req.Type); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePersonEmail(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err := service.DeletePersonEmail(uint(pid)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func AddPersonBankCard(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
		AccountHolder string `json:"account_holder"`
	}
	c.ShouldBindJSON(&req)
	if err := service.AddPersonBankCard(uint(id), req.BankName, req.AccountNumber, req.AccountHolder); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "添加成功", nil)
}

func UpdatePersonBankCard(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	var req struct {
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
		AccountHolder string `json:"account_holder"`
	}
	c.ShouldBindJSON(&req)
	if err := service.UpdatePersonBankCard(uint(pid), req.BankName, req.AccountNumber, req.AccountHolder); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePersonBankCard(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err := service.DeletePersonBankCard(uint(pid)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func GetAllPersonsList(c *gin.Context) {
	list, err := service.GetAllPersons()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	var opts []gin.H
	for _, p := range list {
		opts = append(opts, gin.H{"id": p.ID, "name": p.Name})
	}
	utils.Success(c, opts)
}
