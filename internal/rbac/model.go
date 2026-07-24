package rbac

import "gorm.io/gorm"

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"type:varchar(128);not null" json:"-"`
	PersonID     *uint          `json:"personId"`
	IsFirstLogin bool           `gorm:"default:true" json:"isFirstLogin"`
	Status       int            `gorm:"default:1" json:"status"`
	Roles        []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	CreatedAt    string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt    string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(64);not null" json:"name"`
	Remark      string         `gorm:"type:varchar(256)" json:"remark"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt   string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	PermissionKey string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"permissionKey"`
	Name          string         `gorm:"type:varchar(64);not null" json:"name"`
	Module        string         `gorm:"type:varchar(32);not null" json:"module"`
	Action        string         `gorm:"type:varchar(32);not null" json:"action"`
	CreatedAt     string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt     string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Permission) TableName() string { return "permissions" }

type UserRole struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;index"`
	RoleID uint `gorm:"not null;index"`
}

func (UserRole) TableName() string { return "user_roles" }

type RolePermission struct {
	ID           uint `gorm:"primaryKey"`
	RoleID       uint `gorm:"not null;index"`
	PermissionID uint `gorm:"not null;index"`
}

func (RolePermission) TableName() string { return "role_permissions" }
