package dao

import (
	"probig/internal/models"
)

func GetUserByUsername(username string) (*models.User, error) {
	var u models.User
	if err := DB().Preload("Roles.Permissions").Preload("Person").Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(id uint) (*models.User, error) {
	var u models.User
	if err := DB().Preload("Roles.Permissions").Preload("Person").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserList(page, pageSize int, keyword string) ([]models.User, int64, error) {
	var list []models.User
	var total int64
	q := DB().Model(&models.User{}).Preload("Person")
	if keyword != "" {
		q = q.Where("username LIKE ?", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func CreateUser(u *models.User) error {
	return DB().Create(u).Error
}

func UpdateUser(u *models.User) error {
	return DB().Save(u).Error
}

func DeleteUser(id uint) error {
	tx := DB().Begin()
	if err := tx.Where("user_id = ?", id).Delete(&models.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&models.User{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func GetRoleList() ([]models.Role, error) {
	var list []models.Role
	if err := DB().Preload("Permissions").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetRoleByID(id uint) (*models.Role, error) {
	var r models.Role
	if err := DB().Preload("Permissions").First(&r, id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateRole(r *models.Role) error {
	return DB().Create(r).Error
}

func UpdateRole(r *models.Role) error {
	tx := DB().Begin()
	if err := tx.Model(r).Association("Permissions").Replace(r.Permissions); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(r).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func DeleteRole(id uint) error {
	tx := DB().Begin()
	if err := tx.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("role_id = ?", id).Delete(&models.UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&models.Role{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func GetAllPermissions() ([]models.Permission, error) {
	var list []models.Permission
	if err := DB().Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetUserPermissions(userID uint) ([]string, error) {
	var keys []string
	if err := DB().Raw(`
		SELECT DISTINCT p.permission_key
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND ur.deleted_at IS NULL AND rp.deleted_at IS NULL
	`, userID).Pluck("permission_key", &keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}
