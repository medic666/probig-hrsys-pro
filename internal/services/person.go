package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/models"
)

type PersonService struct {
	db      *sqlx.DB
	audit   *AuditService
	engine  *SnapshotEngine
}

func NewPersonService(db *sqlx.DB, audit *AuditService, engine *SnapshotEngine) *PersonService {
	return &PersonService{db: db, audit: audit, engine: engine}
}

func (s *PersonService) List(page, pageSize int, keyword, status string) ([]map[string]interface{}, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM entities WHERE type = 'person'"
	args := []interface{}{}

	if keyword != "" {
		countSQL += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	if status != "" {
		countSQL += " AND status = ?"
		args = append(args, status)
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := fmt.Sprintf(
		`SELECT e.id, e.name, e.status, e.created_at, e.updated_at,
		 ps.snapshot_data
		 FROM entities e
		 LEFT JOIN person_snapshots ps ON ps.person_id = e.id AND ps.id = (
		   SELECT id FROM person_snapshots WHERE person_id = e.id ORDER BY effective_date DESC, id DESC LIMIT 1
		 )
		 WHERE e.type = 'person'`,
	)
	if keyword != "" {
		querySQL += " AND e.name LIKE ?"
	}
	if status != "" {
		querySQL += " AND e.status = ?"
	}
	querySQL += " ORDER BY e.id DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	rows, err := s.db.Queryx(querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return nil, 0, err
		}

		if snapshotData, ok := row["snapshot_data"]; ok && snapshotData != nil {
			var data models.PersonSnapshotData
			if bs, ok2 := snapshotData.([]byte); ok2 {
				json.Unmarshal(bs, &data)
			} else if str, ok2 := snapshotData.(string); ok2 {
				json.Unmarshal([]byte(str), &data)
			}
			row["info"] = data
		}
		delete(row, "snapshot_data")
		result = append(result, row)
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, total, nil
}

func (s *PersonService) Create(req models.PersonEventRequest, userID uint, ip string) (*models.Entity, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	entityName := req.Data.Name
	if entityName == "" {
		entityName = req.Data.Alias
	}

	result, err := tx.Exec(
		"INSERT INTO entities (type, name, status, created_at, updated_at) VALUES ('person', ?, 'active', ?, ?)",
		entityName, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}
	entityID, _ := result.LastInsertId()

	payloadJSON, _ := json.Marshal(req.Data)
	evtResult, err := tx.Exec(
		`INSERT INTO person_events (person_id, effective_date, event_type, payload, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entityID, req.EffectiveDate, "onboard", string(payloadJSON), userID, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create person event: %w", err)
	}
	eventID, _ := evtResult.LastInsertId()

	if err := s.engine.RebuildPersonSnapshots(tx, uint(entityID)); err != nil {
		return nil, err
	}

	s.audit.Log(tx, userID, "create", "person", ptrUint(uint(entityID)), req, ip)

	entity := &models.Entity{
		ID:        uint(entityID),
		Type:      "person",
		Name:      entityName,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_ = eventID

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *PersonService) GetDetail(personID uint) (*models.Entity, []models.PersonEvent, []models.PersonSnapshot, error) {
	var entity models.Entity
	if err := s.db.Get(&entity, "SELECT * FROM entities WHERE id = ? AND type = 'person'", personID); err != nil {
		return nil, nil, nil, err
	}

	var events []models.PersonEvent
	if err := s.db.Select(&events, "SELECT * FROM person_events WHERE person_id = ? ORDER BY effective_date DESC, id DESC", personID); err != nil {
		return nil, nil, nil, err
	}
	if events == nil {
		events = []models.PersonEvent{}
	}

	var snapshots []models.PersonSnapshot
	if err := s.db.Select(&snapshots, "SELECT * FROM person_snapshots WHERE person_id = ? ORDER BY effective_date DESC, id DESC", personID); err != nil {
		return nil, nil, nil, err
	}
	if snapshots == nil {
		snapshots = []models.PersonSnapshot{}
	}

	return &entity, events, snapshots, nil
}

func (s *PersonService) UpdateStatus(personID uint, status string, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE entities SET status = ?, updated_at = ? WHERE id = ? AND type = 'person'",
		status, time.Now(), personID,
	)
	if err != nil {
		return err
	}
	s.audit.Log(tx, userID, "update", "person", ptrUint(personID), map[string]string{"status": status}, ip)
	return tx.Commit()
}

func (s *PersonService) CreateEvent(req models.PersonEventRequest, userID uint, ip string) (*models.PersonEvent, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	payloadJSON, _ := json.Marshal(req.Data)
	result, err := tx.Exec(
		`INSERT INTO person_events (person_id, effective_date, event_type, payload, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.PersonID, req.EffectiveDate, req.EventType, string(payloadJSON), userID, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create person event: %w", err)
	}
	eventID, _ := result.LastInsertId()

	if err := s.engine.RebuildPersonSnapshots(tx, req.PersonID); err != nil {
		return nil, err
	}

	s.audit.Log(tx, userID, "create", "person_event", ptrUint(uint(eventID)), req, ip)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	event := &models.PersonEvent{
		ID:            uint(eventID),
		PersonID:      req.PersonID,
		EffectiveDate: req.EffectiveDate,
		EventType:     req.EventType,
		Payload:       string(payloadJSON),
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	return event, nil
}

func (s *PersonService) UpdateEvent(eventID uint, req models.PersonEventRequest, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	payloadJSON, _ := json.Marshal(req.Data)
	_, err = tx.Exec(
		`UPDATE person_events SET effective_date = ?, event_type = ?, payload = ?, updated_at = ? WHERE id = ?`,
		req.EffectiveDate, req.EventType, string(payloadJSON), time.Now(), eventID,
	)
	if err != nil {
		return err
	}

	if err := s.engine.RebuildPersonSnapshots(tx, req.PersonID); err != nil {
		return err
	}

	s.audit.Log(tx, userID, "update", "person_event", ptrUint(eventID), req, ip)
	return tx.Commit()
}

func (s *PersonService) DeleteEvent(eventID uint, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var evt models.PersonEvent
	if err := tx.Get(&evt, "SELECT * FROM person_events WHERE id = ?", eventID); err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM person_events WHERE id = ?", eventID)
	if err != nil {
		return err
	}

	if err := s.engine.RebuildPersonSnapshots(tx, evt.PersonID); err != nil {
		return err
	}

	s.audit.Log(tx, userID, "delete", "person_event", ptrUint(eventID), evt, ip)
	return tx.Commit()
}

func ptrUint(v uint) *uint { return &v }
