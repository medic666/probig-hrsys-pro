package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/medic666/probig/internal/models"
	"github.com/medic666/probig/internal/response"
	"github.com/medic666/probig/internal/services"
)

type FileHandler struct {
	svc *services.FileService
}

func NewFileHandler(svc *services.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

func (h *FileHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploadedFile, err := h.svc.Upload(
		c.GetUint("user_id"),
		header.Filename,
		file,
		header.Size,
		mimeType,
		c.ClientIP(),
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, uploadedFile)
}

func (h *FileHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	list, total, err := h.svc.List(page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPage(c, list, total, page, pageSize)
}

func (h *FileHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	if err := h.svc.Delete(uint(id), c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *FileHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	list, _, err := h.svc.List(1, 100000, "")
	if err != nil {
		response.NotFound(c, "文件不存在")
		return
	}
	for _, f := range list {
		if f.ID == uint(id) {
			c.FileAttachment(f.Path, f.OriginalName)
			return
		}
	}
	response.NotFound(c, "文件不存在")
}

func (h *FileHandler) CreateAssociation(c *gin.Context) {
	var req models.FileAssociationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	assoc, err := h.svc.CreateAssociation(req, c.GetUint("user_id"), c.ClientIP())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, assoc)
}

func (h *FileHandler) DeleteAssociation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "ID 无效")
		return
	}
	if err := h.svc.DeleteAssociation(uint(id), c.GetUint("user_id"), c.ClientIP()); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *FileHandler) GetAssociations(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID, err := strconv.ParseUint(c.Query("target_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "target_id 无效")
		return
	}
	assocs, err := h.svc.GetAssociations(targetType, uint(targetID))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, assocs)
}
