package utils

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(1, "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}
}

func TestParseToken(t *testing.T) {
	token, _ := GenerateToken(1, "admin")
	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("UserID = %d, want 1", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("Username = %s, want admin", claims.Username)
	}
}

func TestParseInvalidToken(t *testing.T) {
	_, err := ParseToken("invalid-token")
	if err == nil {
		t.Fatal("ParseToken should fail for invalid token")
	}
}
