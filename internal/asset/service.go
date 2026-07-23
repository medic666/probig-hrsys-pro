package asset

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

func (s *Service) Create(asset *Asset, operatorID int64, remark string) error {
	tx, err := common.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.CreateTx(tx, asset); err != nil {
		return err
	}

	payload, _ := json.Marshal(asset)
	if err := s.eventService.RecordEventTx(tx, "asset", asset.ID, "create", string(payload), operatorID, remark); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) Update(asset *Asset, operatorID int64, remark string) error {
	existing, err := s.repo.GetByID(asset.ID)
	if err != nil {
		return common.ErrNotFound
	}

	asset.Version = existing.Version

	if err := s.repo.AddVersion(asset); err != nil {
		return err
	}

	payload, _ := json.Marshal(asset)
	return s.eventService.RecordEvent("asset", asset.ID, "update", string(payload), operatorID, remark)
}

func (s *Service) Delete(id int64, operatorID int64, remark string) error {
	asset, err := s.repo.GetByID(id)
	if err != nil {
		return common.ErrNotFound
	}

	if err := s.repo.SoftDelete(id); err != nil {
		return err
	}

	payload, _ := json.Marshal(asset)
	return s.eventService.RecordEvent("asset", id, "delete", string(payload), operatorID, remark)
}

func (s *Service) GetByID(id int64) (*Asset, error) {
	return s.repo.GetByID(id)
}

func (s *Service) List(assetType, search string, page, pageSize int) ([]Asset, int64, error) {
	return s.repo.List(assetType, search, page, pageSize)
}

func (s *Service) GetVersions(assetID int64) ([]Asset, error) {
	return s.repo.GetVersions(assetID)
}

func (s *Service) GetTimeline(id int64) ([]event.Event, error) {
	return s.eventService.GetByEntity("asset", id)
}
