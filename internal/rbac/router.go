package rbac

import (
	"github.com/gin-gonic/gin"
	"probig/internal/pkg/middleware"
)

func RegisterRoutes(api *gin.RouterGroup) {
	h := NewHandler()

	auth := api.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/change-password", h.ChangePassword)
		auth.GET("/user-info", h.GetUserInfo)
	}

	users := api.Group("/users")
	{
		users.GET("", middleware.PermissionMiddleware("rbac:read"), h.ListUsers)
		users.POST("", middleware.PermissionMiddleware("rbac:write"), h.CreateUser)
		users.PUT("/:id", middleware.PermissionMiddleware("rbac:write"), h.UpdateUser)
		users.DELETE("/:id", middleware.PermissionMiddleware("rbac:delete"), h.DeleteUser)
		users.GET("/:id/roles", middleware.PermissionMiddleware("rbac:read"), h.GetUserRoles)
		users.PUT("/:id/roles", middleware.PermissionMiddleware("rbac:write"), h.AssignUserRoles)
		users.PUT("/:id/reset-password", middleware.PermissionMiddleware("rbac:write"), h.ResetPassword)
	}

	roles := api.Group("/roles")
	{
		roles.GET("", middleware.PermissionMiddleware("rbac:read"), h.ListRoles)
		roles.POST("", middleware.PermissionMiddleware("rbac:write"), h.CreateRole)
		roles.PUT("/:id", middleware.PermissionMiddleware("rbac:write"), h.UpdateRole)
		roles.DELETE("/:id", middleware.PermissionMiddleware("rbac:delete"), h.DeleteRole)
		roles.GET("/:id/permissions", middleware.PermissionMiddleware("rbac:read"), h.GetRolePermissions)
		roles.PUT("/:id/permissions", middleware.PermissionMiddleware("rbac:write"), h.AssignRolePermissions)
	}

	api.GET("/permissions", middleware.PermissionMiddleware("rbac:read"), h.ListPermissions)
}
