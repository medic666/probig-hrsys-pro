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

	return r
}
