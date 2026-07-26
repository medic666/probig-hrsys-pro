package config

import (
	"os"
	"sync"

	"gorm.io/gorm"
)

var (
	DB            *gorm.DB
	ConfigCache   = make(map[string]string)
	ConfigCacheMu sync.RWMutex
)

var (
	DBPath          string
	JWTSecret       string
	ServerPort      string
	FileStoragePath string
	EncryptKey      string
)

func LoadEnv() {
	DBPath = getEnv("DB_PATH", "./hr.db")
	JWTSecret = getEnv("JWT_SECRET", "probig-secret-key-change-in-production")
	ServerPort = getEnv("SERVER_PORT", "8080")
	FileStoragePath = getEnv("FILE_STORAGE_PATH", "./upload")
	EncryptKey = getEnv("ENCRYPT_KEY", "")
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func GetConfig(key string) string {
	ConfigCacheMu.RLock()
	defer ConfigCacheMu.RUnlock()
	return ConfigCache[key]
}

func SetConfig(key, value string) {
	ConfigCacheMu.Lock()
	defer ConfigCacheMu.Unlock()
	ConfigCache[key] = value
}

func GetConfigNumber(key string) float64 {
	v := GetConfig(key)
	if v == "" {
		return 0
	}
	var f float64
	for _, c := range v {
		if c == '.' {
			continue
		}
		f = f*10 + float64(c-'0')
	}
	if v[0] != '.' {
		return f
	}
	return 0
}

func LoadConfigCache(db *gorm.DB) {
	type SysConfig struct {
		ConfigKey   string `gorm:"column:config_key"`
		ConfigValue string `gorm:"column:config_value"`
	}
	var configs []SysConfig
	if err := db.Table("sys_config").Select("config_key, config_value").Find(&configs).Error; err != nil {
		return
	}
	ConfigCacheMu.Lock()
	defer ConfigCacheMu.Unlock()
	for _, c := range configs {
		ConfigCache[c.ConfigKey] = c.ConfigValue
	}
}
