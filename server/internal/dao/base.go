package dao

import (
	"gorm.io/gorm"
)

type PaginateQuery struct {
	PageNum  int
	PageSize int
	Filters  map[string]interface{}
	Like     map[string]string
	In       map[string][]interface{}
	Order    string
	Preload  []string
}

func BuildPageQuery(db *gorm.DB, q PaginateQuery) *gorm.DB {
	tx := db
	for col, val := range q.Filters {
		tx = tx.Where(col+" = ?", val)
	}
	for col, val := range q.Like {
		tx = tx.Where(col+" LIKE ?", "%"+val+"%")
	}
	for col, vals := range q.In {
		tx = tx.Where(col+" IN ?", vals)
	}
	if q.Order != "" {
		tx = tx.Order(q.Order)
	}
	for _, p := range q.Preload {
		tx = tx.Preload(p)
	}
	return tx
}

func Paginate[T any](db *gorm.DB, q PaginateQuery) ([]T, int64, error) {
	var list []T
	var total int64

	baseTx := BuildPageQuery(db.Model(new(T)), q)
	if err := baseTx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (q.PageNum - 1) * q.PageSize
	if err := baseTx.Offset(offset).Limit(q.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func FindByID[T any](db *gorm.DB, id uint) (*T, error) {
	var entity T
	if err := db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func CreateEntity[T any](db *gorm.DB, entity *T) error {
	return db.Create(entity).Error
}

func UpdateEntity[T any](db *gorm.DB, id uint, updates map[string]interface{}) error {
	var entity T
	return db.Model(&entity).Where("id = ?", id).Updates(updates).Error
}

func DeleteEntity[T any](db *gorm.DB, id uint) error {
	var entity T
	return db.Delete(&entity, id).Error
}

func RestoreEntity[T any](db *gorm.DB, id uint) error {
	var entity T
	return db.Unscoped().Model(&entity).Where("id = ?", id).Update("deleted_at", nil).Error
}

func FindByIDWithTrashedEntity[T any](db *gorm.DB, id uint) (*T, error) {
	var entity T
	if err := db.Unscoped().First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func FindByField[T any](db *gorm.DB, field string, value interface{}) (*T, error) {
	var entity T
	if err := db.Where(field+" = ?", value).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}
