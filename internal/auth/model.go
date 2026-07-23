package auth

import "time"

type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	Status       int       `json:"status" db:"status"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token       string   `json:"token"`
	User        UserInfo `json:"user"`
	Permissions []string `json:"permissions"`
}

type UserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type RoleConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Permissions []string `yaml:"permissions"`
}

type MenuItem struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Icon        string     `json:"icon,omitempty"`
	Path        string     `json:"path,omitempty"`
	Permission  string     `json:"permission,omitempty"`
	Children    []MenuItem `json:"children,omitempty"`
}
