package file

import (
	"probig/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	files := rg.Group("/files")
	{
		files.GET("", middleware.PermissionMiddleware("file:read"), ListFilesHandler)
		files.POST("/upload", middleware.PermissionMiddleware("file:write"), UploadFileHandler)
		files.DELETE("/:id", middleware.PermissionMiddleware("file:delete"), DeleteFileHandler)
		files.PUT("/:id/restore", middleware.PermissionMiddleware("file:write"), RestoreFileHandler)
		files.GET("/trash", middleware.PermissionMiddleware("file:read"), ListTrashHandler)
		files.GET("/:id/download", middleware.PermissionMiddleware("file:read"), DownloadFileHandler)
		files.GET("/:id/relations", middleware.PermissionMiddleware("file:read"), GetRelationsHandler)
		files.PUT("/:id/relations", middleware.PermissionMiddleware("file:write"), UpdateRelationsHandler)
	}
}
