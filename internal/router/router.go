package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"probig/internal/config"
	"probig/internal/handler"
	"probig/internal/middleware"
)

func Setup(cfg *config.Config, frontendFS embed.FS) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	authH := handler.NewAuthHandler(cfg)
	personnelH := handler.NewPersonnelHandler()
	orgH := handler.NewOrganizationHandler()
	attendanceH := handler.NewAttendanceHandler()
	salaryH := handler.NewSalaryHandler()
	fileH := handler.NewFileHandler(cfg)
	auditH := handler.NewAuditHandler()

	api := r.Group("/api")
	{
		api.POST("/auth/login", authH.Login)

		auth := api.Group("")
		auth.Use(middleware.JWTAuth(cfg))
		{
			auth.GET("/auth/me", authH.Me)
			auth.GET("/auth/permissions", authH.Permissions)

			auth.GET("/personnel", middleware.RBAC("personnel", "read"), personnelH.List)
			auth.GET("/personnel/:id", middleware.RBAC("personnel", "read"), personnelH.Get)
			auth.GET("/personnel/:id/history", middleware.RBAC("personnel", "read"), personnelH.History)
			auth.GET("/personnel/events", middleware.RBAC("personnel", "read"), personnelH.ListEvents)
			auth.POST("/personnel/events", middleware.RBAC("personnel", "write"), personnelH.CreateEvent)
			auth.PUT("/personnel/events/:id", middleware.RBAC("personnel", "write"), personnelH.UpdateEvent)
			auth.DELETE("/personnel/events/:id", middleware.RBAC("personnel", "delete"), personnelH.DeleteEvent)

			auth.GET("/organizations", middleware.RBAC("organization", "read"), orgH.List)
			auth.GET("/organizations/:id", middleware.RBAC("organization", "read"), orgH.Get)
			auth.GET("/organizations/:id/history", middleware.RBAC("organization", "read"), orgH.History)
			auth.GET("/organizations/events", middleware.RBAC("organization", "read"), orgH.ListEvents)
			auth.POST("/organizations/events", middleware.RBAC("organization", "write"), orgH.CreateEvent)
			auth.PUT("/organizations/events/:id", middleware.RBAC("organization", "write"), orgH.UpdateEvent)
			auth.DELETE("/organizations/events/:id", middleware.RBAC("organization", "delete"), orgH.DeleteEvent)

			auth.GET("/attendance/events", middleware.RBAC("attendance", "read"), attendanceH.ListEvents)
			auth.POST("/attendance/events", middleware.RBAC("attendance", "write"), attendanceH.CreateEvent)
			auth.PUT("/attendance/events/:id", middleware.RBAC("attendance", "write"), attendanceH.UpdateEvent)
			auth.DELETE("/attendance/events/:id", middleware.RBAC("attendance", "delete"), attendanceH.DeleteEvent)
			auth.GET("/attendance/summaries", middleware.RBAC("attendance", "read"), attendanceH.ListSummaries)
			auth.POST("/attendance/calculate", middleware.RBAC("attendance", "write"), attendanceH.Calculate)

			auth.GET("/salary/events", middleware.RBAC("salary", "read"), salaryH.ListEvents)
			auth.POST("/salary/events", middleware.RBAC("salary", "write"), salaryH.CreateEvent)
			auth.PUT("/salary/events/:id", middleware.RBAC("salary", "write"), salaryH.UpdateEvent)
			auth.DELETE("/salary/events/:id", middleware.RBAC("salary", "delete"), salaryH.DeleteEvent)
			auth.GET("/salary/summaries", middleware.RBAC("salary", "read"), salaryH.ListSummaries)
			auth.POST("/salary/calculate", middleware.RBAC("salary", "write"), salaryH.Calculate)

			auth.GET("/files", middleware.RBAC("file", "read"), fileH.List)
			auth.GET("/files/:id/download", middleware.RBAC("file", "read"), fileH.Download)
			auth.POST("/files/upload", middleware.RBAC("file", "write"), fileH.Upload)
			auth.DELETE("/files/:id", middleware.RBAC("file", "delete"), fileH.Delete)
			auth.POST("/files/:id/associate", middleware.RBAC("file", "write"), fileH.Associate)
			auth.DELETE("/files/:id/associate/:targetId", middleware.RBAC("file", "write"), fileH.Disassociate)

			auth.GET("/audit-logs", middleware.RBAC("audit", "read"), auditH.List)
		}
	}

	if cfg.DevMode {
		r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	} else {
		staticFS, err := fs.Sub(frontendFS, "frontend/dist")
		if err != nil {
			panic("failed to get frontend dist sub filesystem: " + err.Error())
		}
		fsys := http.FS(staticFS)
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"code": 40400, "message": "接口不存在", "data": nil})
				return
			}
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/assets/") {
				serveStaticFile(c, fsys, path)
				return
			}
			serveStaticFile(c, fsys, "index.html")
		})
	}

	return r
}

func serveStaticFile(c *gin.Context, fsys http.FileSystem, name string) {
	f, err := fsys.Open(name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	http.ServeContent(c.Writer, c.Request, name, stat.ModTime(), f)
}
