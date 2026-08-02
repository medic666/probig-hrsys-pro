package router

import (
	"strings"

	"probig/server/internal/config"
	"probig/server/internal/handler"
	"probig/server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS(splitOrigins(config.AppConfig.Server.CorsOrigins)))

	// API 响应禁止缓存，避免浏览器对接口响应做启发式缓存
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
		}
		c.Next()
	})

	r.GET("/api/health", handler.Health)

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", handler.AuthLogin)
		auth.POST("/logout", middleware.AuthRequired(), handler.AuthLogout)
		auth.POST("/change-password", middleware.AuthRequired(), handler.AuthChangePassword)
		auth.GET("/me", middleware.AuthRequired(), handler.AuthMe)
	}

	users := r.Group("/api/users")
	users.Use(middleware.AuthRequired())
	{
		users.GET("", middleware.RequirePermission("user.read"), handler.GetUsers)
		users.POST("", middleware.RequirePermission("user.write"), handler.CreateUser)
		users.PUT("/:id", middleware.RequirePermission("user.write"), handler.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission("user.delete"), handler.DeleteUser)
		users.POST("/:id/reset-password", middleware.RequirePermission("user.write"), handler.ResetUserPassword)
		users.POST("/:id/assign-roles", middleware.RequirePermission("user.write"), handler.AssignUserRoles)
		users.GET("/trash", middleware.RequirePermission("user.read"), handler.GetDeletedUsers)
		users.GET("/:id", middleware.RequirePermission("user.read"), handler.GetUserByID)
		users.POST("/:id/restore", middleware.RequirePermission("user.write"), handler.RestoreUser)
	}

	roles := r.Group("/api/roles")
	roles.Use(middleware.AuthRequired())
	{
		roles.GET("", middleware.RequirePermission("role.read"), handler.GetRoles)
		roles.GET("/all", middleware.RequirePermission("role.read"), handler.GetAllRolesList)
		roles.POST("", middleware.RequirePermission("role.write"), handler.CreateRole)
		roles.PUT("/:id", middleware.RequirePermission("role.write"), handler.UpdateRole)
		roles.DELETE("/:id", middleware.RequirePermission("role.delete"), handler.DeleteRole)
		roles.GET("/trash", middleware.RequirePermission("role.read"), handler.GetDeletedRoles)
		roles.GET("/:id", middleware.RequirePermission("role.read"), handler.GetRoleByID)
		roles.POST("/:id/restore", middleware.RequirePermission("role.write"), handler.RestoreRole)
		roles.POST("/:id/assign-permissions", middleware.RequirePermission("role.write"), handler.AssignRolePermissions)
		roles.GET("/:id/permissions", middleware.RequirePermission("role.read"), handler.GetRolePermissions)
	}

	r.GET("/api/permissions", middleware.AuthRequired(), handler.GetPermissions)

	persons := r.Group("/api/persons")
	persons.Use(middleware.AuthRequired())
	{
		persons.GET("", middleware.RequirePermission("person.read"), handler.GetPersons)
		persons.GET("/all", middleware.RequirePermission("person.read"), handler.GetAllPersonsList)
		persons.GET("/cards", middleware.RequirePermission("person.read"), handler.GetPersonCards)
		persons.GET("/export", middleware.RequirePermission("person.export"), handler.ExportPersons)
		persons.GET("/trash", middleware.RequirePermission("person.read"), handler.GetDeletedPersons)
		persons.GET("/:id", middleware.RequirePermission("person.read"), handler.GetPersonByID)
		persons.POST("/profile", middleware.RequirePermission("person.write"), handler.UpsertPersonProfile)
		persons.DELETE("/:id", middleware.RequirePermission("person.delete"), handler.DeletePerson)
		persons.POST("/:id/restore", middleware.RequirePermission("person.write"), handler.RestorePerson)
		persons.GET("/:id/current-position", middleware.RequirePermission("person.read"), handler.GetPersonCurrentPosition)
		persons.GET("/:id/position-history", middleware.RequirePermission("person.read"), handler.GetPersonPositionHistory)
	}

	companies := r.Group("/api/companies")
	companies.Use(middleware.AuthRequired())
	{
		companies.GET("", middleware.RequirePermission("company.read"), handler.GetCompanies)
		companies.GET("/all", middleware.RequirePermission("company.read"), handler.GetAllCompaniesList)
		companies.GET("/export", middleware.RequirePermission("company.export"), handler.ExportCompanies)
		companies.GET("/trash", middleware.RequirePermission("company.read"), handler.GetDeletedCompanies)
		companies.GET("/:id", middleware.RequirePermission("company.read"), handler.GetCompanyByID)
		companies.POST("", middleware.RequirePermission("company.write"), handler.CreateCompany)
		companies.PUT("/:id", middleware.RequirePermission("company.write"), handler.UpdateCompany)
		companies.DELETE("/:id", middleware.RequirePermission("company.delete"), handler.DeleteCompany)
		companies.POST("/:id/restore", middleware.RequirePermission("company.write"), handler.RestoreCompany)
	}

	files := r.Group("/api/files")
	files.Use(middleware.AuthRequired())
	{
		files.POST("/upload", middleware.RequirePermission("file.write"), handler.UploadFile)
		files.GET("/:id/download", middleware.RequirePermission("file.read"), handler.DownloadFile)
		files.POST("/associate", middleware.RequirePermission("file.write"), handler.AssociateFile)
		files.POST("/disassociate", middleware.RequirePermission("file.write"), handler.DisassociateFile)
		files.GET("/by-target", middleware.RequirePermission("file.read"), handler.GetFilesByTarget)
		files.GET("", middleware.RequirePermission("file.read"), handler.GetFiles)
		files.GET("/trash", middleware.RequirePermission("file.read"), handler.GetDeletedFiles)
		files.DELETE("/:id", middleware.RequirePermission("file.delete"), handler.DeleteFile)
		files.POST("/:id/restore", middleware.RequirePermission("file.write"), handler.RestoreFile)
		files.GET("/:id/associations", middleware.RequirePermission("file.read"), handler.GetFileAssociations)
		files.DELETE("/:id/permanent", middleware.RequirePermission("file.delete"), handler.PermanentDeleteFile)
		files.POST("/clean-orphans", middleware.RequirePermission("file.delete"), handler.CleanOrphanFiles)
	}

	positionEvents := r.Group("/api/position-events")
	positionEvents.Use(middleware.AuthRequired())
	{
		positionEvents.GET("", middleware.RequirePermission("position_event.read"), handler.GetPositionEvents)
		positionEvents.GET("/export", middleware.RequirePermission("position_event.export"), handler.ExportPositionEvents)
		positionEvents.GET("/trash", middleware.RequirePermission("position_event.read"), handler.GetDeletedPositionEvents)
		positionEvents.GET("/:id", middleware.RequirePermission("position_event.read"), handler.GetPositionEventByID)
		positionEvents.POST("", middleware.RequirePermission("position_event.write"), handler.CreatePositionEvent)
		positionEvents.PUT("/:id", middleware.RequirePermission("position_event.write"), handler.UpdatePositionEvent)
		positionEvents.DELETE("/:id", middleware.RequirePermission("position_event.delete"), handler.DeletePositionEvent)
		positionEvents.POST("/:id/restore", middleware.RequirePermission("position_event.write"), handler.RestorePositionEvent)
	}

	attendanceEvents := r.Group("/api/attendance-events")
	attendanceEvents.Use(middleware.AuthRequired())
	{
		attendanceEvents.GET("", middleware.RequirePermission("attendance.read"), handler.GetAttendanceEvents)
		attendanceEvents.GET("/export", middleware.RequirePermission("attendance.export"), handler.ExportAttendanceEvents)
		attendanceEvents.GET("/trash", middleware.RequirePermission("attendance.read"), handler.GetDeletedAttendanceEvents)
		attendanceEvents.GET("/pending", middleware.RequirePermission("attendance.read"), handler.GetPendingDailyList)
		attendanceEvents.GET("/:id", middleware.RequirePermission("attendance.read"), handler.GetAttendanceEventByID)
		attendanceEvents.POST("", middleware.RequirePermission("attendance.write"), handler.CreateAttendanceEvent)
		attendanceEvents.POST("/batch", middleware.RequirePermission("attendance.write"), handler.CreateBatchAttendanceEvents)
		attendanceEvents.POST("/:id/confirm", middleware.RequirePermission("attendance.write"), handler.ConfirmAttendanceDaily)
		attendanceEvents.POST("/import-dingtalk/preview", middleware.RequirePermission("attendance.write"), handler.DingTalkPreview)
		attendanceEvents.POST("/import-dingtalk/execute", middleware.RequirePermission("attendance.write"), handler.DingTalkExecute)
		attendanceEvents.DELETE("/:id", middleware.RequirePermission("attendance.delete"), handler.DeleteAttendanceEvent)
		attendanceEvents.POST("/:id/restore", middleware.RequirePermission("attendance.write"), handler.RestoreAttendanceEvent)
	}

	attendanceDaily := r.Group("/api/attendance-daily")
	attendanceDaily.Use(middleware.AuthRequired())
	{
		attendanceDaily.GET("", middleware.RequirePermission("attendance.read"), handler.GetDailyProjections)
		attendanceDaily.GET("/export", middleware.RequirePermission("attendance.export"), handler.ExportDailyProjections)
		attendanceDaily.GET("/:personId/:date/events", middleware.RequirePermission("attendance.read"), handler.GetEventsByPersonDate)
	}

	attendanceMonthly := r.Group("/api/attendance-monthly")
	attendanceMonthly.Use(middleware.AuthRequired())
	{
		attendanceMonthly.GET("", middleware.RequirePermission("attendance.read"), handler.GetMonthlyList)
		attendanceMonthly.GET("/export", middleware.RequirePermission("attendance.export"), handler.ExportAttendanceMonthly)
		attendanceMonthly.POST("/calculate", middleware.RequirePermission("attendance.write"), handler.CalculateMonthly)
	}

	annualLeave := r.Group("/api/annual-leave-events")
	annualLeave.Use(middleware.AuthRequired())
	{
		annualLeave.GET("", middleware.RequirePermission("annual_leave.read"), handler.GetAnnualLeaveEvents)
		annualLeave.GET("/export", middleware.RequirePermission("annual_leave.export"), handler.ExportAnnualLeaveEvents)
		annualLeave.GET("/trash", middleware.RequirePermission("annual_leave.read"), handler.GetDeletedAnnualLeaveEvents)
		annualLeave.GET("/:id", middleware.RequirePermission("annual_leave.read"), handler.GetAnnualLeaveEventByID)
		annualLeave.POST("", middleware.RequirePermission("annual_leave.write"), handler.CreateAnnualLeaveEvent)
		annualLeave.PUT("/:id", middleware.RequirePermission("annual_leave.write"), handler.UpdateAnnualLeaveEvent)
		annualLeave.DELETE("/:id", middleware.RequirePermission("annual_leave.delete"), handler.DeleteAnnualLeaveEvent)
		annualLeave.POST("/:id/restore", middleware.RequirePermission("annual_leave.write"), handler.RestoreAnnualLeaveEvent)
	}

	leaveBalance := r.Group("/api")
	leaveBalance.Use(middleware.AuthRequired())
	{
		leaveBalance.GET("/persons/:id/annual-leave-balance", middleware.RequirePermission("annual_leave.read"), handler.GetPersonAnnualLeaveBalance)
		leaveBalance.GET("/persons/:id/annual-leave-balance-history", middleware.RequirePermission("annual_leave.read"), handler.GetPersonAnnualLeaveHistory)
		leaveBalance.GET("/persons/:id/annual-leave-balance-detail", middleware.RequirePermission("annual_leave.read"), handler.GetALBalanceDetail)
		leaveBalance.GET("/persons/:id/lil-balance", middleware.RequirePermission("annual_leave.read"), handler.GetPersonLILBalance)
		leaveBalance.GET("/persons/:id/lil-balance-history", middleware.RequirePermission("annual_leave.read"), handler.GetPersonLILHistory)
		leaveBalance.GET("/persons/:id/lil-balance-detail", middleware.RequirePermission("annual_leave.read"), handler.GetLILBalanceDetail)
		leaveBalance.GET("/lil-events", middleware.RequirePermission("annual_leave.read"), handler.GetLILEvents)
		leaveBalance.GET("/annual-leave-balances", middleware.RequirePermission("annual_leave.read"), handler.GetAllALBalances)
		leaveBalance.GET("/lil-balances", middleware.RequirePermission("annual_leave.read"), handler.GetAllLILBalances)
	}

	carryover := r.Group("/api/annual-leave-carryover")
	carryover.Use(middleware.AuthRequired())
	{
		carryover.POST("", middleware.RequirePermission("annual_leave.write"), handler.ExecuteAnnualLeaveCarryover)
		carryover.POST("/:batchId/cancel", middleware.RequirePermission("annual_leave.write"), handler.CancelCarryover)
		carryover.GET("/batches", middleware.RequirePermission("annual_leave.read"), handler.GetCarryoverBatches)
		carryover.GET("/batches/:batchId/events", middleware.RequirePermission("annual_leave.read"), handler.GetBatchEvents)
	}

	salaryEvents := r.Group("/api/salary-events")
	salaryEvents.Use(middleware.AuthRequired())
	{
		salaryEvents.GET("", middleware.RequirePermission("salary.read"), handler.GetSalaryEvents)
		salaryEvents.GET("/export", middleware.RequirePermission("salary.export"), handler.ExportSalaryEvents)
		salaryEvents.GET("/trash", middleware.RequirePermission("salary.read"), handler.GetDeletedSalaryEvents)
		salaryEvents.GET("/:id", middleware.RequirePermission("salary.read"), handler.GetSalaryEventByID)
		salaryEvents.POST("", middleware.RequirePermission("salary.write"), handler.CreateSalaryEvent)
		salaryEvents.PUT("/:id", middleware.RequirePermission("salary.write"), handler.UpdateSalaryEvent)
		salaryEvents.DELETE("/:id", middleware.RequirePermission("salary.delete"), handler.DeleteSalaryEvent)
		salaryEvents.POST("/:id/restore", middleware.RequirePermission("salary.write"), handler.RestoreSalaryEvent)
	}

	salarySummaries := r.Group("/api/salary-summaries")
	salarySummaries.Use(middleware.AuthRequired())
	{
		salarySummaries.GET("", middleware.RequirePermission("salary.read"), handler.GetSalarySummaries)
		salarySummaries.POST("/calculate", middleware.RequirePermission("salary.write"), handler.CalculateSalarySummaries)
		salarySummaries.GET("/export", middleware.RequirePermission("salary.export"), handler.ExportSalarySummaries)
		salarySummaries.GET("/:personId/:month/versions", middleware.RequirePermission("salary.read"), handler.GetSalaryVersions)
		salarySummaries.GET("/:personId/:month/trace", middleware.RequirePermission("salary.read"), handler.GetSalaryTrace)
		salarySummaries.GET("/versions/:vid", middleware.RequirePermission("salary.read"), handler.GetSalaryVersionDetail)
	}

	audit := r.Group("/api/audit-logs")
	audit.Use(middleware.AuthRequired())
	{
		audit.GET("", middleware.RequirePermission("audit.read"), handler.GetAuditLogs)
		audit.GET("/export", middleware.RequirePermission("audit.export"), handler.ExportAuditLogs)
		audit.GET("/:id", middleware.RequirePermission("audit.read"), handler.GetAuditLogDetail)
	}

	sysCfg := r.Group("/api/system-configs")
	sysCfg.Use(middleware.AuthRequired())
	{
		sysCfg.GET("", middleware.RequirePermission("system_config.read"), handler.GetSystemConfigs)
		sysCfg.PUT("/:key", middleware.RequirePermission("system_config.write"), handler.UpdateSystemConfig)
	}

	return r
}


func splitOrigins(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
