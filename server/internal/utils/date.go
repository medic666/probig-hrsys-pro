package utils

import (
	"time"
)

func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func ParseMonth(s string) (time.Time, error) {
	return time.Parse("2006-01", s)
}

func MonthStart(month string) (time.Time, error) {
	t, err := ParseMonth(month)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func MonthEnd(month string) (time.Time, error) {
	t, err := MonthStart(month)
	if err != nil {
		return time.Time{}, err
	}
	return t.AddDate(0, 1, -1), nil
}

func DaysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}
