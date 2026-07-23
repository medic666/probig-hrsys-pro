package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	JWT      JWTConfig      `yaml:"jwt"`
	Database DatabaseConfig `yaml:"database"`
	Upload   UploadConfig   `yaml:"upload"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type UploadConfig struct {
	Dir     string `yaml:"dir"`
	MaxSize int64  `yaml:"max_size"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "probig.db"
	}
	if cfg.Upload.Dir == "" {
		cfg.Upload.Dir = "uploads"
	}
	if cfg.Upload.MaxSize == 0 {
		cfg.Upload.MaxSize = 10 * 1024 * 1024
	}
	return cfg, nil
}
