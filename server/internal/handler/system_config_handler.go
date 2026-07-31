package handler

import (
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetSystemConfigs(c *gin.Context) {
	configs, err := service.GetSystemConfigItems()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, configs)
}

type updateConfigReq struct {
	Value string `json:"value" binding:"required"`
}

func UpdateSystemConfig(c *gin.Context) {
	key := c.Param("key")
	var req updateConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.SetConfig(c.Request.Context(), key, req.Value); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "配置已更新", nil)
}
