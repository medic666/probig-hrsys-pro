package system

import (
	"probig/internal/pkg/database"
)

type SysConfig = database.SysConfig

type ConfigUpdateRequest struct {
	ConfigValue string `json:"config_value" binding:"required"`
}
