package system

import (
	"github.com/gin-gonic/gin"

	"probig/internal/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) GetAll(c *gin.Context) {
	configs, err := h.service.GetAllConfigs()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	for i := range configs {
		if configs[i].ValueType != "select" {
			configs[i].OptionValues = ""
		}
	}

	response.Success(c, configs)
}

func (h *Handler) Update(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := h.service.UpdateConfig(req.Key, req.Value); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}
