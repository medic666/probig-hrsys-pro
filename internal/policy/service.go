package policy

import (
	"encoding/json"

	"probig/internal/common"
	"probig/internal/event"
)

type Service struct {
	repo         *Repository
	eventService *event.Service
}

func NewService(repo *Repository, eventService *event.Service) *Service {
	return &Service{repo: repo, eventService: eventService}
}

func (s *Service) Create(policy *Policy, operatorID int64, remark string) error {
	tx, err := common.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.CreateTx(tx, policy); err != nil {
		return err
	}

	payload, _ := json.Marshal(policy)
	if err := s.eventService.RecordEventTx(tx, "policy", policy.ID, "create", string(payload), operatorID, remark); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) Update(policy *Policy, operatorID int64, remark string) error {
	existing, err := s.repo.GetByID(policy.ID)
	if err != nil {
		return common.ErrNotFound
	}

	policy.Version = existing.Version
	policy.ParentID = 0

	if err := s.repo.AddVersion(policy); err != nil {
		return err
	}

	payload, _ := json.Marshal(policy)
	return s.eventService.RecordEvent("policy", policy.ID, "update", string(payload), operatorID, remark)
}

func (s *Service) Delete(id int64, operatorID int64, remark string) error {
	policy, err := s.repo.GetByID(id)
	if err != nil {
		return common.ErrNotFound
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	payload, _ := json.Marshal(policy)
	return s.eventService.RecordEvent("policy", id, "delete", string(payload), operatorID, remark)
}

func (s *Service) GetByID(id int64) (*Policy, error) {
	return s.repo.GetByID(id)
}

func (s *Service) List(policyType, search string, page, pageSize int) ([]Policy, int64, error) {
	return s.repo.List(policyType, search, page, pageSize)
}

func (s *Service) GetVersions(policyID int64) ([]Policy, error) {
	return s.repo.GetVersions(policyID)
}

func (s *Service) GetTimeline(id int64) ([]event.Event, error) {
	return s.eventService.GetByEntity("policy", id)
}
