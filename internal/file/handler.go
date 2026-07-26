package file

import (
	"io"
	"os"
	"strconv"

	"probig/internal/pkg/config"
	"probig/internal/pkg/middleware"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.Use(middleware.Auth)
	r.GET("", middleware.Permission("file:read"), List)
	r.POST("/upload", middleware.Permission("file:write"), Upload)
	r.DELETE("/:id", middleware.Permission("file:delete"), Delete)
	r.POST("/:id/restore", middleware.Permission("file:write"), Restore)
	r.GET("/:id/download", middleware.Permission("file:read"), Download)
	r.POST("/:id/relations", middleware.Permission("file:write"), AddRelation)
	r.DELETE("/:id/relations/:relationID", middleware.Permission("file:write"), DeleteRelation)
	r.GET("/:id/relations", middleware.Permission("file:read"), GetRelations)
	r.GET("/target/:targetType/:targetID", middleware.Permission("file:read"), GetTargetFiles)
}

func List(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.List(pageNum, pageSize, c.Query("file_name"), c.Query("file_type"))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, "请选择文件")
		return
	}
	defer file.Close()

	uploadLimitStr := config.GetConfig("system.upload_limit")
	if uploadLimitStr != "" {
		limitMB, _ := strconv.ParseInt(uploadLimitStr, 10, 64)
		if limitMB > 0 && header.Size > limitMB*1024*1024 {
			response.Error(c, "文件大小超过限制")
			return
		}
	}

	uname, _ := c.Get("username")
	svc := NewService()
	id, err := svc.Upload(header.Filename, header.Size, file, uname.(string))
	if err != nil {
		response.Error(c, "上传失败")
		return
	}
	response.Success(c, gin.H{"id": id})
}

func Delete(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.Delete(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func Restore(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.Restore(id); err != nil {
		response.Error(c, "恢复失败")
		return
	}
	response.SuccessMsg(c, "恢复成功")
}

func Download(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	f, err := svc.GetByID(id)
	if err != nil {
		response.Error(c, "文件不存在")
		return
	}
	fileData, err := os.Open(f.FilePath)
	if err != nil {
		response.Error(c, "文件读取失败")
		return
	}
	defer fileData.Close()

	c.Header("Content-Disposition", "attachment; filename="+f.FileName)
	c.Header("Content-Type", "application/octet-stream")
	io.Copy(c.Writer, fileData)
}

func AddRelation(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		TargetType string `json:"target_type" binding:"required"`
		TargetID   uint   `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.AddRelation(id, req.TargetType, req.TargetID); err != nil {
		response.Error(c, "关联失败")
		return
	}
	response.SuccessMsg(c, "关联成功")
}

func DeleteRelation(c *gin.Context) {
	relationID := utils.ParseUint(c.Param("relationID"))
	svc := NewService()
	if err := svc.DeleteRelation(relationID); err != nil {
		response.Error(c, "删除关联失败")
		return
	}
	response.SuccessMsg(c, "删除关联成功")
}

func GetRelations(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	relations, err := svc.GetRelations(id)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, relations)
}

func GetTargetFiles(c *gin.Context) {
	targetType := c.Param("targetType")
	targetID := utils.ParseUint(c.Param("targetID"))
	svc := NewService()
	files, err := svc.GetTargetFiles(targetType, targetID)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, files)
}
