package company

import (
	"errors"

	"gorm.io/gorm"

	"probig/internal/pkg/audit"
)

type Service struct {
	dao *DAO
}

var globalService *Service

func NewService(db *gorm.DB, encryptKey string) *Service {
	svc := &Service{dao: NewDAO(db, encryptKey)}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) Create(company *Company, operatorID uint, operatorName string) error {
	existing, _ := s.dao.GetByCreditCode(company.CreditCode)
	if existing != nil {
		return errors.New("credit code already exists")
	}
	if err := s.dao.Create(company); err != nil {
		return err
	}
	if audit.GlobalAuditService != nil {
		audit.GlobalAuditService.Log(operatorID, operatorName, "company", company.ID, "create", "", "", "", "")
	}
	return nil
}

func (s *Service) GetByID(id uint) (*Company, error) {
	return s.dao.GetByID(id)
}

func (s *Service) List(page, pageSize int, name, creditCode string) ([]Company, int64, error) {
	return s.dao.List(page, pageSize, name, creditCode)
}

func (s *Service) Update(company *Company, operatorID uint, operatorName string) error {
	existing, _ := s.dao.GetByCreditCode(company.CreditCode)
	if existing != nil && existing.ID != company.ID {
		return errors.New("credit code already exists")
	}
	if err := s.dao.Update(company); err != nil {
		return err
	}
	if audit.GlobalAuditService != nil {
		audit.GlobalAuditService.Log(operatorID, operatorName, "company", company.ID, "update", "", "", "", "")
	}
	return nil
}

func (s *Service) Delete(id uint, operatorID uint, operatorName string) error {
	if err := s.dao.Delete(id); err != nil {
		return err
	}
	if audit.GlobalAuditService != nil {
		audit.GlobalAuditService.Log(operatorID, operatorName, "company", id, "delete", "", "", "", "")
	}
	return nil
}

func (s *Service) Restore(id uint, operatorID uint, operatorName string) error {
	if err := s.dao.Restore(id); err != nil {
		return err
	}
	if audit.GlobalAuditService != nil {
		audit.GlobalAuditService.Log(operatorID, operatorName, "company", id, "restore", "", "", "", "")
	}
	return nil
}
