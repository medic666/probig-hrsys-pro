package service

import (
	"errors"
	"fmt"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"
)

type CreateUserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	PersonID *uint  `json:"person_id"`
	IsActive *bool  `json:"is_active"`
}

type UpdateUserReq struct {
	Username *string `json:"username"`
	PersonID *uint   `json:"person_id"`
	IsActive *bool   `json:"is_active"`
}

func GetUserList(pageNum, pageSize int, username string, isActive *bool) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.User{})
	if username != "" {
		tx = tx.Where("username LIKE ?", "%"+username+"%")
	}
	if isActive != nil {
		tx = tx.Where("is_active = ?", *isActive)
	}

	var total int64
	tx.Count(&total)

	var users []model.User
	offset := (pageNum - 1) * pageSize
	tx.Preload("Roles").Offset(offset).Limit(pageSize).Order("id DESC").Find(&users)

	var result []map[string]interface{}
	for _, user := range users {
		item := map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"person_id":  user.PersonID,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
		}

		roleNames := make([]string, 0, len(user.Roles))
		for _, r := range user.Roles {
			roleNames = append(roleNames, r.Name)
		}
		item["roles"] = roleNames

		if user.PersonID != nil {
			var personName string
			dao.DB.Table("persons").Select("name").Where("id = ?", *user.PersonID).Scan(&personName)
			item["person_name"] = personName
		}

		result = append(result, item)
	}

	return result, total, nil
}

func CreateUser(req CreateUserReq) (*model.User, error) {
	var count int64
	dao.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	user := model.User{
		Username:     req.Username,
		Password:     hash,
		PersonID:     req.PersonID,
		IsActive:     active,
		IsFirstLogin: true,
	}
	if err := dao.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(id uint, req UpdateUserReq) error {
	var user model.User
	if err := dao.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	updates := map[string]interface{}{}
	if req.Username != nil {
		var count int64
		dao.DB.Model(&model.User{}).Where("username = ? AND id != ?", *req.Username, id).Count(&count)
		if count > 0 {
			return errors.New("用户名已存在")
		}
		updates["username"] = *req.Username
	}
	if req.PersonID != nil {
		updates["person_id"] = req.PersonID
	}
	if req.IsActive != nil {
		if id == 1 && !*req.IsActive {
			return errors.New("不能禁用超级管理员账号")
		}
		updates["is_active"] = *req.IsActive
	}

	return dao.DB.Model(&user).Updates(updates).Error
}

func DeleteUser(id, operatorID uint) error {
	if id == 1 {
		return errors.New("不能删除超级管理员账号")
	}
	if id == operatorID {
		return errors.New("不能删除自己的账号")
	}
	var user model.User
	if err := dao.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	return dao.DB.Delete(&user).Error
}

func RestoreUser(id uint) error {
	return dao.RestoreEntity[model.User](dao.DB, id)
}

func AssignUserRoles(userID uint, roleIDs []uint) error {
	if userID == 1 {
		var defaultRole model.Role
		if err := dao.DB.Where("is_default = ?", true).First(&defaultRole).Error; err != nil {
			return errors.New("默认角色不存在")
		}
		hasDefault := false
		for _, rid := range roleIDs {
			if rid == defaultRole.ID {
				hasDefault = true
				break
			}
		}
		if !hasDefault {
			roleIDs = append(roleIDs, defaultRole.ID)
		}
	}

	dao.DB.Where("user_id = ?", userID).Delete(&model.UserRole{})

	for _, roleID := range roleIDs {
		ur := model.UserRole{UserID: userID, RoleID: roleID}
		dao.DB.Create(&ur)
	}
	return nil
}

func GetUserWithRoles(userID uint) (*model.User, error) {
	var user model.User
	if err := dao.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := dao.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetDeletedUserList(pageNum, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	tx := dao.DB.Unscoped().Model(&model.User{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&users)

	return users, total, nil
}

func GetUserRoleIDs(userID uint) []uint {
	var urs []model.UserRole
	dao.DB.Where("user_id = ?", userID).Find(&urs)
	ids := make([]uint, len(urs))
	for i, ur := range urs {
		ids[i] = ur.RoleID
	}
	return ids
}

func FormatSalary(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
