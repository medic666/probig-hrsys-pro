package file

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

func ListFilesHandler(c *gin.Context) {
	var filter FileFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	files, total, err := List(filter)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: files, Total: total})
}

func UploadFileHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, response.ParamError, "请选择文件")
		return
	}

	targetType := c.PostForm("target_type")
	targetIDStr := c.PostForm("target_id")
	targetID, _ := strconv.ParseUint(targetIDStr, 10, 64)

	uploaderID := utils.GetUserID(c)
	uploaderName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	f, err := Upload(fileHeader, targetType, uint(targetID), uploaderID, uploaderName, clientIP)
	if err != nil {
		response.Error(c, response.InternalError, "上传失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "上传成功", FileVO{
		ID:         f.ID,
		FileName:   f.FileName,
		FileType:   f.FileType,
		FileSize:   f.FileSize,
		FilePath:   f.FilePath,
		UploaderID: f.UploaderID,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
	})
}

func DownloadFileHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	file, err := Download(id)
	if err != nil {
		response.Error(c, response.NotFound, "文件不存在")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.FileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(file.FileSize, 10))

	f, err := os.Open(file.FilePath)
	if err != nil {
		response.Error(c, response.InternalError, "读取文件失败")
		return
	}
	defer f.Close()

	io.Copy(c.Writer, f)
}

func DeleteFileHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	if err := Delete(id, operatorID, operatorName, clientIP); err != nil {
		response.Error(c, response.InternalError, "删除失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreFileHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	if err := Restore(id, operatorID, operatorName, clientIP); err != nil {
		response.Error(c, response.InternalError, "恢复失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "恢复成功", nil)
}

func ListTrashHandler(c *gin.Context) {
	var filter FileFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	files, total, err := ListTrash(filter)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: files, Total: total})
}

func GetRelationsHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	relations, err := GetFileRelations(id)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, relations)
}

func UpdateRelationsHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	var req RelationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	if err := UpdateRelations(id, &req, operatorID, operatorName, clientIP); err != nil {
		response.Error(c, response.InternalError, "更新失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}
