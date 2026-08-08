package service

import (
	"context"
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
)

func GetRoleList(pageNum, pageSize int, name string) ([]model.Role, int64, error) {
	tx := dao.DB.Model(&model.Role{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}

	var total int64
	tx.Count(&total)

	var roles []model.Role
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("id ASC").Find(&roles)

	return roles, total, nil
}

func GetAllRoles() ([]model.Role, error) {
	var roles []model.Role
	if err := dao.DB.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func CreateRole(ctx context.Context, name, remark string) (*model.Role, error) {
	var count int64
	dao.DB.Model(&model.Role{}).Where("name = ?", name).Count(&count)
	if count > 0 {
		return nil, errors.New("角色名已存在")
	}

	role := model.Role{Name: name, Remark: remark}
	if err := dao.DB.Create(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func UpdateRole(ctx context.Context, id uint, name, remark string) error {
	var role model.Role
	if err := dao.DB.First(&role, id).Error; err != nil {
		return errors.New("角色不存在")
	}
	if role.IsDefault {
		return errors.New("默认角色不可修改")
	}

	var count int64
	dao.DB.Model(&model.Role{}).Where("name = ? AND id != ?", name, id).Count(&count)
	if count > 0 {
		return errors.New("角色名已存在")
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if remark != "" {
		updates["remark"] = remark
	}
	return dao.DBFromContext(ctx).Model(&role).Updates(updates).Error
}

func DeleteRole(ctx context.Context, id uint) error {
	var role model.Role
	if err := dao.DB.First(&role, id).Error; err != nil {
		return errors.New("角色不存在")
	}
	if role.IsDefault {
		return errors.New("默认角色不可删除")
	}
	if err := dao.DBFromContext(ctx).Delete(&role).Error; err != nil {
		return err
	}
	InvalidateRolePermissionCache(id)
	return nil
}

func RestoreRole(ctx context.Context, id uint) error {
	return dao.RestoreEntity[model.Role](dao.DBFromContext(ctx), id)
}

func GetDeletedRoleList(pageNum, pageSize int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	tx := dao.DB.Unscoped().Model(&model.Role{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&roles)

	return roles, total, nil
}

func AssignRolePermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	var role model.Role
	if err := dao.DB.First(&role, roleID).Error; err != nil {
		return errors.New("角色不存在")
	}
	if role.IsDefault {
		return errors.New("默认角色不可修改权限")
	}

	dao.DBFromContext(ctx).Where("role_id = ?", roleID).Delete(&model.RolePermission{})

	for _, permID := range permIDs {
		rp := model.RolePermission{RoleID: roleID, PermissionID: permID}
		dao.DBFromContext(ctx).Create(&rp)
	}
	InvalidateRolePermissionCache(roleID)
	return nil
}

func GetAllPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	if err := dao.DB.Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func GetRolePermissionIDs(roleID uint) []uint {
	var rps []model.RolePermission
	dao.DB.Where("role_id = ?", roleID).Find(&rps)
	ids := make([]uint, len(rps))
	for i, rp := range rps {
		ids[i] = rp.PermissionID
	}
	return ids
}

func GetRoleByID(roleID uint) (*model.Role, error) {
	var role model.Role
	if err := dao.DB.First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
