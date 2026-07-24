package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"probig/internal/models"
)

var DB *gorm.DB

func Init(dbPath string) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := DB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		log.Fatalf("failed to enable foreign keys: %v", err)
	}

	autoMigrate()
}

func autoMigrate() {
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Entity{},
		&models.File{},
		&models.PersonnelEvent{},
		&models.OrganizationEvent{},
		&models.PersonnelSnapshot{},
		&models.OrganizationSnapshot{},
		&models.AttendanceEvent{},
		&models.AttendanceSummary{},
		&models.SalaryEvent{},
		&models.SalarySummary{},
		&models.FileAssociation{},
		&models.AuditLog{},
	); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}
}
