package batch

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

type SysBatch = database.SysBatch

const (
	StatusPending   int8 = 1
	StatusExecuted  int8 = 2
	StatusCanceled  int8 = 3
	StatusFailed    int8 = 4
)

func CreateBatch(tx *gorm.DB, businessType, businessPeriod string, operatorID uint, operatorName string, totalCount int, remark string) (*SysBatch, error) {
	batchNo := fmt.Sprintf("BATCH-%s-%s-%d", businessType, businessPeriod, time.Now().UnixMilli())

	b := &SysBatch{
		BatchNo:        batchNo,
		BusinessType:   businessType,
		BusinessPeriod: businessPeriod,
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		Status:         StatusPending,
		TotalCount:     totalCount,
		Remark:         remark,
	}

	db := tx
	if db == nil {
		db = database.DB
	}

	if err := db.Create(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

func ExecuteBatch(batchID uint) error {
	now := time.Now()
	return database.DB.Model(&SysBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
		"status":      StatusExecuted,
		"executed_at": now,
	}).Error
}

func CancelBatch(batchID uint) error {
	now := time.Now()
	return database.DB.Model(&SysBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
		"status":      StatusCanceled,
		"canceled_at": now,
	}).Error
}
