package rbac

import (
	"probig/internal/pkg/database"

	"gorm.io/gorm"
)

func CreateUser(tx *gorm.DB, user *database.User) error {
	return tx.Create(user).Error
}

func UpdateUser(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&database.User{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteUser(tx *gorm.DB, id uint) error {
	return tx.Delete(&database.User{}, id).Error
}

func RestoreUser(tx *gorm.DB, id uint) error {
	return tx.Unscoped().Model(&database.User{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func GetUserByID(id uint) (*database.User, error) {
	var user database.User
	err := database.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUsername(username string) (*database.User, error) {
	var user database.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UserExistsByUsername(username string, excludeID uint) (bool, error) {
	var count int64
	query := database.DB.Model(&database.User{}).Where("username = ?", username)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

type ListUsersFilter struct {
	Username string
	Status   *int8
}

func ListUsers(pageNum, pageSize int, filter ListUsersFilter) ([]database.User, int64, error) {
	var users []database.User
	var total int64

	query := database.DB.Model(&database.User{})

	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func GetUserRoles(userID uint) ([]database.Role, error) {
	var roles []database.Role
	err := database.DB.Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Find(&roles).Error
	return roles, err
}

func UpdateUserRoles(tx *gorm.DB, userID uint, roleIDs []uint) error {
	if err := tx.Where("user_id = ?", userID).Delete(&database.UserRole{}).Error; err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		ur := database.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := tx.Create(&ur).Error; err != nil {
			return err
		}
	}
	return nil
}

func GetUserPermissions(userID uint) ([]string, error) {
	var permKeys []string
	err := database.DB.Table("permissions").
		Select("DISTINCT permissions.perm_key").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("permissions.perm_key", &permKeys).Error
	return permKeys, err
}

func CreateRole(tx *gorm.DB, role *database.Role) error {
	return tx.Create(role).Error
}

func UpdateRole(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&database.Role{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteRole(tx *gorm.DB, id uint) error {
	return tx.Delete(&database.Role{}, id).Error
}

func GetRoleByID(id uint) (*database.Role, error) {
	var role database.Role
	err := database.DB.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func ListRoles() ([]database.Role, error) {
	var roles []database.Role
	err := database.DB.Order("id ASC").Find(&roles).Error
	return roles, err
}

func IsSuperAdminRole(roleID uint) bool {
	var role database.Role
	if err := database.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		return false
	}
	return role.Name == "超级管理员"
}

func GetRolePermissions(roleID uint) ([]database.Permission, error) {
	var permissions []database.Permission
	err := database.DB.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

func UpdateRolePermissions(tx *gorm.DB, roleID uint, permIDs []uint) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&database.RolePermission{}).Error; err != nil {
		return err
	}
	for _, permID := range permIDs {
		rp := database.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		}
		if err := tx.Create(&rp).Error; err != nil {
			return err
		}
	}
	return nil
}

func ListAllPermissions() ([]database.Permission, error) {
	var permissions []database.Permission
	err := database.DB.Order("module ASC, action ASC").Find(&permissions).Error
	return permissions, err
}

func GetPermissionByKey(key string) (*database.Permission, error) {
	var perm database.Permission
	err := database.DB.Where("perm_key = ?", key).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func ResetUserPassword(tx *gorm.DB, userID uint, newHash string) error {
	return tx.Model(&database.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":       newHash,
		"is_first_login": true,
	}).Error
}
