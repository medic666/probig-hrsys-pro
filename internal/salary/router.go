package salary

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	events := r.Group("/salary-events")
	events.Use(authMiddleware)
	events.GET("", permMiddleware("salary.read"), handler.ListEvents)
	events.POST("", permMiddleware("salary.write"), handler.CreateEvent)
	events.PUT("/:id", permMiddleware("salary.write"), handler.UpdateEvent)
	events.DELETE("/:id", permMiddleware("salary.write"), handler.DeleteEvent)

	summaries := r.Group("/salary-summaries")
	summaries.Use(authMiddleware)
	summaries.GET("", permMiddleware("salary.read"), handler.ListSummaries)
	summaries.GET("/detail", permMiddleware("salary.read"), handler.GetSummary)
	summaries.POST("/calculate", permMiddleware("salary.write"), handler.CalculateSalary)
	summaries.POST("/batch-calculate", permMiddleware("salary.write"), handler.CalculateSalaries)
	summaries.PUT("/lock", permMiddleware("salary.write"), handler.LockSummary)
	summaries.PUT("/unlock", permMiddleware("salary.write"), handler.UnlockSummary)
}
