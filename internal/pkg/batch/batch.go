package batch

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type SysBatch struct {
	ID             uint       `gorm:"primaryKey"`
	BatchNo        string     `gorm:"size:64;uniqueIndex"`
	BusinessType   string     `gorm:"size:32;index"`
	BusinessPeriod string     `gorm:"size:32"`
	OperatorID     uint
	OperatorName   string     `gorm:"size:64"`
	Status         int        `gorm:"default:1"`
	TotalCount     int        `gorm:"default:0"`
	Remark         string     `gorm:"size:256"`
	CreatedAt      time.Time
	ExecutedAt     *time.Time
	CanceledAt     *time.Time
}

const (
	BatchStatusPending   = 1
	BatchStatusExecuted  = 2
	BatchStatusCanceled  = 3
	BatchStatusFailed    = 4
)

func GenerateBatchNo() string {
	now := time.Now().Format("20060102150405")
	r := rand.Intn(10000)
	return fmt.Sprintf("B%s%04d", now, r)
}

func CreateBatch(db *gorm.DB, businessType, businessPeriod string, operatorID uint, operatorName string) (*SysBatch, error) {
	b := &SysBatch{
		BatchNo:        GenerateBatchNo(),
		BusinessType:   businessType,
		BusinessPeriod: businessPeriod,
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		Status:         BatchStatusPending,
	}
	if err := db.Create(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

func ExecuteBatch(db *gorm.DB, batchID uint, totalCount int) error {
	now := time.Now()
	return db.Model(&SysBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
		"status":      BatchStatusExecuted,
		"total_count": totalCount,
		"executed_at": now,
	}).Error
}

func CancelBatch(db *gorm.DB, batchID uint) error {
	now := time.Now()
	return db.Model(&SysBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
		"status":      BatchStatusCanceled,
		"canceled_at": now,
	}).Error
}

func FailBatch(db *gorm.DB, batchID uint) error {
	return db.Model(&SysBatch{}).Where("id = ?", batchID).Update("status", BatchStatusFailed).Error
}

func GetBatch(db *gorm.DB, batchID uint) (*SysBatch, error) {
	var b SysBatch
	err := db.Where("id = ?", batchID).First(&b).Error
	return &b, err
}
