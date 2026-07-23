package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/models"
)

type OrganizationService struct {
	db     *sqlx.DB
	audit  *AuditService
	engine *SnapshotEngine
}

func NewOrganizationService(db *sqlx.DB, audit *AuditService, engine *SnapshotEngine) *OrganizationService {
	return &OrganizationService{db: db, audit: audit, engine: engine}
}

func (s *OrganizationService) List(page, pageSize int, keyword string) ([]map[string]interface{}, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM entities WHERE type = 'organization'"
	args := []interface{}{}
	if keyword != "" {
		countSQL += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := `SELECT e.id, e.name, e.status, e.created_at, e.updated_at,
	 os.snapshot_data
	 FROM entities e
	 LEFT JOIN org_snapshots os ON os.org_id = e.id AND os.id = (
	   SELECT id FROM org_snapshots WHERE org_id = e.id ORDER BY effective_date DESC, id DESC LIMIT 1
	 )
	 WHERE e.type = 'organization'`
	if keyword != "" {
		querySQL += " AND e.name LIKE ?"
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
		if sd, ok := row["snapshot_data"]; ok && sd != nil {
			var data models.OrgSnapshotData
			if bs, ok2 := sd.([]byte); ok2 {
				json.Unmarshal(bs, &data)
			} else if str, ok2 := sd.(string); ok2 {
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

func (s *OrganizationService) Create(req models.OrgEventRequest, userID uint, ip string) (*models.Entity, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	entityName := req.Data.CompanyName
	result, err := tx.Exec(
		"INSERT INTO entities (type, name, status, created_at, updated_at) VALUES ('organization', ?, 'active', ?, ?)",
		entityName, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create entity: %w", err)
	}
	entityID, _ := result.LastInsertId()

	payloadJSON, _ := json.Marshal(req.Data)
	_, err = tx.Exec(
		`INSERT INTO org_events (org_id, effective_date, event_type, payload, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entityID, req.EffectiveDate, "establish", string(payloadJSON), userID, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, err
	}

	if err := s.engine.RebuildOrgSnapshots(tx, uint(entityID)); err != nil {
		return nil, err
	}

	s.audit.Log(tx, userID, "create", "organization", ptrUint(uint(entityID)), req, ip)

	entity := &models.Entity{
		ID:        uint(entityID),
		Type:      "organization",
		Name:      entityName,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *OrganizationService) GetDetail(orgID uint) (*models.Entity, []models.OrgEvent, []models.OrgSnapshot, error) {
	var entity models.Entity
	if err := s.db.Get(&entity, "SELECT * FROM entities WHERE id = ? AND type = 'organization'", orgID); err != nil {
		return nil, nil, nil, err
	}

	var events []models.OrgEvent
	if err := s.db.Select(&events, "SELECT * FROM org_events WHERE org_id = ? ORDER BY effective_date DESC, id DESC", orgID); err != nil {
		return nil, nil, nil, err
	}
	if events == nil {
		events = []models.OrgEvent{}
	}

	var snapshots []models.OrgSnapshot
	if err := s.db.Select(&snapshots, "SELECT * FROM org_snapshots WHERE org_id = ? ORDER BY effective_date DESC, id DESC", orgID); err != nil {
		return nil, nil, nil, err
	}
	if snapshots == nil {
		snapshots = []models.OrgSnapshot{}
	}

	return &entity, events, snapshots, nil
}

func (s *OrganizationService) CreateEvent(req models.OrgEventRequest, userID uint, ip string) (*models.OrgEvent, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	payloadJSON, _ := json.Marshal(req.Data)
	result, err := tx.Exec(
		`INSERT INTO org_events (org_id, effective_date, event_type, payload, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.OrgID, req.EffectiveDate, req.EventType, string(payloadJSON), userID, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, err
	}
	eventID, _ := result.LastInsertId()

	if err := s.engine.RebuildOrgSnapshots(tx, req.OrgID); err != nil {
		return nil, err
	}

	s.audit.Log(tx, userID, "create", "org_event", ptrUint(uint(eventID)), req, ip)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.OrgEvent{
		ID:            uint(eventID),
		OrgID:         req.OrgID,
		EffectiveDate: req.EffectiveDate,
		EventType:     req.EventType,
		Payload:       string(payloadJSON),
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func (s *OrganizationService) UpdateEvent(eventID uint, req models.OrgEventRequest, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	payloadJSON, _ := json.Marshal(req.Data)
	_, err = tx.Exec(
		`UPDATE org_events SET effective_date = ?, event_type = ?, payload = ?, updated_at = ? WHERE id = ?`,
		req.EffectiveDate, req.EventType, string(payloadJSON), time.Now(), eventID,
	)
	if err != nil {
		return err
	}

	if err := s.engine.RebuildOrgSnapshots(tx, req.OrgID); err != nil {
		return err
	}

	s.audit.Log(tx, userID, "update", "org_event", ptrUint(eventID), req, ip)
	return tx.Commit()
}

func (s *OrganizationService) DeleteEvent(eventID uint, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var evt models.OrgEvent
	if err := tx.Get(&evt, "SELECT * FROM org_events WHERE id = ?", eventID); err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM org_events WHERE id = ?", eventID)
	if err != nil {
		return err
	}

	if err := s.engine.RebuildOrgSnapshots(tx, evt.OrgID); err != nil {
		return err
	}

	s.audit.Log(tx, userID, "delete", "org_event", ptrUint(eventID), evt, ip)
	return tx.Commit()
}
