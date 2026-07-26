package company

import (
	"probig/internal/pkg/database"

	"gorm.io/gorm"
)

func CreateCompany(tx *gorm.DB, company *database.Company) error {
	return tx.Create(company).Error
}

func UpdateCompany(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&database.Company{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteCompany(tx *gorm.DB, id uint) error {
	return tx.Delete(&database.Company{}, id).Error
}

func RestoreCompany(tx *gorm.DB, id uint) error {
	return tx.Unscoped().Model(&database.Company{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func GetCompanyByID(id uint) (*database.Company, error) {
	var company database.Company
	err := database.DB.Where("id = ?", id).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func CreditCodeExists(creditCode string, excludeID uint) (bool, error) {
	var count int64
	query := database.DB.Model(&database.Company{}).Where("credit_code = ?", creditCode)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

type ListCompaniesFilter struct {
	Name       string
	CreditCode string
}

func ListCompanies(pageNum, pageSize int, filter ListCompaniesFilter) ([]database.Company, int64, error) {
	var companies []database.Company
	var total int64

	query := database.DB.Model(&database.Company{})

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.CreditCode != "" {
		query = query.Where("credit_code LIKE ?", "%"+filter.CreditCode+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&companies).Error; err != nil {
		return nil, 0, err
	}

	return companies, total, nil
}

func ListTrashCompanies(pageNum, pageSize int) ([]database.Company, int64, error) {
	var companies []database.Company
	var total int64

	query := database.DB.Unscoped().Where("deleted_at IS NOT NULL")

	if err := query.Model(&database.Company{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&companies).Error; err != nil {
		return nil, 0, err
	}

	return companies, total, nil
}

func CascadeDeleteCompany(tx *gorm.DB, companyID uint) error {
	if err := tx.Table("file_relation").Where("target_type = ? AND target_id = ?", "company", companyID).Update("deleted_at", gorm.Expr("datetime('now')")).Error; err != nil {
		return err
	}

	return tx.Delete(&database.Company{}, companyID).Error
}

func CascadeRestoreCompany(tx *gorm.DB, companyID uint) error {
	if err := tx.Table("file_relation").Where("target_type = ? AND target_id = ?", "company", companyID).Update("deleted_at", nil).Error; err != nil {
		return err
	}

	return tx.Unscoped().Model(&database.Company{}).Where("id = ?", companyID).Update("deleted_at", nil).Error
}
