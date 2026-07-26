package projection

import "time"

type Status int

const (
	StatusNotCalculated Status = iota
	StatusCalculated
	StatusDataChanged
)

func (s Status) String() string {
	switch s {
	case StatusNotCalculated:
		return "未核算"
	case StatusCalculated:
		return "已核算"
	case StatusDataChanged:
		return "数据已变动"
	default:
		return "未知"
	}
}

func CheckProjectionStatus(lastCalcAt *time.Time, sourceUpdatedAt *time.Time) Status {
	if lastCalcAt == nil {
		return StatusNotCalculated
	}
	if sourceUpdatedAt != nil && sourceUpdatedAt.After(*lastCalcAt) {
		return StatusDataChanged
	}
	return StatusCalculated
}

func IsDataStale(lastCalcAt time.Time, sourceTimes ...time.Time) bool {
	for _, t := range sourceTimes {
		if t.After(lastCalcAt) {
			return true
		}
	}
	return false
}
