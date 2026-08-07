package dao

import (
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// GetSQLiteDialector SQLite 连接：_txlock=immediate 使事务以 BEGIN IMMEDIATE 开启——
// 写锁在事务首语句即获取，避免并发事务基于旧读快照覆盖重建结果（丢失更新）。
// 普通查询为单语句自动提交不受影响；SQLite 本就单写者，锁持有总时长不变。
func GetSQLiteDialector(path string) gorm.Dialector {
	dsn := path
	if strings.Contains(dsn, "?") {
		dsn += "&_txlock=immediate"
	} else {
		dsn += "?_txlock=immediate"
	}
	return sqlite.Open(dsn)
}
