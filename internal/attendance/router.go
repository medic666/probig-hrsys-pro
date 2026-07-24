package attendance

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	events := r.Group("/attendance-events")
	events.Use(authMiddleware)
	events.GET("", permMiddleware("attendance.read"), handler.ListEvents)
	events.POST("", permMiddleware("attendance.write"), handler.CreateEvent)
	events.PUT("/:id", permMiddleware("attendance.write"), handler.UpdateEvent)
	events.DELETE("/:id", permMiddleware("attendance.write"), handler.DeleteEvent)
	events.POST("/cross-day", permMiddleware("attendance.write"), handler.CreateCrossDayEvents)
	events.POST("/batch", permMiddleware("attendance.write"), handler.BatchCreateEvents)

	summaries := r.Group("/attendance-summaries")
	summaries.Use(authMiddleware)
	summaries.GET("", permMiddleware("attendance.read"), handler.ListSummaries)
	summaries.GET("/detail", permMiddleware("attendance.read"), handler.GetSummary)
	summaries.POST("/calculate", permMiddleware("attendance.write"), handler.CalculateSummary)
	summaries.POST("/batch-calculate", permMiddleware("attendance.write"), handler.CalculateSummaries)
	summaries.PUT("/lock", permMiddleware("attendance.write"), handler.LockSummary)
	summaries.PUT("/unlock", permMiddleware("attendance.write"), handler.UnlockSummary)
}
