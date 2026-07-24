package company

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	r.GET("/companies", authMiddleware, permMiddleware("company.read"), handler.List)
	r.GET("/companies/:id", authMiddleware, permMiddleware("company.read"), handler.GetByID)
	r.POST("/companies", authMiddleware, permMiddleware("company.write"), handler.Create)
	r.PUT("/companies/:id", authMiddleware, permMiddleware("company.write"), handler.Update)
	r.DELETE("/companies/:id", authMiddleware, permMiddleware("company.delete"), handler.Delete)
	r.PUT("/companies/:id/restore", authMiddleware, permMiddleware("company.write"), handler.Restore)
}
