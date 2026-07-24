package position

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	events := r.Group("/position-events")
	events.Use(authMiddleware)
	events.GET("", permMiddleware("position.read"), handler.ListEvents)
	events.GET("/:id", permMiddleware("position.read"), handler.GetEventByID)
	events.POST("", permMiddleware("position.write"), handler.CreateEvent)
	events.PUT("/:id", permMiddleware("position.write"), handler.UpdateEvent)
	events.DELETE("/:id", permMiddleware("position.write"), handler.DeleteEvent)

	snapshots := r.Group("/position-snapshots")
	snapshots.Use(authMiddleware)
	snapshots.GET("", permMiddleware("position.read"), handler.GetSnapshot)
	snapshots.POST("/rebuild/:personId", permMiddleware("position.write"), handler.RebuildSnapshots)
}
