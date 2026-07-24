package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex" json:"username"`
	PasswordHash string         `gorm:"type:varchar(128)" json:"-"`
	PersonID     *uint          `json:"person_id"`
	IsFirstLogin bool           `json:"is_first_login"`
	Status       int8           `gorm:"type:tinyint;default:1" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Person *Person `gorm:"foreignKey:PersonID" json:"person"`
	Roles  []Role  `gorm:"many2many:user_roles;" json:"roles"`
}

func (User) TableName() string {
	return "users"
}

type Role struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(64)" json:"name"`
	Remark    string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

func (Role) TableName() string {
	return "roles"
}

type Permission struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	PermissionKey string `gorm:"type:varchar(64);uniqueIndex" json:"permission_key"`
	Name          string `gorm:"type:varchar(64)" json:"name"`
	Module        string `gorm:"type:varchar(32)" json:"module"`
	Action        string `gorm:"type:varchar(32)" json:"action"`
}

func (Permission) TableName() string {
	return "permissions"
}

type UserRole struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	RoleID    uint           `gorm:"not null;index" json:"role_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

type RolePermission struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	RoleID       uint           `gorm:"not null;index" json:"role_id"`
	PermissionID uint           `gorm:"not null;index" json:"permission_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
