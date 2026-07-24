package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"probig/internal/config"
	"probig/internal/database"
	"probig/internal/models"
)

type FileService struct {
	cfg *config.Config
}

func NewFileService(cfg *config.Config) *FileService {
	return &FileService{cfg: cfg}
}

func (s *FileService) Upload(audit *AuditService, ctx EventContext, originalName, mimeType string, size int64, reader io.Reader) (*models.File, error) {
	ext := filepath.Ext(originalName)
	storedName := uuid.New().String() + ext

	uploadDir := s.cfg.UploadDir
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败")
	}

	storagePath := filepath.Join(uploadDir, storedName)
	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		os.Remove(storagePath)
		return nil, fmt.Errorf("保存文件失败")
	}

	file := &models.File{
		Filename:     storedName,
		OriginalName: originalName,
		MimeType:     mimeType,
		Size:         size,
		StoragePath:  storagePath,
		UploadedBy:   ctx.UserID,
	}
	if err := database.DB.Create(file).Error; err != nil {
		os.Remove(storagePath)
		return nil, err
	}

	audit.Log(ctx.UserID, ctx.Username, "create", "file", file.ID, originalName, "上传文件", models.JSONMap{
		"original_name": originalName,
		"size":          size,
	})
	return file, nil
}

func (s *FileService) Delete(audit *AuditService, ctx EventContext, fileID uint) error {
	var file models.File
	if err := database.DB.First(&file, fileID).Error; err != nil {
		return err
	}

	database.DB.Where("file_id = ?", fileID).Delete(&models.FileAssociation{})

	if err := os.Remove(file.StoragePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := database.DB.Delete(&file).Error; err != nil {
		return err
	}

	audit.Log(ctx.UserID, ctx.Username, "delete", "file", fileID, file.OriginalName, "删除文件", models.JSONMap{
		"original_name": file.OriginalName,
	})
	return nil
}

func (s *FileService) Get(fileID uint) (*models.File, error) {
	var file models.File
	if err := database.DB.First(&file, fileID).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *FileService) List(page, pageSize int) ([]models.File, int64, error) {
	var files []models.File
	var total int64
	database.DB.Model(&models.File{}).Count(&total)
	if err := database.DB.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, total, nil
}

func (s *FileService) Associate(audit *AuditService, ctx EventContext, fileID uint, targetType string, targetID uint) error {
	assoc := models.FileAssociation{
		FileID:     fileID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	if err := database.DB.Create(&assoc).Error; err != nil {
		return err
	}
	audit.Log(ctx.UserID, ctx.Username, "create", "file_association", assoc.ID, fmt.Sprintf("文件%d关联%d", fileID, targetID), "文件关联", models.JSONMap{
		"file_id":     fileID,
		"target_type": targetType,
		"target_id":   targetID,
	})
	return nil
}

func (s *FileService) Disassociate(audit *AuditService, ctx EventContext, fileID uint, targetID uint) error {
	result := database.DB.Where("file_id = ? AND target_id = ?", fileID, targetID).Delete(&models.FileAssociation{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("关联不存在")
	}
	audit.Log(ctx.UserID, ctx.Username, "delete", "file_association", 0, fmt.Sprintf("文件%d取消关联%d", fileID, targetID), "取消关联", models.JSONMap{
		"file_id":   fileID,
		"target_id": targetID,
	})
	return nil
}

func (s *FileService) GetAssociations(fileID uint) ([]models.FileAssociation, error) {
	var assocs []models.FileAssociation
	if err := database.DB.Where("file_id = ?", fileID).Find(&assocs).Error; err != nil {
		return nil, err
	}
	return assocs, nil
}
