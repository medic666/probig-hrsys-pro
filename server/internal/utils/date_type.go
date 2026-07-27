package utils

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type DateOnly time.Time

func (d DateOnly) MarshalJSON() ([]byte, error) {
	s := time.Time(d).Format("2006-01-02")
	return []byte(`"` + s + `"`), nil
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = DateOnly(t)
	return nil
}

func (d DateOnly) Value() (driver.Value, error) {
	return time.Time(d).Format("2006-01-02"), nil
}

func (d *DateOnly) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*d = DateOnly(v)
	case string:
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		*d = DateOnly(t)
	default:
		return fmt.Errorf("cannot scan %T into DateOnly", value)
	}
	return nil
}

func (d DateOnly) Time() time.Time { return time.Time(d) }

func (d DateOnly) String() string { return time.Time(d).Format("2006-01-02") }

func (d DateOnly) IsZero() bool { return time.Time(d).IsZero() }

func (d DateOnly) Equal(other DateOnly) bool { return time.Time(d).Equal(time.Time(other)) }

func (d DateOnly) Before(other DateOnly) bool { return time.Time(d).Before(time.Time(other)) }

func (d DateOnly) AddDate(years, months, days int) DateOnly {
	return DateOnly(time.Time(d).AddDate(years, months, days))
}

func DateOnlyFromTime(t time.Time) DateOnly { return DateOnly(t) }
