package rbac

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password     string         `gorm:"size:256;not null" json:"-"`
	PersonID     uint           `gorm:"default:0" json:"person_id"`
	Status       int            `gorm:"default:1" json:"status"`
	IsFirstLogin bool           `gorm:"default:true" json:"is_first_login"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Role struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	Remark    string         `gorm:"size:256" json:"remark"`
	IsAdmin   bool           `gorm:"default:false" json:"is_admin"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Permission struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Module   string `gorm:"size:64" json:"module"`
	PermKey  string `gorm:"size:64;uniqueIndex" json:"perm_key"`
	PermName string `gorm:"size:64" json:"perm_name"`
	Action   string `gorm:"size:32" json:"action"`
}

type UserRole struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"index"`
	RoleID uint `gorm:"index"`
}

type RolePermission struct {
	ID           uint `gorm:"primaryKey"`
	RoleID       uint `gorm:"index"`
	PermissionID uint `gorm:"index"`
}

func (UserRole) TableName() string     { return "user_roles" }
func (RolePermission) TableName() string { return "role_permissions" }
