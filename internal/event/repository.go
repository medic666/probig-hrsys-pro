package event

import (
	"fmt"
	"strings"

	"probig/internal/common"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository() *Repository {
	return &Repository{db: common.DB}
}

func (r *Repository) Create(event *Event) error {
	_, err := r.db.Exec(`INSERT INTO events (entity_type, entity_id, event_type, payload, operator_id, remark)
		VALUES (?, ?, ?, ?, ?, ?)`,
		event.EntityType, event.EntityID, event.EventType, event.Payload, event.OperatorID, event.Remark)
	return err
}

func (r *Repository) CreateTx(tx *sqlx.Tx, event *Event) error {
	_, err := tx.Exec(`INSERT INTO events (entity_type, entity_id, event_type, payload, operator_id, remark)
		VALUES (?, ?, ?, ?, ?, ?)`,
		event.EntityType, event.EntityID, event.EventType, event.Payload, event.OperatorID, event.Remark)
	return err
}

func (r *Repository) List(filter EventFilter, page, pageSize int) ([]Event, int64, error) {
	where := []string{"is_deleted = 0"}
	args := []interface{}{}

	if filter.EntityType != "" {
		where = append(where, "entity_type = ?")
		args = append(args, filter.EntityType)
	}
	if filter.EntityID > 0 {
		where = append(where, "entity_id = ?")
		args = append(args, filter.EntityID)
	}
	if filter.EventType != "" {
		where = append(where, "event_type = ?")
		args = append(args, filter.EventType)
	}
	if filter.OperatorID > 0 {
		where = append(where, "operator_id = ?")
		args = append(args, filter.OperatorID)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events WHERE %s", whereClause)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM events WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereClause)
	queryArgs := append(args, pageSize, offset)

	var events []Event
	err := r.db.Select(&events, query, queryArgs...)
	return events, total, err
}

func (r *Repository) GetByID(id int64) (*Event, error) {
	var e Event
	err := r.db.Get(&e, "SELECT * FROM events WHERE id = ? AND is_deleted = 0", id)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) SoftDelete(id int64) error {
	_, err := r.db.Exec("UPDATE events SET is_deleted = 1 WHERE id = ?", id)
	return err
}

func (r *Repository) UpdateRemark(id int64, remark string) error {
	_, err := r.db.Exec("UPDATE events SET remark = ? WHERE id = ?", remark, id)
	return err
}

func (r *Repository) GetByEntity(entityType string, entityID int64) ([]Event, error) {
	var events []Event
	err := r.db.Select(&events,
		"SELECT * FROM events WHERE entity_type = ? AND entity_id = ? AND is_deleted = 0 ORDER BY created_at DESC",
		entityType, entityID)
	return events, err
}

func (r *Repository) GetSnapshotBefore(entityType string, entityID int64, beforeTime string) (*Event, error) {
	var e Event
	err := r.db.Get(&e,
		`SELECT * FROM events WHERE entity_type = ? AND entity_id = ? AND created_at <= ? AND is_deleted = 0
		 ORDER BY created_at DESC LIMIT 1`,
		entityType, entityID, beforeTime)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
