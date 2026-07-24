package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"time"
)

func ParseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func ParseUint(s string) uint {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(v)
}

func ParseFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func IntToStr(i int) string {
	return strconv.Itoa(i)
}

func UintToStr(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

func FloatToStr(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

func NowTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func NowDateStr() string {
	return time.Now().Format("2006-01-02")
}

func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func GenBatchID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x%x", time.Now().UnixNano(), b)
}

func GenerateRandomKey(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

func RoundFloat(f float64, precision int) float64 {
	shift := math.Pow(10, float64(precision))
	return math.Round(f*shift) / shift
}

func GetMonthStartEnd(year, month int) (time.Time, time.Time) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, -1)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, time.Local)
	return start, end
}

func GetCurrentMonthEnd() time.Time {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, -1)
	return time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, time.Local)
}

func GetMonthsBetween(start, end time.Time) []string {
	var months []string
	year := start.Year()
	month := int(start.Month())
	endYear := end.Year()
	endMonth := int(end.Month())
	for year < endYear || (year == endYear && month <= endMonth) {
		months = append(months, fmt.Sprintf("%04d-%02d", year, month))
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
	return months
}

func StringPtr(s string) *string {
	return &s
}

func UintPtr(i uint) *uint {
	return &i
}

func FloatPtr(f float64) *float64 {
	return &f
}

func BoolPtr(b bool) *bool {
	return &b
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func IsEmptyStr(s *string) bool {
	return s == nil || *s == ""
}
