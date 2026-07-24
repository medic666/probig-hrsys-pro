package file

import "gorm.io/gorm"

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) Create(file *File) error {
	return d.db.Create(file).Error
}

func (d *DAO) GetByID(id uint) (*File, error) {
	var file File
	if err := d.db.First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *DAO) GetByIDWithoutContent(id uint) (*File, error) {
	var file File
	if err := d.db.Select("id", "name", "size", "mime_type", "uploader_id", "created_at", "updated_at").First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *DAO) List(page, pageSize int, name, mimeType string) ([]File, int64, error) {
	var files []File
	var total int64
	query := d.db.Model(&File{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if mimeType != "" {
		query = query.Where("mime_type LIKE ?", "%"+mimeType+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.
		Select("id", "name", "size", "mime_type", "uploader_id", "created_at", "updated_at").
		Offset(offset).Limit(pageSize).Order("id DESC").
		Find(&files).Error
	if err != nil {
		return nil, 0, err
	}
	return files, total, nil
}

func (d *DAO) Update(file *File) error {
	return d.db.Save(file).Error
}

func (d *DAO) Delete(id uint) error {
	return d.db.Delete(&File{}, id).Error
}

func (d *DAO) Restore(id uint) error {
	return d.db.Unscoped().Model(&File{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (d *DAO) DeleteFile(id uint) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&File{}, id).Error; err != nil {
			return err
		}
		if err := tx.Where("file_id = ?", id).Delete(&FileRelation{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DAO) RestoreFile(id uint) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&File{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&FileRelation{}).Where("file_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DAO) CreateRelation(relation *FileRelation) error {
	return d.db.Create(relation).Error
}

func (d *DAO) DeleteRelation(id uint) error {
	return d.db.Delete(&FileRelation{}, id).Error
}

func (d *DAO) RestoreRelation(id uint) error {
	return d.db.Unscoped().Model(&FileRelation{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (d *DAO) GetRelationsByFileID(fileID uint) ([]FileRelation, error) {
	var relations []FileRelation
	err := d.db.Where("file_id = ?", fileID).Find(&relations).Error
	return relations, err
}

func (d *DAO) GetRelationsByTarget(targetType string, targetID uint) ([]FileRelation, error) {
	var relations []FileRelation
	err := d.db.Where("target_type = ? AND target_id = ?", targetType, targetID).Find(&relations).Error
	return relations, err
}

func (d *DAO) GetFilesByTarget(targetType string, targetID uint) ([]File, error) {
	var files []File
	err := d.db.Table("files").
		Select("files.id, files.name, files.size, files.mime_type, files.uploader_id, files.created_at, files.updated_at").
		Joins("INNER JOIN file_relations ON file_relations.file_id = files.id AND file_relations.deleted_at IS NULL").
		Where("file_relations.target_type = ? AND file_relations.target_id = ? AND file_relations.deleted_at IS NULL", targetType, targetID).
		Find(&files).Error
	return files, err
}

func (d *DAO) BatchCreateRelations(relations []FileRelation) error {
	if len(relations) == 0 {
		return nil
	}
	return d.db.Create(&relations).Error
}

func (d *DAO) DeleteRelationsByFileID(fileID uint) error {
	return d.db.Where("file_id = ?", fileID).Delete(&FileRelation{}).Error
}

func (d *DAO) DeleteRelationsByTarget(targetType string, targetID uint) error {
	return d.db.Where("target_type = ? AND target_id = ?", targetType, targetID).Delete(&FileRelation{}).Error
}

func (d *DAO) RestoreRelationsByFileID(fileID uint) error {
	return d.db.Unscoped().Model(&FileRelation{}).Where("file_id = ?", fileID).Update("deleted_at", nil).Error
}

func (d *DAO) RestoreRelationsByTarget(targetType string, targetID uint) error {
	return d.db.Unscoped().Model(&FileRelation{}).Where("target_type = ? AND target_id = ?", targetType, targetID).Update("deleted_at", nil).Error
}
