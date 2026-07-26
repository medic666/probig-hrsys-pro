package database

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"probig/internal/pkg/encrypt"
)

var DB *gorm.DB
var encryptKey string

type SysConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ConfigKey    string    `gorm:"uniqueIndex;size:64;not null" json:"config_key"`
	ConfigValue  string    `gorm:"type:text" json:"config_value"`
	ConfigName   string    `gorm:"size:128" json:"config_name"`
	ConfigDesc   string    `gorm:"size:256" json:"config_desc"`
	ValueType    string    `gorm:"size:16;default:string" json:"value_type"`
	OptionValues string    `gorm:"type:text" json:"option_values"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func Init(dbPath string) (*gorm.DB, error) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := DB.AutoMigrate(&SysConfig{}); err != nil {
		return nil, fmt.Errorf("failed to migrate sys_config: %w", err)
	}

	initEncryptKey()

	RegisterEncryptHooks(DB)

	AutoMigrateAll(DB)

	return DB, nil
}

func initEncryptKey() {
	var cfg SysConfig
	result := DB.Where("config_key = ?", "system.encrypt_key").First(&cfg)
	if result.Error != nil {
		key, err := encrypt.GenerateRandomKey()
		if err != nil {
			log.Printf("failed to generate encrypt key: %v", err)
			return
		}
		cfg = SysConfig{
			ConfigKey:   "system.encrypt_key",
			ConfigValue: key,
			ConfigName:  "系统加密密钥",
			ConfigDesc:  "系统自动生成的加密密钥，用于敏感字段加解密，禁止手动修改",
			ValueType:   "string",
		}
		if err := DB.Create(&cfg).Error; err != nil {
			log.Printf("failed to save encrypt key: %v", err)
			return
		}
	}
	encryptKey = cfg.ConfigValue
}

func GetEncryptKey() string {
	return encryptKey
}

func encryptField(field reflect.Value, key string) error {
	if field.Kind() != reflect.String {
		return nil
	}
	plaintext := field.String()
	if plaintext == "" {
		return nil
	}
	if isEncrypted(plaintext) {
		return nil
	}
	encrypted, err := encrypt.Encrypt(plaintext, key)
	if err != nil {
		return err
	}
	field.SetString(encrypted)
	return nil
}

func decryptField(field reflect.Value, key string) error {
	if field.Kind() != reflect.String {
		return nil
	}
	ciphertext := field.String()
	if ciphertext == "" {
		return nil
	}
	if !isEncrypted(ciphertext) {
		return nil
	}
	plaintext, err := encrypt.Decrypt(ciphertext, key)
	if err != nil {
		return err
	}
	field.SetString(plaintext)
	return nil
}

func isEncrypted(s string) bool {
	if len(s) < 24 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

func RegisterEncryptHooks(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("encrypt_before_create", func(db *gorm.DB) {
		if db.Statement.Schema == nil {
			return
		}
		processEncryptFields(db.Statement.ReflectValue, encryptField, db)
	})

	db.Callback().Update().Before("gorm:update").Register("encrypt_before_update", func(db *gorm.DB) {
		if db.Statement.Schema == nil || db.Statement.ReflectValue.Kind() != reflect.Struct {
			return
		}
		processEncryptFields(db.Statement.ReflectValue, encryptField, db)
	})

	db.Callback().Query().After("gorm:query").Register("encrypt_after_find", func(db *gorm.DB) {
		if db.Statement.Schema == nil {
			return
		}
		processEncryptFields(db.Statement.ReflectValue, decryptField, db)
	})
}

func processEncryptFields(reflectValue reflect.Value, processor func(reflect.Value, string) error, db *gorm.DB) {
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	if reflectValue.Kind() == reflect.Slice {
		for i := 0; i < reflectValue.Len(); i++ {
			elem := reflectValue.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() == reflect.Struct {
				processStructFields(elem, processor, db)
			}
		}
		return
	}
	if reflectValue.Kind() == reflect.Struct {
		processStructFields(reflectValue, processor, db)
	}
}

func processStructFields(structValue reflect.Value, processor func(reflect.Value, string) error, db *gorm.DB) {
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		encryptTag, ok := fieldType.Tag.Lookup("encrypt")
		if ok && encryptTag == "true" {
			if field.CanSet() {
				if err := processor(field, encryptKey); err != nil {
					_ = db.AddError(err)
				}
			}
		}
	}
}
