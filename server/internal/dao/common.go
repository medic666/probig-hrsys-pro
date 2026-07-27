package dao

import (
	"gorm.io/gorm"
)

func SoftDeleteByID(db *gorm.DB, model interface{}, id uint) error {
	return db.Delete(model, id).Error
}

func SoftDelete(db *gorm.DB, model interface{}) error {
	return db.Delete(model).Error
}

func RestoreByID(db *gorm.DB, model interface{}, id uint) error {
	return db.Unscoped().Model(model).Where("id = ?", id).Update("deleted_at", nil).Error
}

func FindByIDWithTrashed(db *gorm.DB, model interface{}, id uint) error {
	return db.Unscoped().First(model, id).Error
}

func IsDeleted(db *gorm.DB, model interface{}, id uint) bool {
	var count int64
	db.Unscoped().Model(model).Where("id = ? AND deleted_at IS NOT NULL", id).Count(&count)
	return count > 0
}
