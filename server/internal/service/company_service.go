package service

import (
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
)

func GetCompanyList(pageNum, pageSize int, name, creditCode string) ([]model.Company, int64, error) {
	tx := dao.DB.Model(&model.Company{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if creditCode != "" {
		tx = tx.Where("credit_code LIKE ?", "%"+creditCode+"%")
	}
	var total int64
	tx.Count(&total)
	var list []model.Company
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list)
	return list, total, nil
}

func GetCompanyByID(id uint) (*model.Company, error) {
	var c model.Company
	if err := dao.DB.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateCompany(c *model.Company) error {
	if c.CreditCode != "" {
		var count int64
		dao.DB.Model(&model.Company{}).Where("credit_code = ?", c.CreditCode).Count(&count)
		if count > 0 {
			return errors.New("统一社会信用代码已存在")
		}
	}
	return dao.DB.Create(c).Error
}

func UpdateCompany(id uint, c *model.Company) error {
	var existing model.Company
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("公司不存在")
	}
	if c.CreditCode != "" && c.CreditCode != existing.CreditCode {
		var count int64
		dao.DB.Model(&model.Company{}).Where("credit_code = ? AND id != ?", c.CreditCode, id).Count(&count)
		if count > 0 {
			return errors.New("统一社会信用代码已存在")
		}
	}
	updates := map[string]interface{}{
		"name":          c.Name,
		"credit_code":   c.CreditCode,
		"address":       c.Address,
		"contact_phone": c.ContactPhone,
		"bank_name":     c.BankName,
		"bank_account":  c.BankAccount,
	}
	return dao.DB.Model(&existing).Updates(updates).Error
}

func DeleteCompany(id uint) error {
	var c model.Company
	if err := dao.DB.First(&c, id).Error; err != nil {
		return err
	}
	return dao.DB.Delete(&c).Error
}

func RestoreCompany(id uint) error {
	var c model.Company
	if err := dao.DB.Unscoped().First(&c, id).Error; err != nil {
		return err
	}
	if c.CreditCode != "" {
		var count int64
		dao.DB.Model(&model.Company{}).Where("credit_code = ? AND id != ?", c.CreditCode, id).Count(&count)
		if count > 0 {
			return errors.New("统一社会信用代码已被占用，无法恢复")
		}
	}
	return dao.DB.Unscoped().Model(&c).Update("deleted_at", nil).Error
}

func GetDeletedCompanies(pageNum, pageSize int) ([]model.Company, int64, error) {
	var list []model.Company
	var total int64
	tx := dao.DB.Unscoped().Model(&model.Company{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}

func GetAllCompanies() ([]model.Company, error) {
	var list []model.Company
	if err := dao.DB.Order("name").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
