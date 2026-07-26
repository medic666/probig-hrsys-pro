package system

import (
	"errors"

	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListConfigsHandler(c *gin.Context) {
	configs, err := ListAllConfigs()
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, configs)
}

func UpdateConfigHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	var req ConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	err = UpdateConfig(id, req.ConfigValue, operatorID, operatorName, clientIP)
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			response.Error(c, response.Forbidden, "系统加密密钥为只读项，禁止手动修改")
			return
		}
		response.Error(c, response.InternalError, "更新失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}
