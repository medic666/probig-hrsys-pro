package handlers

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/auth"
	"github.com/medic666/probig/internal/middleware"
	"github.com/medic666/probig/internal/response"
	"github.com/medic666/probig/internal/services"
)

type RouterConfig struct {
	DB           *sqlx.DB
	JWTManager   *auth.JWTManager
	UploadDir    string
	MaxUploadSize int64
	FrontendFS   fs.FS
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	auditSvc := services.NewAuditService(cfg.DB)
	snapshotEngine := services.NewSnapshotEngine(cfg.DB)

	personSvc := services.NewPersonService(cfg.DB, auditSvc, snapshotEngine)
	orgSvc := services.NewOrganizationService(cfg.DB, auditSvc, snapshotEngine)
	attendanceSvc := services.NewAttendanceService(cfg.DB, auditSvc)
	salarySvc := services.NewSalaryService(cfg.DB, auditSvc)
	fileSvc := services.NewFileService(cfg.DB, auditSvc, cfg.UploadDir, cfg.MaxUploadSize)

	authH := NewAuthHandler(cfg.DB, cfg.JWTManager)
	personH := NewPersonHandler(personSvc)
	orgH := NewOrganizationHandler(orgSvc)
	attendanceH := NewAttendanceHandler(attendanceSvc)
	salaryH := NewSalaryHandler(salarySvc)
	fileH := NewFileHandler(fileSvc)
	auditH := NewAuditHandler(auditSvc)
	exportH := NewExportHandler(cfg.DB)

	api := r.Group("/api")
	{
		api.POST("/auth/login", authH.Login)

		authGroup := api.Group("", middleware.AuthRequired(cfg.JWTManager))
		{
			authGroup.GET("/auth/me", authH.Me)

			authGroup.GET("/persons", middleware.RBAC("person", "read"), personH.List)
			authGroup.POST("/persons", middleware.RBAC("person", "write"), personH.Create)
			authGroup.GET("/persons/:id", middleware.RBAC("person", "read"), personH.GetDetail)
			authGroup.PUT("/persons/:id/status", middleware.RBAC("person", "write"), personH.UpdateStatus)
			authGroup.GET("/persons/:id/snapshots", middleware.RBAC("person", "read"), personH.GetSnapshots)
			authGroup.POST("/person-events", middleware.RBAC("person", "write"), personH.CreateEvent)
			authGroup.PUT("/person-events/:id", middleware.RBAC("person", "write"), personH.UpdateEvent)
			authGroup.DELETE("/person-events/:id", middleware.RBAC("person", "delete"), personH.DeleteEvent)

			authGroup.GET("/organizations", middleware.RBAC("organization", "read"), orgH.List)
			authGroup.POST("/organizations", middleware.RBAC("organization", "write"), orgH.Create)
			authGroup.GET("/organizations/:id", middleware.RBAC("organization", "read"), orgH.GetDetail)
			authGroup.POST("/org-events", middleware.RBAC("organization", "write"), orgH.CreateEvent)
			authGroup.PUT("/org-events/:id", middleware.RBAC("organization", "write"), orgH.UpdateEvent)
			authGroup.DELETE("/org-events/:id", middleware.RBAC("organization", "delete"), orgH.DeleteEvent)

			authGroup.GET("/attendance-events", middleware.RBAC("attendance", "read"), attendanceH.ListEvents)
			authGroup.POST("/attendance-events", middleware.RBAC("attendance", "write"), attendanceH.CreateEvent)
			authGroup.PUT("/attendance-events/:id", middleware.RBAC("attendance", "write"), attendanceH.UpdateEvent)
			authGroup.DELETE("/attendance-events/:id", middleware.RBAC("attendance", "delete"), attendanceH.DeleteEvent)
			authGroup.POST("/attendance/calculate", middleware.RBAC("attendance", "write"), attendanceH.Calculate)
			authGroup.GET("/attendance-summaries", middleware.RBAC("attendance", "read"), attendanceH.ListSummaries)

			authGroup.GET("/salary-events", middleware.RBAC("salary", "read"), salaryH.ListEvents)
			authGroup.POST("/salary-events", middleware.RBAC("salary", "write"), salaryH.CreateEvent)
			authGroup.PUT("/salary-events/:id", middleware.RBAC("salary", "write"), salaryH.UpdateEvent)
			authGroup.DELETE("/salary-events/:id", middleware.RBAC("salary", "delete"), salaryH.DeleteEvent)
			authGroup.POST("/salary/calculate", middleware.RBAC("salary", "write"), salaryH.Calculate)
			authGroup.GET("/salary-summaries", middleware.RBAC("salary", "read"), salaryH.ListSummaries)

			authGroup.POST("/files/upload", middleware.RBAC("file", "write"), fileH.Upload)
			authGroup.GET("/files", middleware.RBAC("file", "read"), fileH.List)
			authGroup.DELETE("/files/:id", middleware.RBAC("file", "delete"), fileH.Delete)
			authGroup.GET("/files/:id/download", middleware.RBAC("file", "read"), fileH.Download)
			authGroup.POST("/file-associations", middleware.RBAC("file", "write"), fileH.CreateAssociation)
			authGroup.DELETE("/file-associations/:id", middleware.RBAC("file", "delete"), fileH.DeleteAssociation)
			authGroup.GET("/file-associations", middleware.RBAC("file", "read"), fileH.GetAssociations)

			authGroup.GET("/audit-logs", middleware.RBAC("audit", "read"), auditH.List)

			authGroup.GET("/export/attendance", middleware.RBAC("attendance", "read"), exportH.ExportAttendance)
			authGroup.GET("/export/salary", middleware.RBAC("salary", "read"), exportH.ExportSalary)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		if cfg.FrontendFS == nil {
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			response.NotFound(c, "接口不存在")
			return
		}
		path := c.Request.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		f, err := cfg.FrontendFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			c.FileFromFS(path, http.FS(cfg.FrontendFS))
			return
		}
		c.FileFromFS("/index.html", http.FS(cfg.FrontendFS))
	})

	return r
}
