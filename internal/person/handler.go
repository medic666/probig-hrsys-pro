package person

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/pkg/config"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"
)

type Handler struct {
	svc *Service
}

func NewHandler() *Handler {
	return &Handler{svc: DefaultService}
}

func getPageParams(c *gin.Context) (int, int) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	if pageNum < 1 {
		pageNum = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", config.Get("system.page_size")))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageNum, pageSize
}

func (h *Handler) ListPersons(c *gin.Context) {
	pageNum, pageSize := getPageParams(c)
	name := c.Query("name")
	idCard := c.Query("idcard")
	attendanceGroup := c.Query("attendance_group")
	employmentStatus := c.Query("employment_status")

	persons, total, err := h.svc.ListPersons(pageNum, pageSize, name, idCard, attendanceGroup, employmentStatus)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}

	response.Success(c, &response.PageResult{
		List:  persons,
		Total: total,
	})
}

func (h *Handler) CreatePerson(c *gin.Context) {
	var req CreatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	person, err := h.svc.CreatePerson(&req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "创建成功", person)
}

func (h *Handler) UpdatePerson(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID无效")
		return
	}

	var req UpdatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdatePerson(id, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

func (h *Handler) DeletePerson(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID无效")
		return
	}

	if err := h.svc.DeletePerson(id); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func (h *Handler) GetPersonDetail(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID无效")
		return
	}

	detail, err := h.svc.GetPersonDetail(id)
	if err != nil {
		response.ErrorWithMsg(c, "人员不存在")
		return
	}

	response.Success(c, detail)
}

func (h *Handler) ListTrashPersons(c *gin.Context) {
	pageNum, pageSize := getPageParams(c)

	persons, total, err := h.svc.ListTrashPersons(pageNum, pageSize)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}

	response.Success(c, &response.PageResult{
		List:  persons,
		Total: total,
	})
}

func (h *Handler) RestorePerson(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID无效")
		return
	}

	if err := h.svc.RestorePerson(id); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "恢复成功", nil)
}

func (h *Handler) CreatePhone(c *gin.Context) {
	personID, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID无效")
		return
	}

	var req CreatePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	phone, err := h.svc.CreatePhone(personID, &req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "添加成功", phone)
}

func (h *Handler) UpdatePhone(c *gin.Context) {
	phoneID, err := utils.ParseID(c, "phoneId")
	if err != nil {
		response.Error(c, response.ParamError, "电话ID无效")
		return
	}

	var req UpdatePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdatePhone(phoneID, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

func (h *Handler) DeletePhone(c *gin.Context) {
	phoneID, err := utils.ParseID(c, "phoneId")
	if err != nil {
		response.Error(c, response.ParamError, "电话ID无效")
		return
	}

	if err := h.svc.DeletePhone(phoneID); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func (h *Handler) CreateEmail(c *gin.Context) {
	personID, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID无效")
		return
	}

	var req CreateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	email, err := h.svc.CreateEmail(personID, &req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "添加成功", email)
}

func (h *Handler) UpdateEmail(c *gin.Context) {
	emailID, err := utils.ParseID(c, "emailId")
	if err != nil {
		response.Error(c, response.ParamError, "邮箱ID无效")
		return
	}

	var req UpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateEmail(emailID, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

func (h *Handler) DeleteEmail(c *gin.Context) {
	emailID, err := utils.ParseID(c, "emailId")
	if err != nil {
		response.Error(c, response.ParamError, "邮箱ID无效")
		return
	}

	if err := h.svc.DeleteEmail(emailID); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func (h *Handler) CreateBankCard(c *gin.Context) {
	personID, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID无效")
		return
	}

	var req CreateBankCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	card, err := h.svc.CreateBankCard(personID, &req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "添加成功", card)
}

func (h *Handler) UpdateBankCard(c *gin.Context) {
	cardID, err := utils.ParseID(c, "cardId")
	if err != nil {
		response.Error(c, response.ParamError, "银行卡ID无效")
		return
	}

	var req UpdateBankCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateBankCard(cardID, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

func (h *Handler) DeleteBankCard(c *gin.Context) {
	cardID, err := utils.ParseID(c, "cardId")
	if err != nil {
		response.Error(c, response.ParamError, "银行卡ID无效")
		return
	}

	if err := h.svc.DeleteBankCard(cardID); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}
