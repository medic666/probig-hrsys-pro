package file

import (
	"fmt"
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

func (h *Handler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ParamError(c, "请选择文件")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.ServerError(c, "文件读取失败")
		return
	}
	defer src.Close()

	content := make([]byte, file.Size)
	if _, err := src.Read(content); err != nil {
		response.ServerError(c, "文件读取失败")
		return
	}

	uploaded, err := h.service.Upload(file.Filename, file.Size, file.Header.Get("Content-Type"), content, 0)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, uploaded)
}

func (h *Handler) Download(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	file, err := h.service.GetFile(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, "文件不存在")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))
	c.Header("Content-Type", file.MimeType)
	c.Data(200, file.MimeType, file.Content)
}

func (h *Handler) GetFileInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	file, err := h.service.GetFileInfo(uint(id))
	if err != nil {
		response.ErrorWithMsg(c, "文件不存在")
		return
	}

	response.Success(c, file)
}

func (h *Handler) ListFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	name := c.Query("name")
	mimeType := c.Query("mimeType")

	files, total, err := h.service.ListFiles(page, pageSize, name, mimeType)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Page(c, files, total, page, pageSize)
}

func (h *Handler) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	if err := h.service.DeleteFile(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RestoreFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	if err := h.service.RestoreFile(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) AddRelation(c *gin.Context) {
	var relation FileRelation
	if err := c.ShouldBindJSON(&relation); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if relation.FileID == 0 || relation.TargetID == 0 || relation.TargetType == "" {
		response.ParamError(c, "参数不完整")
		return
	}

	if err := h.service.AddRelation(relation.FileID, relation.TargetID, relation.TargetType); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) RemoveRelation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效ID")
		return
	}

	if err := h.service.RemoveRelation(uint(id)); err != nil {
		response.ErrorWithMsg(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *Handler) GetFilesByTarget(c *gin.Context) {
	targetType := c.Query("targetType")
	targetIDStr := c.Query("targetId")
	if targetType == "" || targetIDStr == "" {
		response.ParamError(c, "参数不完整")
		return
	}

	targetID, err := strconv.ParseUint(targetIDStr, 10, 64)
	if err != nil {
		response.ParamError(c, "无效targetId")
		return
	}

	files, err := h.service.GetFilesByTarget(targetType, uint(targetID))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, files)
}
