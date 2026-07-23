package person

import (
	"encoding/json"
	"probig/internal/common"
	"probig/internal/event"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo        *Repository
	eventService *event.Service
}

func NewService(repo *Repository, eventService *event.Service) *Service {
	return &Service{repo: repo, eventService: eventService}
}

func (s *Service) Create(person *Person, operatorID int64, remark string) error {
	tx, err := common.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.CreateTx(tx, person); err != nil {
		return err
	}

	payload, _ := json.Marshal(person)
	if err := s.eventService.RecordEventTx(tx, "person", person.ID, "create", string(payload), operatorID, remark); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) Update(person *Person, operatorID int64, remark string) error {
	existing, err := s.repo.GetByID(person.ID)
	if err != nil {
		return common.ErrNotFound
	}

	person.ID = existing.ID

	tx, err := common.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.UpdateTx(tx, person); err != nil {
		return err
	}

	payload, _ := json.Marshal(person)
	if err := s.eventService.RecordEventTx(tx, "person", person.ID, "update", string(payload), operatorID, remark); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) Delete(id int64, operatorID int64, remark string) error {
	person, err := s.repo.GetByID(id)
	if err != nil {
		return common.ErrNotFound
	}

	tx, err := common.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.DeleteTx(tx, id); err != nil {
		return err
	}

	payload, _ := json.Marshal(person)
	if err := s.eventService.RecordEventTx(tx, "person", id, "delete", string(payload), operatorID, remark); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) GetByID(id int64) (*Person, error) {
	return s.repo.GetByID(id)
}

func (s *Service) List(search string, page, pageSize int) ([]Person, int64, error) {
	return s.repo.List(search, page, pageSize)
}

func (s *Service) GetSnapshot(id int64, at string) (*Person, error) {
	evt, err := s.eventService.GetSnapshotBefore("person", id, at)
	if err != nil {
		return nil, common.ErrNotFound
	}

	var p Person
	if err := json.Unmarshal([]byte(evt.Payload), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) GetTimeline(id int64) ([]event.Event, error) {
	return s.eventService.GetByEntity("person", id)
}

func (s *Service) All() ([]Person, error) {
	return s.repo.All()
}

func (s *Service) UpdatePartial(id int64, updates map[string]interface{}, operatorID int64, remark string) error {
	person, err := s.repo.GetByID(id)
	if err != nil {
		return common.ErrNotFound
	}

	tx, err := common.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.applyUpdateTx(tx, id, updates); err != nil {
		return err
	}

	payload, _ := json.Marshal(person)
	if err := s.eventService.RecordEventTx(tx, "person", id, "update", string(payload), operatorID, remark); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) applyUpdateTx(tx *sqlx.Tx, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	var sets []string
	var args []interface{}

	for k, v := range updates {
		sets = append(sets, k+" = ?")
		args = append(args, v)
	}

	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")

	query := "UPDATE persons SET "
	for i, s := range sets {
		if i > 0 {
			query += ", "
		}
		query += s
	}
	query += " WHERE id = ?"
	args = append(args, id)

	_, err := tx.Exec(query, args...)
	return err
}
