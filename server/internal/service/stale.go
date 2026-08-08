package service

import (
	"time"

	"probig/server/internal/dao"
)

// stale.go 核算结果过期判定收敛入口：
// 列表/详情 status（calculated/data_changed）与徽章标橙（green/orange）语义同源——
// 任一派生层源的最后计算/变更时间晚于核算结果时间，即视为过期（data_changed）。
// 取数口径由 StaleSource 声明式源定义统一：行级判定（RowDataChanged）与批量徽章
// （PersonLatestTimes）共用同一份源定义，杜绝行级/批量口径分叉；NULL 处理
// （Nullable 列自动过滤 IS NOT NULL）与错误检查统一收敛进引擎。

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

// StaleSource 派生层"最后变更时间"数据源定义。
// 引擎统一处理：person_id 按需追加、Nullable 列自动过滤 IS NOT NULL
// （规避 Pluck/Scan NULL 报错）、错误检查。行级判定与批量徽章共用同一份源定义。
type StaleSource struct {
	Model    interface{}
	Column   string
	Where    string
	Args     []interface{}
	Unscoped bool
	Nullable bool
}

// sourceTimeRow 批量取数行：人员 + 源时间
type sourceTimeRow struct {
	PersonID uint
	At       time.Time
}

// RowDataChanged 行级 stale 判定：任一源最后时间晚于结果时间 → true；
// 查询失败返回 error（调用方保守标 data_changed）
func RowDataChanged(lastCalcAt time.Time, personID uint, sources []StaleSource) (bool, error) {
	for _, s := range sources {
		times, err := sourceTimes(s, personID)
		if err != nil {
			return false, err
		}
		if LatestTime(times).After(lastCalcAt) {
			return true, nil
		}
	}
	return false, nil
}

// PersonLatestTimes 批量 stale 判定：按人员聚合各源最后时间（一次查询/源，消除 N+1）
func PersonLatestTimes(sources []StaleSource) (map[uint]time.Time, error) {
	latest := make(map[uint]time.Time)
	for _, s := range sources {
		rows, err := sourceTimesByPerson(s)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			if r.At.After(latest[r.PersonID]) {
				latest[r.PersonID] = r.At
			}
		}
	}
	return latest, nil
}

// sourceTimes 行级取数（单人）
func sourceTimes(s StaleSource, personID uint) ([]*time.Time, error) {
	db := dao.DB.Model(s.Model)
	if s.Unscoped {
		db = db.Unscoped()
	}
	q := db.Where(s.Where, s.Args...)
	if personID > 0 {
		q = q.Where("person_id = ?", personID)
	}
	if s.Nullable {
		q = q.Where(s.Column + " IS NOT NULL")
	}
	var times []*time.Time
	err := q.Pluck(s.Column, &times).Error
	return times, err
}

// sourceTimesByPerson 批量取数（按人员分组）
func sourceTimesByPerson(s StaleSource) ([]sourceTimeRow, error) {
	db := dao.DB.Model(s.Model)
	if s.Unscoped {
		db = db.Unscoped()
	}
	q := db.Select("person_id, "+s.Column+" AS at").Where(s.Where, s.Args...)
	if s.Nullable {
		q = q.Where(s.Column + " IS NOT NULL")
	}
	var rows []sourceTimeRow
	err := q.Scan(&rows).Error
	return rows, err
}
