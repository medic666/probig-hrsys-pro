package config

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	FileStorage FileStorageConfig `yaml:"file_storage"`
	Jwt         JwtConfig         `yaml:"jwt"`
	Log         LogConfig         `yaml:"log"`
}

type ServerConfig struct {
	Port        int    `yaml:"port"`
	Mode        string `yaml:"mode"`
	CorsOrigins string `yaml:"cors_origins"`
}

type DatabaseConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

type FileStorageConfig struct {
	Path string `yaml:"path"`
}

type JwtConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	RetainDays int    `yaml:"retain_days"`
}

var AppConfig *Config

func GetExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func ResolvePath(relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(GetExeDir(), relPath)
}

func LoadConfig() (loadedFromFile bool) {
	configPath := filepath.Join(GetExeDir(), "config.yaml")

	cfg, ok := tryLoadFile(configPath)
	if ok {
		applyDefaults(cfg)
		AppConfig = cfg
		log.Printf("已加载配置文件: %s", configPath)
		return true
	}

	log.Println("未找到配置文件，使用默认配置")
	cfg = &Config{}
	applyDefaults(cfg)
	AppConfig = cfg
	return false
}

func WriteDefaultConfig() error {
	configPath := filepath.Join(GetExeDir(), "config.yaml")

	cfg := &Config{}
	applyDefaults(cfg)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}
	log.Printf("已生成默认配置文件: %s", configPath)
	return nil
}

func tryLoadFile(path string) (*Config, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Printf("配置文件 %s 解析失败: %v", path, err)
		return nil, false
	}
	return cfg, true
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/probig.db"
	}
	if cfg.FileStorage.Path == "" {
		cfg.FileStorage.Path = "./data/uploads"
	}
	if cfg.Jwt.Secret == "" {
		cfg.Jwt.Secret = "probig-jwt-secret-change-in-production"
	}
	if cfg.Jwt.ExpireHours <= 0 {
		cfg.Jwt.ExpireHours = 8
	}
	if cfg.Log.File == "" {
		cfg.Log.File = "./data/probig.log"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.RetainDays == 0 {
		cfg.Log.RetainDays = 30
	}
}
