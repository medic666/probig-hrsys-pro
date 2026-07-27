package utils

import (
	"regexp"
	"strings"
)

var (
	idCardPattern = regexp.MustCompile(`^\d{17}[\dXx]$`)
	phonePattern  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	emailPattern  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func ValidateIDCard(idCard string) bool {
	if !idCardPattern.MatchString(strings.ToUpper(idCard)) {
		return false
	}

	weight := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkMap := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	idCard = strings.ToUpper(idCard)
	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(idCard[i]-'0') * weight[i]
	}

	return idCard[17] == checkMap[sum%11]
}

func ValidatePhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

func ValidateEmail(email string) bool {
	return emailPattern.MatchString(email)
}

func ValidateNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}
