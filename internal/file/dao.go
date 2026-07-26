package file

import (
	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

func db() *gorm.DB {
	return database.DB
}

func CreateFile(file *File) error {
	return db().Create(file).Error
}

func GetFileByID(id uint) (*File, error) {
	var f File
	err := db().Where("id = ?", id).First(&f).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func DeleteFile(id uint) error {
	return db().Delete(&File{}, id).Error
}

func RestoreFile(id uint) error {
	return db().Unscoped().Model(&File{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func ListFiles(filter FileFilter) ([]File, int64, error) {
	query := db().Model(&File{})
	if filter.FileName != "" {
		query = query.Where("file_name LIKE ?", "%"+filter.FileName+"%")
	}
	if filter.FileType != "" {
		query = query.Where("file_type = ?", filter.FileType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var files []File
	pageNum := filter.PageNum
	pageSize := filter.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func ListTrashFiles(filter FileFilter) ([]File, int64, error) {
	query := db().Unscoped().Model(&File{}).Where("deleted_at IS NOT NULL")
	if filter.FileName != "" {
		query = query.Where("file_name LIKE ?", "%"+filter.FileName+"%")
	}
	if filter.FileType != "" {
		query = query.Where("file_type = ?", filter.FileType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var files []File
	pageNum := filter.PageNum
	pageSize := filter.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	if err := query.Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func CreateRelation(relation *FileRelation) error {
	return db().Create(relation).Error
}

func DeleteRelation(id uint) error {
	return db().Delete(&FileRelation{}, id).Error
}

func GetRelations(fileID uint) ([]FileRelation, error) {
	var relations []FileRelation
	err := db().Where("file_id = ?", fileID).Find(&relations).Error
	return relations, err
}

func GetRelationByFileAndTarget(fileID uint, targetType string, targetID uint) (*FileRelation, error) {
	var relation FileRelation
	err := db().Where("file_id = ? AND target_type = ? AND target_id = ?", fileID, targetType, targetID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

func GetFilesByTarget(targetType string, targetID uint) ([]File, error) {
	var files []File
	err := db().Joins("JOIN file_relations ON file_relations.file_id = files.id").
		Where("file_relations.target_type = ? AND file_relations.target_id = ? AND file_relations.deleted_at IS NULL AND files.deleted_at IS NULL",
			targetType, targetID).
		Find(&files).Error
	return files, err
}
