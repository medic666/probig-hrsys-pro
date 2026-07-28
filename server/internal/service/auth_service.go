package service

import (
	"errors"
	"fmt"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

const DefaultPassword = "admin123"

func Login(username, password string) (string, *model.User, error) {
	var user model.User
	if err := dao.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	if !user.IsActive {
		return "", nil, fmt.Errorf("账号已被禁用")
	}

	if !utils.CheckPassword(password, user.Password) {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", nil, fmt.Errorf("生成Token失败")
	}

	return token, &user, nil
}

func ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user model.User
	if err := dao.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if !user.IsFirstLogin {
		if !utils.CheckPassword(oldPassword, user.Password) {
			return errors.New("原密码错误")
		}
	}

	return setUserPassword(&user, newPassword)
}

func ResetPassword(userID uint) (string, error) {
	var user model.User
	if err := dao.DB.First(&user, userID).Error; err != nil {
		return "", errors.New("用户不存在")
	}

	newPwd := DefaultPassword
	if err := setUserPassword(&user, newPwd); err != nil {
		return "", err
	}
	return newPwd, nil
}

func setUserPassword(user *model.User, password string) error {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	user.Password = hash
	user.IsFirstLogin = false
	return dao.DB.Save(user).Error
}

func GetUserPermissions(userID uint) ([]string, []map[string]interface{}, error) {
	var permissions []model.Permission
	dao.DB.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&permissions)

	var permKeys []string
	for _, p := range permissions {
		permKeys = append(permKeys, p.Module+"."+p.Action)
	}

	menus := buildMenuTree(permKeys)

	return permKeys, menus, nil
}

func buildMenuTree(permKeys []string) []map[string]interface{} {
	has := func(key string) bool {
		for _, k := range permKeys {
			if k == key {
				return true
			}
		}
		return false
	}

	var menus []map[string]interface{}

	if has("home") {
		menus = append(menus, map[string]interface{}{
			"path": "/home", "title": "首页", "icon": "HomeFilled",
		})
	}

	if has("person.read") || has("company.read") || has("position_event.read") {
		children := []map[string]interface{}{}
		if has("person.read") {
			children = append(children, map[string]interface{}{"path": "/person", "title": "人员管理", "icon": "User"})
		}
		if has("company.read") {
			children = append(children, map[string]interface{}{"path": "/company", "title": "公司管理", "icon": "OfficeBuilding"})
		}
		if has("position_event.read") {
			children = append(children, map[string]interface{}{"path": "/position-event", "title": "职务事件", "icon": "Document"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/data", "title": "主数据管理", "icon": "OfficeBuilding", "children": children,
		})
	}

	if has("attendance.read") {
		children := []map[string]interface{}{}
		if has("attendance.read") {
			children = append(children, map[string]interface{}{"path": "/attendance", "title": "考勤事件", "icon": "Clock"})
			children = append(children, map[string]interface{}{"path": "/attendance-daily", "title": "日记工时", "icon": "Clock"})
			children = append(children, map[string]interface{}{"path": "/attendance-monthly", "title": "月度考勤核算", "icon": "Clock"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/attendance-group", "title": "考勤管理", "icon": "Clock", "children": children,
		})
	}

	if has("annual_leave.read") {
		children := []map[string]interface{}{}
		if has("annual_leave.read") {
			children = append(children, map[string]interface{}{"path": "/annual-leave-events", "title": "年假事件", "icon": "Clock"})
			children = append(children, map[string]interface{}{"path": "/annual-leave-balance", "title": "年假余额", "icon": "Clock"})
			children = append(children, map[string]interface{}{"path": "/annual-leave-carryover", "title": "周年结转", "icon": "Clock"})
			children = append(children, map[string]interface{}{"path": "/lil-events", "title": "调休事件", "icon": "Clock"})
			children = append(children, map[string]interface{}{"path": "/lil-balance", "title": "调休余额", "icon": "Clock"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/leave-group", "title": "假期管理", "icon": "Clock", "children": children,
		})
	}

	if has("salary.read") {
		children := []map[string]interface{}{}
		if has("salary.read") {
			children = append(children, map[string]interface{}{"path": "/salary-events", "title": "工资事件", "icon": "Money"})
			children = append(children, map[string]interface{}{"path": "/salary-summaries", "title": "月度工资汇总", "icon": "Money"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/salary-group", "title": "薪资管理", "icon": "Money", "children": children,
		})
	}

	if has("user.read") || has("role.read") || has("system_config.read") {
		children := []map[string]interface{}{}
		if has("user.read") {
			children = append(children, map[string]interface{}{"path": "/system/users", "title": "用户管理", "icon": "Setting"})
		}
		if has("role.read") {
			children = append(children, map[string]interface{}{"path": "/system/roles", "title": "角色管理", "icon": "Setting"})
		}
		if has("system_config.read") {
			children = append(children, map[string]interface{}{"path": "/system/config", "title": "系统配置", "icon": "Setting"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/system", "title": "系统管理", "icon": "Setting", "children": children,
		})
	}

	if has("file.read") {
		menus = append(menus, map[string]interface{}{"path": "/files", "title": "文件管理", "icon": "Document"})
	}

	if has("audit.read") {
		menus = append(menus, map[string]interface{}{"path": "/audit-logs", "title": "审计日志", "icon": "Document"})
	}

	return menus
}

func SeedDefaultAdmin(db *gorm.DB) error {
	modules := []struct {
		Module  string
		Name    string
		Actions []string
	}{
		{"home", "首页", []string{"read"}},
		{"person", "人员管理", []string{"read", "write", "delete", "export"}},
		{"company", "公司管理", []string{"read", "write", "delete", "export"}},
		{"position_event", "职务事件", []string{"read", "write", "delete", "export"}},
		{"attendance", "考勤管理", []string{"read", "write", "delete", "export"}},
		{"annual_leave", "年假管理", []string{"read", "write", "delete", "export"}},
		{"salary", "工资管理", []string{"read", "write", "delete", "export"}},
		{"file", "文件管理", []string{"read", "write", "delete", "export"}},
		{"audit", "审计日志", []string{"read", "export"}},
		{"user", "用户管理", []string{"read", "write", "delete", "export"}},
		{"role", "角色管理", []string{"read", "write", "delete", "export"}},
		{"system_config", "系统配置", []string{"read", "write"}},
	}

	var adminRole model.Role
	result := db.Where("name = ? AND is_default = ?", "超级管理员", true).First(&adminRole)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		adminRole = model.Role{
			Name:      "超级管理员",
			Remark:    "系统默认超级管理员角色，不可删除",
			IsDefault: true,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			return err
		}
	}

	var permCount int64
	db.Model(&model.Permission{}).Count(&permCount)
	if permCount == 0 {
		for _, mod := range modules {
			for _, action := range mod.Actions {
				actionNames := map[string]string{
					"read": "查看", "write": "编辑", "delete": "删除", "export": "导出",
				}
				p := model.Permission{
					Module: mod.Module,
					Action: action,
					Name:   mod.Name + actionNames[action],
				}
				db.Create(&p)

				rp := model.RolePermission{RoleID: adminRole.ID, PermissionID: p.ID}
				db.Create(&rp)
			}
		}
	} else {
		var allPerms []model.Permission
		db.Find(&allPerms)
		for _, p := range allPerms {
			var count int64
			db.Model(&model.RolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, p.ID).Count(&count)
			if count == 0 {
				db.Create(&model.RolePermission{RoleID: adminRole.ID, PermissionID: p.ID})
			}
		}
	}

	hash, err := utils.HashPassword(DefaultPassword)
	if err != nil {
		return err
	}

	var adminUser model.User
	result = db.Where("username = ?", "admin").First(&adminUser)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		adminUser = model.User{
			Username:     "admin",
			Password:     hash,
			IsActive:     true,
			IsFirstLogin: true,
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}

		userRole := model.UserRole{UserID: adminUser.ID, RoleID: adminRole.ID}
		db.Create(&userRole)
	} else {
		var urCount int64
		db.Model(&model.UserRole{}).Where("user_id = ? AND role_id = ?", adminUser.ID, adminRole.ID).Count(&urCount)
		if urCount == 0 {
			db.Create(&model.UserRole{UserID: adminUser.ID, RoleID: adminRole.ID})
		}
	}

	return nil
}
