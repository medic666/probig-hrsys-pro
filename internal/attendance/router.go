package attendance

import (
	"github.com/gin-gonic/gin"
	"probig/internal/pkg/middleware"
)

func RegisterRoutes(r *gin.RouterGroup) {
	InitDB()

	events := r.Group("/attendance-events")
	{
		events.GET("", middleware.PermissionMiddleware("attendance:read"), ListEventsHandler)
		events.GET("/:id", middleware.PermissionMiddleware("attendance:read"), GetEventHandler)
		events.POST("", middleware.PermissionMiddleware("attendance:write"), CreateEventHandler)
		events.PUT("/:id", middleware.PermissionMiddleware("attendance:write"), UpdateEventHandler)
		events.DELETE("/:id", middleware.PermissionMiddleware("attendance:delete"), DeleteEventHandler)
		events.GET("/trash", middleware.PermissionMiddleware("attendance:read"), ListTrashHandler)
		events.PUT("/:id/restore", middleware.PermissionMiddleware("attendance:write"), RestoreEventHandler)
		events.POST("/batch", middleware.PermissionMiddleware("attendance:write"), BatchCreateHandler)
		events.POST("/cross-day", middleware.PermissionMiddleware("attendance:write"), CrossDayCreateHandler)
	}

	daily := r.Group("/attendance-daily")
	{
		daily.GET("", middleware.PermissionMiddleware("attendance:read"), ListDailyHandler)
	}

	monthly := r.Group("/attendance-monthly")
	{
		monthly.GET("", middleware.PermissionMiddleware("attendance:read"), ListMonthlyHandler)
		monthly.POST("/calc", middleware.PermissionMiddleware("attendance:calc"), CalcMonthlyHandler)
		monthly.GET("/export", middleware.PermissionMiddleware("attendance:export"), ExportMonthlyHandler)
	}
}
