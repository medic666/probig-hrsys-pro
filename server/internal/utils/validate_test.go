package utils

import (
	"testing"
)

func TestValidateIDCard(t *testing.T) {
	tests := []struct {
		id     string
		expect bool
	}{
		{"110101199001011234", false},
		{"123456789012345678", false},
		{"", false},
		{"abcdefghijklmnopqr", false},
	}
	for _, tt := range tests {
		result := ValidateIDCard(tt.id)
		if result != tt.expect {
			t.Errorf("ValidateIDCard(%s) = %v, want %v", tt.id, result, tt.expect)
		}
	}
}

func TestValidatePhone(t *testing.T) {
	if !ValidatePhone("13800138000") {
		t.Fatal("13800138000 should be valid")
	}
	if ValidatePhone("12345678901") {
		t.Fatal("12345678901 should be invalid")
	}
	if ValidatePhone("") {
		t.Fatal("empty should be invalid")
	}
}

func TestValidateEmail(t *testing.T) {
	if !ValidateEmail("test@example.com") {
		t.Fatal("test@example.com should be valid")
	}
	if ValidateEmail("not-an-email") {
		t.Fatal("not-an-email should be invalid")
	}
}

func TestValidateNotEmpty(t *testing.T) {
	if ValidateNotEmpty("") {
		t.Fatal("empty string should be invalid")
	}
	if !ValidateNotEmpty("hello") {
		t.Fatal("hello should be valid")
	}
}
