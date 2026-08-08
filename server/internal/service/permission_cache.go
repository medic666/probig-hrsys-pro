package service

import (
	"sync"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
)

const permCacheTTL = 5 * time.Minute

type cachedPerms struct {
	keys     []string
	cachedAt time.Time
}

var permCache sync.Map // userID → cachedPerms

var scopeCache sync.Map // userID → cachedScope

type cachedScope struct {
	scope    string
	cachedAt time.Time
}

// GetUserEffectiveScope 用户有效数据范围 = 角色范围并集（最宽优先，与权限并集语义一致）：
// 任一角色为 all → all；否则 own；无任何角色 → all（无角色本就无权限）。
// 缓存与权限缓存同构（TTL + 角色变更主动失效）。
func GetUserEffectiveScope(userID uint) string {
	if v, ok := scopeCache.Load(userID); ok {
		c := v.(cachedScope)
		if time.Since(c.cachedAt) < permCacheTTL {
			return c.scope
		}
	}
	scope := dao.DataScopeAll
	var ownCount int64
	dao.DB.Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
		Where("roles.data_scope = ?", dao.DataScopeAll).
		Count(&ownCount)
	if ownCount == 0 {
		dao.DB.Table("roles").
			Joins("JOIN user_roles ON user_roles.role_id = roles.id").
			Where("user_roles.user_id = ? AND roles.deleted_at IS NULL", userID).
			Where("roles.data_scope = ?", dao.DataScopeOwn).
			Count(&ownCount)
		if ownCount > 0 {
			scope = dao.DataScopeOwn
		}
	}
	scopeCache.Store(userID, cachedScope{scope: scope, cachedAt: time.Now()})
	return scope
}

// GetUserPermissionKeys 带缓存的用户权限查询（TTL + 主动失效双保险）
func GetUserPermissionKeys(userID uint) []string {
	if v, ok := permCache.Load(userID); ok {
		c := v.(cachedPerms)
		if time.Since(c.cachedAt) < permCacheTTL {
			return c.keys
		}
	}
	var permissions []model.Permission
	dao.DB.Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&permissions)
	keys := make([]string, 0, len(permissions))
	for _, p := range permissions {
		keys = append(keys, p.Module+"."+p.Action)
	}
	permCache.Store(userID, cachedPerms{keys: keys, cachedAt: time.Now()})
	return keys
}

// InvalidateUserPermissionCache 权限变更后主动失效（角色分配/用户角色变更/角色删除）
func InvalidateUserPermissionCache(userID uint) {
	permCache.Delete(userID)
	scopeCache.Delete(userID)
}

// InvalidateRolePermissionCache 角色权限/数据范围变更后失效该角色全部关联用户
func InvalidateRolePermissionCache(roleID uint) {
	var userIDs []uint
	dao.DB.Table("user_roles").Where("role_id = ?", roleID).Pluck("user_id", &userIDs)
	for _, uid := range userIDs {
		InvalidateUserPermissionCache(uid)
	}
}
