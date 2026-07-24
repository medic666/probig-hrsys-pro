package dao

import (
	"gorm.io/gorm"
	"probig/pkg/crypto"
	"probig/pkg/database"
)

func encrypt(s string) string {
	if s == "" {
		return ""
	}
	result, err := crypto.Encrypt(s)
	if err != nil {
		return s
	}
	return result
}

func decrypt(s string) string {
	if s == "" {
		return ""
	}
	result, err := crypto.Decrypt(s)
	if err != nil {
		return s
	}
	return result
}

func DB() *gorm.DB {
	return database.DB
}
