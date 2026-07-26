package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"os"
	"reflect"

	"probig/internal/pkg/config"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var EncryptKey []byte

func Init() *gorm.DB {
	dbPath := config.DBPath
	if dbPath == "" {
		dbPath = "./hr.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	config.DB = db
	return db
}

func InitEncryptKey(db *gorm.DB) {
	type SysConfig struct {
		ID          uint   `gorm:"primaryKey"`
		ConfigKey   string `gorm:"size:64;uniqueIndex"`
		ConfigValue string
		ConfigName  string
		ConfigDesc  string
		ValueType   string `gorm:"size:16"`
		OptionValues string
	}

	var c SysConfig
	err := db.Where("config_key = ?", "system.encrypt_key").First(&c).Error
	if err != nil {
		if config.EncryptKey != "" {
			EncryptKey = []byte(config.EncryptKey)
		} else {
			key := make([]byte, 32)
			io.ReadFull(rand.Reader, key)
			EncryptKey = key
		}
		db.Create(&SysConfig{
			ConfigKey:   "system.encrypt_key",
			ConfigValue: base64.StdEncoding.EncodeToString(EncryptKey),
			ConfigName:  "系统加密密钥",
			ConfigDesc:  "系统加密密钥，首次启动自动生成，请勿修改",
			ValueType:   "string",
		})
	} else {
		EncryptKey, err = base64.StdEncoding.DecodeString(c.ConfigValue)
		if err != nil || len(EncryptKey) < 16 {
			EncryptKey = make([]byte, 32)
			io.ReadFull(rand.Reader, EncryptKey)
			db.Model(&c).Update("config_value", base64.StdEncoding.EncodeToString(EncryptKey))
		}
	}

	os.MkdirAll(config.FileStoragePath, 0755)
}

func Encrypt(plaintext string) string {
	if plaintext == "" || len(EncryptKey) == 0 {
		return plaintext
	}
	block, err := aes.NewCipher(EncryptKey)
	if err != nil {
		return plaintext
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ct := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ct)
}

func Decrypt(ciphertext string) string {
	if ciphertext == "" || len(EncryptKey) == 0 {
		return ciphertext
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return ciphertext
	}
	block, err := aes.NewCipher(EncryptKey)
	if err != nil {
		return ciphertext
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertext
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return ciphertext
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ct, nil)
	if err != nil {
		return ciphertext
	}
	return string(plaintext)
}

var encryptFields = map[string][]string{}

func RegisterEncryptField(modelName, field string) {
	encryptFields[modelName] = append(encryptFields[modelName], field)
}

func EncryptStruct(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()
	modelName := rt.Name()
	fields, ok := encryptFields[modelName]
	if !ok {
		return
	}
	for _, fieldName := range fields {
		f := rv.FieldByName(fieldName)
		if f.IsValid() && f.Kind() == reflect.String {
			val := f.String()
			if val != "" {
				f.SetString(Encrypt(val))
			}
		}
	}
}

func DecryptStruct(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()
	modelName := rt.Name()
	fields, ok := encryptFields[modelName]
	if !ok {
		return
	}
	for _, fieldName := range fields {
		f := rv.FieldByName(fieldName)
		if f.IsValid() && f.Kind() == reflect.String {
			val := f.String()
			if val != "" {
				f.SetString(Decrypt(val))
			}
		}
	}
}
