package rbac

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	auth := r.Group("/auth")
	auth.POST("/login", handler.Login)
	auth.GET("/me", authMiddleware, handler.GetCurrentUser)
	auth.POST("/change-password", authMiddleware, handler.ChangePassword)

	users := r.Group("/users")
	users.Use(authMiddleware)
	users.GET("", permMiddleware("rbac.read"), handler.ListUsers)
	users.POST("", permMiddleware("rbac.write"), handler.CreateUser)
	users.PUT("/:id", permMiddleware("rbac.write"), handler.UpdateUser)
	users.DELETE("/:id", permMiddleware("rbac.delete"), handler.DeleteUser)
	users.PUT("/:id/reset-password", permMiddleware("rbac.write"), handler.ResetPassword)
	users.PUT("/:id/roles", permMiddleware("rbac.write"), handler.AssignRoles)

	roles := r.Group("/roles")
	roles.Use(authMiddleware)
	roles.GET("", permMiddleware("rbac.read"), handler.ListRoles)
	roles.POST("", permMiddleware("rbac.write"), handler.CreateRole)
	roles.PUT("/:id", permMiddleware("rbac.write"), handler.UpdateRole)
	roles.DELETE("/:id", permMiddleware("rbac.delete"), handler.DeleteRole)
	roles.PUT("/:id/permissions", permMiddleware("rbac.write"), handler.AssignPermissions)

	perms := r.Group("/permissions")
	perms.Use(authMiddleware)
	perms.GET("", permMiddleware("rbac.read"), handler.GetAllPermissions)
}
