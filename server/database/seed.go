package database

import (
	"encoding/json"
	"time"

	"probig/config"
	"probig/models"
	"probig/utils"

	"golang.org/x/crypto/bcrypt"
)

func seedAll() {
	var count int64
	DB.Model(&models.Permission{}).Count(&count)
	if count > 0 {
		return
	}

	tx := DB.Begin()

	modules := []string{"person", "company", "attendance", "salary", "file", "audit", "system", "user"}
	actions := []string{"read", "write", "delete", "export"}
	var perms []models.Permission
	for _, m := range modules {
		for _, a := range actions {
			perms = append(perms, models.Permission{
				PermissionKey: m + "." + a,
				Name:          moduleName(m) + actionName(a),
				Module:        m,
				Action:        a,
			})
		}
	}
	if err := tx.Create(&perms).Error; err != nil {
		tx.Rollback()
		panic("seed permissions failed: " + err.Error())
	}

	adminRole := models.Role{Name: "超级管理员", Remark: "系统超级管理员，拥有全部权限"}
	if err := tx.Create(&adminRole).Error; err != nil {
		tx.Rollback()
		panic("seed role failed: " + err.Error())
	}

	for _, p := range perms {
		if err := tx.Create(&models.RolePermission{RoleID: adminRole.ID, PermissionID: p.ID}).Error; err != nil {
			tx.Rollback()
			panic("seed role_perm failed: " + err.Error())
		}
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := models.User{
		Username:     "admin",
		PasswordHash: string(hash),
		IsFirstLogin: true,
		Status:       1,
	}
	if err := tx.Create(&adminUser).Error; err != nil {
		tx.Rollback()
		panic("seed user failed: " + err.Error())
	}

	if err := tx.Create(&models.UserRole{UserID: adminUser.ID, RoleID: adminRole.ID}).Error; err != nil {
		tx.Rollback()
		panic("seed user_role failed: " + err.Error())
	}

	// encrypt default admin person
	encCard, _ := utils.Encrypt("000000000000000000", config.EncryptKeyBytes)
	encPhone, _ := utils.Encrypt("13800000000", config.EncryptKeyBytes)
	encEmail, _ := utils.Encrypt("admin@example.com", config.EncryptKeyBytes)
	encBank, _ := utils.Encrypt("0000000000000000", config.EncryptKeyBytes)
	birthday := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	adminPerson := models.Person{
		Name:     "系统管理员",
		IdCard:   encCard,
		Birthday: &birthday,
		Phones:   []models.PersonPhone{{Phone: encPhone, Remark: "工作号"}},
		Emails:   []models.PersonEmail{{Email: encEmail, Remark: "工作邮箱"}},
		BankCards: []models.PersonBankCard{{BankCard: encBank, BankName: "测试银行", Remark: "工资卡"}},
	}
	if err := tx.Create(&adminPerson).Error; err != nil {
		tx.Rollback()
		panic("seed person failed: " + err.Error())
	}
	tx.Model(&adminUser).Update("person_id", adminPerson.ID)

	configs := defaultConfigs()
	for _, c := range configs {
		if err := tx.Create(&c).Error; err != nil {
			tx.Rollback()
			panic("seed config failed: " + err.Error())
		}
	}

	tx.Commit()
}

func moduleName(m string) string {
	names := map[string]string{
		"person": "人员管理", "company": "公司管理", "attendance": "假勤管理",
		"salary": "工资管理", "file": "文件管理", "audit": "审计日志",
		"system": "系统配置", "user": "用户管理",
	}
	return names[m]
}

func actionName(a string) string {
	names := map[string]string{"read": "查看", "write": "编辑", "delete": "删除", "export": "导出"}
	return names[a]
}

func defaultConfigs() []models.SysConfig {
	return []models.SysConfig{
		{ConfigKey: "attendance.overtime_control", ConfigValue: "true", ConfigName: "调休管控模式", ConfigDesc: "开启后使用加班时长累计调休额度，关闭后调休不校验额度", ValueType: "bool"},
		{ConfigKey: "attendance.leave_special_approval", ConfigValue: "true", ConfigName: "假期特批开关", ConfigDesc: "开启后年假/调休额度不足时支持特批", ValueType: "bool"},
		{ConfigKey: "attendance.leave_min_unit", ConfigValue: "0.5", ConfigName: "最小请假单位", ConfigDesc: "最小请假单位（小时）", ValueType: "number"},
		{ConfigKey: "attendance.default_annual_leave", ConfigValue: "5", ConfigName: "默认年假额度（天）", ConfigDesc: "入职满周年自动配发的默认年假天数", ValueType: "number"},
		{ConfigKey: "salary.sick_leave_ratio", ConfigValue: "0.6", ConfigName: "病假工资系数", ConfigDesc: "病假出勤工资折算系数", ValueType: "number"},
		{ConfigKey: "salary.workday_overtime_ratio", ConfigValue: "1.5", ConfigName: "工作日加班系数", ConfigDesc: "工作日加班工资倍率", ValueType: "number"},
		{ConfigKey: "salary.holiday_overtime_ratio", ConfigValue: "2.0", ConfigName: "节假日加班系数", ConfigDesc: "节假日加班工资倍率", ValueType: "number"},
		{ConfigKey: "salary.attendance_bonus_ratio", ConfigValue: "50", ConfigName: "全勤奖系数", ConfigDesc: "全勤奖 = (计薪天数 - 违纪次数) × 此系数", ValueType: "number"},
		{ConfigKey: "salary.high_temp_months", ConfigValue: `["06","07","08","09"]`, ConfigName: "高温补贴发放月份", ConfigDesc: "高温补贴仅在配置月份发放", ValueType: "select", OptionValues: `["06","07","08","09"]`},
		{ConfigKey: "general.page_size", ConfigValue: "20", ConfigName: "分页默认条数", ConfigDesc: "列表页默认每页显示条数", ValueType: "number"},
		{ConfigKey: "general.export_max", ConfigValue: "10000", ConfigName: "导出最大条数", ConfigDesc: "导出Excel的最大记录数", ValueType: "number"},
		{ConfigKey: "general.upload_max_size", ConfigValue: "10485760", ConfigName: "文件上传大小限制(字节)", ConfigDesc: "文件上传大小限制，默认10M", ValueType: "number"},
	}
}

func LoadConfigCache() map[string]string {
	var configs []models.SysConfig
	DB.Find(&configs)
	cache := make(map[string]string)
	v, _ := json.Marshal(configs)
	_ = v
	for _, c := range configs {
		cache[c.ConfigKey] = c.ConfigValue
	}
	return cache
}
