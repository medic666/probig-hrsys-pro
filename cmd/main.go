package main

import (
	"fmt"
	"os"

	"probig/internal/pkg/audit"
	"probig/internal/pkg/batch"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
	"probig/internal/pkg/encrypt"
	"probig/internal/pkg/middleware"
	"probig/internal/attendance"
	"probig/internal/audit_log"
	"probig/internal/company"
	"probig/internal/file"
	"probig/internal/leave_account"
	"probig/internal/person"
	"probig/internal/position"
	"probig/internal/rbac"
	"probig/internal/salary"
	"probig/internal/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	config.LoadEnv()

	db := database.Init()

	autoMigrate(db)
	initDefaultData(db)
	database.InitEncryptKey(db)

	config.DB = db
	config.LoadConfigCache(db)

	attendance.RebuildLeaveBalanceCallback = leave_account.NewService().RebuildBalance

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api")
	{
		rbac.RegisterRoutes(api)
		person.RegisterRoutes(api.Group("/persons"))
		company.RegisterRoutes(api.Group("/companies"))
		position.RegisterRoutes(api.Group("/positions"))
		attendance.RegisterRoutes(api.Group("/attendances"))
		leave_account.RegisterRoutes(api.Group("/leave-accounts"))
		salary.RegisterRoutes(api.Group("/salaries"))
		file.RegisterRoutes(api.Group("/files"))
		audit_log.RegisterRoutes(api.Group("/audits"))
		system.RegisterRoutes(api.Group("/configs"))
	}

	serveStatic(r)

	port := config.ServerPort
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&rbac.Permission{},
		&rbac.Role{},
		&rbac.User{},
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
		&attendance.AttendanceDaily{},
		&attendance.AttendanceSalary{},
		&leave_account.LeaveAccountEvent{},
		&leave_account.LeaveAccountBalance{},
		&salary.SalaryEvent{},
		&salary.SalarySummary{},
		&file.FileModel{},
		&file.FileRelation{},
		&audit.AuditLog{},
		&system.SysConfig{},
		&batch.SysBatch{},
	)
}

func initDefaultData(db *gorm.DB) {
	var count int64
	db.Model(&rbac.User{}).Count(&count)
	if count > 0 {
		return
	}

	hash, _ := encrypt.HashPassword("admin123")

	superAdminRole := &rbac.Role{
		Name:    "超级管理员",
		Remark:  "系统超级管理员，拥有所有权限",
		IsAdmin: true,
	}
	db.Create(superAdminRole)

	adminUser := &rbac.User{
		Username:     "admin",
		Password:     hash,
		PersonID:     0,
		Status:       1,
		IsFirstLogin: true,
	}
	db.Create(adminUser)

	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", adminUser.ID, superAdminRole.ID)

	permissions := []rbac.Permission{
		{Module: "人员管理", PermKey: "person:read", PermName: "查看人员", Action: "read"},
		{Module: "人员管理", PermKey: "person:write", PermName: "编辑人员", Action: "write"},
		{Module: "人员管理", PermKey: "person:delete", PermName: "删除人员", Action: "delete"},
		{Module: "人员管理", PermKey: "person:export", PermName: "导出人员", Action: "export"},
		{Module: "公司管理", PermKey: "company:read", PermName: "查看公司", Action: "read"},
		{Module: "公司管理", PermKey: "company:write", PermName: "编辑公司", Action: "write"},
		{Module: "公司管理", PermKey: "company:delete", PermName: "删除公司", Action: "delete"},
		{Module: "公司管理", PermKey: "company:export", PermName: "导出公司", Action: "export"},
		{Module: "职务管理", PermKey: "position:read", PermName: "查看职务", Action: "read"},
		{Module: "职务管理", PermKey: "position:write", PermName: "编辑职务事件", Action: "write"},
		{Module: "职务管理", PermKey: "position:delete", PermName: "删除职务事件", Action: "delete"},
		{Module: "职务管理", PermKey: "position:export", PermName: "导出职务", Action: "export"},
		{Module: "考勤管理", PermKey: "attendance:read", PermName: "查看考勤", Action: "read"},
		{Module: "考勤管理", PermKey: "attendance:write", PermName: "编辑考勤事件", Action: "write"},
		{Module: "考勤管理", PermKey: "attendance:delete", PermName: "删除考勤事件", Action: "delete"},
		{Module: "考勤管理", PermKey: "attendance:export", PermName: "导出考勤", Action: "export"},
		{Module: "考勤管理", PermKey: "attendance:calc", PermName: "核算假勤工资", Action: "write"},
		{Module: "假期管理", PermKey: "leave:read", PermName: "查看假期", Action: "read"},
		{Module: "假期管理", PermKey: "leave:write", PermName: "编辑假期事件", Action: "write"},
		{Module: "假期管理", PermKey: "leave:delete", PermName: "删除假期事件", Action: "delete"},
		{Module: "假期管理", PermKey: "leave:export", PermName: "导出假期", Action: "export"},
		{Module: "假期管理", PermKey: "leave:carryover", PermName: "年假结转", Action: "write"},
		{Module: "工资管理", PermKey: "salary:read", PermName: "查看工资", Action: "read"},
		{Module: "工资管理", PermKey: "salary:write", PermName: "编辑工资事件", Action: "write"},
		{Module: "工资管理", PermKey: "salary:delete", PermName: "删除工资事件", Action: "delete"},
		{Module: "工资管理", PermKey: "salary:export", PermName: "导出工资", Action: "export"},
		{Module: "工资管理", PermKey: "salary:calc", PermName: "核算工资", Action: "write"},
		{Module: "文件管理", PermKey: "file:read", PermName: "查看文件", Action: "read"},
		{Module: "文件管理", PermKey: "file:write", PermName: "上传文件", Action: "write"},
		{Module: "文件管理", PermKey: "file:delete", PermName: "删除文件", Action: "delete"},
		{Module: "审计日志", PermKey: "audit:read", PermName: "查看审计", Action: "read"},
		{Module: "审计日志", PermKey: "audit:export", PermName: "导出审计", Action: "export"},
		{Module: "用户管理", PermKey: "rbac:read", PermName: "查看用户", Action: "read"},
		{Module: "用户管理", PermKey: "rbac:write", PermName: "编辑用户角色", Action: "write"},
		{Module: "系统配置", PermKey: "system:read", PermName: "查看配置", Action: "read"},
		{Module: "系统配置", PermKey: "system:write", PermName: "修改配置", Action: "write"},
	}
	for _, p := range permissions {
		db.Where("perm_key = ?", p.PermKey).FirstOrCreate(&p)
	}

	var allPerms []rbac.Permission
	db.Find(&allPerms)
	for _, p := range allPerms {
		db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", superAdminRole.ID, p.ID)
	}

	sysConfigs := []system.SysConfig{
		{ConfigKey: "system.page_size", ConfigValue: "20", ConfigName: "分页默认条数", ConfigDesc: "列表默认每页显示条数", ValueType: "number"},
		{ConfigKey: "system.export_max", ConfigValue: "10000", ConfigName: "导出最大条数", ConfigDesc: "单次导出最大数据条数", ValueType: "number"},
		{ConfigKey: "system.upload_limit", ConfigValue: "50", ConfigName: "文件上传大小限制(MB)", ConfigDesc: "单个文件上传大小限制", ValueType: "number"},
		{ConfigKey: "system.work_hours_per_day", ConfigValue: "8", ConfigName: "计薪小时基准", ConfigDesc: "每天标准工时", ValueType: "number"},
		{ConfigKey: "attendance.sick_leave_ratio", ConfigValue: "0.6", ConfigName: "病假系数", ConfigDesc: "病假记出勤工时的折算系数", ValueType: "number"},
		{ConfigKey: "attendance.overtime_workday_ratio", ConfigValue: "1.5", ConfigName: "工作日加班系数", ConfigDesc: "工作日加班工资折算系数", ValueType: "number"},
		{ConfigKey: "attendance.overtime_holiday_ratio", ConfigValue: "2.0", ConfigName: "节假日加班系数", ConfigDesc: "节假日加班工资折算系数", ValueType: "number"},
		{ConfigKey: "attendance.bonus_daily", ConfigValue: "50", ConfigName: "全勤奖日标准", ConfigDesc: "全勤奖每日标准金额", ValueType: "number"},
		{ConfigKey: "attendance.high_temp_months", ConfigValue: `["06","07","08","09"]`, ConfigName: "高温补贴发放月份", ConfigDesc: "高温补贴发放的月份列表", ValueType: "string"},
		{ConfigKey: "attendance.special_approval", ConfigValue: "true", ConfigName: "假期特批开关", ConfigDesc: "是否允许特批请假", ValueType: "bool"},
		{ConfigKey: "attendance.min_leave_unit", ConfigValue: "0.5", ConfigName: "最小请假单位", ConfigDesc: "最小请假时长单位（小时）", ValueType: "number"},
		{ConfigKey: "leave.annual_quota", ConfigValue: "40", ConfigName: "年假年度额度", ConfigDesc: "每年配发的年假标准额度（小时）", ValueType: "number"},
		{ConfigKey: "leave.cycle_rule", ConfigValue: "entry_anniversary", ConfigName: "年假周期规则", ConfigDesc: "入职周年/自然周年", ValueType: "select", OptionValues: `["entry_anniversary","calendar_year"]`},
	}
	for _, c := range sysConfigs {
		db.Where("config_key = ?", c.ConfigKey).FirstOrCreate(&c)
	}
}
