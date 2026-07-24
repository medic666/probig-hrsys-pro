package person

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, permMiddleware func(string) gin.HandlerFunc) {
	handler := NewHandler(GetService())

	persons := r.Group("/persons")
	persons.GET("", authMiddleware, permMiddleware("person.read"), handler.ListPersons)
	persons.GET("/simple", authMiddleware, handler.GetAllPersonsSimple)
	persons.GET("/:id", authMiddleware, permMiddleware("person.read"), handler.GetPersonByID)
	persons.POST("", authMiddleware, permMiddleware("person.write"), handler.CreatePerson)
	persons.PUT("/:id", authMiddleware, permMiddleware("person.write"), handler.UpdatePerson)
	persons.DELETE("/:id", authMiddleware, permMiddleware("person.delete"), handler.DeletePerson)
	persons.PUT("/:id/restore", authMiddleware, permMiddleware("person.write"), handler.RestorePerson)

	persons.POST("/:id/phones", authMiddleware, permMiddleware("person.write"), handler.AddPhone)
	persons.PUT("/:id/phones/:phoneId", authMiddleware, permMiddleware("person.write"), handler.UpdatePhone)
	persons.DELETE("/:id/phones/:phoneId", authMiddleware, permMiddleware("person.write"), handler.DeletePhone)

	persons.POST("/:id/emails", authMiddleware, permMiddleware("person.write"), handler.AddEmail)
	persons.PUT("/:id/emails/:emailId", authMiddleware, permMiddleware("person.write"), handler.UpdateEmail)
	persons.DELETE("/:id/emails/:emailId", authMiddleware, permMiddleware("person.write"), handler.DeleteEmail)

	persons.POST("/:id/bank-cards", authMiddleware, permMiddleware("person.write"), handler.AddBankCard)
	persons.PUT("/:id/bank-cards/:cardId", authMiddleware, permMiddleware("person.write"), handler.UpdateBankCard)
	persons.DELETE("/:id/bank-cards/:cardId", authMiddleware, permMiddleware("person.write"), handler.DeleteBankCard)
}
