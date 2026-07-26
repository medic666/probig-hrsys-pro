package person

import (
	"github.com/gin-gonic/gin"
	"probig/internal/pkg/middleware"
)

func RegisterRoutes(api *gin.RouterGroup) {
	h := NewHandler()

	persons := api.Group("/persons")
	{
		persons.GET("", middleware.PermissionMiddleware("person:read"), h.ListPersons)
		persons.POST("", middleware.PermissionMiddleware("person:write"), h.CreatePerson)
		persons.GET("/trash", middleware.PermissionMiddleware("person:read"), h.ListTrashPersons)
		persons.GET("/:id", middleware.PermissionMiddleware("person:read"), h.GetPersonDetail)
		persons.PUT("/:id", middleware.PermissionMiddleware("person:write"), h.UpdatePerson)
		persons.DELETE("/:id", middleware.PermissionMiddleware("person:delete"), h.DeletePerson)
		persons.PUT("/:id/restore", middleware.PermissionMiddleware("person:write"), h.RestorePerson)

		persons.POST("/:id/phones", middleware.PermissionMiddleware("person:write"), h.CreatePhone)
		persons.PUT("/:id/phones/:phoneId", middleware.PermissionMiddleware("person:write"), h.UpdatePhone)
		persons.DELETE("/:id/phones/:phoneId", middleware.PermissionMiddleware("person:delete"), h.DeletePhone)

		persons.POST("/:id/emails", middleware.PermissionMiddleware("person:write"), h.CreateEmail)
		persons.PUT("/:id/emails/:emailId", middleware.PermissionMiddleware("person:write"), h.UpdateEmail)
		persons.DELETE("/:id/emails/:emailId", middleware.PermissionMiddleware("person:delete"), h.DeleteEmail)

		persons.POST("/:id/bankcards", middleware.PermissionMiddleware("person:write"), h.CreateBankCard)
		persons.PUT("/:id/bankcards/:cardId", middleware.PermissionMiddleware("person:write"), h.UpdateBankCard)
		persons.DELETE("/:id/bankcards/:cardId", middleware.PermissionMiddleware("person:delete"), h.DeleteBankCard)
	}
}
