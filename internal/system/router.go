package system

import (
	"probig/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	configs := rg.Group("/system-configs")
	{
		configs.GET("", middleware.PermissionMiddleware("system:read"), ListConfigsHandler)
		configs.PUT("/:id", middleware.PermissionMiddleware("system:write"), UpdateConfigHandler)
	}
}
