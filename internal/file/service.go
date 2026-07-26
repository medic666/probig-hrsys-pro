package file

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
)

func getStoragePath() string {
	p := config.Get("system.file_storage_path")
	if p == "" {
		p = os.Getenv("FILE_STORAGE_PATH")
		if p == "" {
			p = "./upload"
		}
	}
	return p
}

func Upload(fileHeader *multipart.FileHeader, targetType string, targetID uint, uploaderID uint, uploaderName string, clientIP string) (*File, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileType := strings.TrimPrefix(ext, ".")
	if fileType == "" {
		fileType = "unknown"
	}

	storagePath := getStoragePath()
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	savedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
	savedPath := filepath.Join(storagePath, savedName)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(savedPath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	file := &File{
		FileName:   fileHeader.Filename,
		FileType:   fileType,
		FileSize:   fileHeader.Size,
		FilePath:   savedPath,
		UploaderID: uploaderID,
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(file).Error; err != nil {
			return err
		}

		audit.CreateAuditLog(tx, uploaderID, uploaderName, "file", file.ID, file.FileName, "新增", nil, file, clientIP)

		if targetType != "" && targetID > 0 {
			relation := &FileRelation{
				FileID:     file.ID,
				TargetType: targetType,
				TargetID:   targetID,
			}
			if err := tx.Create(relation).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		os.Remove(savedPath)
		return nil, err
	}

	return file, nil
}

func GetFile(id uint) (*File, error) {
	return GetFileByID(id)
}

func Download(id uint) (*File, error) {
	file, err := GetFileByID(id)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在")
	}

	return file, nil
}

func Delete(id uint, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var f File
		if err := tx.First(&f, id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&f).Error; err != nil {
			return err
		}

		audit.CreateAuditLog(tx, operatorID, operatorName, "file", f.ID, f.FileName, "删除", f, nil, clientIP)
		return nil
	})
}

func Restore(id uint, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&File{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}

		var f File
		if err := tx.Unscoped().First(&f, id).Error; err != nil {
			return err
		}

		audit.CreateAuditLog(tx, operatorID, operatorName, "file", f.ID, f.FileName, "恢复", nil, f, clientIP)
		return nil
	})
}

func List(filter FileFilter) ([]FileVO, int64, error) {
	files, total, err := ListFiles(filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]FileVO, 0, len(files))
	for _, f := range files {
		result = append(result, FileVO{
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

	return result, total, nil
}

func ListTrash(filter FileFilter) ([]FileVO, int64, error) {
	files, total, err := ListTrashFiles(filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]FileVO, 0, len(files))
	for _, f := range files {
		result = append(result, FileVO{
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

	return result, total, nil
}

func GetFileRelations(fileID uint) ([]FileRelationVO, error) {
	relations, err := GetRelations(fileID)
	if err != nil {
		return nil, err
	}

	result := make([]FileRelationVO, 0, len(relations))
	for _, r := range relations {
		targetName := getTargetName(r.TargetType, r.TargetID)
		result = append(result, FileRelationVO{
			ID:         r.ID,
			TargetType: r.TargetType,
			TargetID:   r.TargetID,
			TargetName: targetName,
		})
	}

	return result, nil
}

func UpdateRelations(fileID uint, req *RelationUpdateRequest, operatorID uint, operatorName string, clientIP string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, target := range req.Add {
			var count int64
			tx.Model(&FileRelation{}).Where("file_id = ? AND target_type = ? AND target_id = ?", fileID, target.TargetType, target.TargetID).Count(&count)
			if count == 0 {
				relation := &FileRelation{
					FileID:     fileID,
					TargetType: target.TargetType,
					TargetID:   target.TargetID,
				}
				if err := tx.Create(relation).Error; err != nil {
					return err
				}
			}
		}

		for _, target := range req.Remove {
			var relation FileRelation
			if err := tx.Where("file_id = ? AND target_type = ? AND target_id = ?", fileID, target.TargetType, target.TargetID).First(&relation).Error; err == nil {
				if err := tx.Delete(&relation).Error; err != nil {
					return err
				}
			}
		}

		var f File
		if err := tx.First(&f, fileID).Error; err == nil {
			audit.CreateAuditLog(tx, operatorID, operatorName, "file_relation", fileID, f.FileName, "修改关联", nil, nil, clientIP)
		}

		return nil
	})
}

func getTargetName(targetType string, targetID uint) string {
	db := database.DB
	switch targetType {
	case "person":
		var person database.Person
		if err := db.Unscoped().Where("id = ?", targetID).First(&person).Error; err == nil {
			return person.Name
		}
	case "company":
		var company database.Company
		if err := db.Unscoped().Where("id = ?", targetID).First(&company).Error; err == nil {
			return company.Name
		}
	}
	return fmt.Sprintf("%s-%d", targetType, targetID)
}

func GetFilesByTargetService(targetType string, targetID uint) ([]File, error) {
	return GetFilesByTarget(targetType, targetID)
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
}
