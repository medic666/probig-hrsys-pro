package router

import (
	"probig/handlers"
	"probig/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ErrorHandler())

	api := r.Group("/api")
	{
		api.POST("/login", handlers.Login)

		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/user/info", handlers.GetUserInfo)
			auth.PUT("/user/change-password", handlers.ChangePassword)
			auth.PUT("/user/:id/reset-password", handlers.ResetUserPassword)

			auth.GET("/persons", handlers.ListPersons)
			auth.GET("/persons/deleted", handlers.ListDeletedPersons)
			auth.GET("/persons/:id", handlers.GetPerson)
			auth.POST("/persons", handlers.CreatePerson)
			auth.PUT("/persons/:id", handlers.UpdatePerson)
			auth.DELETE("/persons/:id", handlers.DeletePerson)
			auth.PUT("/persons/:id/restore", handlers.RestorePerson)
			auth.POST("/persons/:id/phones", handlers.AddPersonPhone)
			auth.PUT("/person-phones/:id", handlers.UpdatePersonPhone)
			auth.DELETE("/person-phones/:id", handlers.DeletePersonPhone)
			auth.POST("/persons/:id/emails", handlers.AddPersonEmail)
			auth.PUT("/person-emails/:id", handlers.UpdatePersonEmail)
			auth.DELETE("/person-emails/:id", handlers.DeletePersonEmail)
			auth.POST("/persons/:id/bank-cards", handlers.AddPersonBankCard)
			auth.PUT("/person-bank-cards/:id", handlers.UpdatePersonBankCard)
			auth.DELETE("/person-bank-cards/:id", handlers.DeletePersonBankCard)

			auth.GET("/companies", handlers.ListCompanies)
			auth.GET("/companies/deleted", handlers.ListDeletedCompanies)
			auth.GET("/companies/:id", handlers.GetCompany)
			auth.POST("/companies", handlers.CreateCompany)
			auth.PUT("/companies/:id", handlers.UpdateCompany)
			auth.DELETE("/companies/:id", handlers.DeleteCompany)
			auth.PUT("/companies/:id/restore", handlers.RestoreCompany)

			auth.GET("/position-events", handlers.ListPositionEvents)
			auth.GET("/position-events/:id", handlers.GetPositionEvent)
			auth.POST("/position-events", handlers.CreatePositionEvent)
			auth.PUT("/position-events/:id", handlers.UpdatePositionEvent)
			auth.DELETE("/position-events/:id", handlers.DeletePositionEvent)
			auth.GET("/position-snapshots/:person_id", handlers.ListPositionSnapshots)
			auth.GET("/position-snapshots/:person_id/latest", handlers.GetLatestSnapshot)
			auth.POST("/position-snapshots/:person_id/rebuild", handlers.RebuildSnapshots)

			auth.GET("/attendance-events", handlers.ListAttendanceEvents)
			auth.POST("/attendance-events", handlers.CreateAttendanceEvent)
			auth.POST("/attendance-events/batch", handlers.BatchCreateAttendanceEvents)
			auth.PUT("/attendance-events/:id", handlers.UpdateAttendanceEvent)
			auth.DELETE("/attendance-events/:id", handlers.DeleteAttendanceEvent)
			auth.POST("/attendance-summaries/calc", handlers.CalcAttendanceSummary)
			auth.GET("/attendance-summaries", handlers.ListAttendanceSummaries)
			auth.PUT("/attendance-summaries/:id/lock", handlers.LockAttendanceSummary)
			auth.GET("/annual-leave/:person_id/balance", handlers.GetAnnualLeaveBalance)
			auth.POST("/annual-leave/anniversary", handlers.AnnualLeaveAnniversary)

			auth.GET("/salary-events", handlers.ListSalaryEvents)
			auth.POST("/salary-events", handlers.CreateSalaryEvent)
			auth.PUT("/salary-events/:id", handlers.UpdateSalaryEvent)
			auth.DELETE("/salary-events/:id", handlers.DeleteSalaryEvent)
			auth.POST("/salary-summaries/calc", handlers.CalcSalary)
			auth.GET("/salary-summaries", handlers.ListSalarySummaries)
			auth.PUT("/salary-summaries/:id/lock", handlers.LockSalarySummary)

			auth.POST("/files/upload", handlers.UploadFile)
			auth.GET("/files", handlers.ListFiles)
			auth.GET("/files/:id", handlers.GetFile)
			auth.DELETE("/files/:id", handlers.DeleteFile)
			auth.PUT("/files/:id/restore", handlers.RestoreFile)
			auth.POST("/file-relations", handlers.AddFileRelation)
			auth.DELETE("/file-relations/:id", handlers.RemoveFileRelation)
			auth.GET("/file-relations", handlers.GetFileRelations)

			auth.GET("/audit-logs", handlers.ListAuditLogs)

			auth.GET("/configs", handlers.ListConfigs)
			auth.PUT("/configs", handlers.UpdateConfig)

			auth.GET("/users", handlers.ListUsers)
			auth.GET("/users/simple", handlers.ListUsersSimple)
			auth.GET("/users/:id", handlers.GetUser)
			auth.POST("/users", handlers.CreateUser)
			auth.PUT("/users/:id", handlers.UpdateUser)
			auth.DELETE("/users/:id", handlers.DeleteUser)
			auth.PUT("/users/:id/roles", handlers.AssignUserRoles)

			auth.GET("/roles", handlers.ListRoles)
			auth.GET("/roles/all", handlers.GetAllRoles)
			auth.GET("/roles/:id", handlers.GetRole)
			auth.POST("/roles", handlers.CreateRole)
			auth.PUT("/roles/:id", handlers.UpdateRole)
			auth.DELETE("/roles/:id", handlers.DeleteRole)
			auth.PUT("/roles/:id/permissions", handlers.AssignRolePermissions)

			auth.GET("/permissions", handlers.ListPermissions)
			auth.GET("/files/:id/download", handlers.DownloadFile)
		}

		api.GET("/public/download/:id", handlers.PublicDownloadFile)
	}

	return r
}
