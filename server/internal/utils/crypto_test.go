package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}
	if hash == "test123" {
		t.Fatal("hash should not equal plain text")
	}
}

func TestCheckPassword(t *testing.T) {
	hash, _ := HashPassword("admin123")
	if !CheckPassword("admin123", hash) {
		t.Fatal("CheckPassword should return true for correct password")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("CheckPassword should return false for wrong password")
	}
}
