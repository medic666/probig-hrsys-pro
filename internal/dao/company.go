package dao

import (
	"probig/internal/models"
)

func GetCompanyList(page, pageSize int, keyword string) ([]models.Company, int64, error) {
	var list []models.Company
	var total int64
	q := DB().Model(&models.Company{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR credit_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	for i := range list {
		list[i].BankAccountPlain = decrypt(list[i].BankAccount)
	}
	return list, total, nil
}

func GetCompanyByID(id uint) (*models.Company, error) {
	var c models.Company
	if err := DB().First(&c, id).Error; err != nil {
		return nil, err
	}
	c.BankAccountPlain = decrypt(c.BankAccount)
	return &c, nil
}

func GetCompanyByCreditCode(code string) (*models.Company, error) {
	var c models.Company
	if err := DB().Where("credit_code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateCompany(c *models.Company) error {
	c.BankAccount = encrypt(c.BankAccountPlain)
	return DB().Create(c).Error
}

func UpdateCompany(c *models.Company) error {
	c.BankAccount = encrypt(c.BankAccountPlain)
	tx := DB().Begin()
	if err := tx.Save(c).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func SoftDeleteCompany(id uint) error {
	tx := DB().Begin()
	if err := tx.Where("target_type = ? AND target_id = ?", "company", id).Delete(&models.FileRelation{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&models.Company{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func RestoreCompany(id uint) error {
	tx := DB().Begin()
	if err := tx.Unscoped().Model(&models.Company{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.FileRelation{}).Where("target_type = ? AND target_id = ?", "company", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
