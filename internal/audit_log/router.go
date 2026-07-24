package audit_log

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())
	audit := r.Group("/audit-logs")
	audit.Use(authMiddleware)
	audit.GET("", permMiddleware("audit.read"), handler.List)
}
