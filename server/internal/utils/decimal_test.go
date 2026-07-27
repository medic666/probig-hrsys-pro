package utils

import (
	"math"
	"testing"
)

func TestRoundTwoDecimal(t *testing.T) {
	if RoundTwoDecimal(3.333) != 3.33 {
		t.Errorf("RoundTwoDecimal(3.333) = %f, want 3.33", RoundTwoDecimal(3.333))
	}
	if RoundTwoDecimal(3.336) != 3.34 {
		t.Errorf("RoundTwoDecimal(3.336) = %f, want 3.34", RoundTwoDecimal(3.336))
	}
}

func TestDecimalAdd(t *testing.T) {
	result := DecimalAdd(1.11, 2.22)
	if math.Abs(result-3.33) > 0.001 {
		t.Errorf("DecimalAdd = %f, want 3.33", result)
	}
}

func TestDecimalSub(t *testing.T) {
	result := DecimalSub(5.55, 2.22)
	if math.Abs(result-3.33) > 0.001 {
		t.Errorf("DecimalSub = %f, want 3.33", result)
	}
}

func TestDecimalMul(t *testing.T) {
	result := DecimalMul(1.5, 2.0)
	if math.Abs(result-3.0) > 0.001 {
		t.Errorf("DecimalMul = %f, want 3.0", result)
	}
}

func TestDecimalDiv(t *testing.T) {
	result := DecimalDiv(10.0, 3.0)
	if math.Abs(result-3.33) > 0.01 {
		t.Errorf("DecimalDiv = %f, want ~3.33", result)
	}
	if DecimalDiv(10.0, 0) != 0 {
		t.Fatal("divide by zero should return 0")
	}
}
