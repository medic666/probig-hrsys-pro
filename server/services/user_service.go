package services

import (
	"probig/database"
	"probig/models"
)

func ListUsers(query string, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	db := database.DB.Model(&models.User{}).Preload("Person").Preload("Roles")
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("username LIKE ?", like)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error
	return users, total, err
}

func GetUser(id uint) (*models.User, error) {
	var user models.User
	err := database.DB.Preload("Person").Preload("Roles").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(user *models.User) error {
	return database.DB.Create(user).Error
}

func UpdateUser(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteUser(id uint) error {
	return database.DB.Delete(&models.User{}, id).Error
}

func AssignUserRoles(userID uint, roleIDs []uint) error {
	database.DB.Where("user_id = ?", userID).Delete(&models.UserRole{})
	for _, rid := range roleIDs {
		database.DB.Create(&models.UserRole{UserID: userID, RoleID: rid})
	}
	return nil
}

func ListRoles(query string, offset, limit int) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64
	db := database.DB.Model(&models.Role{}).Preload("Permissions")
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("name LIKE ?", like)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&roles).Error
	return roles, total, err
}

func GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := database.DB.Find(&roles).Error
	return roles, err
}

func GetRole(id uint) (*models.Role, error) {
	var role models.Role
	err := database.DB.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func CreateRole(role *models.Role) error {
	return database.DB.Create(role).Error
}

func UpdateRole(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteRole(id uint) error {
	database.DB.Where("role_id = ?", id).Delete(&models.RolePermission{})
	database.DB.Where("role_id = ?", id).Delete(&models.UserRole{})
	return database.DB.Delete(&models.Role{}, id).Error
}

func AssignRolePermissions(roleID uint, permIDs []uint) error {
	database.DB.Where("role_id = ?", roleID).Delete(&models.RolePermission{})
	for _, pid := range permIDs {
		database.DB.Create(&models.RolePermission{RoleID: roleID, PermissionID: pid})
	}
	return nil
}

func ListPermissions() ([]models.Permission, error) {
	var perms []models.Permission
	err := database.DB.Order("module ASC, action ASC").Find(&perms).Error
	return perms, err
}

func ListUsersSimple() ([]models.User, error) {
	var users []models.User
	err := database.DB.Select("id, username").Where("status = 1").Find(&users).Error
	return users, err
}
