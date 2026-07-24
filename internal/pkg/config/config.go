package config

import (
	"os"
	"sync"
)

type Config struct {
	DBPath      string
	JwtSecret   string
	EncryptKey  string
	ServerPort  string
	SysConfigs  map[string]string
}

type SysConfigItem struct {
	ConfigKey   string
	ConfigValue string
}

var (
	cfg  Config
	mu   sync.RWMutex
	once sync.Once
)

func Load() Config {
	once.Do(func() {
		cfg.DBPath = envOrDefault("DB_PATH", "./hr.db")
		cfg.JwtSecret = envOrDefault("JWT_SECRET", "")
		cfg.EncryptKey = envOrDefault("ENCRYPT_KEY", "")
		cfg.ServerPort = envOrDefault("SERVER_PORT", "8080")
		cfg.SysConfigs = make(map[string]string)
	})
	return cfg
}

func GetSysConfig(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.SysConfigs[key]
}

func SetSysConfig(key, value string) {
	mu.Lock()
	defer mu.Unlock()
	cfg.SysConfigs[key] = value
}

func LoadSysConfigs(items []SysConfigItem) {
	mu.Lock()
	defer mu.Unlock()
	for _, item := range items {
		cfg.SysConfigs[item.ConfigKey] = item.ConfigValue
	}
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
