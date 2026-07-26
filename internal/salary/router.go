package salary

import (
	"probig/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	events := rg.Group("/salary-events")
	{
		events.GET("", middleware.PermissionMiddleware("salary:read"), ListEventsHandler)
		events.GET("/:id", middleware.PermissionMiddleware("salary:read"), GetEventHandler)
		events.POST("", middleware.PermissionMiddleware("salary:write"), CreateEventHandler)
		events.PUT("/:id", middleware.PermissionMiddleware("salary:write"), UpdateEventHandler)
		events.DELETE("/:id", middleware.PermissionMiddleware("salary:delete"), DeleteEventHandler)
		events.GET("/trash", middleware.PermissionMiddleware("salary:read"), ListTrashHandler)
		events.PUT("/:id/restore", middleware.PermissionMiddleware("salary:write"), RestoreEventHandler)
	}

	summaries := rg.Group("/salary-summaries")
	{
		summaries.GET("", middleware.PermissionMiddleware("salary:read"), ListSummariesHandler)
		summaries.POST("/calc", middleware.PermissionMiddleware("salary:calc"), CalcSummaryHandler)
		summaries.GET("/:id", middleware.PermissionMiddleware("salary:read"), GetSummaryDetailHandler)
		summaries.GET("/export", middleware.PermissionMiddleware("salary:export"), ExportSummariesHandler)
	}
}
