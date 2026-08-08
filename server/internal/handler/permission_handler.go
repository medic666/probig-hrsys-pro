package handler

import (
	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

type permActionItem struct {
	ID   uint   `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type permModuleItem struct {
	Module  string           `json:"module"`
	Name    string           `json:"name"`
	Actions []permActionItem `json:"actions"`
}

// GetPermissions 权限清单：按 ModuleActions 定义顺序输出，Name 为中文模块名，
// 供角色权限分配页按「模块 × 动作」表格渲染。
func GetPermissions(c *gin.Context) {
	perms, err := service.GetAllPermissions()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	byModule := make(map[string][]permActionItem)
	for _, p := range perms {
		byModule[p.Module] = append(byModule[p.Module], permActionItem{
			ID: p.ID, Key: p.Module + "." + p.Action, Name: p.Name,
		})
	}

	result := make([]permModuleItem, 0, len(model.ModuleActions))
	for _, mod := range model.ModuleActions {
		if actions, ok := byModule[mod.Module]; ok {
			result = append(result, permModuleItem{Module: mod.Module, Name: mod.Name, Actions: actions})
		}
	}
	utils.Success(c, result)
}
