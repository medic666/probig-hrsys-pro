package dao

import (
	"fmt"

	"probig/server/internal/config"

	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 校验配置的数据库类型并建立连接：当前仅支持 sqlite（单文件、单二进制部署），
// 配置 type 显式校验，为未来扩展其它数据库预留语义。
func InitDB(path string) (*gorm.DB, error) {
	t := config.AppConfig.Database.Type
	if t != "" && t != "sqlite" {
		return nil, fmt.Errorf("不支持的数据库类型: %s（当前仅支持 sqlite）", t)
	}
	var err error
	DB, err = NewDB(path)
	return DB, err
}

func NewDB(path string) (*gorm.DB, error) {
	return gorm.Open(GetSQLiteDialector(path), &gorm.Config{})
}
