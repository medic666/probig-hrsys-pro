package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"probig/internal/attendance"
	"probig/internal/audit_log"
	"probig/internal/company"
	"probig/internal/file"
	"probig/internal/person"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
	"probig/internal/pkg/middleware"
	_ "probig/internal/pkg/utils"
	"probig/internal/position"
	"probig/internal/rbac"
	"probig/internal/salary"
	"probig/internal/system"
)

//go:embed web/dist
var webDist embed.FS

func main() {
	cfg := config.Load()

	if cfg.JwtSecret == "" {
		cfg.JwtSecret = "default-jwt-secret-change-in-production"
	}
	if cfg.EncryptKey == "" {
		cfg.EncryptKey = "default-encrypt-key-16b!"
	}

	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	if err := database.SeedDefaultData(db); err != nil {
		log.Fatalf("failed to seed default data: %v", err)
	}

	audit.GlobalAuditService = audit.NewAuditService(db)

	system.NewService(db).LoadToCache()
	audit_log.NewService(db)
	rbac.NewService(db)
	person.NewService(db, cfg.EncryptKey)
	company.NewService(db, cfg.EncryptKey)
	position.NewService(db, person.GetService())
	attendance.NewService(db)
	salary.NewService(db)
	file.NewService(db)

	engine := gin.Default()
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.ErrorRecoveryMiddleware())

	api := engine.Group("/api")

	authMiddleware := middleware.AuthMiddleware(cfg.JwtSecret)

	rbac.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	system.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	audit_log.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	person.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	company.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	position.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	attendance.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	salary.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)
	file.RegisterRoutes(api, authMiddleware, middleware.PermissionMiddleware)

	serveFrontend(engine)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func serveFrontend(engine *gin.Engine) {
	distFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Printf("no frontend dist found, serving API only")
		return
	}

	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found"})
			return
		}

		filePath := strings.TrimPrefix(path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		if f, err := distFS.Open(filePath); err == nil {
			f.Close()
			c.FileFromFS(filePath, http.FS(distFS))
			return
		}

		c.FileFromFS("index.html", http.FS(distFS))
	})
}
