package company

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

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	name := c.Query("name")
	creditCode := c.Query("creditCode")

	companies, total, err := h.service.List(page, pageSize, name, creditCode)
	if err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Page(c, companies, total, page, pageSize)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	company, err := h.service.GetByID(uint(id))
	if err != nil {
		response.Error(c, response.CodeNotFound, "company not found")
		return
	}

	response.Success(c, company)
}

func (h *Handler) Create(c *gin.Context) {
	var company Company
	if err := c.ShouldBindJSON(&company); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	operatorID, _ := userID.(uint)
	operatorName, _ := username.(string)

	if err := h.service.Create(&company, operatorID, operatorName); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, company)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	var company Company
	if err := c.ShouldBindJSON(&company); err != nil {
		response.ParamError(c, err.Error())
		return
	}
	company.ID = uint(id)

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	operatorID, _ := userID.(uint)
	operatorName, _ := username.(string)

	if err := h.service.Update(&company, operatorID, operatorName); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, company)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	operatorID, _ := userID.(uint)
	operatorName, _ := username.(string)

	if err := h.service.Delete(uint(id), operatorID, operatorName); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) Restore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "invalid id")
		return
	}

	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	operatorID, _ := userID.(uint)
	operatorName, _ := username.(string)

	if err := h.service.Restore(uint(id), operatorID, operatorName); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}
