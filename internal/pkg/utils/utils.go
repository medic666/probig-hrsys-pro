package utils

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DateFormat  = "2006-01-02"
	MonthFormat = "2006-01"
)

func ParseID(c *gin.Context, param string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func GetUserID(c *gin.Context) uint {
	val, exists := c.Get("userID")
	if !exists {
		return 0
	}
	id, ok := val.(uint)
	if !ok {
		return 0
	}
	return id
}

func GetUsername(c *gin.Context) string {
	val, exists := c.Get("username")
	if !exists {
		return ""
	}
	name, ok := val.(string)
	if !ok {
		return ""
	}
	return name
}

func TimeNow() time.Time {
	return time.Now()
}

func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

func FormatMonth(t time.Time) string {
	return t.Format(MonthFormat)
}

func StrPtr(s string) *string {
	return &s
}

func FloatToStr(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}
