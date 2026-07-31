package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"probig/server/internal/config"
	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/router"
	"probig/server/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	loaded := config.LoadConfig()

	if !loaded {
		if err := config.WriteDefaultConfig(); err != nil {
			log.Fatalf("生成默认配置文件失败: %v", err)
		}
	}

	dbPath := config.ResolvePath(config.AppConfig.Database.Path)
	uploadDir := config.ResolvePath(config.AppConfig.FileStorage.Path)
	logPath := config.ResolvePath(config.AppConfig.Log.File)

	for _, p := range []string{filepath.Dir(dbPath), uploadDir, filepath.Dir(logPath)} {
		if err := os.MkdirAll(p, 0755); err != nil {
			log.Fatalf("创建目录失败 %s: %v", p, err)
		}
	}

	log.Printf("数据库路径: %s", dbPath)
	log.Printf("文件存储路径: %s", uploadDir)

	db, err := dao.InitDB(dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := initSystem(db); err != nil {
		log.Fatalf("初始化系统失败: %v", err)
	}

	r := router.SetupRouter()

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Status(http.StatusNotFound)
			return
		}
		serveEmbeddedFile(c)
	})

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("服务启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func initSystem(db *gorm.DB) error {
	if err := autoMigrate(db); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	if err := service.InitSysConfig(db); err != nil {
		return fmt.Errorf("初始化系统配置失败: %w", err)
	}

	dao.RegisterAuditHooks(db)

	if err := service.SeedDefaultAdmin(db); err != nil {
		return fmt.Errorf("初始化默认管理员失败: %w", err)
	}

	log.Println("系统初始化完成")
	return nil
}

var contentTypeMap = map[string]string{
	".html": "text/html; charset=utf-8",
	".js":   "application/javascript",
	".css":  "text/css",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".woff": "font/woff",
	".woff2": "font/woff2",
}

func serveEmbeddedFile(c *gin.Context) {
	urlPath := c.Request.URL.Path
	filePath := "static" + urlPath
	if urlPath == "/" {
		filePath = "static/index.html"
	}

	data, err := staticFiles.ReadFile(filePath)
	if err != nil {
		filePath = "static/index.html"
		data, err = staticFiles.ReadFile(filePath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
	}

	contentType := "text/html; charset=utf-8"
	for ext, ct := range contentTypeMap {
		if strings.HasSuffix(filePath, ext) {
			contentType = ct
			break
		}
	}
	c.Data(http.StatusOK, contentType, data)
}

func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.AuditLog{},
		&model.SysConfig{},
		&model.Person{},
		&model.PersonPhone{},
		&model.PersonEmail{},
		&model.PersonBankCard{},
		&model.Company{},
		&model.File{},
		&model.FileRelation{},
		&model.PositionEvent{},
		&model.PositionSnapshot{},
		&model.AttendanceDaily{},
		&model.AttendanceEventDetail{},
		&model.AttendanceDailyProjection{},
		&model.AttendanceCalculationMonthly{},
		&model.AnnualLeaveAccountEvent{},
		&model.AnnualLeaveBalanceSnapshot{},
		&model.LeaveInLieuBalanceSnapshot{},
		&model.SysBatch{},
		&model.SalaryEvent{},
		&model.SalarySummary{},
		&model.SalarySummaryVersion{},
	); err != nil {
		return err
	}

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_persons_id_card ON persons(id_card) WHERE deleted_at IS NULL AND id_card != ''")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_credit_code ON companies(credit_code) WHERE deleted_at IS NULL AND credit_code != ''")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_position_events_person_seq ON position_events(person_id, seq) WHERE deleted_at IS NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_daily_person_date ON attendance_daily(person_id, event_date) WHERE deleted_at IS NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_salary_summaries_person_month ON salary_summaries(person_id, belong_month)")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_annual_leave_events_person_seq ON annual_leave_account_events(person_id, seq) WHERE deleted_at IS NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_salary_events_person_seq ON salary_events(person_id, seq) WHERE deleted_at IS NULL")

	return nil
}
