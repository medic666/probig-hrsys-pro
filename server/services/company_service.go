package services

import (
	"errors"

	"probig/database"
	"probig/models"

	"gorm.io/gorm"
)

func ListCompanies(query string, offset, limit int) ([]models.Company, int64, error) {
	var companies []models.Company
	var total int64
	db := database.DB.Model(&models.Company{})
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("name LIKE ? OR credit_code LIKE ?", like, like)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&companies).Error
	return companies, total, err
}

func GetCompany(id uint) (*models.Company, error) {
	var company models.Company
	err := database.DB.First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func CreateCompany(company *models.Company) error {
	if company.CreditCode != "" {
		var existing models.Company
		if err := database.DB.Where("credit_code = ?", company.CreditCode).First(&existing).Error; err == nil {
			return errors.New("统一社会信用代码已存在")
		}
	}
	return database.DB.Create(company).Error
}

func UpdateCompany(id uint, updates map[string]interface{}) error {
	if code, ok := updates["credit_code"].(string); ok && code != "" {
		var existing models.Company
		if err := database.DB.Where("credit_code = ? AND id != ?", code, id).First(&existing).Error; err == nil {
			return errors.New("统一社会信用代码已存在")
		}
	}
	return database.DB.Model(&models.Company{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteCompany(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Company{}, id).Error; err != nil {
			return err
		}
		tx.Where("target_type = 'company' AND target_id = ?", id).Delete(&models.FileRelation{})
		return nil
	})
}

func RestoreCompany(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.Company{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Where("target_type = 'company' AND target_id = ?", id).Model(&models.FileRelation{}).Update("deleted_at", nil)
		return nil
	})
}

func ListDeletedCompanies(query string, offset, limit int) ([]models.Company, int64, error) {
	var companies []models.Company
	var total int64
	db := database.DB.Unscoped().Model(&models.Company{}).Where("deleted_at IS NOT NULL")
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("name LIKE ? OR credit_code LIKE ?", like, like)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("deleted_at DESC").Find(&companies).Error
	return companies, total, err
}
