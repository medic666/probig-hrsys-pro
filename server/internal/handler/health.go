package handler

import (
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	utils.Success(c, nil)
}
