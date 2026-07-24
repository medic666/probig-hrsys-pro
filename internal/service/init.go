package service

import (
	"log"

	"probig/internal/dao"
	"probig/internal/models"
	"probig/pkg/crypto"
)

var ConfigCache = make(map[string]string)

func InitSystem() error {
	if err := initSysConfig(); err != nil {
		return err
	}
	if err := initPermissions(); err != nil {
		return err
	}
	if err := initAdmin(); err != nil {
		return err
	}
	log.Println("system initialized successfully")
	return nil
}

func initSysConfig() error {
	configs := []models.SysConfig{
		{ConfigKey: "attendance.overtime_control", ConfigValue: "true", ConfigName: "调休管控模式", ConfigDesc: "开启后加班累计调休额度，提交调休需校验余额", ValueType: "bool"},
		{ConfigKey: "attendance.special_approval", ConfigValue: "true", ConfigName: "假期特批开关", ConfigDesc: "开启后年假/调休额度不足时可特批绕过校验", ValueType: "bool"},
		{ConfigKey: "attendance.min_leave_unit", ConfigValue: "0.5", ConfigName: "最小请假单位(小时)", ConfigDesc: "请假最小时间单位", ValueType: "number"},
		{ConfigKey: "salary.sick_leave_ratio", ConfigValue: "0.6", ConfigName: "病假工资系数", ConfigDesc: "病假期间工资发放比例", ValueType: "number"},
		{ConfigKey: "salary.workday_overtime_ratio", ConfigValue: "1.5", ConfigName: "工作日加班系数", ConfigDesc: "工作日加班工资倍数", ValueType: "number"},
		{ConfigKey: "salary.holiday_overtime_ratio", ConfigValue: "2.0", ConfigName: "节假日加班系数", ConfigDesc: "节假日加班工资倍数", ValueType: "number"},
		{ConfigKey: "salary.attendance_bonus_ratio", ConfigValue: "50", ConfigName: "全勤奖系数", ConfigDesc: "全勤奖 = (出勤计薪天数-违纪次数) × 此系数", ValueType: "number"},
		{ConfigKey: "salary.high_temp_months", ConfigValue: `["6","7","8","9"]`, ConfigName: "高温补贴发放月份", ConfigDesc: "高温补贴发放的月份列表", ValueType: "select", OptionValues: `["6","7","8","9"]`},
		{ConfigKey: "salary.default_annual_leave", ConfigValue: "5", ConfigName: "默认年假额度(天)", ConfigDesc: "员工默认享有的年假天数", ValueType: "number"},
		{ConfigKey: "general.page_size", ConfigValue: "20", ConfigName: "分页默认条数", ConfigDesc: "列表页默认每页显示条数", ValueType: "number"},
		{ConfigKey: "general.export_max", ConfigValue: "10000", ConfigName: "导出最大条数", ConfigDesc: "单次导出最大数据条数限制", ValueType: "number"},
		{ConfigKey: "general.upload_max_size", ConfigValue: "10", ConfigName: "文件上传大小限制(MB)", ConfigDesc: "单文件上传最大大小（MB）", ValueType: "number"},
	}
	for _, c := range configs {
		if err := dao.UpsertSysConfig(&c); err != nil {
			return err
		}
	}
	list, err := dao.GetAllSysConfigs()
	if err != nil {
		return err
	}
	for _, c := range list {
		ConfigCache[c.ConfigKey] = c.ConfigValue
	}
	log.Printf("loaded %d system configs", len(list))
	return nil
}

func initPermissions() error {
	modules := []struct {
		module string
		name   string
	}{
		{"person", "人员管理"},
		{"company", "公司管理"},
		{"attendance", "假勤管理"},
		{"salary", "工资管理"},
		{"file", "文件管理"},
		{"audit", "操作审计"},
		{"user", "用户管理"},
		{"system", "系统配置"},
	}
	actions := []struct {
		action string
		name   string
	}{
		{"read", "查看"},
		{"write", "编辑"},
		{"delete", "删除"},
		{"export", "导出"},
	}
	for _, m := range modules {
		for _, a := range actions {
			perm := models.Permission{
				PermissionKey: m.module + "." + a.action,
				Name:          m.name + a.name,
				Module:        m.module,
				Action:        a.action,
			}
			dao.DB().Where(models.Permission{PermissionKey: perm.PermissionKey}).
				FirstOrCreate(&perm)
		}
	}
	return nil
}

func initAdmin() error {
	_, err := dao.GetUserByUsername("admin")
	if err == nil {
		return nil
	}
	hash, err := crypto.HashPassword("admin123")
	if err != nil {
		return err
	}
	user := &models.User{
		Username:     "admin",
		PasswordHash: hash,
		IsFirstLogin: true,
		Status:       1,
	}
	if err := dao.CreateUser(user); err != nil {
		return err
	}
	role := &models.Role{
		Name:   "超级管理员",
		Remark: "系统默认超级管理员角色，拥有全部权限",
	}
	if err := dao.CreateRole(role); err != nil {
		return err
	}
	perms, err := dao.GetAllPermissions()
	if err != nil {
		return err
	}
	role.Permissions = perms
	if err := dao.UpdateRole(role); err != nil {
		return err
	}
	ur := models.UserRole{UserID: user.ID, RoleID: role.ID}
	dao.DB().Create(&ur)
	log.Println("admin account created: admin/admin123")
	return nil
}

func RefreshConfigCache() error {
	ConfigCache = make(map[string]string)
	list, err := dao.GetAllSysConfigs()
	if err != nil {
		return err
	}
	for _, c := range list {
		ConfigCache[c.ConfigKey] = c.ConfigValue
	}
	return nil
}
