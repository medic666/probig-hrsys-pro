package service

import (
	"context"
	"errors"
	"fmt"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

const (
	// DefaultPassword 非 admin 用户新建/重置的默认密码（首登强制改密）
	DefaultPassword = "123456"
	// AdminInitialPassword 超级管理员初始/重置密码（仅 admin 使用，首登强制改密）
	AdminInitialPassword = "admin123"
)

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

	CleanExpiredTokens()

	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", nil, fmt.Errorf("生成Token失败")
	}

	return token, &user, nil
}

func ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	var user model.User
	if err := dao.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if !user.IsFirstLogin {
		if !utils.CheckPassword(oldPassword, user.Password) {
			return errors.New("原密码错误")
		}
	}

	// 防绕过：新密码不允许为任何默认密码（首登强制改密才有效）
	if newPassword == DefaultPassword || newPassword == AdminInitialPassword {
		return errors.New("新密码不能与默认密码相同")
	}

	return setUserPassword(ctx, &user, newPassword)
}

// ResetPassword 管理员重置密码：admin 重置回 AdminInitialPassword，其它用户重置回
// DefaultPassword；重置后置首登强制改密（IsFirstLogin=true），任何用户重置后首次登录必须改密。
func ResetPassword(ctx context.Context, userID uint) (string, error) {
	var user model.User
	if err := dao.DB.First(&user, userID).Error; err != nil {
		return "", errors.New("用户不存在")
	}

	newPwd := DefaultPassword
	if user.ID == 1 {
		newPwd = AdminInitialPassword
	}

	hash, err := utils.HashPassword(newPwd)
	if err != nil {
		return "", err
	}
	user.Password = hash
	user.IsFirstLogin = true
	if err := dao.DBFromContext(ctx).Save(&user).Error; err != nil {
		return "", err
	}
	return newPwd, nil
}

func setUserPassword(ctx context.Context, user *model.User, password string) error {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	user.Password = hash
	user.IsFirstLogin = false
	return dao.DBFromContext(ctx).Save(user).Error
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

	// 主数据管理
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

	// 考勤管理
	if has("attendance_event.read") || has("attendance_daily.read") || has("attendance_monthly.read") {
		children := []map[string]interface{}{}
		if has("attendance_event.read") {
			children = append(children, map[string]interface{}{"path": "/attendance", "title": "考勤事件", "icon": "Clock"})
		}
		if has("attendance_daily.read") {
			children = append(children, map[string]interface{}{"path": "/attendance-daily", "title": "日记工时", "icon": "Clock"})
		}
		if has("attendance_monthly.read") {
			children = append(children, map[string]interface{}{"path": "/attendance-monthly", "title": "月度考勤核算", "icon": "Clock"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/attendance-group", "title": "考勤管理", "icon": "Clock", "children": children,
		})
	}

	// 假期管理
	if has("annual_leave_event.read") || has("leave_in_lieu.read") || has("annual_leave_carryover.read") {
		children := []map[string]interface{}{}
		if has("annual_leave_event.read") {
			children = append(children, map[string]interface{}{"path": "/annual-leave-events", "title": "年假事件", "icon": "Clock"})
		}
		if has("leave_in_lieu.read") {
			children = append(children, map[string]interface{}{"path": "/lil-events", "title": "调休事件", "icon": "Clock"})
		}
		if has("annual_leave_carryover.read") {
			children = append(children, map[string]interface{}{"path": "/annual-leave-carryover", "title": "年假配发结转", "icon": "Clock"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/leave-group", "title": "假期管理", "icon": "Clock", "children": children,
		})
	}

	// 薪资管理
	if has("salary_event.read") || has("salary_summary.read") {
		children := []map[string]interface{}{}
		if has("salary_event.read") {
			children = append(children, map[string]interface{}{"path": "/salary-events", "title": "工资事件", "icon": "Money"})
		}
		if has("salary_summary.read") {
			children = append(children, map[string]interface{}{"path": "/salary-summaries", "title": "月度工资汇总", "icon": "Money"})
		}
		menus = append(menus, map[string]interface{}{
			"path": "/salary-group", "title": "薪资管理", "icon": "Money", "children": children,
		})
	}

	// 系统管理
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
	// 权限行已由 SyncPermissionRows 保证存在（启动链路先于本函数执行）；
	// 本函数职责 = 确保默认角色存在 + admin 用户存在 + admin 全量挂接权限

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

	// 超级管理员默认包含全部权限：对现存权限行逐条幂等补挂
	var allPerms []model.Permission
	if err := db.Find(&allPerms).Error; err != nil {
		return err
	}
	for _, p := range allPerms {
		var count int64
		db.Model(&model.RolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, p.ID).Count(&count)
		if count == 0 {
			if err := db.Create(&model.RolePermission{RoleID: adminRole.ID, PermissionID: p.ID}).Error; err != nil {
				return err
			}
		}
	}

	hash, err := utils.HashPassword(AdminInitialPassword)
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
