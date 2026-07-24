package config

import "os"

type Config struct {
	DBPath     string
	JWTSecret  string
	EncryptKey string
	ServerPort string
}

func Load() *Config {
	return &Config{
		DBPath:     getEnv("DB_PATH", "./hr.db"),
		JWTSecret:  getEnv("JWT_SECRET", "probig-default-jwt-secret-key-2024"),
		EncryptKey: getEnv("ENCRYPT_KEY", "probig-encrypt-key-16b"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
