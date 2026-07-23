package asset

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

func (r *Repository) Create(asset *Asset) error {
	result, err := r.db.Exec(`INSERT INTO assets (asset_type, name, description, content, status, version, parent_id, is_current)
		VALUES (?, ?, ?, ?, ?, 1, ?, 1)`,
		asset.AssetType, asset.Name, asset.Description, asset.Content, asset.Status, asset.ParentID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	asset.ID = id
	asset.Version = 1
	return nil
}

func (r *Repository) CreateTx(tx *sqlx.Tx, asset *Asset) error {
	result, err := tx.Exec(`INSERT INTO assets (asset_type, name, description, content, status, version, parent_id, is_current)
		VALUES (?, ?, ?, ?, ?, 1, ?, 1)`,
		asset.AssetType, asset.Name, asset.Description, asset.Content, asset.Status, asset.ParentID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	asset.ID = id
	asset.Version = 1
	return nil
}

func (r *Repository) AddVersion(asset *Asset) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE assets SET is_current = 0 WHERE id = ?", asset.ID); err != nil {
		return err
	}

	result, err := tx.Exec(`INSERT INTO assets (asset_type, name, description, content, status, version, parent_id, is_current)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
		asset.AssetType, asset.Name, asset.Description, asset.Content, asset.Status, asset.Version+1, asset.ID)
	if err != nil {
		return err
	}

	newID, _ := result.LastInsertId()
	asset.ParentID = asset.ID
	asset.ID = newID
	asset.Version = asset.Version + 1

	return tx.Commit()
}

func (r *Repository) GetByID(id int64) (*Asset, error) {
	var a Asset
	err := r.db.Get(&a, "SELECT * FROM assets WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) List(assetType, search string, page, pageSize int) ([]Asset, int64, error) {
	where := "is_current = 1"
	args := []interface{}{}

	if assetType != "" {
		where += " AND asset_type = ?"
		args = append(args, assetType)
	}
	if search != "" {
		where += " AND (name LIKE ? OR description LIKE ?)"
		s := "%" + search + "%"
		args = append(args, s, s)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM assets WHERE %s", where)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM assets WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", where)
	queryArgs := append(args, pageSize, offset)

	var assets []Asset
	err := r.db.Select(&assets, query, queryArgs...)
	return assets, total, err
}

func (r *Repository) GetVersions(assetID int64) ([]Asset, error) {
	var versions []Asset
	err := r.db.Select(&versions, `WITH RECURSIVE version_chain AS (
		SELECT * FROM assets WHERE id = ?
		UNION ALL
		SELECT a.* FROM assets a JOIN version_chain v ON a.id = v.parent_id
	) SELECT * FROM version_chain ORDER BY version DESC`, assetID)
	return versions, err
}

func (r *Repository) SoftDelete(id int64) error {
	_, err := r.db.Exec("UPDATE assets SET status = 0, is_current = 0 WHERE id = ?", id)
	return err
}
