package utils

import "math"

func RoundTwoDecimal(value float64) float64 {
	return math.Round(value*100) / 100
}

func DecimalAdd(a, b float64) float64 {
	return RoundTwoDecimal(a + b)
}

func DecimalSub(a, b float64) float64 {
	return RoundTwoDecimal(a - b)
}

func DecimalMul(a, b float64) float64 {
	return RoundTwoDecimal(a * b)
}

func DecimalDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return RoundTwoDecimal(a / b)
}
