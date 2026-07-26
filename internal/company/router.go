package company

import (
	"github.com/gin-gonic/gin"
	"probig/internal/pkg/middleware"
)

func RegisterRoutes(api *gin.RouterGroup) {
	h := NewHandler()

	companies := api.Group("/companies")
	{
		companies.GET("", middleware.PermissionMiddleware("company:read"), h.ListCompanies)
		companies.POST("", middleware.PermissionMiddleware("company:write"), h.CreateCompany)
		companies.GET("/trash", middleware.PermissionMiddleware("company:read"), h.ListTrashCompanies)
		companies.GET("/:id", middleware.PermissionMiddleware("company:read"), h.GetCompany)
		companies.PUT("/:id", middleware.PermissionMiddleware("company:write"), h.UpdateCompany)
		companies.DELETE("/:id", middleware.PermissionMiddleware("company:delete"), h.DeleteCompany)
		companies.PUT("/:id/restore", middleware.PermissionMiddleware("company:write"), h.RestoreCompany)
	}
}
