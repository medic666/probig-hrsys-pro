package utils

import (
	"strconv"
	"time"
)

func ParseUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

func ParseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func StrToUintPtr(s string) *uint {
	if s == "" {
		return nil
	}
	v := ParseUint(s)
	if v == 0 {
		return nil
	}
	return &v
}

func ParseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func DateStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func ParseDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse("2006-01-02", s)
	return t
}

const InfiniteDate = "9999-12-31"

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func UintPtr(v uint) *uint {
	return &v
}

func BoolPtr(v bool) *bool {
	return &v
}

func FloatPtr(v float64) *float64 {
	return &v
}

func IntPtr(v int) *int {
	return &v
}

func ContainsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
