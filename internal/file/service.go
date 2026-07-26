package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"probig/internal/pkg/config"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) Upload(fileName string, fileSize int64, reader io.Reader, uploadUser string) (uint, error) {
	storagePath := config.FileStoragePath
	os.MkdirAll(storagePath, 0755)

	ext := filepath.Ext(fileName)
	storedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	storedPath := filepath.Join(storagePath, storedName)

	dst, err := os.Create(storedPath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return 0, err
	}

	fileModel := FileModel{
		FileName:   fileName,
		FilePath:   storedPath,
		FileSize:   fileSize,
		FileType:   ext,
		UploadUser: uploadUser,
	}
	if err := s.DB.Create(&fileModel).Error; err != nil {
		return 0, err
	}
	return fileModel.ID, nil
}

func (s *Service) List(pageNum, pageSize int, fileName, fileType string) ([]FileModel, int64, error) {
	var list []FileModel
	var total int64
	db := s.DB.Model(&FileModel{})
	if fileName != "" {
		db = db.Where("file_name like ?", "%"+fileName+"%")
	}
	if fileType != "" {
		db = db.Where("file_type = ?", fileType)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *Service) Delete(id uint) error {
	return s.DB.Delete(&FileModel{}, id).Error
}

func (s *Service) Restore(id uint) error {
	return s.DB.Unscoped().Model(&FileModel{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (s *Service) GetByID(id uint) (*FileModel, error) {
	var f FileModel
	err := s.DB.First(&f, id).Error
	return &f, err
}

func (s *Service) AddRelation(fileID uint, targetType string, targetID uint) error {
	return s.DB.Create(&FileRelation{FileID: fileID, TargetType: targetType, TargetID: targetID}).Error
}

func (s *Service) DeleteRelation(relationID uint) error {
	return s.DB.Delete(&FileRelation{}, relationID).Error
}

func (s *Service) GetRelations(fileID uint) ([]FileRelation, error) {
	var list []FileRelation
	err := s.DB.Where("file_id = ?", fileID).Find(&list).Error
	return list, err
}

func (s *Service) GetTargetFiles(targetType string, targetID uint) ([]FileModel, error) {
	var relations []FileRelation
	s.DB.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&relations)
	var files []FileModel
	for _, rel := range relations {
		var f FileModel
		if err := s.DB.First(&f, rel.FileID).Error; err == nil {
			files = append(files, f)
		}
	}
	return files, nil
}
