package company

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"probig/internal/pkg/encrypt"
)

type DAO struct {
	db         *gorm.DB
	encryptKey string
}

func NewDAO(db *gorm.DB, encryptKey string) *DAO {
	return &DAO{db: db, encryptKey: encryptKey}
}

func (d *DAO) encryptFields(company *Company) error {
	if company.CreditCode != "" {
		encrypted, err := encrypt.Encrypt(company.CreditCode, d.encryptKey)
		if err != nil {
			return err
		}
		company.CreditCode = encrypted
	}
	if company.BankAccount != "" {
		encrypted, err := encrypt.Encrypt(company.BankAccount, d.encryptKey)
		if err != nil {
			return err
		}
		company.BankAccount = encrypted
	}
	return nil
}

func (d *DAO) decryptFields(company *Company) error {
	if company.CreditCode != "" {
		decrypted, err := encrypt.Decrypt(company.CreditCode, d.encryptKey)
		if err != nil {
			return err
		}
		company.CreditCode = decrypted
	}
	if company.BankAccount != "" {
		decrypted, err := encrypt.Decrypt(company.BankAccount, d.encryptKey)
		if err != nil {
			return err
		}
		company.BankAccount = decrypted
	}
	return nil
}

func (d *DAO) Create(company *Company) error {
	if err := d.encryptFields(company); err != nil {
		return err
	}
	if err := d.db.Create(company).Error; err != nil {
		return err
	}
	return d.decryptFields(company)
}

func (d *DAO) GetByID(id uint) (*Company, error) {
	var company Company
	err := d.db.First(&company, id).Error
	if err != nil {
		return nil, err
	}
	if err := d.decryptFields(&company); err != nil {
		return nil, err
	}
	return &company, nil
}

func (d *DAO) List(page, pageSize int, name, creditCode string) ([]Company, int64, error) {
	var companies []Company
	var total int64
	query := d.db.Model(&Company{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&companies).Error
	if err != nil {
		return nil, 0, err
	}

	decrypted := make([]Company, 0, len(companies))
	for i := range companies {
		if err := d.decryptFields(&companies[i]); err != nil {
			return nil, 0, err
		}
		if creditCode != "" && !strings.Contains(companies[i].CreditCode, creditCode) {
			continue
		}
		decrypted = append(decrypted, companies[i])
	}

	return decrypted, total, nil
}

func (d *DAO) Update(company *Company) error {
	if err := d.encryptFields(company); err != nil {
		return err
	}
	return d.db.Model(company).Select("name", "credit_code", "address", "contact_phone", "bank_name", "bank_account").Updates(company).Error
}

func (d *DAO) Delete(id uint) error {
	return d.db.Delete(&Company{}, id).Error
}

func (d *DAO) Restore(id uint) error {
	return d.db.Unscoped().Model(&Company{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (d *DAO) GetByCreditCode(code string) (*Company, error) {
	var companies []Company
	if err := d.db.Find(&companies).Error; err != nil {
		return nil, err
	}
	for i := range companies {
		if err := d.decryptFields(&companies[i]); err != nil {
			return nil, err
		}
		if companies[i].CreditCode == code {
			return &companies[i], nil
		}
	}
	return nil, errors.New("company not found")
}
