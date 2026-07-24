package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/middleware"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetFileList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total, err := service.GetFileList(page, pageSize, keyword)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func GetFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	f, err := service.GetFile(uint(id))
	if err != nil {
		response.Error(c, "文件不存在")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+f.Name)
	c.Header("Content-Type", f.MimeType)
	c.Data(200, f.MimeType, f.Content)
}

func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, "请选择文件")
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, "读取文件失败")
		return
	}
	claims := middleware.GetUserClaims(c)
	f := &models.File{
		Name:       header.Filename,
		Size:       header.Size,
		MimeType:   header.Header.Get("Content-Type"),
		Content:    content,
		UploaderID: 0,
	}
	if claims != nil {
		f.UploaderID = claims.UserID
	}
	if err := service.CreateFile(c, f); err != nil {
		response.Error(c, "上传失败")
		return
	}
	response.Success(c, f)
}

func DeleteFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteFile(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func RestoreFile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreFile(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func CreateFileRelation(c *gin.Context) {
	var fr models.FileRelation
	if err := c.ShouldBindJSON(&fr); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.CreateFileRelation(c, &fr); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, fr)
}

func GetFileRelationsByTarget(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID, _ := strconv.ParseUint(c.Query("target_id"), 10, 64)
	relations, err := service.GetFileRelationsByTarget(targetType, uint(targetID))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, relations)
}

func DeleteFileRelation(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteFileRelation(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}
