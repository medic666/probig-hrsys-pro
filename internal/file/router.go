package file

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	r.POST("/files/upload", authMiddleware, permMiddleware("file.write"), handler.Upload)
	r.GET("/files/:id/download", authMiddleware, permMiddleware("file.read"), handler.Download)
	r.GET("/files/:id", authMiddleware, permMiddleware("file.read"), handler.GetFileInfo)
	r.GET("/files", authMiddleware, permMiddleware("file.read"), handler.ListFiles)
	r.DELETE("/files/:id", authMiddleware, permMiddleware("file.delete"), handler.DeleteFile)
	r.PUT("/files/:id/restore", authMiddleware, permMiddleware("file.write"), handler.RestoreFile)
	r.POST("/file-relations", authMiddleware, permMiddleware("file.write"), handler.AddRelation)
	r.DELETE("/file-relations/:id", authMiddleware, permMiddleware("file.write"), handler.RemoveRelation)
	r.GET("/files/by-target", authMiddleware, permMiddleware("file.read"), handler.GetFilesByTarget)
}
