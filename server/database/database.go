package database

import (
	"probig/config"
	"probig/models"
	"probig/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	registerHooks()

	if err := DB.AutoMigrate(
		&models.Person{},
		&models.PersonPhone{},
		&models.PersonEmail{},
		&models.PersonBankCard{},
		&models.Company{},
		&models.File{},
		&models.FileRelation{},
		&models.PositionEvent{},
		&models.AttendanceEvent{},
		&models.SalaryEvent{},
		&models.PositionSnapshot{},
		&models.AttendanceSummary{},
		&models.SalarySummary{},
		&models.AuditLog{},
		&models.SysConfig{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
	); err != nil {
		panic("failed to migrate: " + err.Error())
	}

	seedAll()
}

func registerHooks() {
	encryptHook := func(db *gorm.DB) {
		if db.Statement.Schema == nil {
			return
		}
		for _, field := range db.Statement.Schema.Fields {
			if tag, ok := field.Tag.Lookup("encrypted"); ok && tag == "true" {
				v, zero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
				if zero {
					continue
				}
				val, ok := v.(string)
				if ok && val != "" {
					enc, err := utils.Encrypt(val, config.EncryptKeyBytes)
					if err == nil {
						field.Set(db.Statement.Context, db.Statement.ReflectValue, enc)
					}
				}
			}
		}
	}

	decryptHook := func(db *gorm.DB) {
		if db.Statement.Schema == nil {
			return
		}
		rv := db.Statement.ReflectValue
		if rv.Kind().String() != "struct" {
			return
		}
		for _, field := range db.Statement.Schema.Fields {
			if tag, ok := field.Tag.Lookup("encrypted"); ok && tag == "true" {
				v, zero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
				if zero {
					continue
				}
				val, ok := v.(string)
				if ok && val != "" {
					dec, err := utils.Decrypt(val, config.EncryptKeyBytes)
					if err == nil {
						field.Set(db.Statement.Context, db.Statement.ReflectValue, dec)
					}
				}
			}
		}
	}

	DB.Callback().Create().Before("gorm:create").Register("encrypt_before_create", encryptHook)
	DB.Callback().Update().Before("gorm:update").Register("encrypt_before_update", encryptHook)
	DB.Callback().Query().After("gorm:query").Register("decrypt_after_query", decryptHook)
}
