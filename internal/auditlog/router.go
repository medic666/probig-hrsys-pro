package auditlog

import (
	"probig/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	logs := rg.Group("/audit-logs")
	{
		logs.GET("", middleware.PermissionMiddleware("audit:read"), ListHandler)
		logs.GET("/:id", middleware.PermissionMiddleware("audit:read"), GetDetailHandler)
		logs.GET("/export", middleware.PermissionMiddleware("audit:export"), ExportHandler)
	}
}
