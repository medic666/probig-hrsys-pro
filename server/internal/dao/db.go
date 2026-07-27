package dao

import (
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(path string) (*gorm.DB, error) {
	var err error
	DB, err = NewDB(path)
	return DB, err
}

func NewDB(path string) (*gorm.DB, error) {
	return gorm.Open(GetSQLiteDialector(path), &gorm.Config{})
}
