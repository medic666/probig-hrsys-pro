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
// 安全约束：own 范围但未关联人员 → 查询过滤为空、写操作拒绝（防止漏成全量）。

// OwnFilter 查询按数据范围过滤：own 时追加 personCol = 本人；
// own 且未关联人员 → 恒空过滤（WHERE 1=0），防泄漏。
func OwnFilter(ctx context.Context, tx *gorm.DB, personCol string) *gorm.DB {
	info := dao.ScopeFromContext(ctx)
	if info.DataScope != dao.DataScopeOwn {
		return tx
	}
	if info.PersonID == nil {
		return tx.Where("1 = 0")
	}
	return tx.Where(personCol+" = ?", *info.PersonID)
}

// EnsureOwnPerson 写操作归属校验：own 用户操作他人数据 → 拒绝；
// own 且未关联人员 → 拒绝（无可见数据，无从操作）
func EnsureOwnPerson(ctx context.Context, personID uint) error {
	info := dao.ScopeFromContext(ctx)
	if info.DataScope != dao.DataScopeOwn {
		return nil
	}
	if info.PersonID == nil {
		return errors.New("当前账号未关联人员，无法操作")
	}
	if *info.PersonID != personID {
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

// ScopePersonIDs 批量操作人选过滤：仅自己范围时只保留本人（不在名单中则返回空）；
// own 且未关联人员 → 空（无可操作对象）
func ScopePersonIDs(ctx context.Context, personIDs []uint) []uint {
	info := dao.ScopeFromContext(ctx)
	if info.DataScope != dao.DataScopeOwn {
		return personIDs
	}
	if info.PersonID == nil {
		return nil
	}
	for _, p := range personIDs {
		if p == *info.PersonID {
			return []uint{*info.PersonID}
		}
	}
	return nil
}
