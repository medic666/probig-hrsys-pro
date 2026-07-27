package router

import (
	"probig/server/internal/handler"
	"probig/server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(gin.Logger())

	r.GET("/api/health", handler.Health)

	test := r.Group("/api/test")
	{
		test.GET("/pagination", handler.TestPagination)
		test.GET("/names", handler.TestNames)
		test.POST("/form-submit", handler.TestFormSubmit)
		test.GET("/trash-list", handler.TestTrashList)
		test.POST("/trash-restore", handler.TestTrashRestore)
	}

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
		persons.GET("/export", middleware.RequirePermission("person.export"), handler.ExportPersons)
		persons.GET("/trash", middleware.RequirePermission("person.read"), handler.GetDeletedPersons)
		persons.GET("/:id", middleware.RequirePermission("person.read"), handler.GetPersonByID)
		persons.POST("", middleware.RequirePermission("person.write"), handler.CreatePerson)
		persons.PUT("/:id", middleware.RequirePermission("person.write"), handler.UpdatePerson)
		persons.DELETE("/:id", middleware.RequirePermission("person.delete"), handler.DeletePerson)
		persons.POST("/:id/restore", middleware.RequirePermission("person.write"), handler.RestorePerson)
		persons.POST("/:id/phones", middleware.RequirePermission("person.write"), handler.AddPersonPhone)
		persons.PUT("/:id/phones/:pid", middleware.RequirePermission("person.write"), handler.UpdatePersonPhone)
		persons.DELETE("/:id/phones/:pid", middleware.RequirePermission("person.write"), handler.DeletePersonPhone)
		persons.POST("/:id/emails", middleware.RequirePermission("person.write"), handler.AddPersonEmail)
		persons.PUT("/:id/emails/:pid", middleware.RequirePermission("person.write"), handler.UpdatePersonEmail)
		persons.DELETE("/:id/emails/:pid", middleware.RequirePermission("person.write"), handler.DeletePersonEmail)
		persons.POST("/:id/bank-cards", middleware.RequirePermission("person.write"), handler.AddPersonBankCard)
		persons.PUT("/:id/bank-cards/:pid", middleware.RequirePermission("person.write"), handler.UpdatePersonBankCard)
		persons.DELETE("/:id/bank-cards/:pid", middleware.RequirePermission("person.write"), handler.DeletePersonBankCard)
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
	}

	positionEvents := r.Group("/api/position-events")
	positionEvents.Use(middleware.AuthRequired())
	{
		positionEvents.GET("", middleware.RequirePermission("position_event.read"), handler.GetPositionEvents)
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
		attendanceEvents.GET("/trash", middleware.RequirePermission("attendance.read"), handler.GetDeletedAttendanceEvents)
		attendanceEvents.GET("/:id", middleware.RequirePermission("attendance.read"), handler.GetAttendanceEventByID)
		attendanceEvents.POST("", middleware.RequirePermission("attendance.write"), handler.CreateAttendanceEvent)
		attendanceEvents.POST("/batch", middleware.RequirePermission("attendance.write"), handler.CreateBatchAttendanceEvents)
		attendanceEvents.PUT("/:id", middleware.RequirePermission("attendance.write"), handler.UpdateAttendanceEvent)
		attendanceEvents.DELETE("/:id", middleware.RequirePermission("attendance.delete"), handler.DeleteAttendanceEvent)
		attendanceEvents.POST("/:id/restore", middleware.RequirePermission("attendance.write"), handler.RestoreAttendanceEvent)
	}

	attendanceDaily := r.Group("/api/attendance-daily")
	attendanceDaily.Use(middleware.AuthRequired())
	{
		attendanceDaily.GET("", middleware.RequirePermission("attendance.read"), handler.GetDailyProjections)
		attendanceDaily.GET("/:personId/:date/events", middleware.RequirePermission("attendance.read"), handler.GetEventsByPersonDate)
	}

	attendanceMonthly := r.Group("/api/attendance-monthly")
	attendanceMonthly.Use(middleware.AuthRequired())
	{
		attendanceMonthly.GET("", middleware.RequirePermission("attendance.read"), handler.GetMonthlyList)
		attendanceMonthly.POST("/calculate", middleware.RequirePermission("attendance.write"), handler.CalculateMonthly)
	}

	return r
}
