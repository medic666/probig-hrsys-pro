package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string    `gorm:"size:256;not null" json:"-"`
	RealName  string    `gorm:"size:64;not null" json:"real_name"`
	Status    string    `gorm:"size:16;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Roles     []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Description string       `gorm:"size:256" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type Permission struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Module string `gorm:"size:32;not null" json:"module"`
	Action string `gorm:"size:16;not null" json:"action"`
}
