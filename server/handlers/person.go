package handlers

import (
	"strconv"

	"probig/middleware"
	"probig/models"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func ListPersons(c *gin.Context) {
	query := c.Query("query")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	persons, total, err := services.ListPersons(query, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, persons, total, page, pageSize)
}

func ListDeletedPersons(c *gin.Context) {
	query := c.Query("query")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	persons, total, err := services.ListDeletedPersons(query, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, persons, total, page, pageSize)
}

func GetPerson(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}

	person, err := services.GetPerson(uint(id))
	if err != nil {
		utils.ErrBadRequest(c, "人员不存在")
		return
	}
	utils.Success(c, person)
}

func CreatePerson(c *gin.Context) {
	var person models.Person
	if err := c.ShouldBindJSON(&person); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	before := "{}"
	if err := services.CreatePerson(&person); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "person", person.ID, "新增", before, person)
	utils.Success(c, person)
}

func UpdatePerson(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	before, _ := services.GetPerson(uint(id))
	if err := services.UpdatePerson(uint(id), updates); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	after, _ := services.GetPerson(uint(id))
	middleware.AuditAction(c, "person", uint(id), "修改", before, after)
	utils.Success(c, nil)
}

func DeletePerson(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	before, _ := services.GetPerson(uint(id))
	if err := services.DeletePerson(uint(id)); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "person", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func RestorePerson(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.RestorePerson(uint(id)); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "person", uint(id), "恢复", "{}", "restored")
	utils.Success(c, nil)
}

func AddPersonPhone(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var phone models.PersonPhone
	if err := c.ShouldBindJSON(&phone); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.AddPersonPhone(uint(id), &phone); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	utils.Success(c, phone)
}

func UpdatePersonPhone(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	if err := services.UpdatePersonPhone(uint(id), updates); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

func DeletePersonPhone(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	services.DeletePersonPhone(uint(id))
	utils.Success(c, nil)
}

func AddPersonEmail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var email models.PersonEmail
	if err := c.ShouldBindJSON(&email); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.AddPersonEmail(uint(id), &email); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	utils.Success(c, email)
}

func UpdatePersonEmail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	services.UpdatePersonEmail(uint(id), updates)
	utils.Success(c, nil)
}

func DeletePersonEmail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	services.DeletePersonEmail(uint(id))
	utils.Success(c, nil)
}

func AddPersonBankCard(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var card models.PersonBankCard
	if err := c.ShouldBindJSON(&card); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	if err := services.AddPersonBankCard(uint(id), &card); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	utils.Success(c, card)
}

func UpdatePersonBankCard(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	services.UpdatePersonBankCard(uint(id), updates)
	utils.Success(c, nil)
}

func DeletePersonBankCard(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	services.DeletePersonBankCard(uint(id))
	utils.Success(c, nil)
}
