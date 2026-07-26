package rbac

import (
	"fmt"

	"probig/internal/pkg/config"
	"probig/internal/pkg/encrypt"
	"probig/internal/pkg/jwt"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) Login(username, password string) (*User, string, error) {
	var user User
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, "", err
	}
	if user.Status != 1 {
		return nil, "", gorm.ErrRecordNotFound
	}
	if !encrypt.CheckPassword(password, user.Password) {
		return nil, "", gorm.ErrRecordNotFound
	}
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", err
	}
	return &user, token, nil
}

func (s *Service) ChangePassword(userID uint, oldPW, newPW string) error {
	var user User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return err
	}
	if !encrypt.CheckPassword(oldPW, user.Password) {
		return gorm.ErrRecordNotFound
	}
	hash, err := encrypt.HashPassword(newPW)
	if err != nil {
		return err
	}
	return s.DB.Model(&user).Updates(map[string]interface{}{
		"password":       hash,
		"is_first_login": false,
	}).Error
}

func (s *Service) ResetPassword(userID uint) (string, error) {
	hash, err := encrypt.HashPassword("admin123")
	if err != nil {
		return "", err
	}
	err = s.DB.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":       hash,
		"is_first_login": true,
	}).Error
	return "admin123", err
}

func (s *Service) GetUserList(pageNum, pageSize int, username string) ([]map[string]interface{}, int64, error) {
	var users []User
	var total int64
	db := s.DB.Model(&User{})
	if username != "" {
		db = db.Where("username like ?", "%"+username+"%")
	}
	db.Count(&total)
	if err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var result []map[string]interface{}
	for _, u := range users {
		var roleNames []string
		var roleIDs []UserRole
		s.DB.Where("user_id = ?", u.ID).Find(&roleIDs)
		for _, ur := range roleIDs {
			var role Role
			if err := s.DB.First(&role, ur.RoleID).Error; err == nil {
				roleNames = append(roleNames, role.Name)
			}
		}

		rolesStr := ""
		for i, rn := range roleNames {
			if i > 0 {
				rolesStr += ", "
			}
			rolesStr += rn
		}

		result = append(result, map[string]interface{}{
			"id":             u.ID,
			"username":       u.Username,
			"person_id":      u.PersonID,
			"status":         u.Status,
			"is_first_login": u.IsFirstLogin,
			"roles":          rolesStr,
			"created_at":     u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, total, nil
}

func (s *Service) CreateUser(username, password string, roleIDs []uint) error {
	hash, err := encrypt.HashPassword(password)
	if err != nil {
		return err
	}
	user := &User{
		Username:     username,
		Password:     hash,
		Status:       1,
		IsFirstLogin: true,
	}
	if err := s.DB.Create(user).Error; err != nil {
		return err
	}
	for _, rid := range roleIDs {
		s.DB.Create(&UserRole{UserID: user.ID, RoleID: rid})
	}
	return nil
}

func (s *Service) UpdateUser(userID uint, username string, roleIDs []uint) error {
	if err := s.DB.Model(&User{}).Where("id = ?", userID).Update("username", username).Error; err != nil {
		return err
	}
	s.DB.Where("user_id = ?", userID).Delete(&UserRole{})
	for _, rid := range roleIDs {
		s.DB.Create(&UserRole{UserID: userID, RoleID: rid})
	}
	return nil
}

func (s *Service) DeleteUser(userID uint) error {
	return s.DB.Delete(&User{}, userID).Error
}

func (s *Service) ToggleUserStatus(userID uint) error {
	var user User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return err
	}
	newStatus := 1
	if user.Status == 1 {
		newStatus = 0
	}
	return s.DB.Model(&user).Update("status", newStatus).Error
}

func (s *Service) GetUserRoles(userID uint) ([]Role, error) {
	var ur []UserRole
	s.DB.Where("user_id = ?", userID).Find(&ur)
	var roles []Role
	for _, r := range ur {
		var role Role
		if err := s.DB.First(&role, r.RoleID).Error; err == nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (s *Service) GetCurrentUser(userID uint) (map[string]interface{}, error) {
	var user User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	perms, _ := s.GetUserPermissions(userID)
	return map[string]interface{}{
		"id":             user.ID,
		"username":       user.Username,
		"is_first_login": user.IsFirstLogin,
		"permissions":    perms,
	}, nil
}

func (s *Service) GetUserPermissions(userID uint) ([]string, error) {
	var roleIDs []uint
	var urs []UserRole
	s.DB.Where("user_id = ?", userID).Find(&urs)
	for _, ur := range urs {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	var isAdmin bool
	var roles []Role
	s.DB.Where("id in ?", roleIDs).Find(&roles)
	for _, r := range roles {
		if r.IsAdmin {
			isAdmin = true
			break
		}
	}

	if isAdmin {
		var allPerms []Permission
		s.DB.Find(&allPerms)
		var keys []string
		for _, p := range allPerms {
			keys = append(keys, p.PermKey)
		}
		return keys, nil
	}

	var permIDs []uint
	var rps []RolePermission
	s.DB.Where("role_id in ?", roleIDs).Find(&rps)
	for _, rp := range rps {
		permIDs = append(permIDs, rp.PermissionID)
	}

	var perms []Permission
	s.DB.Where("id in ?", permIDs).Find(&perms)
	var keys []string
	for _, p := range perms {
		keys = append(keys, p.PermKey)
	}
	return keys, nil
}

func (s *Service) GetRoles() ([]Role, error) {
	var roles []Role
	err := s.DB.Find(&roles).Error
	return roles, err
}

func (s *Service) CreateRole(name, remark string) error {
	return s.DB.Create(&Role{Name: name, Remark: remark}).Error
}

func (s *Service) UpdateRole(roleID uint, name, remark string) error {
	var role Role
	if err := s.DB.First(&role, roleID).Error; err != nil {
		return err
	}
	if role.IsAdmin {
		return fmt.Errorf("超级管理员角色不可修改")
	}
	return s.DB.Model(&Role{}).Where("id = ?", roleID).Updates(map[string]interface{}{
		"name":   name,
		"remark": remark,
	}).Error
}

func (s *Service) DeleteRole(roleID uint) error {
	var role Role
	if err := s.DB.First(&role, roleID).Error; err != nil {
		return err
	}
	if role.IsAdmin {
		return nil
	}
	s.DB.Where("role_id = ?", roleID).Delete(&RolePermission{})
	s.DB.Where("role_id = ?", roleID).Delete(&UserRole{})
	return s.DB.Delete(&Role{}, roleID).Error
}

func (s *Service) GetRolePermissions(roleID uint) ([]uint, error) {
	var rps []RolePermission
	s.DB.Where("role_id = ?", roleID).Find(&rps)
	var ids []uint
	for _, rp := range rps {
		ids = append(ids, rp.PermissionID)
	}
	return ids, nil
}

func (s *Service) SetRolePermissions(roleID uint, permIDs []uint) error {
	var role Role
	if err := s.DB.First(&role, roleID).Error; err != nil {
		return err
	}
	if role.IsAdmin {
		return fmt.Errorf("超级管理员角色权限不可修改")
	}
	s.DB.Where("role_id = ?", roleID).Delete(&RolePermission{})
	for _, pid := range permIDs {
		s.DB.Create(&RolePermission{RoleID: roleID, PermissionID: pid})
	}
	return nil
}

func (s *Service) GetAllPermissions() ([]Permission, error) {
	var perms []Permission
	err := s.DB.Find(&perms).Error
	return perms, err
}

func (s *Service) BindPerson(userID, personID uint) error {
	return s.DB.Model(&User{}).Where("id = ?", userID).Update("person_id", personID).Error
}
