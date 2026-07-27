package dao

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func GetSQLiteDialector(path string) gorm.Dialector {
	return sqlite.Open(path)
}
