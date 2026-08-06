package utils

import (
	"testing"
	"time"
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

func TestShouldRenew(t *testing.T) {
	ttl := JwtTTL()
	half := ttl / 2
	now := time.Now()

	if !ShouldRenew(now.Add(half - time.Minute)) {
		t.Error("remaining < half TTL should renew")
	}
	if ShouldRenew(now.Add(half + time.Minute)) {
		t.Error("remaining >= half TTL should not renew")
	}
	if !ShouldRenew(now.Add(-time.Minute)) {
		t.Error("expired token should renew")
	}
}
