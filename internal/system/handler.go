package system

import (
	"fmt"

	"probig/internal/pkg/config"
	"probig/internal/pkg/middleware"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) GetAllConfigs() ([]SysConfig, error) {
	var configs []SysConfig
	err := s.DB.Find(&configs).Error
	return configs, err
}

func (s *Service) UpdateConfig(id uint, value string) error {
	var c SysConfig
	if err := s.DB.First(&c, id).Error; err != nil {
		return err
	}
	if c.ConfigKey == "system.encrypt_key" {
		return fmt.Errorf("系统加密密钥不可修改")
	}
	if err := s.DB.Model(&SysConfig{}).Where("id = ?", id).Update("config_value", value).Error; err != nil {
		return err
	}
	config.SetConfig(c.ConfigKey, value)
	return nil
}

func RegisterRoutes(r *gin.RouterGroup) {
	r.Use(middleware.Auth)
	r.GET("", middleware.Permission("system:read"), GetAll)
	r.PUT("/:id", middleware.Permission("system:write"), Update)
}

func GetAll(c *gin.Context) {
	svc := NewService()
	configs, err := svc.GetAllConfigs()
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	grouped := make(map[string][]SysConfig)
	for _, c := range configs {
		grouped[c.ConfigKey] = append(grouped[c.ConfigKey], c)
	}
	response.Success(c, configs)
}

func Update(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.UpdateConfig(id, req.Value); err != nil {
		response.Error(c, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}
