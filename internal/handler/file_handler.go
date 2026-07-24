package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"probig/internal/config"
	"probig/internal/service"
	"probig/internal/utils"
)

type FileHandler struct {
	svc *service.FileService
}

func NewFileHandler(cfg *config.Config) *FileHandler {
	return &FileHandler{svc: service.NewFileService(cfg)}
}

func (h *FileHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	files, total, err := h.svc.List(page, pageSize)
	if err != nil {
		utils.InternalError(c, "查询文件列表失败")
		return
	}
	utils.SuccessPage(c, files, total, page, pageSize)
}

func (h *FileHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)

	result, err := h.svc.Upload(auditSvc, ctx, header.Filename, header.Header.Get("Content-Type"), header.Size, file)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.Success(c, result)
}

func (h *FileHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}
	f, err := h.svc.Get(uint(id))
	if err != nil {
		utils.NotFound(c, "文件不存在")
		return
	}
	c.FileAttachment(f.StoragePath, f.OriginalName)
}

func (h *FileHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的ID")
		return
	}

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)
	if err := h.svc.Delete(auditSvc, ctx, uint(id)); err != nil {
		utils.InternalError(c, "删除文件失败")
		return
	}
	utils.Success(c, nil)
}

func (h *FileHandler) Associate(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的文件ID")
		return
	}

	var req struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   uint   `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)
	if err := h.svc.Associate(auditSvc, ctx, uint(fileID), req.TargetType, req.TargetID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

func (h *FileHandler) Disassociate(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的文件ID")
		return
	}
	targetID, err := strconv.ParseUint(c.Param("targetId"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的关联ID")
		return
	}

	auditSvc := service.NewAuditService()
	ctx := getEventContext(c)
	if err := h.svc.Disassociate(auditSvc, ctx, uint(fileID), uint(targetID)); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}
