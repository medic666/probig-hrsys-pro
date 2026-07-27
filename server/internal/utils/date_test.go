package utils

import (
	"testing"
)

func TestParseDate(t *testing.T) {
	d, err := ParseDate("2025-01-15")
	if err != nil {
		t.Fatalf("ParseDate failed: %v", err)
	}
	if d.Year() != 2025 || d.Month() != 1 || d.Day() != 15 {
		t.Fatal("ParseDate returned wrong date")
	}
}

func TestMonthStartEnd(t *testing.T) {
	start, err := MonthStart("2025-02")
	if err != nil {
		t.Fatalf("MonthStart failed: %v", err)
	}
	if start.Day() != 1 {
		t.Fatal("MonthStart should be day 1")
	}

	end, err := MonthEnd("2025-02")
	if err != nil {
		t.Fatalf("MonthEnd failed: %v", err)
	}
	if end.Day() != 28 {
		t.Fatalf("MonthEnd for 2025-02 should be 28, got %d", end.Day())
	}
}

func TestDaysBetween(t *testing.T) {
	start, _ := ParseDate("2025-01-01")
	end, _ := ParseDate("2025-01-10")
	d := DaysBetween(start, end)
	if d != 9 {
		t.Fatalf("DaysBetween should be 9, got %d", d)
	}
}
