package config

import (
	"strconv"
	"sync"

	"probig/internal/pkg/database"
)

var (
	cache = make(map[string]string)
	mu    sync.RWMutex
)

func Init() error {
	return Reload()
}

func Reload() error {
	var configs []database.SysConfig
	if err := database.DB.Find(&configs).Error; err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	cache = make(map[string]string, len(configs))
	for _, c := range configs {
		cache[c.ConfigKey] = c.ConfigValue
	}
	return nil
}

func Get(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	return cache[key]
}

func Set(key, value string) error {
	if err := database.DB.Model(&database.SysConfig{}).Where("config_key = ?", key).Update("config_value", value).Error; err != nil {
		return err
	}
	mu.Lock()
	cache[key] = value
	mu.Unlock()
	return nil
}

func GetInt(key string) int {
	v := Get(key)
	if v == "" {
		return 0
	}
	i, _ := strconv.Atoi(v)
	return i
}

func GetFloat(key string) float64 {
	v := Get(key)
	if v == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func GetBool(key string) bool {
	v := Get(key)
	return v == "true" || v == "1"
}
