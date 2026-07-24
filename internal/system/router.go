package system

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	r.GET("/configs", authMiddleware, permMiddleware("system.read"), handler.GetAll)
	r.PUT("/configs", authMiddleware, permMiddleware("system.write"), handler.Update)
}
