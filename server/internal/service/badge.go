package service

import (
	"context"
	"fmt"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/utils"
)

// PersonBadge 人员徽章状态（卡片右上角颜色点）：
// gray 无数据 / green 正常 / orange 异常 / red 特定状态。
// 各模块徽章接口统一返回该结构，前端仅渲染，不参与运算。
type PersonBadge struct {
	PersonID uint   `json:"person_id"`
	Level    string `json:"level"`
}

// PersonBalance 人员数值型余额（如工资预支累计），meta 位徽章/展示通用
type PersonBalance struct {
	PersonID uint    `json:"person_id"`
	Balance  float64 `json:"balance"`
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

// GetAnnualLeaveEventBadges 年假事件徽章：当前月 == 员工入职周年月（每年同月且已满一周年）
// 且上月最后一日末年假余额快照 > 0（年假未被结转）→ orange 提醒；否则 green（点恒在体系）。
// 参数按 GORM 铁则分别传给所属子句方法（Select 的参数传 Select，Joins 的参数传 Joins）。
func GetAnnualLeaveEventBadges(ctx context.Context) ([]PersonBadge, error) {
	now := time.Now()
	prevMonthEnd := utils.DateOnlyFromTime(time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1))
	currentMonth := fmt.Sprintf("%02d", int(now.Month()))
	currentYear := fmt.Sprintf("%d", now.Year())

	var rows []struct {
		PersonID uint
		Level    string
	}
	db := dao.DBFromContext(ctx).Table("persons").
		Select(`persons.id AS person_id,
			CASE
				WHEN strftime('%m', s.entry_date) = ? AND CAST(strftime('%Y', s.entry_date) AS INTEGER) < ?
					AND b.balance_hours > 0 THEN 'orange'
				ELSE 'green'
			END AS level`, currentMonth, currentYear).
		Joins(`LEFT JOIN position_snapshots s ON s.person_id = persons.id AND s.effective_end_date = '9999-12-31'`).
		Joins(`LEFT JOIN annual_leave_balance_snapshots b ON b.person_id = persons.id
			AND b.effective_start_date <= ? AND b.effective_end_date >= ?`, prevMonthEnd, prevMonthEnd).
		Where("persons.deleted_at IS NULL")
	if pid, ok := dao.OwnPersonID(ctx); ok {
		db = db.Where("persons.id = ?", pid)
	}
	err := db.Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPersonBadges(rows), nil
}
