package handler

import (
	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetAllConfigs(c *gin.Context) {
	list, err := service.GetAllSysConfigs()
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func UpdateConfig(c *gin.Context) {
	var config models.SysConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if err := service.UpdateSysConfig(c, &config); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, config)
}
