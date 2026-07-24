package dao

import (
	"probig/internal/models"
)

func GetFileList(page, pageSize int, keyword string) ([]models.File, int64, error) {
	var list []models.File
	var total int64
	q := DB().Model(&models.File{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetFileByID(id uint) (*models.File, error) {
	var f models.File
	if err := DB().First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func CreateFile(f *models.File) error {
	return DB().Create(f).Error
}

func DeleteFile(id uint) error {
	tx := DB().Begin()
	if err := tx.Where("file_id = ?", id).Delete(&models.FileRelation{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&models.File{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func RestoreFile(id uint) error {
	tx := DB().Begin()
	if err := tx.Unscoped().Model(&models.File{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.FileRelation{}).Where("file_id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func CreateFileRelation(fr *models.FileRelation) error {
	return DB().Create(fr).Error
}

func GetFileRelationsByTarget(targetType string, targetID uint) ([]models.FileRelation, error) {
	var list []models.FileRelation
	if err := DB().Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func DeleteFileRelation(id uint) error {
	return DB().Delete(&models.FileRelation{}, id).Error
}
