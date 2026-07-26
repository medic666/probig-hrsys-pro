package position

import (
	"github.com/gin-gonic/gin"
	"probig/internal/pkg/middleware"
)

func RegisterRoutes(r *gin.RouterGroup) {
	InitDB()

	events := r.Group("/position-events")
	{
		events.GET("", middleware.PermissionMiddleware("position:read"), ListEventsHandler)
		events.GET("/:id", middleware.PermissionMiddleware("position:read"), GetEventHandler)
		events.POST("", middleware.PermissionMiddleware("position:write"), CreateEventHandler)
		events.PUT("/:id", middleware.PermissionMiddleware("position:write"), UpdateEventHandler)
		events.DELETE("/:id", middleware.PermissionMiddleware("position:delete"), DeleteEventHandler)
		events.GET("/trash", middleware.PermissionMiddleware("position:read"), ListTrashEventsHandler)
		events.PUT("/:id/restore", middleware.PermissionMiddleware("position:write"), RestoreEventHandler)
	}

	snapshots := r.Group("/position-snapshots")
	{
		snapshots.GET("", middleware.PermissionMiddleware("position:read"), ListSnapshotsHandler)
		snapshots.GET("/:id", middleware.PermissionMiddleware("position:read"), GetSnapshotHandler)
		snapshots.GET("/person/:personId/current", middleware.PermissionMiddleware("position:read"), GetCurrentSnapshotHandler)
		snapshots.GET("/person/:personId/status", middleware.PermissionMiddleware("position:read"), GetEmploymentStatusHandler)
	}
}
