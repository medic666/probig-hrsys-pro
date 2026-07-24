package rbac

import (
	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) CreateUser(user *User) error {
	return d.db.Create(user).Error
}

func (d *DAO) GetUserByUsername(username string) (*User, error) {
	var user User
	err := d.db.Where("username = ?", username).Preload("Roles.Permissions").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *DAO) GetUserByID(id uint) (*User, error) {
	var user User
	err := d.db.Preload("Roles.Permissions").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *DAO) ListUsers(page, pageSize int, username string) ([]User, int64, error) {
	var users []User
	var total int64
	query := d.db.Model(&User{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Preload("Roles").Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (d *DAO) UpdateUser(user *User) error {
	return d.db.Model(user).Select("username", "person_id", "status", "is_first_login", "password_hash").Updates(user).Error
}

func (d *DAO) DeleteUser(id uint) error {
	return d.db.Delete(&User{}, id).Error
}

func (d *DAO) UpdateUserRoles(userID uint, roleIDs []uint) error {
	var user User
	user.ID = userID

	if len(roleIDs) == 0 {
		return d.db.Model(&user).Association("Roles").Clear()
	}

	var roles []Role
	if err := d.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return err
	}
	return d.db.Model(&user).Association("Roles").Replace(roles)
}

func (d *DAO) GetAllUsers() ([]User, error) {
	var users []User
	err := d.db.Find(&users).Error
	return users, err
}

func (d *DAO) CreateRole(role *Role) error {
	return d.db.Create(role).Error
}

func (d *DAO) GetRoleByID(id uint) (*Role, error) {
	var role Role
	err := d.db.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (d *DAO) ListRoles(page, pageSize int, name string) ([]Role, int64, error) {
	var roles []Role
	var total int64
	query := d.db.Model(&Role{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&roles).Error
	return roles, total, err
}

func (d *DAO) UpdateRole(role *Role) error {
	return d.db.Model(role).Select("name", "remark").Updates(role).Error
}

func (d *DAO) DeleteRole(id uint) error {
	return d.db.Delete(&Role{}, id).Error
}

func (d *DAO) UpdateRolePermissions(roleID uint, permIDs []uint) error {
	var role Role
	role.ID = roleID

	if len(permIDs) == 0 {
		return d.db.Model(&role).Association("Permissions").Clear()
	}

	var permissions []Permission
	if err := d.db.Where("id IN ?", permIDs).Find(&permissions).Error; err != nil {
		return err
	}
	return d.db.Model(&role).Association("Permissions").Replace(permissions)
}

func (d *DAO) GetAllRoles() ([]Role, error) {
	var roles []Role
	err := d.db.Find(&roles).Error
	return roles, err
}

func (d *DAO) GetAllPermissions() ([]Permission, error) {
	var permissions []Permission
	err := d.db.Find(&permissions).Error
	return permissions, err
}

func (d *DAO) CheckPermission(userID uint, permissionKey string) (bool, error) {
	var count int64
	err := d.db.Raw(`
		SELECT COUNT(*) FROM user_roles ur
		JOIN role_permissions rp ON ur.role_id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = ? AND p.permission_key = ?
	`, userID, permissionKey).Scan(&count).Error
	return count > 0, err
}
