package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"probig/internal/handler"
	"probig/internal/middleware"
	"probig/internal/service"
	"probig/pkg/config"
	frontend "probig/internal/frontend"
)

func Setup(r *gin.Engine, cfg *config.Config) {
	if err := service.InitSystem(); err != nil {
		panic("failed to initialize system: " + err.Error())
	}

	r.Use(middleware.CORS())

	api := r.Group("/api")

	api.POST("/login", handler.Login)

	auth := api.Group("")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/user/info", handler.GetUserInfo)
		auth.POST("/user/change-password", handler.ChangePassword)
		auth.POST("/user/reset-password", handler.ResetUserPassword)

		personGroup := auth.Group("/persons")
		{
			personGroup.GET("", middleware.RequirePermission("person.read"), handler.GetPersonList)
			personGroup.GET("/:id", middleware.RequirePermission("person.read"), handler.GetPerson)
			personGroup.POST("", middleware.RequirePermission("person.write"), handler.CreatePerson)
			personGroup.PUT("/:id", middleware.RequirePermission("person.write"), handler.UpdatePerson)
			personGroup.DELETE("/:id", middleware.RequirePermission("person.delete"), handler.DeletePerson)
			personGroup.POST("/:id/restore", middleware.RequirePermission("person.delete"), handler.RestorePerson)
		}

		companyGroup := auth.Group("/companies")
		{
			companyGroup.GET("", middleware.RequirePermission("company.read"), handler.GetCompanyList)
			companyGroup.GET("/:id", middleware.RequirePermission("company.read"), handler.GetCompany)
			companyGroup.POST("", middleware.RequirePermission("company.write"), handler.CreateCompany)
			companyGroup.PUT("/:id", middleware.RequirePermission("company.write"), handler.UpdateCompany)
			companyGroup.DELETE("/:id", middleware.RequirePermission("company.delete"), handler.DeleteCompany)
			companyGroup.POST("/:id/restore", middleware.RequirePermission("company.delete"), handler.RestoreCompany)
		}

		positionGroup := auth.Group("/position-events")
		{
			positionGroup.GET("", middleware.RequirePermission("person.read"), handler.GetPositionEvents)
			positionGroup.GET("/:id", middleware.RequirePermission("person.read"), handler.GetPositionEvent)
			positionGroup.POST("", middleware.RequirePermission("person.write"), handler.CreatePositionEvent)
			positionGroup.PUT("/:id", middleware.RequirePermission("person.write"), handler.UpdatePositionEvent)
			positionGroup.DELETE("/:id", middleware.RequirePermission("person.write"), handler.DeletePositionEvent)
		}

		positionSnapshotGroup := auth.Group("/position-snapshots")
		{
			positionSnapshotGroup.GET("", middleware.RequirePermission("person.read"), handler.GetPositionSnapshots)
			positionSnapshotGroup.POST("/rebuild", middleware.RequirePermission("person.write"), handler.RebuildSnapshots)
		}

		attendanceGroup := auth.Group("/attendance-events")
		{
			attendanceGroup.GET("", middleware.RequirePermission("attendance.read"), handler.GetAttendanceEventList)
			attendanceGroup.GET("/:id", middleware.RequirePermission("attendance.read"), handler.GetAttendanceEvent)
			attendanceGroup.POST("", middleware.RequirePermission("attendance.write"), handler.CreateAttendanceEvent)
			attendanceGroup.PUT("/:id", middleware.RequirePermission("attendance.write"), handler.UpdateAttendanceEvent)
			attendanceGroup.DELETE("/:id", middleware.RequirePermission("attendance.write"), handler.DeleteAttendanceEvent)
		}

		attendanceSummaryGroup := auth.Group("/attendance-summaries")
		{
			attendanceSummaryGroup.GET("", middleware.RequirePermission("attendance.read"), handler.GetAttendanceSummaryList)
			attendanceSummaryGroup.POST("/calculate", middleware.RequirePermission("attendance.write"), handler.CalculateAttendance)
			attendanceSummaryGroup.POST("/lock", middleware.RequirePermission("attendance.write"), handler.LockAttendanceSummary)
		}

		salaryEventGroup := auth.Group("/salary-events")
		{
			salaryEventGroup.GET("", middleware.RequirePermission("salary.read"), handler.GetSalaryEventList)
			salaryEventGroup.GET("/:id", middleware.RequirePermission("salary.read"), handler.GetSalaryEvent)
			salaryEventGroup.POST("", middleware.RequirePermission("salary.write"), handler.CreateSalaryEvent)
			salaryEventGroup.PUT("/:id", middleware.RequirePermission("salary.write"), handler.UpdateSalaryEvent)
			salaryEventGroup.DELETE("/:id", middleware.RequirePermission("salary.write"), handler.DeleteSalaryEvent)
		}

		salarySummaryGroup := auth.Group("/salary-summaries")
		{
			salarySummaryGroup.GET("", middleware.RequirePermission("salary.read"), handler.GetSalarySummaryList)
			salarySummaryGroup.POST("/calculate", middleware.RequirePermission("salary.write"), handler.CalculateSalary)
			salarySummaryGroup.POST("/lock", middleware.RequirePermission("salary.write"), handler.LockSalarySummary)
		}

		fileGroup := auth.Group("/files")
		{
			fileGroup.GET("", middleware.RequirePermission("file.read"), handler.GetFileList)
			fileGroup.GET("/:id", middleware.RequirePermission("file.read"), handler.GetFile)
			fileGroup.POST("/upload", middleware.RequirePermission("file.write"), handler.UploadFile)
			fileGroup.DELETE("/:id", middleware.RequirePermission("file.delete"), handler.DeleteFile)
			fileGroup.POST("/:id/restore", middleware.RequirePermission("file.delete"), handler.RestoreFile)
			fileGroup.GET("/relations", middleware.RequirePermission("file.read"), handler.GetFileRelationsByTarget)
			fileGroup.POST("/relations", middleware.RequirePermission("file.write"), handler.CreateFileRelation)
			fileGroup.DELETE("/relations/:id", middleware.RequirePermission("file.write"), handler.DeleteFileRelation)
		}

		auditGroup := auth.Group("/audit-logs")
		{
			auditGroup.GET("", middleware.RequirePermission("audit.read"), handler.GetAuditLogList)
		}

		userGroup := auth.Group("/users")
		{
			userGroup.GET("", middleware.RequirePermission("user.read"), handler.GetUserList)
			userGroup.GET("/:id", middleware.RequirePermission("user.read"), handler.GetUser)
			userGroup.POST("", middleware.RequirePermission("user.write"), handler.CreateUser)
			userGroup.PUT("/:id", middleware.RequirePermission("user.write"), handler.UpdateUser)
			userGroup.DELETE("/:id", middleware.RequirePermission("user.write"), handler.DeleteUser)
		}

		roleGroup := auth.Group("/roles")
		{
			roleGroup.GET("", middleware.RequirePermission("user.read"), handler.GetRoleList)
			roleGroup.GET("/:id", middleware.RequirePermission("user.read"), handler.GetRole)
			roleGroup.POST("", middleware.RequirePermission("user.write"), handler.CreateRole)
			roleGroup.PUT("/:id", middleware.RequirePermission("user.write"), handler.UpdateRole)
			roleGroup.DELETE("/:id", middleware.RequirePermission("user.write"), handler.DeleteRole)
		}

		auth.GET("/permissions", middleware.RequirePermission("user.read"), handler.GetAllPermissions)

		configGroup := auth.Group("/configs")
		{
			configGroup.GET("", middleware.RequirePermission("system.read"), handler.GetAllConfigs)
			configGroup.PUT("/:key", middleware.RequirePermission("system.write"), handler.UpdateConfig)
		}
	}

	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "pong"})
	})

	serveFrontend(r)
}

func serveFrontend(r *gin.Engine) {
	fileSystem := frontend.GetFileSystem()

	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, fileSystem)
	})

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "接口不存在"})
			return
		}
		c.FileFromFS("index.html", fileSystem)
	})
}
