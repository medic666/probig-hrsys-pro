package handler

import (
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetPermissions(c *gin.Context) {
	perms, err := service.GetAllPermissions()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	grouped := make(map[string][]map[string]interface{})
	for _, p := range perms {
		grouped[p.Module] = append(grouped[p.Module], map[string]interface{}{
			"id":   p.ID,
			"key":  p.Module + "." + p.Action,
			"name": p.Name,
		})
	}

	utils.Success(c, grouped)
}
