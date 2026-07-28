package model

import "time"

type SysBatch struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	BatchNo        string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"batch_no"`
	BusinessType   string     `gorm:"type:varchar(32);not null" json:"business_type"`
	BusinessPeriod string     `gorm:"type:varchar(32)" json:"business_period"`
	OperatorID     uint       `json:"operator_id"`
	OperatorName   string     `gorm:"type:varchar(64)" json:"operator_name"`
	Status         int8       `gorm:"default:1" json:"status"`
	TotalCount     int        `gorm:"default:0" json:"total_count"`
	Remark         string     `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt      time.Time  `json:"created_at"`
	ExecutedAt     *time.Time `json:"executed_at"`
	CanceledAt     *time.Time `json:"canceled_at"`
}

func (SysBatch) TableName() string { return "sys_batches" }
