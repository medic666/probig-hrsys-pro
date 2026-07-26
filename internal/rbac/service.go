package rbac

import (
	"errors"
	"fmt"

	"probig/internal/pkg/database"
	"probig/internal/pkg/encrypt"
	jwtPkg "probig/internal/pkg/jwt"

	"gorm.io/gorm"
)

type Service struct{}

var DefaultService = &Service{}

func (s *Service) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return database.DB
}

func (s *Service) Login(username, password, ip string) (*LoginResponse, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用，请联系管理员")
	}

	if !encrypt.CheckPassword(password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := jwtPkg.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	return &LoginResponse{
		Token:        token,
		Username:     user.Username,
		IsFirstLogin: user.IsFirstLogin,
	}, nil
}

func (s *Service) ChangePassword(userID uint, oldPwd, newPwd string) error {
	user, err := GetUserByID(userID)
	if err != nil {
		return err
	}

	if !encrypt.CheckPassword(oldPwd, user.Password) {
		return errors.New("原密码错误")
	}

	if oldPwd == newPwd {
		return errors.New("新密码不能与原密码相同")
	}

	if len(newPwd) < 6 {
		return errors.New("新密码长度不能少于6位")
	}

	hashed, err := encrypt.HashPassword(newPwd)
	if err != nil {
		return err
	}

	return UpdateUser(nil, userID, map[string]interface{}{
		"password":       hashed,
		"is_first_login": false,
	})
}

func (s *Service) CreateUser(username, password string, personID *uint) (*database.User, error) {
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if password == "" {
		return nil, errors.New("密码不能为空")
	}
	if len(password) < 6 {
		return nil, errors.New("密码长度不能少于6位")
	}

	exists, err := UserExistsByUsername(username, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	hashed, err := encrypt.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &database.User{
		Username:     username,
		Password:     hashed,
		PersonID:     personID,
		IsFirstLogin: true,
		Status:       1,
	}

	if err := CreateUser(nil, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateUser(id uint, req *UpdateUserRequest) error {
	user, err := GetUserByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if req.PersonID != nil {
		updates["person_id"] = *req.PersonID
	}
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			return errors.New("用户状态值无效")
		}
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return nil
	}

	return UpdateUser(nil, user.ID, updates)
}

func (s *Service) DeleteUser(id uint) error {
	_, err := GetUserByID(id)
	if err != nil {
		return err
	}
	return DeleteUser(nil, id)
}

func (s *Service) RestoreUser(id uint) error {
	return RestoreUser(nil, id)
}

func (s *Service) ListUsers(pageNum, pageSize int, username string, status *int8) ([]database.User, int64, error) {
	return ListUsers(pageNum, pageSize, ListUsersFilter{
		Username: username,
		Status:   status,
	})
}

func (s *Service) GetUserInfo(userID uint) (map[string]interface{}, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	permissions, err := GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	roles, err := GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user":        user,
		"permissions": permissions,
		"roles":       roles,
	}, nil
}

func (s *Service) GetUserRoles(userID uint) ([]database.Role, error) {
	return GetUserRoles(userID)
}

func (s *Service) AssignUserRoles(userID uint, roleIDs []uint) error {
	if len(roleIDs) == 0 {
		return errors.New("角色列表不能为空")
	}

	for _, roleID := range roleIDs {
		_, err := GetRoleByID(roleID)
		if err != nil {
			return fmt.Errorf("角色ID %d 不存在", roleID)
		}
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		return UpdateUserRoles(tx, userID, roleIDs)
	})
}

func (s *Service) ResetUserPassword(userID uint, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("新密码长度不能少于6位")
	}

	hashed, err := encrypt.HashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = GetUserByID(userID)
	if err != nil {
		return err
	}

	return ResetUserPassword(nil, userID, hashed)
}

func (s *Service) CreateRole(name, remark string) (*database.Role, error) {
	if name == "" {
		return nil, errors.New("角色名称不能为空")
	}

	role := &database.Role{
		Name:   name,
		Remark: remark,
	}

	if err := CreateRole(nil, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) UpdateRole(id uint, req *UpdateRoleRequest) error {
	role, err := GetRoleByID(id)
	if err != nil {
		return err
	}

	if IsSuperAdminRole(role.ID) {
		return errors.New("超级管理员角色不允许修改")
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	if len(updates) == 0 {
		return nil
	}

	return UpdateRole(nil, id, updates)
}

func (s *Service) DeleteRole(id uint) error {
	if IsSuperAdminRole(id) {
		return errors.New("超级管理员角色不允许删除")
	}

	_, err := GetRoleByID(id)
	if err != nil {
		return err
	}

	return DeleteRole(nil, id)
}

func (s *Service) ListRoles() ([]database.Role, error) {
	return ListRoles()
}

func (s *Service) GetRolePermissions(roleID uint) ([]database.Permission, error) {
	return GetRolePermissions(roleID)
}

func (s *Service) AssignRolePermissions(roleID uint, permIDs []uint) error {
	if IsSuperAdminRole(roleID) {
		return errors.New("超级管理员角色不允许修改权限")
	}

	for _, permID := range permIDs {
		var perm database.Permission
		if err := database.DB.First(&perm, permID).Error; err != nil {
			return fmt.Errorf("权限ID %d 不存在", permID)
		}
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		return UpdateRolePermissions(tx, roleID, permIDs)
	})
}

func (s *Service) ListAllPermissions() ([]database.Permission, error) {
	return ListAllPermissions()
}
