package service

import (
	"probig/internal/dao"
	"probig/internal/middleware"
	"probig/internal/models"

	"github.com/gin-gonic/gin"
)

func GetAllSysConfigs() ([]models.SysConfig, error) {
	return dao.GetAllSysConfigs()
}

func UpdateSysConfig(c *gin.Context, config *models.SysConfig) error {
	old, err := dao.GetSysConfigByKey(config.ConfigKey)
	if err != nil {
		return dao.UpsertSysConfig(config)
	}
	if err := dao.UpdateSysConfig(config); err != nil {
		return err
	}
	RefreshConfigCache()
	middleware.RecordAudit(c, "配置修改", "sys_config", config.ID, old, config, "")
	return nil
}
