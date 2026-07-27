package model

type RolePermission struct {
	RoleID       uint `gorm:"primaryKey" json:"role_id"`
	PermissionID uint `gorm:"primaryKey" json:"permission_id"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
