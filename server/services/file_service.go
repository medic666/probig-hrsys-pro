package services

import (
	"probig/database"
	"probig/models"

	"gorm.io/gorm"
)

func ListFiles(query, mimeType string, offset, limit int) ([]models.File, int64, error) {
	var files []models.File
	var total int64
	db := database.DB.Model(&models.File{})
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("name LIKE ?", like)
	}
	if mimeType != "" {
		db = db.Where("mime_type LIKE ?", "%"+mimeType+"%")
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error
	return files, total, err
}

func GetFile(id uint) (*models.File, error) {
	var file models.File
	err := database.DB.First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func UploadFile(file *models.File) error {
	return database.DB.Create(file).Error
}

func DeleteFile(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.File{}, id).Error; err != nil {
			return err
		}
		tx.Where("file_id = ?", id).Delete(&models.FileRelation{})
		return nil
	})
}

func RestoreFile(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.File{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Where("file_id = ?", id).Model(&models.FileRelation{}).Update("deleted_at", nil)
		return nil
	})
}

func AddFileRelation(fileID uint, targetType string, targetID uint) error {
	rel := models.FileRelation{
		FileID:     fileID,
		TargetType: targetType,
		TargetID:   targetID,
	}
	return database.DB.Create(&rel).Error
}

func RemoveFileRelation(relationID uint) error {
	return database.DB.Delete(&models.FileRelation{}, relationID).Error
}

func GetFileRelations(targetType string, targetID uint) ([]models.FileRelation, error) {
	var relations []models.FileRelation
	err := database.DB.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&relations).Error
	return relations, err
}

func GetFileRelationsWithFiles(targetType string, targetID uint) ([]models.File, error) {
	var files []models.File
	database.DB.Table("files f").
		Joins("JOIN file_relations fr ON f.id = fr.file_id").
		Where("fr.target_type = ? AND fr.target_id = ? AND fr.deleted_at IS NULL", targetType, targetID).
		Find(&files)
	return files, nil
}
