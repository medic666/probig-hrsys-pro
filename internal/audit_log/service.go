package audit_log

import (
	"gorm.io/gorm"
)

type Service struct {
	dao *DAO
}

var globalService *Service

func NewService(db *gorm.DB) *Service {
	svc := &Service{dao: NewDAO(db)}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) List(params AuditListParams) ([]AuditLog, int64, error) {
	return s.dao.List(params)
}
