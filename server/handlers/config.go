package handlers

import (
	"strconv"

	"probig/middleware"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func ListConfigs(c *gin.Context) {
	configs := services.ListConfigs()
	utils.Success(c, configs)
}

func UpdateConfig(c *gin.Context) {
	var input struct {
		ConfigKey   string `json:"config_key"`
		ConfigValue string `json:"config_value"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	oldVal, _ := services.GetConfig(input.ConfigKey)
	if err := services.UpdateConfig(input.ConfigKey, input.ConfigValue); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "sys_config", 0, "配置修改",
		gin.H{"key": input.ConfigKey, "old": oldVal},
		gin.H{"key": input.ConfigKey, "new": input.ConfigValue})
	utils.Success(c, nil)
}

func ListAuditLogs(c *gin.Context) {
	operatorIDStr := c.Query("operator_id")
	var operatorID uint
	if operatorIDStr != "" {
		n, _ := strconv.ParseUint(operatorIDStr, 10, 64)
		operatorID = uint(n)
	}
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	logs, total, err := services.ListAuditLogs(
		operatorID,
		c.Query("target_type"),
		c.Query("action"),
		c.Query("batch_id"),
		c.Query("start_date"),
		c.Query("end_date"),
		offset, pageSize,
	)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, logs, total, page, pageSize)
}
