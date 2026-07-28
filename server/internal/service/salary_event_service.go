package service

import (
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreateSalaryEvent(event *model.SalaryEvent) error {
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		var maxSeq int
		tx.Unscoped().Model(&model.SalaryEvent{}).Where("person_id = ?", event.PersonID).
			Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
		event.Seq = maxSeq + 1
		return tx.Create(event).Error
	})
}

func UpdateSalaryEvent(id uint, event *model.SalaryEvent) error {
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
	return dao.DB.Model(&existing).Updates(updates).Error
}

func DeleteSalaryEvent(id uint) error {
	var event model.SalaryEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	return dao.DB.Delete(&event).Error
}

func RestoreSalaryEvent(id uint) error {
	var event model.SalaryEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	return dao.DB.Unscoped().Model(&event).Update("deleted_at", nil).Error
}

func GetSalaryEvent(id uint) (*model.SalaryEvent, error) {
	var event model.SalaryEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func GetSalaryEventList(pageNum, pageSize int, personID uint, belongMonth, eventType string) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.SalaryEvent{})
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	if belongMonth != "" {
		tx = tx.Where("belong_month = ?", belongMonth)
	}
	if eventType != "" {
		tx = tx.Where("event_type = ?", eventType)
	}
	var total int64
	tx.Count(&total)
	var events []model.SalaryEvent
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("belong_month DESC, seq DESC").Find(&events)
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
		var personName string
		dao.DB.Table("persons").Select("name").Where("id = ?", e.PersonID).Scan(&personName)
		item["person_name"] = personName
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
