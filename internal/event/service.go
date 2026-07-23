package event

import (
	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordEvent(entityType string, entityID int64, eventType string, payload string, operatorID int64, remark string) error {
	e := &Event{
		EntityType: entityType,
		EntityID:   entityID,
		EventType:  eventType,
		Payload:    payload,
		OperatorID: operatorID,
		Remark:     remark,
	}
	return s.repo.Create(e)
}

func (s *Service) RecordEventTx(tx *sqlx.Tx, entityType string, entityID int64, eventType string, payload string, operatorID int64, remark string) error {
	e := &Event{
		EntityType: entityType,
		EntityID:   entityID,
		EventType:  eventType,
		Payload:    payload,
		OperatorID: operatorID,
		Remark:     remark,
	}
	return s.repo.CreateTx(tx, e)
}

func (s *Service) List(filter EventFilter, page, pageSize int) ([]Event, int64, error) {
	return s.repo.List(filter, page, pageSize)
}

func (s *Service) GetByID(id int64) (*Event, error) {
	return s.repo.GetByID(id)
}

func (s *Service) SoftDelete(id int64) error {
	return s.repo.SoftDelete(id)
}

func (s *Service) UpdateRemark(id int64, remark string) error {
	return s.repo.UpdateRemark(id, remark)
}

func (s *Service) GetByEntity(entityType string, entityID int64) ([]Event, error) {
	return s.repo.GetByEntity(entityType, entityID)
}

func (s *Service) GetSnapshotBefore(entityType string, entityID int64, beforeTime string) (*Event, error) {
	return s.repo.GetSnapshotBefore(entityType, entityID, beforeTime)
}
