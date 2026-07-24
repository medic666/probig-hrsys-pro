package handlers

import (
	"io"
	"strconv"

	"probig/middleware"
	"probig/models"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		utils.ErrBadRequest(c, "请选择文件")
		return
	}
	f, err := fh.Open()
	if err != nil {
		utils.ErrInternal(c, "打开文件失败")
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		utils.ErrInternal(c, "读取文件失败")
		return
	}

	file := models.File{
		Name:       fh.Filename,
		Size:       fh.Size,
		MimeType:   fh.Header.Get("Content-Type"),
		Content:    content,
		UploaderID: middleware.GetUserID(c),
	}
	if err := services.UploadFile(&file); err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}

	targetType := c.PostForm("target_type")
	targetIDStr := c.PostForm("target_id")
	if targetType != "" && targetIDStr != "" {
		targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)
		services.AddFileRelation(file.ID, targetType, uint(targetID))
	}

	middleware.AuditAction(c, "file", file.ID, "上传", "{}", file)
	utils.Success(c, file)
}

func ListFiles(c *gin.Context) {
	query := c.Query("query")
	mimeType := c.Query("mime_type")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	files, total, err := services.ListFiles(query, mimeType, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, files, total, page, pageSize)
}

func GetFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	file, err := services.GetFile(uint(id))
	if err != nil {
		utils.ErrBadRequest(c, "文件不存在")
		return
	}
	utils.Success(c, file)
}

func DeleteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	before, _ := services.GetFile(uint(id))
	services.DeleteFile(uint(id))
	middleware.AuditAction(c, "file", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func RestoreFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	services.RestoreFile(uint(id))
	middleware.AuditAction(c, "file", uint(id), "恢复", "{}", "restored")
	utils.Success(c, nil)
}

func AddFileRelation(c *gin.Context) {
	var input struct {
		FileID     uint   `json:"file_id"`
		TargetType string `json:"target_type"`
		TargetID   uint   `json:"target_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	services.AddFileRelation(input.FileID, input.TargetType, input.TargetID)
	utils.Success(c, nil)
}

func RemoveFileRelation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	services.RemoveFileRelation(uint(id))
	utils.Success(c, nil)
}

func GetFileRelations(c *gin.Context) {
	files, _ := services.GetFileRelationsWithFiles(
		c.Query("target_type"),
		parseUint(c.Query("target_id")),
	)
	utils.Success(c, files)
}

func parseUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}
