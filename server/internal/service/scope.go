package service

import (
	"context"
	"errors"

	"probig/server/internal/dao"

	"gorm.io/gorm"
)

// 人员维度数据范围统一助手：
// 查询侧 OwnFilter 追加 own 过滤；写操作侧 EnsureOwnPerson 校验归属。
// 全局模块（company/file/audit/user/role/system_config）不参与人员维度。

// OwnFilter 查询按数据范围过滤：own 时追加 personCol = 本人（personCol 通常为 person_id）。
func OwnFilter(ctx context.Context, tx *gorm.DB, personCol string) *gorm.DB {
	if pid, ok := dao.OwnPersonID(ctx); ok {
		return tx.Where(personCol+" = ?", pid)
	}
	return tx
}

// EnsureOwnPerson 写操作归属校验：own 用户操作他人数据 → 拒绝
func EnsureOwnPerson(ctx context.Context, personID uint) error {
	if pid, ok := dao.OwnPersonID(ctx); ok && pid != personID {
		return errors.New("无权操作他人数据")
	}
	return nil
}

// RequireAllScope 全局性操作校验：own 范围拒绝（如公司级年假结转）
func RequireAllScope(ctx context.Context, action string) error {
	if _, ok := dao.OwnPersonID(ctx); ok {
		return errors.New(action + "为全局操作，需要全部人员数据范围")
	}
	return nil
}

// ScopePersonIDs 批量操作人选过滤：仅自己范围时只保留本人（不在名单中则返回空）
func ScopePersonIDs(ctx context.Context, personIDs []uint) []uint {
	if pid, ok := dao.OwnPersonID(ctx); ok {
		for _, p := range personIDs {
			if p == pid {
				return []uint{pid}
			}
		}
		return nil
	}
	return personIDs
}
