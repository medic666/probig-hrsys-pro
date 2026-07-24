package database

import (
	"probig/internal/attendance"
	"probig/internal/audit_log"
	"probig/internal/company"
	"probig/internal/file"
	"probig/internal/person"
	"probig/internal/position"
	"probig/internal/rbac"
	"probig/internal/salary"
	"probig/internal/system"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	DB = db
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&rbac.User{},
		&rbac.Role{},
		&rbac.Permission{},
		&rbac.UserRole{},
		&rbac.RolePermission{},
		&person.Person{},
		&person.PersonPhone{},
		&person.PersonEmail{},
		&person.PersonBankCard{},
		&company.Company{},
		&position.PositionEvent{},
		&position.PositionSnapshot{},
		&attendance.AttendanceEvent{},
		&attendance.AttendanceSummary{},
		&salary.SalaryEvent{},
		&salary.SalarySummary{},
		&file.File{},
		&file.FileRelation{},
		&audit_log.AuditLog{},
		&system.SysConfig{},
	)
}

func SeedDefaultData(db *gorm.DB) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var existingUser rbac.User
	if err := db.Where("username = ?", "admin").First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			adminUser := rbac.User{
				Username:      "admin",
				PasswordHash:  string(passwordHash),
				IsFirstLogin:  true,
				Status:        1,
			}
			if err := db.Create(&adminUser).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var adminRole rbac.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			adminRole = rbac.Role{
				Name:   "admin",
				Remark: "Super Administrator",
			}
			if err := db.Create(&adminRole).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	modules := []string{"person", "company", "position", "attendance", "salary", "file", "audit", "system", "rbac"}
	actions := []string{"read", "write", "delete", "export"}

	for _, module := range modules {
		for _, action := range actions {
			var perm rbac.Permission
			key := module + "." + action
			if err := db.Where("permission_key = ?", key).First(&perm).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					perm = rbac.Permission{
						PermissionKey: key,
						Name:          key,
						Module:        module,
						Action:        action,
					}
					if err := db.Create(&perm).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	}

	var permissions []rbac.Permission
	if err := db.Find(&permissions).Error; err != nil {
		return err
	}
	for _, perm := range permissions {
		var rp rbac.RolePermission
		if err := db.Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).First(&rp).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				rp = rbac.RolePermission{
					RoleID:       adminRole.ID,
					PermissionID: perm.ID,
				}
				if err := db.Create(&rp).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	db.First(&existingUser, "username = ?", "admin")
	var ur rbac.UserRole
	if err := db.Where("user_id = ? AND role_id = ?", existingUser.ID, adminRole.ID).First(&ur).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ur = rbac.UserRole{
				UserID: existingUser.ID,
				RoleID: adminRole.ID,
			}
			if err := db.Create(&ur).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	defaultConfigs := []system.SysConfig{
		{ConfigKey: "attendance.comp_time_control", ConfigValue: "true", ConfigName: "调休管控模式", ConfigDesc: "是否开启调休管控模式", ValueType: "bool"},
		{ConfigKey: "attendance.special_approval", ConfigValue: "true", ConfigName: "假期特批开关", ConfigDesc: "是否允许特批假期", ValueType: "bool"},
		{ConfigKey: "attendance.min_leave_unit", ConfigValue: "0.5", ConfigName: "最小请假单位", ConfigDesc: "最小请假单位（天）", ValueType: "number"},
		{ConfigKey: "salary.sick_leave_rate", ConfigValue: "0.6", ConfigName: "病假系数", ConfigDesc: "病假工资发放比例", ValueType: "number"},
		{ConfigKey: "salary.overtime_workday_rate", ConfigValue: "1.5", ConfigName: "工作日加班系数", ConfigDesc: "工作日加班工资倍数", ValueType: "number"},
		{ConfigKey: "salary.overtime_holiday_rate", ConfigValue: "2.0", ConfigName: "节假日加班系数", ConfigDesc: "节假日加班工资倍数", ValueType: "number"},
		{ConfigKey: "salary.attendance_bonus_rate", ConfigValue: "200", ConfigName: "全勤奖系数", ConfigDesc: "全勤奖金额系数", ValueType: "number"},
		{ConfigKey: "salary.high_temp_months", ConfigValue: `["6","7","8","9"]`, ConfigName: "高温补贴发放月份", ConfigDesc: "高温补贴发放的月份", ValueType: "select"},
		{ConfigKey: "salary.default_annual_leave", ConfigValue: "5", ConfigName: "默认年假额度", ConfigDesc: "员工默认年假天数", ValueType: "number"},
		{ConfigKey: "general.page_size", ConfigValue: "20", ConfigName: "分页默认条数", ConfigDesc: "列表分页默认每页条数", ValueType: "number"},
		{ConfigKey: "general.export_max", ConfigValue: "10000", ConfigName: "导出最大条数", ConfigDesc: "单次导出最大记录数", ValueType: "number"},
		{ConfigKey: "general.upload_max_size", ConfigValue: "10", ConfigName: "文件上传大小限制(MB)", ConfigDesc: "单文件上传最大大小（MB）", ValueType: "number"},
	}

	for _, cfg := range defaultConfigs {
		var existing system.SysConfig
		if err := db.Where("config_key = ?", cfg.ConfigKey).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&cfg).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
