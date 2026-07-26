package company

import (
	"errors"

	"probig/internal/pkg/database"

	"gorm.io/gorm"
)

type Service struct{}

var DefaultService = &Service{}

func (s *Service) CreateCompany(req *CreateCompanyRequest) (*database.Company, error) {
	if req.Name == "" {
		return nil, errors.New("公司名称不能为空")
	}

	if req.CreditCode != "" {
		exists, err := CreditCodeExists(req.CreditCode, 0)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("统一社会信用代码已存在")
		}
	}

	company := &database.Company{
		Name:         req.Name,
		CreditCode:   req.CreditCode,
		Address:      req.Address,
		ContactPhone: req.ContactPhone,
		BankName:     req.BankName,
		BankAccount:  req.BankAccount,
	}

	if err := CreateCompany(nil, company); err != nil {
		return nil, err
	}
	return company, nil
}

func (s *Service) UpdateCompany(id uint, req *UpdateCompanyRequest) error {
	company, err := GetCompanyByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.CreditCode != "" && req.CreditCode != company.CreditCode {
		exists, err := CreditCodeExists(req.CreditCode, id)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("统一社会信用代码已存在")
		}
		updates["credit_code"] = req.CreditCode
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.ContactPhone != "" {
		updates["contact_phone"] = req.ContactPhone
	}
	if req.BankName != "" {
		updates["bank_name"] = req.BankName
	}
	if req.BankAccount != "" {
		updates["bank_account"] = req.BankAccount
	}

	if len(updates) == 0 {
		return nil
	}

	return UpdateCompany(nil, id, updates)
}

func (s *Service) DeleteCompany(id uint) error {
	_, err := GetCompanyByID(id)
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		return CascadeDeleteCompany(tx, id)
	})
}

func (s *Service) RestoreCompany(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return CascadeRestoreCompany(tx, id)
	})
}

func (s *Service) GetCompany(id uint) (*database.Company, error) {
	return GetCompanyByID(id)
}

func (s *Service) ListCompanies(pageNum, pageSize int, name, creditCode string) ([]database.Company, int64, error) {
	return ListCompanies(pageNum, pageSize, ListCompaniesFilter{
		Name:       name,
		CreditCode: creditCode,
	})
}

func (s *Service) ListTrashCompanies(pageNum, pageSize int) ([]database.Company, int64, error) {
	return ListTrashCompanies(pageNum, pageSize)
}
