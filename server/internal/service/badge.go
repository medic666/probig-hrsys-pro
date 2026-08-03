package service

import "time"

// PersonBadge 人员徽章状态（卡片右上角颜色点）：
// gray 无数据 / green 正常 / orange 异常 / red 特定状态。
// 各模块徽章接口统一返回该结构，前端仅渲染，不参与运算。
type PersonBadge struct {
	PersonID uint   `json:"person_id"`
	Level    string `json:"level"`
}

// DefaultBadgeMonth 徽章默认统计月份：当前时间的上一个月（考勤核算必核月份）
func DefaultBadgeMonth() string {
	return time.Now().AddDate(0, -1, 0).Format("2006-01")
}

// toPersonBadges 统一转换聚合查询结果（各徽章接口共用）
func toPersonBadges(rows []struct {
	PersonID uint
	Level    string
}) []PersonBadge {
	result := make([]PersonBadge, 0, len(rows))
	for _, r := range rows {
		result = append(result, PersonBadge{PersonID: r.PersonID, Level: r.Level})
	}
	return result
}
