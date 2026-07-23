package auth

import (
	"probig/internal/common"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository() *Repository {
	return &Repository{db: common.DB}
}

func (r *Repository) FindByUsername(username string) (*User, error) {
	var u User
	err := r.db.Get(&u, "SELECT * FROM users WHERE username = ? AND status = 1", username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindByID(id int64) (*User, error) {
	var u User
	err := r.db.Get(&u, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(user *User) error {
	_, err := r.db.Exec(`INSERT INTO users (username, password_hash, role, status) VALUES (?, ?, ?, ?)`,
		user.Username, user.PasswordHash, user.Role, user.Status)
	return err
}

func (r *Repository) ListUsers(page, pageSize int) ([]User, int64, error) {
	var total int64
	if err := r.db.Get(&total, "SELECT COUNT(*) FROM users"); err != nil {
		return nil, 0, err
	}

	var users []User
	offset := (page - 1) * pageSize
	err := r.db.Select(&users, "SELECT * FROM users ORDER BY id ASC LIMIT ? OFFSET ?", pageSize, offset)
	return users, total, err
}

func (r *Repository) SyncRolePermissions(roleName string, permissions []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM role_permissions WHERE role_name = ?", roleName); err != nil {
		return err
	}

	if _, err := tx.Exec("INSERT OR IGNORE INTO roles (name, description) VALUES (?, ?)", roleName, roleName); err != nil {
		return err
	}

	for _, p := range permissions {
		if _, err := tx.Exec("INSERT OR IGNORE INTO role_permissions (role_name, permission_code) VALUES (?, ?)", roleName, p); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetRolePermissions(roleName string) ([]string, error) {
	var perms []string
	err := r.db.Select(&perms, "SELECT permission_code FROM role_permissions WHERE role_name = ?", roleName)
	return perms, err
}

func (r *Repository) GetAllRoles() ([]RoleConfig, error) {
	rows, err := r.db.Query("SELECT DISTINCT role_name FROM role_permissions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roleNames = append(roleNames, name)
	}

	var roles []RoleConfig
	for _, name := range roleNames {
		perms, err := r.GetRolePermissions(name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, RoleConfig{Name: name, Permissions: perms})
	}
	return roles, nil
}
