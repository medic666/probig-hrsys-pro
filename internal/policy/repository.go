package policy

import (
	"fmt"

	"probig/internal/common"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository() *Repository {
	return &Repository{db: common.DB}
}

func (r *Repository) Create(policy *Policy) error {
	result, err := r.db.Exec(`INSERT INTO policies (policy_type, title, content, status, version, parent_id, is_current)
		VALUES (?, ?, ?, ?, ?, ?, 1)`,
		policy.PolicyType, policy.Title, policy.Content, policy.Status, 1, policy.ParentID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	policy.ID = id
	policy.Version = 1
	return nil
}

func (r *Repository) CreateTx(tx *sqlx.Tx, policy *Policy) error {
	result, err := tx.Exec(`INSERT INTO policies (policy_type, title, content, status, version, parent_id, is_current)
		VALUES (?, ?, ?, ?, ?, ?, 1)`,
		policy.PolicyType, policy.Title, policy.Content, policy.Status, policy.Version, policy.ParentID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	policy.ID = id
	policy.Version = 1
	return nil
}

func (r *Repository) AddVersion(policy *Policy) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE policies SET is_current = 0 WHERE id = ?", policy.ID); err != nil {
		return err
	}

	result, err := tx.Exec(`INSERT INTO policies (policy_type, title, content, status, version, parent_id, is_current)
		VALUES (?, ?, ?, ?, ?, ?, 1)`,
		policy.PolicyType, policy.Title, policy.Content, policy.Status, policy.Version+1, policy.ID)
	if err != nil {
		return err
	}

	newID, _ := result.LastInsertId()
	policy.ParentID = policy.ID
	policy.ID = newID
	policy.Version = policy.Version + 1

	return tx.Commit()
}

func (r *Repository) GetByID(id int64) (*Policy, error) {
	var p Policy
	err := r.db.Get(&p, "SELECT * FROM policies WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) List(policyType, search string, page, pageSize int) ([]Policy, int64, error) {
	where := "is_current = 1"
	args := []interface{}{}

	if policyType != "" {
		where += " AND policy_type = ?"
		args = append(args, policyType)
	}
	if search != "" {
		where += " AND (title LIKE ? OR content LIKE ?)"
		s := "%" + search + "%"
		args = append(args, s, s)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM policies WHERE %s", where)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM policies WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", where)
	queryArgs := append(args, pageSize, offset)

	var policies []Policy
	err := r.db.Select(&policies, query, queryArgs...)
	return policies, total, err
}

func (r *Repository) GetVersions(policyID int64) ([]Policy, error) {
	var versions []Policy
	err := r.db.Select(&versions, `WITH RECURSIVE version_chain AS (
		SELECT * FROM policies WHERE id = ?
		UNION ALL
		SELECT p.* FROM policies p JOIN version_chain v ON p.id = v.parent_id
	) SELECT * FROM version_chain ORDER BY version DESC`, policyID)
	return versions, err
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("UPDATE policies SET is_current = 0 WHERE id = ?", id)
	return err
}
