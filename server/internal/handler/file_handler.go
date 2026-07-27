package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "请选择文件")
		return
	}
	defer file.Close()

	uploadDir := "./data/uploads"
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

	f := model.File{
		Name:         saveName,
		OriginalName: header.Filename,
		Path:         savePath,
		Size:         size,
		MimeType:     header.Header.Get("Content-Type"),
	}
	dao.DB.Create(&f)

	utils.Success(c, gin.H{
		"id":            f.ID,
		"name":          f.OriginalName,
		"size":          f.Size,
		"mime_type":     f.MimeType,
	})
}

func DownloadFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var f model.File
	if err := dao.DB.First(&f, id).Error; err != nil {
		utils.Error(c, "文件不存在")
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
	r := model.FileRelation{FileID: req.FileID, TargetType: req.TargetType, TargetID: req.TargetID}
	dao.DB.Create(&r)
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
	dao.DB.Delete(&model.FileRelation{}, req.ID)
	utils.SuccessWithMsg(c, "解除关联成功", nil)
}

func GetFilesByTarget(c *gin.Context) {
	targetType := c.Query("target_type")
	targetIDStr := c.Query("target_id")
	if targetType == "" || targetIDStr == "" {
		utils.BadRequest(c, "参数错误")
		return
	}
	targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)

	var relations []model.FileRelation
	dao.DB.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&relations)

	var fileIDs []uint
	for _, r := range relations {
		fileIDs = append(fileIDs, r.FileID)
	}

	var files []model.File
	if len(fileIDs) > 0 {
		dao.DB.Where("id IN ?", fileIDs).Find(&files)
	}

	type result struct {
		Relation model.FileRelation `json:"relation"`
		File     model.File         `json:"file"`
	}
	var list []result
	for _, rel := range relations {
		for _, f := range files {
			if f.ID == rel.FileID {
				list = append(list, result{Relation: rel, File: f})
			}
		}
	}
	utils.Success(c, list)
}
