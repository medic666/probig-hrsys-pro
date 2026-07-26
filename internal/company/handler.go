package company

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

func (h *Handler) ListCompanies(c *gin.Context) {
	pageNum, pageSize := getPageParams(c)
	name := c.Query("name")
	creditCode := c.Query("credit_code")

	companies, total, err := h.svc.ListCompanies(pageNum, pageSize, name, creditCode)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}

	response.Success(c, &response.PageResult{
		List:  companies,
		Total: total,
	})
}

func (h *Handler) CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	company, err := h.svc.CreateCompany(&req)
	if err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "创建成功", company)
}

func (h *Handler) UpdateCompany(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "公司ID无效")
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateCompany(id, &req); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

func (h *Handler) DeleteCompany(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "公司ID无效")
		return
	}

	if err := h.svc.DeleteCompany(id); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func (h *Handler) GetCompany(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "公司ID无效")
		return
	}

	company, err := h.svc.GetCompany(id)
	if err != nil {
		response.ErrorWithMsg(c, "公司不存在")
		return
	}

	response.Success(c, company)
}

func (h *Handler) ListTrashCompanies(c *gin.Context) {
	pageNum, pageSize := getPageParams(c)

	companies, total, err := h.svc.ListTrashCompanies(pageNum, pageSize)
	if err != nil {
		response.Error(c, response.InternalError, err.Error())
		return
	}

	response.Success(c, &response.PageResult{
		List:  companies,
		Total: total,
	})
}

func (h *Handler) RestoreCompany(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "公司ID无效")
		return
	}

	if err := h.svc.RestoreCompany(id); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "恢复成功", nil)
}
