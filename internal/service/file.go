package service

import (
	"probig/internal/dao"
	"probig/internal/middleware"
	"probig/internal/models"

	"github.com/gin-gonic/gin"
)

func GetFileList(page, pageSize int, keyword string) ([]models.File, int64, error) {
	return dao.GetFileList(page, pageSize, keyword)
}

func GetFile(id uint) (*models.File, error) {
	return dao.GetFileByID(id)
}

func CreateFile(c *gin.Context, f *models.File) error {
	if err := dao.CreateFile(f); err != nil {
		return err
	}
	middleware.RecordAudit(c, "上传", "file", f.ID, nil, f, "")
	return nil
}

func DeleteFile(c *gin.Context, id uint) error {
	f, err := dao.GetFileByID(id)
	if err != nil {
		return err
	}
	if err := dao.DeleteFile(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "file", id, f, nil, "")
	return nil
}

func RestoreFile(c *gin.Context, id uint) error {
	if err := dao.RestoreFile(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "恢复", "file", id, nil, nil, "")
	return nil
}

func CreateFileRelation(c *gin.Context, fr *models.FileRelation) error {
	if err := dao.CreateFileRelation(fr); err != nil {
		return err
	}
	middleware.RecordAudit(c, "关联", "file_relation", fr.ID, nil, fr, "")
	return nil
}

func GetFileRelationsByTarget(targetType string, targetID uint) ([]models.FileRelation, error) {
	return dao.GetFileRelationsByTarget(targetType, targetID)
}

func DeleteFileRelation(c *gin.Context, id uint) error {
	if err := dao.DeleteFileRelation(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "解除关联", "file_relation", id, nil, nil, "")
	return nil
}
