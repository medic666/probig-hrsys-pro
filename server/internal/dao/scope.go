package dao

import "context"

// 数据范围常量（用户人员维度权限：全部 / 仅自己）
const (
	DataScopeAll = "all"
	DataScopeOwn = "own"
)

// ScopeInfo 请求级数据范围上下文，由中间件注入，随 context 贯穿整个请求。
// 人员维度权限：own（仅自己）时所有人员维度查询按 PersonID 过滤，写操作校验归属。
type ScopeInfo struct {
	UserID    uint
	DataScope string
	PersonID  *uint
}

type scopeCtxKey struct{}

// WithScopeInfo 将数据范围信息注入 context
func WithScopeInfo(ctx context.Context, info ScopeInfo) context.Context {
	return context.WithValue(ctx, scopeCtxKey{}, info)
}

// ScopeFromContext 从 context 读取数据范围信息，无则返回零值（视为全部范围）
func ScopeFromContext(ctx context.Context) ScopeInfo {
	if ctx == nil {
		return ScopeInfo{}
	}
	info, ok := ctx.Value(scopeCtxKey{}).(ScopeInfo)
	if !ok {
		return ScopeInfo{}
	}
	return info
}

// OwnPersonID 返回"仅自己"范围下可见的 person_id；
// own 且已关联人员时返回 (personID, true)，否则 (0, false)。
func OwnPersonID(ctx context.Context) (uint, bool) {
	info := ScopeFromContext(ctx)
	if info.DataScope != DataScopeOwn || info.PersonID == nil {
		return 0, false
	}
	return *info.PersonID, true
}
