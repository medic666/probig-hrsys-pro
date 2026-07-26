package projection

import "time"

const (
	StatusNotCalc    = 0
	StatusUpToDate   = 1
	StatusOutdated   = 2
)

func GetStatus(lastCalcAt *time.Time, sourceUpdatedAt *time.Time) int {
	if lastCalcAt == nil || lastCalcAt.IsZero() {
		return StatusNotCalc
	}
	if sourceUpdatedAt != nil && sourceUpdatedAt.After(*lastCalcAt) {
		return StatusOutdated
	}
	return StatusUpToDate
}

func MaxTime(a, b *time.Time) *time.Time {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.After(*b) {
		return a
	}
	return b
}
