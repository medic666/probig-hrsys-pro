package service

import "time"

// stale.go 核算结果过期判定收敛入口：
// 列表/详情 status（calculated/data_changed）与徽章标橙（green/orange）语义同源——
// 任一派生层源的最后计算时间晚于核算结果时间，即视为过期（data_changed）。
// 各模块差异仅在于"源收集 SQL"，判定骨架统一由 IsStaleAfter 承担。

// LatestTime 取时间切片最大值（nil 忽略；空切片返回零值）
func LatestTime(times []*time.Time) time.Time {
	var max time.Time
	for _, t := range times {
		if t != nil && t.After(max) {
			max = *t
		}
	}
	return max
}

// IsStaleAfter 任一源时间晚于结果时间 → 结果已过期（data_changed）
func IsStaleAfter(lastCalcAt time.Time, sources ...[]*time.Time) bool {
	for _, s := range sources {
		if LatestTime(s).After(lastCalcAt) {
			return true
		}
	}
	return false
}
