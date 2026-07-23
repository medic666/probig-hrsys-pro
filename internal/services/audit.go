package services

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/models"
)

type AuditService struct {
	db *sqlx.DB
}

func NewAuditService(db *sqlx.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Log(tx *sqlx.Tx, userID uint, action, entityType string, entityID *uint, payload interface{}, ipAddress string) error {
	payloadJSON := "{}"
	if payload != nil {
		b, err := json.Marshal(payload)
		if err == nil {
			payloadJSON = string(b)
		}
	}
	_, err := tx.Exec(
		`INSERT INTO audit_logs (user_id, action, entity_type, entity_id, payload, ip_address, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, action, entityType, entityID, payloadJSON, ipAddress, time.Now(),
	)
	if err != nil {
		auditFallback(s.db, userID, action, entityType, entityID, payloadJSON, ipAddress)
	}
	return nil
}

func auditFallback(db *sqlx.DB, userID uint, action, entityType string, entityID *uint, payload, ipAddress string) {
	db.Exec(
		`INSERT INTO audit_logs (user_id, action, entity_type, entity_id, payload, ip_address, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, action, entityType, entityID, payload, ipAddress, time.Now(),
	)
}

func (s *AuditService) List(page, pageSize int, userID uint, entityType, action string) ([]models.AuditLog, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM audit_logs al WHERE 1=1"
	args := []interface{}{}

	if userID > 0 {
		countSQL += " AND al.user_id = ?"
		args = append(args, userID)
	}
	if entityType != "" {
		countSQL += " AND al.entity_type = ?"
		args = append(args, entityType)
	}
	if action != "" {
		countSQL += " AND al.action = ?"
		args = append(args, action)
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := "SELECT al.*, u.username FROM audit_logs al LEFT JOIN users u ON u.id = al.user_id WHERE 1=1"
	if userID > 0 {
		querySQL += " AND al.user_id = ?"
	}
	if entityType != "" {
		querySQL += " AND al.entity_type = ?"
	}
	if action != "" {
		querySQL += " AND al.action = ?"
	}
	querySQL += " ORDER BY al.created_at DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	var logs []models.AuditLog
	if err := s.db.Select(&logs, querySQL, queryArgs...); err != nil {
		return nil, 0, err
	}
	if logs == nil {
		logs = []models.AuditLog{}
	}
	return logs, total, nil
}
