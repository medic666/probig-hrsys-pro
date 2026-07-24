package config

import "os"

type Config struct {
	Port      string
	DBPath    string
	JWTSecret string
	DevMode   bool
	UploadDir string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBPath:    getEnv("DB_PATH", "data/probig.db"),
		JWTSecret: getEnv("JWT_SECRET", "probig-jwt-secret-change-me"),
		DevMode:   os.Getenv("DEV_MODE") == "true",
		UploadDir: getEnv("UPLOAD_DIR", "data/uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
