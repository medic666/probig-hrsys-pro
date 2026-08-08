package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"probig/server/internal/config"
	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetFiles(c *gin.Context) {
	pageReq := utils.BindPage(c)
	dateStart, dateEnd := utils.BindDateRange(c)
	list, total, err := service.GetFileList(pageReq.PageNum, pageReq.PageSize,
		c.Query("name"), c.Query("mime_type"), dateStart, dateEnd)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
}

func DeleteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteFileByID(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreFileByID(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedFiles(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedFileList(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
}

func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	maxSizeMB := service.GetFileMaxSizeMB()
	if header.Size > maxSizeMB*1024*1024 {
		utils.BadRequest(c, fmt.Sprintf("文件大小超过限制（最大 %dMB）", maxSizeMB))
		return
	}
	if service.IsBlockedFileExt(filepath.Ext(header.Filename)) {
		utils.BadRequest(c, "不允许上传该类型文件")
		return
	}

	uploadDir := config.ResolvePath(config.AppConfig.FileStorage.Path)
	os.MkdirAll(uploadDir, 0755)
	ext := filepath.Ext(header.Filename)
	saveName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, saveName)
	out, err := os.Create(savePath)
	if err != nil {
		utils.Error(c, "文件保存失败")
		return
	}
	defer out.Close()
	size, _ := io.Copy(out, file)
	md5hash, _ := service.ComputeFileMD5(savePath)

	if dup, err := service.FindFileByMD5(md5hash); err == nil {
		os.Remove(savePath)
		utils.SuccessWithMsg(c, "文件已存在(复用已有记录)", gin.H{"id": dup.ID, "name": dup.OriginalName, "duplicate": true})
		return
	}

	f := model.File{Name: saveName, OriginalName: header.Filename, Path: savePath, Size: size, MimeType: header.Header.Get("Content-Type"), MD5: md5hash}
	dao.DBFromContext(c.Request.Context()).Create(&f)
	utils.Success(c, gin.H{"id": f.ID, "name": f.OriginalName, "size": f.Size, "mime_type": f.MimeType, "duplicate": false})
}

func DownloadFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	f, err := service.GetFileUnscoped(uint(id))
	if err != nil {
		c.String(404, "文件不存在")
		return
	}
	c.File(f.Path)
}

type associateReq struct {
	FileID     uint   `json:"file_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
}

func AssociateFile(c *gin.Context) {
	var req associateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if service.IsFileAssociated(req.FileID, req.TargetType, req.TargetID) {
		utils.SuccessWithMsg(c, "已存在关联", nil)
		return
	}
	dao.DBFromContext(c.Request.Context()).Create(&model.FileRelation{FileID: req.FileID, TargetType: req.TargetType, TargetID: req.TargetID})
	utils.SuccessWithMsg(c, "关联成功", nil)
}

func DisassociateFile(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	dao.DBFromContext(c.Request.Context()).Delete(&model.FileRelation{}, req.ID)
	utils.SuccessWithMsg(c, "解除关联成功", nil)
}

func GetFilesByTarget(c *gin.Context) {
	targetType, targetIDStr := c.Query("target_type"), c.Query("target_id")
	if targetType == "" || targetIDStr == "" {
		utils.BadRequest(c, "参数错误")
		return
	}
	targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)
	list, _ := service.GetFilesForTarget(targetType, uint(targetID))
	utils.Success(c, list)
}

func GetFileAssociations(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := service.GetFileAssociations(uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, list)
}

func PermanentDeleteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	usedCount, err := service.PermanentDeleteFile(c.Request.Context(), uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	if usedCount > 0 {
		utils.Error(c, fmt.Sprintf("该文件仍被 %d 个实体使用", usedCount))
		return
	}
	utils.SuccessWithMsg(c, "已彻底删除", nil)
}

func CleanOrphanFiles(c *gin.Context) {
	count, err := service.CleanOrphanFiles(c.Request.Context())
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, fmt.Sprintf("已清理 %d 个孤儿文件", count), gin.H{"count": count})
}
