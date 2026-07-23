package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"probig/internal/asset"
	"probig/internal/attendance"
	"probig/internal/auth"
	"probig/internal/common"
	"probig/internal/event"
	"probig/internal/person"
	"probig/internal/policy"
	"probig/internal/salary"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

//go:embed frontend/dist/*
var frontendFiles embed.FS

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Admin    AdminConfig    `yaml:"admin"`
	Roles    []auth.RoleConfig `yaml:"roles"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Path        string `yaml:"path"`
	WAL         bool   `yaml:"wal"`
	Synchronous string `yaml:"synchronous"`
}

type JWTConfig struct {
	Secret       string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func main() {
	dataDir := flag.String("data", ".", "数据目录路径")
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	dbPath := config.Database.Path
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(*dataDir, dbPath)
	}
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	if err := common.InitDB(dbPath, config.Database.WAL); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer common.CloseDB()

	if err := common.RunMigrations(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	auth.JWTSecret = config.JWT.Secret
	auth.JWTExpireHours = config.JWT.ExpireHours
	auth.RolePermissions = make(map[string][]string)

	authRepo := auth.NewRepository()
	authService := auth.NewService(authRepo)

	for _, role := range config.Roles {
		auth.RolePermissions[role.Name] = role.Permissions
		if err := authRepo.SyncRolePermissions(role.Name, role.Permissions); err != nil {
			log.Printf("同步角色权限失败 [%s]: %v", role.Name, err)
		}
	}

	if err := auth.InitAdminUser(authRepo, config.Admin.Username, config.Admin.Password); err != nil {
		log.Fatalf("创建管理员失败: %v", err)
	}

	authHandler := auth.NewHandler(authService)

	eventRepo := event.NewRepository()
	eventService := event.NewService(eventRepo)
	eventHandler := event.NewHandler(eventService)

	personRepo := person.NewRepository()
	personService := person.NewService(personRepo, eventService)
	personHandler := person.NewHandler(personService)

	policyRepo := policy.NewRepository()
	policyService := policy.NewService(policyRepo, eventService)
	policyHandler := policy.NewHandler(policyService)

	attendanceRepo := attendance.NewRepository()
	attendanceService := attendance.NewService(attendanceRepo, eventService)
	attendanceHandler := attendance.NewHandler(attendanceService)

	salaryRepo := salary.NewRepository()
	salaryService := salary.NewService(salaryRepo, eventService)
	salaryHandler := salary.NewHandler(salaryService)

	assetRepo := asset.NewRepository()
	assetService := asset.NewService(assetRepo, eventService)
	assetHandler := asset.NewHandler(assetService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(corsMiddleware())

	api := r.Group("/api/v1")

	api.POST("/auth/login", authHandler.Login)

	authGroup := api.Group("")
	authGroup.Use(auth.AuthMiddleware())
	{
		authGroup.GET("/auth/me", authHandler.Me)
		authGroup.GET("/auth/menus", authHandler.GetMenus)
		authGroup.GET("/auth/permissions", authHandler.GetPermissions)

		authGroup.GET("/dashboard/stats", func(c *gin.Context) {
			var personCount, eventCount, assetCount int64
			common.DB.Get(&personCount, "SELECT COUNT(*) FROM persons")
			common.DB.Get(&eventCount, "SELECT COUNT(*) FROM events WHERE is_deleted = 0")
			common.DB.Get(&assetCount, "SELECT COUNT(*) FROM assets WHERE is_current = 1")
			common.Success(c, gin.H{
				"personCount": personCount,
				"eventCount":  eventCount,
				"assetCount":  assetCount,
			})
		})

		exports := authGroup.Group("/export")
		{
			exports.GET("/persons", exportPersons)
			exports.GET("/salary", exportSalary)
		}

		events := authGroup.Group("/events")
		{
			events.GET("", eventHandler.List)
			events.GET("/entity/:entityType/:entityId", eventHandler.GetEntityEvents)
			events.PUT("/:id/remark", eventHandler.UpdateRemark)
			events.DELETE("/:id", eventHandler.SoftDelete)
		}

		persons := authGroup.Group("/persons")
		{
			persons.GET("", personHandler.List)
			persons.GET("/all", personHandler.All)
			persons.POST("", personHandler.Create)
			persons.GET("/:id", personHandler.GetByID)
			persons.PUT("/:id", personHandler.Update)
			persons.DELETE("/:id", personHandler.Delete)
			persons.GET("/:id/timeline", personHandler.GetTimeline)
			persons.GET("/:id/snapshot", personHandler.GetSnapshot)
		}

		policies := authGroup.Group("/policies")
		{
			policies.GET("", policyHandler.List)
			policies.POST("", policyHandler.Create)
			policies.GET("/:id", policyHandler.GetByID)
			policies.PUT("/:id", policyHandler.Update)
			policies.DELETE("/:id", policyHandler.Delete)
			policies.GET("/:id/versions", policyHandler.GetVersions)
			policies.GET("/:id/timeline", policyHandler.GetTimeline)
		}

		attendances := authGroup.Group("/attendance")
		{
			attendances.GET("/events", attendanceHandler.ListEvents)
			attendances.POST("/events", attendanceHandler.CreateEvent)
			attendances.PUT("/events/:id", attendanceHandler.UpdateEvent)
			attendances.DELETE("/events/:id", attendanceHandler.DeleteEvent)
			attendances.GET("/leave-balance", attendanceHandler.GetLeaveBalance)
			attendances.POST("/grant-annual-leave", attendanceHandler.GrantAnnualLeave)
			attendances.POST("/close-month", attendanceHandler.CloseMonth)
			attendances.GET("/monthly", attendanceHandler.GetMonthlyEvents)
		}

		salaries := authGroup.Group("/salary")
		{
			salaries.GET("", salaryHandler.List)
			salaries.POST("/calculate", salaryHandler.Calculate)
			salaries.GET("/record", salaryHandler.GetRecord)
			salaries.POST("/adjustments", salaryHandler.AddAdjustment)
			salaries.DELETE("/adjustments/:id", salaryHandler.DeleteAdjustment)
			salaries.GET("/adjustments", salaryHandler.GetAdjustments)
			salaries.GET("/by-month", salaryHandler.ListByMonth)
		}

		assets := authGroup.Group("/assets")
		{
			assets.GET("", assetHandler.List)
			assets.POST("", assetHandler.Create)
			assets.GET("/:id", assetHandler.GetByID)
			assets.PUT("/:id", assetHandler.Update)
			assets.DELETE("/:id", assetHandler.Delete)
			assets.GET("/:id/versions", assetHandler.GetVersions)
			assets.GET("/:id/timeline", assetHandler.GetTimeline)
		}
	}

	r.NoRoute(spaFallback())

	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("服务启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func loadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.JWT.ExpireHours == 0 {
		config.JWT.ExpireHours = 24
	}

	return &config, nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func spaFallback() gin.HandlerFunc {
	fsys := frontendFS()
	fileServer := http.FileServer(http.FS(fsys))
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, common.Response{Code: 404, Message: "not found"})
			return
		}
		trimmed := strings.TrimPrefix(path, "/")
		f, err := fsys.Open(trimmed)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func frontendFS() fs.FS {
	sub, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		log.Printf("frontend dist not found, running API-only mode")
		return nil
	}
	return sub
}

func exportPersons(c *gin.Context) {
	var persons []person.Person
	err := common.DB.Select(&persons, "SELECT * FROM persons ORDER BY id")
	if err != nil {
		common.Error(c, 500, err.Error())
		return
	}

	headers := []string{"ID", "姓名", "入职日期", "基本工资", "绩效工资", "身份证号", "电话", "邮箱"}
	var rows [][]string
	for _, p := range persons {
		rows = append(rows, []string{
			fmt.Sprintf("%d", p.ID), p.Name, p.HireDate,
			fmt.Sprintf("%.2f", p.BaseSalary), fmt.Sprintf("%.2f", p.PerformanceSalary),
			p.IDNumber, p.Phones, p.Emails,
		})
	}

	f, err := common.ExportToExcel(headers, rows)
	if err != nil {
		common.Error(c, 500, err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=persons.xlsx")
	f.Write(c.Writer)
}

func exportSalary(c *gin.Context) {
	yearMonth := c.Query("yearMonth")
	if yearMonth == "" {
		common.Error(c, 400, "缺少月份参数")
		return
	}

	var records []salary.SalaryRecord
	err := common.DB.Select(&records, "SELECT * FROM salary_records WHERE year_month = ?", yearMonth)
	if err != nil {
		common.Error(c, 500, err.Error())
		return
	}

	headers := []string{"ID", "人员ID", "月份", "基本工资", "考勤工资", "绩效工资", "补贴合计", "扣款合计", "实发工资"}
	var rows [][]string
	for _, r := range records {
		rows = append(rows, []string{
			fmt.Sprintf("%d", r.ID), fmt.Sprintf("%d", r.PersonID), r.YearMonth,
			fmt.Sprintf("%.2f", r.BaseSalary), fmt.Sprintf("%.2f", r.AttendanceSalary),
			fmt.Sprintf("%.2f", r.PerformanceSalary), fmt.Sprintf("%.2f", r.TotalAllowances),
			fmt.Sprintf("%.2f", r.TotalDeductions), fmt.Sprintf("%.2f", r.NetSalary),
		})
	}

	f, err := common.ExportToExcel(headers, rows)
	if err != nil {
		common.Error(c, 500, err.Error())
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=salary_"+yearMonth+".xlsx")
	f.Write(c.Writer)
}
