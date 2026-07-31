package service

import (
	"context"
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func GetCompanyList(pageNum, pageSize int, name, creditCode string, id string) ([]model.Company, int64, error) {
	tx := dao.DB.Model(&model.Company{})
	if id != "" {
		tx = tx.Where("id = ?", id)
	}
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

func CreateCompany(ctx context.Context, c *model.Company) error {
	if c.CreditCode != "" {
		var count int64
		dao.DB.Model(&model.Company{}).Where("credit_code = ?", c.CreditCode).Count(&count)
		if count > 0 {
			return errors.New("统一社会信用代码已存在")
		}
	}
	return dao.DBFromContext(ctx).Create(c).Error
}

func UpdateCompany(ctx context.Context, id uint, c *model.Company) error {
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
	return dao.DBFromContext(ctx).Model(&existing).Updates(updates).Error
}

func DeleteCompany(ctx context.Context, id uint) error {
	var c model.Company
	if err := dao.DB.First(&c, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Delete(&c).Error; err != nil {
			return err
		}
		tx.Where("target_type = ? AND target_id = ?", "company", id).Delete(&model.FileRelation{})
		return nil
	})
}

func RestoreCompany(ctx context.Context, id uint) error {
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
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&c).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Model(&model.FileRelation{}).Where("target_type = ? AND target_id = ?", "company", id).Update("deleted_at", nil)
		return nil
	})
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
