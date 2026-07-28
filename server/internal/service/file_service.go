package service

import (
	"probig/server/internal/dao"
	"probig/server/internal/model"
)

func GetFileList(pageNum, pageSize int, name string) ([]model.File, int64, error) {
	tx := dao.DB.Model(&model.File{})
	if name != "" {
		tx = tx.Where("original_name LIKE ?", "%"+name+"%")
	}
	var total int64
	tx.Count(&total)
	var list []model.File
	offset := (pageNum - 1) * pageSize
	err := tx.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func DeleteFileByID(id uint) error {
	return dao.DB.Delete(&model.File{}, id).Error
}

func RestoreFileByID(id uint) error {
	return dao.DB.Unscoped().Model(&model.File{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func GetDeletedFileList(pageNum, pageSize int) ([]model.File, int64, error) {
	var list []model.File
	var total int64
	tx := dao.DB.Unscoped().Model(&model.File{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}
