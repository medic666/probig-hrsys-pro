package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"probig/internal/attendance"
	"probig/internal/auditlog"
	"probig/internal/company"
	"probig/internal/file"
	"probig/internal/leave_account"
	"probig/internal/person"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
	"probig/internal/pkg/middleware"
	"probig/internal/pkg/response"
	"probig/internal/position"
	"probig/internal/rbac"
	"probig/internal/salary"
	"probig/internal/system"
)

//go:embed all:web/dist
var webDistFS embed.FS

func main() {
	dbPath := getEnv("DB_PATH", "./hr.db")
	port := getEnv("SERVER_PORT", "8080")
	_ = getEnv("JWT_SECRET", "")
	_ = getEnv("FILE_STORAGE_PATH", "./upload")
	_ = getEnv("ENCRYPT_KEY", "")

	if err := os.MkdirAll(getEnv("FILE_STORAGE_PATH", "./upload"), 0755); err != nil {
		log.Printf("failed to create upload directory: %v", err)
	}

	_, err := database.Init(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	log.Println("database initialized successfully")

	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize config cache: %v", err)
	}
	log.Println("config cache initialized successfully")

	seedDefaultData()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.AuthMiddleware())

	api := r.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	rbac.RegisterRoutes(api)
	person.RegisterRoutes(api)
	company.RegisterRoutes(api)
	position.RegisterRoutes(api)
	attendance.RegisterRoutes(api)
	leave_account.RegisterRoutes(api)

	salary.RegisterRoutes(api)
	file.RegisterRoutes(api)
	auditlog.RegisterRoutes(api)
	system.RegisterRoutes(api)

	attendance.OnLeaveAccountRecalc = leave_account.RebuildBalanceForPerson

	distFS, err := fs.Sub(webDistFS, "web/dist")
	if err != nil {
		log.Printf("warning: web/dist not found in embedded FS, static file serving disabled: %v", err)
	} else {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api") {
				c.JSON(http.StatusNotFound, response.Response{
					Code: response.NotFound,
					Msg:  "接口不存在",
					Data: nil,
				})
				return
			}
			c.FileFromFS(path, http.FS(distFS))
		})

		if indexFile, err := distFS.Open("index.html"); err == nil {
			indexFile.Close()
			r.GET("/", func(c *gin.Context) {
				c.FileFromFS("index.html", http.FS(distFS))
			})
		}
	}

	log.Printf("server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func seedDefaultData() {
	seedSuperAdminRole()
	seedDefaultPermissions()
	seedDefaultAdmin()
	seedDefaultConfigs()
}

func seedSuperAdminRole() {
	var count int64
	database.DB.Model(&database.Role{}).Where("name = ?", "超级管理员").Count(&count)
	if count > 0 {
		return
	}

	role := database.Role{
		Name:   "超级管理员",
		Remark: "系统内置超级管理员角色，禁止修改权限、禁止删除",
	}
	if err := database.DB.Create(&role).Error; err != nil {
		log.Printf("failed to seed super admin role: %v", err)
	}
}

func seedDefaultPermissions() {
	modules := []struct {
		module  string
		actions []struct {
			key  string
			name string
		}
	}{
		{
			module: "person",
			actions: []struct{ key, name string }{
				{"person:read", "人员查看"},
				{"person:write", "人员编辑"},
				{"person:delete", "人员删除"},
				{"person:export", "人员导出"},
			},
		},
		{
			module: "company",
			actions: []struct{ key, name string }{
				{"company:read", "公司查看"},
				{"company:write", "公司编辑"},
				{"company:delete", "公司删除"},
				{"company:export", "公司导出"},
			},
		},
		{
			module: "position",
			actions: []struct{ key, name string }{
				{"position:read", "职务事件查看"},
				{"position:write", "职务事件编辑"},
				{"position:delete", "职务事件删除"},
				{"position:export", "职务事件导出"},
			},
		},
		{
			module: "attendance",
			actions: []struct{ key, name string }{
				{"attendance:read", "考勤查看"},
				{"attendance:write", "考勤编辑"},
				{"attendance:delete", "考勤删除"},
				{"attendance:export", "考勤导出"},
				{"attendance:calc", "考勤工资核算"},
			},
		},
		{
			module: "leave_account",
			actions: []struct{ key, name string }{
				{"leave_account:read", "假期账户查看"},
				{"leave_account:write", "假期账户编辑"},
				{"leave_account:delete", "假期账户删除"},
				{"leave_account:export", "假期账户导出"},
				{"leave_account:carryover", "假期结转操作"},
			},
		},
		{
			module: "salary",
			actions: []struct{ key, name string }{
				{"salary:read", "工资查看"},
				{"salary:write", "工资编辑"},
				{"salary:delete", "工资删除"},
				{"salary:export", "工资导出"},
				{"salary:calc", "工资核算"},
			},
		},
		{
			module: "file",
			actions: []struct{ key, name string }{
				{"file:read", "文件查看"},
				{"file:write", "文件上传"},
				{"file:delete", "文件删除"},
				{"file:export", "文件导出"},
			},
		},
		{
			module: "audit",
			actions: []struct{ key, name string }{
				{"audit:read", "审计查看"},
				{"audit:export", "审计导出"},
			},
		},
		{
			module: "rbac",
			actions: []struct{ key, name string }{
				{"rbac:read", "权限查看"},
				{"rbac:write", "权限编辑"},
				{"rbac:delete", "权限删除"},
			},
		},
		{
			module: "system",
			actions: []struct{ key, name string }{
				{"system:read", "系统配置查看"},
				{"system:write", "系统配置编辑"},
			},
		},
	}

	for _, m := range modules {
		for _, a := range m.actions {
			var count int64
			database.DB.Model(&database.Permission{}).Where("perm_key = ?", a.key).Count(&count)
			if count > 0 {
				continue
			}
			perm := database.Permission{
				Module:   m.module,
				Action:   strings.TrimPrefix(a.key, m.module+":"),
				PermKey:  a.key,
				PermName: a.name,
			}
			database.DB.Create(&perm)
		}
	}

	var role database.Role
	if database.DB.Where("name = ?", "超级管理员").First(&role).Error != nil {
		return
	}

	var permissions []database.Permission
	database.DB.Find(&permissions)
	for _, p := range permissions {
		var count int64
		database.DB.Model(&database.RolePermission{}).Where("role_id = ? AND permission_id = ?", role.ID, p.ID).Count(&count)
		if count == 0 {
			database.DB.Create(&database.RolePermission{
				RoleID:       role.ID,
				PermissionID: p.ID,
			})
		}
	}
}

func seedDefaultAdmin() {
	var count int64
	database.DB.Model(&database.User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash admin password: %v", err)
		return
	}

	user := database.User{
		Username:     "admin",
		Password:     string(hashedPassword),
		IsFirstLogin: true,
		Status:       1,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		log.Printf("failed to create admin user: %v", err)
		return
	}

	var role database.Role
	if database.DB.Where("name = ?", "超级管理员").First(&role).Error != nil {
		return
	}

	database.DB.Create(&database.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	})
}

func seedDefaultConfigs() {
	defaults := []database.SysConfig{
		{ConfigKey: "system.page_size", ConfigValue: "20", ConfigName: "分页默认条数", ConfigDesc: "列表查询默认分页条数", ValueType: "number"},
		{ConfigKey: "system.export_max", ConfigValue: "10000", ConfigName: "导出最大条数", ConfigDesc: "单次导出最大记录数", ValueType: "number"},
		{ConfigKey: "system.file_max_size", ConfigValue: "50", ConfigName: "文件上传大小限制(MB)", ConfigDesc: "文件上传最大限制，单位MB", ValueType: "number"},
		{ConfigKey: "system.file_storage_path", ConfigValue: "./upload", ConfigName: "文件存储根路径", ConfigDesc: "文件上传存储的根目录路径", ValueType: "string"},
		{ConfigKey: "system.work_hours_per_day", ConfigValue: "8", ConfigName: "计薪小时基准", ConfigDesc: "每天标准计薪小时数", ValueType: "number"},
		{ConfigKey: "attendance.sick_leave_ratio", ConfigValue: "0.6", ConfigName: "病假系数", ConfigDesc: "病假工时计入出勤的比例", ValueType: "number"},
		{ConfigKey: "attendance.overtime_workday_ratio", ConfigValue: "1.5", ConfigName: "工作日加班系数", ConfigDesc: "工作日加班工资计算系数", ValueType: "number"},
		{ConfigKey: "attendance.overtime_holiday_ratio", ConfigValue: "3.0", ConfigName: "节假日加班系数", ConfigDesc: "节假日加班工资计算系数", ValueType: "number"},
		{ConfigKey: "attendance.bonus_daily_standard", ConfigValue: "20", ConfigName: "全勤奖日标准", ConfigDesc: "全勤奖每日标准金额", ValueType: "number"},
		{ConfigKey: "attendance.high_temp_months", ConfigValue: `["06","07","08","09"]`, ConfigName: "高温补贴发放月份", ConfigDesc: "高温补贴发放月份列表", ValueType: "select"},
		{ConfigKey: "attendance.min_leave_unit", ConfigValue: "0.5", ConfigName: "最小请假单位(小时)", ConfigDesc: "最小请假时长单位", ValueType: "number"},
		{ConfigKey: "attendance.special_approval", ConfigValue: "true", ConfigName: "假期特批开关", ConfigDesc: "是否允许假期额度不足时特批", ValueType: "bool"},
		{ConfigKey: "annual_leave.yearly_hours", ConfigValue: "40", ConfigName: "年假年度额度(小时)", ConfigDesc: "每周年配发的年假标准额度", ValueType: "number"},
		{ConfigKey: "annual_leave.cycle_rule", ConfigValue: "entry_anniversary", ConfigName: "年假周期规则", ConfigDesc: "年假周期计算规则：entry_anniversary=入职周年，natural_year=自然年", ValueType: "select"},
	}

	for _, cfg := range defaults {
		var count int64
		database.DB.Model(&database.SysConfig{}).Where("config_key = ?", cfg.ConfigKey).Count(&count)
		if count > 0 {
			continue
		}
		if err := database.DB.Create(&cfg).Error; err != nil {
			log.Printf("failed to seed config %s: %v", cfg.ConfigKey, err)
		}
	}
}
