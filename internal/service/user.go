package service

import (
	"errors"

	"probig/internal/dao"
	"probig/internal/middleware"
	"probig/internal/models"
	"probig/pkg/crypto"

	"github.com/gin-gonic/gin"
)

func GetUserList(page, pageSize int, keyword string) ([]models.User, int64, error) {
	return dao.GetUserList(page, pageSize, keyword)
}

func GetUser(id uint) (*models.User, error) {
	return dao.GetUserByID(id)
}

func CreateUser(c *gin.Context, u *models.User, password string) error {
	existing, _ := dao.GetUserByUsername(u.Username)
	if existing != nil {
		return errors.New("用户名已存在")
	}
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	u.IsFirstLogin = true
	u.Status = 1
	if err := dao.CreateUser(u); err != nil {
		return err
	}
	middleware.RecordAudit(c, "新增", "user", u.ID, nil, u, "")
	return nil
}

func UpdateUser(c *gin.Context, u *models.User) error {
	old, err := dao.GetUserByID(u.ID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if err := dao.UpdateUser(u); err != nil {
		return err
	}
	middleware.RecordAudit(c, "修改", "user", u.ID, old, u, "")
	return nil
}

func DeleteUser(c *gin.Context, id uint) error {
	u, err := dao.GetUserByID(id)
	if err != nil {
		return errors.New("用户不存在")
	}
	if err := dao.DeleteUser(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "user", id, u, nil, "")
	return nil
}

func GetRoleList() ([]models.Role, error) {
	return dao.GetRoleList()
}

func GetRole(id uint) (*models.Role, error) {
	return dao.GetRoleByID(id)
}

func CreateRole(c *gin.Context, r *models.Role) error {
	if err := dao.CreateRole(r); err != nil {
		return err
	}
	middleware.RecordAudit(c, "新增", "role", r.ID, nil, r, "")
	return nil
}

func UpdateRole(c *gin.Context, r *models.Role) error {
	old, err := dao.GetRoleByID(r.ID)
	if err != nil {
		return errors.New("角色不存在")
	}
	if err := dao.UpdateRole(r); err != nil {
		return err
	}
	middleware.RecordAudit(c, "修改", "role", r.ID, old, r, "")
	return nil
}

func DeleteRole(c *gin.Context, id uint) error {
	r, err := dao.GetRoleByID(id)
	if err != nil {
		return errors.New("角色不存在")
	}
	if err := dao.DeleteRole(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "role", id, r, nil, "")
	return nil
}

func GetAllPermissions() ([]models.Permission, error) {
	return dao.GetAllPermissions()
}

func GetUserPermissions(userID uint) ([]string, error) {
	return dao.GetUserPermissions(userID)
}
