package service

import (
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreateAnnualLeaveEvent(event *model.AnnualLeaveAccountEvent) error {
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		var maxSeq int
		tx.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ?", event.PersonID).
			Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
		event.Seq = maxSeq + 1
		event.SourceType = "manual"

		if err := tx.Create(event).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, event.PersonID)
	})
}

func UpdateAnnualLeaveEvent(id uint, event *model.AnnualLeaveAccountEvent) error {
	var existing model.AnnualLeaveAccountEvent
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("年假权益事件不存在")
	}
	if existing.SourceType == "system_period" {
		return errors.New("系统周期事件不可编辑")
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"event_type":     event.EventType,
			"hours":          event.Hours,
			"effective_date": event.EffectiveDate,
			"remark":         event.Remark,
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, existing.PersonID)
	})
}

func DeleteAnnualLeaveEvent(id uint) error {
	var event model.AnnualLeaveAccountEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	if event.SourceType == "system_period" {
		return errors.New("系统周期事件不可删除")
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, event.PersonID)
	})
}

func RestoreAnnualLeaveEvent(id uint) error {
	var event model.AnnualLeaveAccountEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	if event.SourceType == "system_period" {
		return errors.New("系统周期事件不可人工恢复")
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, event.PersonID)
	})
}

func GetAnnualLeaveEvent(id uint) (*model.AnnualLeaveAccountEvent, error) {
	var event model.AnnualLeaveAccountEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func GetAnnualLeaveEventList(pageNum, pageSize int, personID uint, dateStart, dateEnd, eventType string) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.AnnualLeaveAccountEvent{})
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	if dateStart != "" {
		tx = tx.Where("effective_date >= ?", dateStart)
	}
	if dateEnd != "" {
		tx = tx.Where("effective_date <= ?", dateEnd)
	}
	if eventType != "" {
		tx = tx.Where("event_type = ?", eventType)
	}

	var total int64
	tx.Count(&total)

	var events []model.AnnualLeaveAccountEvent
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("person_id ASC, effective_date DESC, seq DESC").Find(&events)

	var result []map[string]interface{}
	for _, e := range events {
		item := map[string]interface{}{
			"id":             e.ID,
			"person_id":      e.PersonID,
			"seq":            e.Seq,
			"event_type":     e.EventType,
			"source_type":    e.SourceType,
			"batch_id":       e.BatchID,
			"hours":          e.Hours,
			"effective_date": e.EffectiveDate,
			"remark":         e.Remark,
			"created_at":     e.CreatedAt,
		}
		var personName string
		dao.DB.Table("persons").Select("name").Where("id = ?", e.PersonID).Scan(&personName)
		item["person_name"] = personName
		result = append(result, item)
	}
	return result, total, nil
}

func GetDeletedAnnualLeaveEvents(pageNum, pageSize int) ([]model.AnnualLeaveAccountEvent, int64, error) {
	var list []model.AnnualLeaveAccountEvent
	var total int64
	tx := dao.DB.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}
