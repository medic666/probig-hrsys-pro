package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

var (
	DBPath      string
	JWTSecret   string
	EncryptKey  string
	ServerPort  string
	EncryptKeyBytes []byte
)

func genRandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Init() {
	DBPath = getEnv("DB_PATH", "./hr.db")
	JWTSecret = getEnv("JWT_SECRET", genRandomHex(32))
	EncryptKey = getEnv("ENCRYPT_KEY", genRandomHex(32))
	ServerPort = getEnv("SERVER_PORT", "8080")

	k, err := hex.DecodeString(EncryptKey)
	if err != nil || len(k) != 32 {
		EncryptKeyBytes = make([]byte, 32)
		rand.Read(EncryptKeyBytes)
	} else {
		EncryptKeyBytes = k
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
