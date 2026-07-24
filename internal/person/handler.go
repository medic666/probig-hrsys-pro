package person

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) ListPersons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	name := c.Query("name")
	idCard := c.Query("idCard")

	persons, total, err := h.service.ListPersons(page, pageSize, name, idCard)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, persons, total, page, pageSize)
}

func (h *Handler) CreatePerson(c *gin.Context) {
	var person Person
	if err := c.ShouldBindJSON(&person); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if person.Name == "" {
		response.ParamError(c, "姓名不能为空")
		return
	}

	if err := h.service.CreatePerson(&person); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, person)
}

func (h *Handler) UpdatePerson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	var person Person
	if err := c.ShouldBindJSON(&person); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	person.ID = uint(id)
	if err := h.service.UpdatePerson(&person); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, person)
}

func (h *Handler) DeletePerson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	if err := h.service.DeletePerson(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RestorePerson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	if err := h.service.RestorePerson(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetPersonByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	person, err := h.service.GetPersonByID(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, "人员不存在")
		return
	}

	response.Success(c, person)
}

func (h *Handler) GetAllPersonsSimple(c *gin.Context) {
	persons, err := h.service.GetAllPersonsSimple()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, persons)
}

func (h *Handler) AddPhone(c *gin.Context) {
	personID, err := parsePersonID(c)
	if err != nil {
		response.ParamError(c, "无效人员ID")
		return
	}

	var phone PersonPhone
	if err := c.ShouldBindJSON(&phone); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	phone.PersonID = personID
	if err := h.service.CreatePhone(&phone); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, phone)
}

func (h *Handler) UpdatePhone(c *gin.Context) {
	phoneID, err := parseSubID(c, "phoneId")
	if err != nil {
		response.ParamError(c, "无效电话ID")
		return
	}

	var phone PersonPhone
	if err := c.ShouldBindJSON(&phone); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	phone.ID = phoneID
	if err := h.service.UpdatePhone(&phone); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeletePhone(c *gin.Context) {
	phoneID, err := parseSubID(c, "phoneId")
	if err != nil {
		response.ParamError(c, "无效电话ID")
		return
	}

	if err := h.service.DeletePhone(phoneID); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) AddEmail(c *gin.Context) {
	personID, err := parsePersonID(c)
	if err != nil {
		response.ParamError(c, "无效人员ID")
		return
	}

	var email PersonEmail
	if err := c.ShouldBindJSON(&email); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	email.PersonID = personID
	if err := h.service.CreateEmail(&email); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, email)
}

func (h *Handler) UpdateEmail(c *gin.Context) {
	emailID, err := parseSubID(c, "emailId")
	if err != nil {
		response.ParamError(c, "无效邮箱ID")
		return
	}

	var email PersonEmail
	if err := c.ShouldBindJSON(&email); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	email.ID = emailID
	if err := h.service.UpdateEmail(&email); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeleteEmail(c *gin.Context) {
	emailID, err := parseSubID(c, "emailId")
	if err != nil {
		response.ParamError(c, "无效邮箱ID")
		return
	}

	if err := h.service.DeleteEmail(emailID); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) AddBankCard(c *gin.Context) {
	personID, err := parsePersonID(c)
	if err != nil {
		response.ParamError(c, "无效人员ID")
		return
	}

	var card PersonBankCard
	if err := c.ShouldBindJSON(&card); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	card.PersonID = personID
	if err := h.service.CreateBankCard(&card); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, card)
}

func (h *Handler) UpdateBankCard(c *gin.Context) {
	cardID, err := parseSubID(c, "cardId")
	if err != nil {
		response.ParamError(c, "无效银行卡ID")
		return
	}

	var card PersonBankCard
	if err := c.ShouldBindJSON(&card); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	card.ID = cardID
	if err := h.service.UpdateBankCard(&card); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) DeleteBankCard(c *gin.Context) {
	cardID, err := parseSubID(c, "cardId")
	if err != nil {
		response.ParamError(c, "无效银行卡ID")
		return
	}

	if err := h.service.DeleteBankCard(cardID); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func parsePersonID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	return uint(id), err
}

func parseSubID(c *gin.Context, key string) (uint, error) {
	idStr := c.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 64)
	return uint(id), err
}
