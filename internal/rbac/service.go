package rbac

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"probig/internal/pkg/audit"
)

type Service struct {
	dao *DAO
}

var globalService *Service

func NewService(db *gorm.DB) *Service {
	svc := &Service{dao: NewDAO(db)}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) CreateUser(user *User) error {
	return s.dao.CreateUser(user)
}

func (s *Service) GetUserByID(id uint) (*User, error) {
	return s.dao.GetUserByID(id)
}

func (s *Service) ListUsers(page, pageSize int, username string) ([]User, int64, error) {
	return s.dao.ListUsers(page, pageSize, username)
}

func (s *Service) UpdateUser(user *User) error {
	return s.dao.UpdateUser(user)
}

func (s *Service) DeleteUser(id uint) error {
	if err := s.dao.DeleteUser(id); err != nil {
		return err
	}
	if audit.GlobalAuditService != nil {
		audit.GlobalAuditService.Log(0, "", "user", id, "delete", "", "", "", "")
	}
	return nil
}

func (s *Service) AssignRoles(userID uint, roleIDs []uint) error {
	return s.dao.UpdateUserRoles(userID, roleIDs)
}

func (s *Service) ResetPassword(userID uint, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.dao.UpdateUser(&User{
		ID:           userID,
		PasswordHash: string(hash),
		IsFirstLogin: true,
	})
}

func (s *Service) Login(username, password string) (*User, error) {
	user, err := s.dao.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}
	if user.Status != 1 {
		return nil, errors.New("account is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}
	return user, nil
}

func (s *Service) ChangePassword(userID uint, oldPwd, newPwd string) error {
	user, err := s.dao.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return errors.New("old password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.dao.UpdateUser(&User{
		ID:           userID,
		PasswordHash: string(hash),
		IsFirstLogin: false,
	})
}

func (s *Service) CreateRole(role *Role) error {
	return s.dao.CreateRole(role)
}

func (s *Service) GetRoleByID(id uint) (*Role, error) {
	return s.dao.GetRoleByID(id)
}

func (s *Service) ListRoles(page, pageSize int, name string) ([]Role, int64, error) {
	return s.dao.ListRoles(page, pageSize, name)
}

func (s *Service) UpdateRole(role *Role) error {
	return s.dao.UpdateRole(role)
}

func (s *Service) DeleteRole(id uint) error {
	return s.dao.DeleteRole(id)
}

func (s *Service) AssignPermissions(roleID uint, permIDs []uint) error {
	return s.dao.UpdateRolePermissions(roleID, permIDs)
}

func (s *Service) CheckPermission(userID uint, permissionKey string) bool {
	ok, err := s.dao.CheckPermission(userID, permissionKey)
	if err != nil {
		return false
	}
	return ok
}

func (s *Service) GetAllPermissions() ([]Permission, error) {
	return s.dao.GetAllPermissions()
}
