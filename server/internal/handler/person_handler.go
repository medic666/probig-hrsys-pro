package handler

import (
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// personListQuery 人员列表筛选解析（列表与导出共用，单一事实来源）
func personListQuery(c *gin.Context) service.PersonListQuery {
	pageReq := utils.BindPage(c)
	companyID, _ := strconv.ParseUint(c.Query("company_id"), 10, 64)
	return service.PersonListQuery{
		PageNum:    pageReq.PageNum,
		PageSize:   pageReq.PageSize,
		Name:       c.Query("name"),
		IDCard:     c.Query("id_card"),
		PersonID:   c.Query("person_id"),
		CompanyID:  uint(companyID),
		Department: c.Query("department"),
		Status:     c.Query("status"),
	}
}

func GetPersons(c *gin.Context) {
	q := personListQuery(c)
	list, total, err := service.GetPersonList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, utils.PageRequest{PageNum: q.PageNum, PageSize: q.PageSize}))
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
	if err := service.CreatePerson(c.Request.Context(), &p); err != nil {
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
	if err := service.UpdatePerson(c.Request.Context(), uint(id), &p); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

// UpdatePersonProfile 聚合更新人员档案（基础字段 + 电话/邮箱/银行卡/紧急联系人，事务内同步）
func UpdatePersonProfile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.PersonProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdatePersonProfile(c.Request.Context(), uint(id), &req); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "保存成功", nil)
}

// UpsertPersonProfile 人员档案统一保存（新增=编辑同一入口，req.id=0 视为新增）
func UpsertPersonProfile(c *gin.Context) {
	var req service.PersonProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	isNew := req.ID == 0
	person, err := service.UpsertPersonProfile(c.Request.Context(), &req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	if isNew {
		utils.SuccessWithMsg(c, "创建成功", gin.H{"id": person.ID})
		return
	}
	utils.SuccessWithMsg(c, "保存成功", gin.H{"id": person.ID})
}

func DeletePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestorePerson(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestorePerson(c.Request.Context(), uint(id)); err != nil {
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

// personExportFilters 人员导出文件名筛选摘要
func personExportFilters(q service.PersonListQuery) []string {
	var parts []string
	if q.CompanyID > 0 {
		parts = append(parts, "公司="+service.CompanyNameMap([]uint{q.CompanyID})[q.CompanyID])
	}
	if q.Department != "" {
		parts = append(parts, "部门="+q.Department)
	}
	switch q.Status {
	case "active":
		parts = append(parts, "状态=在职")
	case "left":
		parts = append(parts, "状态=已离职")
	case "not_entered":
		parts = append(parts, "状态=未入职")
	}
	return parts
}

func ExportPersons(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选（与列表同一解析函数，固定全量拉取）
	q := personListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, err := service.GetPersonList(q)
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	var rows [][]interface{}
	for _, p := range list {
		birthday := ""
		if p.Birthday != nil {
			birthday = p.Birthday.String()
		}
		rows = append(rows, []interface{}{
			p.Name, p.IDCard, genderText(p.Gender), birthday, p.Nation,
			p.NativePlace, p.Address, p.PoliticalStatus, maritalText(p.MaritalStatus), p.Alias,
		})
	}
	writeExcel(c, "人员列表",
		[]string{"姓名", "身份证号", "性别", "生日", "民族", "籍贯", "住址", "政治面貌", "婚姻状态", "别名"}, rows,
		personExportFilters(q)...)
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
	if err := service.AddPersonPhone(c.Request.Context(), uint(id), req.Phone, req.Type); err != nil {
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
	if err := service.UpdatePersonPhone(c.Request.Context(), uint(pid), req.Phone, req.Type); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePersonPhone(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err := service.DeletePersonPhone(c.Request.Context(), uint(pid)); err != nil {
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
	if err := service.AddPersonEmail(c.Request.Context(), uint(id), req.Email, req.Type); err != nil {
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
	if err := service.UpdatePersonEmail(c.Request.Context(), uint(pid), req.Email, req.Type); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePersonEmail(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err := service.DeletePersonEmail(c.Request.Context(), uint(pid)); err != nil {
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
	if err := service.AddPersonBankCard(c.Request.Context(), uint(id), req.BankName, req.AccountNumber, req.AccountHolder); err != nil {
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
	if err := service.UpdatePersonBankCard(c.Request.Context(), uint(pid), req.BankName, req.AccountNumber, req.AccountHolder); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePersonBankCard(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err := service.DeletePersonBankCard(c.Request.Context(), uint(pid)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func AddPersonEmergencyContact(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		ContactName  string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
		Sort         int    `json:"sort"`
	}
	c.ShouldBindJSON(&req)
	if err := service.AddPersonEmergencyContact(c.Request.Context(), uint(id), req.ContactName, req.ContactPhone, req.Sort); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "添加成功", nil)
}

func UpdatePersonEmergencyContact(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	var req struct {
		ContactName  string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
		Sort         int    `json:"sort"`
	}
	c.ShouldBindJSON(&req)
	if err := service.UpdatePersonEmergencyContact(c.Request.Context(), uint(pid), req.ContactName, req.ContactPhone, req.Sort); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePersonEmergencyContact(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("pid"), 10, 64)
	if err := service.DeletePersonEmergencyContact(c.Request.Context(), uint(pid)); err != nil {
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

func GetPersonCards(c *gin.Context) {
	cards, err := service.GetPersonCards()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, cards)
}
