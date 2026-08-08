package service

import (
	"probig/server/internal/model"

	"gorm.io/gorm"
)

// SyncPermissionRows 按 ModuleActions 幂等补齐权限行（每次启动执行）：
// 新增业务模块只需改 model.ModuleActions 一处定义，存量库启动即自动出现新权限行，
// 无需手写迁移。与 20260808_01/03 迁移的补全步骤同构（迁移为一次性历史路径，此为常态机制）。
func SyncPermissionRows(db *gorm.DB) error {
	for _, mod := range model.ModuleActions {
		for _, action := range mod.Actions {
			var count int64
			if err := db.Model(&model.Permission{}).
				Where("module = ? AND action = ?", mod.Module, action).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				if err := db.Create(&model.Permission{
					Module: mod.Module, Action: action,
					Name: mod.Name + model.PermissionActionNames[action],
				}).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}
