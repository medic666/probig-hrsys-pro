package service

import (
	"context"
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreateSalaryEvent(ctx context.Context, event *model.SalaryEvent) error {
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		var maxSeq int
		tx.Unscoped().Model(&model.SalaryEvent{}).Where("person_id = ?", event.PersonID).
			Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
		event.Seq = maxSeq + 1
		return tx.Create(event).Error
	})
}

func UpdateSalaryEvent(ctx context.Context, id uint, event *model.SalaryEvent) error {
	var existing model.SalaryEvent
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("工资事件不存在")
	}
	updates := map[string]interface{}{
		"belong_month": event.BelongMonth,
		"event_type":   event.EventType,
		"amount":       event.Amount,
		"remark":       event.Remark,
	}
	return dao.DBFromContext(ctx).Model(&existing).Updates(updates).Error
}

func DeleteSalaryEvent(ctx context.Context, id uint) error {
	var event model.SalaryEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	return dao.DBFromContext(ctx).Delete(&event).Error
}

func RestoreSalaryEvent(ctx context.Context, id uint) error {
	var event model.SalaryEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	return dao.DBFromContext(ctx).Unscoped().Model(&event).Update("deleted_at", nil).Error
}

func GetSalaryEvent(id uint) (*model.SalaryEvent, error) {
	var event model.SalaryEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// SalaryEventListQuery 工资事件列表查询（列表与导出共用）
type SalaryEventListQuery struct {
	PageNum     int
	PageSize    int
	PersonID    uint
	BelongMonth string
	EventType   string
}

func GetSalaryEventList(q SalaryEventListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.SalaryEvent{})
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}
	if q.BelongMonth != "" {
		tx = tx.Where("belong_month = ?", q.BelongMonth)
	}
	if q.EventType != "" {
		tx = tx.Where("event_type = ?", q.EventType)
	}
	var total int64
	tx.Count(&total)
	var events []model.SalaryEvent
	offset := (q.PageNum - 1) * q.PageSize
	tx.Offset(offset).Limit(q.PageSize).Order("belong_month DESC, seq DESC").Find(&events)
	ids := make([]uint, len(events))
	for i, e := range events {
		ids[i] = e.PersonID
	}
	nameMap := PersonNameMap(ids)

	var result []map[string]interface{}
	for _, e := range events {
		item := map[string]interface{}{
			"id":           e.ID,
			"person_id":    e.PersonID,
			"seq":          e.Seq,
			"belong_month": e.BelongMonth,
			"event_type":   e.EventType,
			"amount":       e.Amount,
			"remark":       e.Remark,
			"created_at":   e.CreatedAt,
		}
		item["person_name"] = nameMap[e.PersonID]
		result = append(result, item)
	}
	return result, total, nil
}

func GetDeletedSalaryEvents(pageNum, pageSize int) ([]model.SalaryEvent, int64, error) {
	var list []model.SalaryEvent
	var total int64
	tx := dao.DB.Unscoped().Model(&model.SalaryEvent{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}
