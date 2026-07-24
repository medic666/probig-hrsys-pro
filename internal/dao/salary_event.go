package dao

import (
	"time"

	"probig/internal/models"
)

func GetSalaryEventList(page, pageSize int, personID uint, belongMonth, eventType string) ([]models.SalaryEvent, int64, error) {
	var list []models.SalaryEvent
	var total int64
	q := DB().Model(&models.SalaryEvent{}).Preload("Person")
	if personID > 0 {
		q = q.Where("person_id = ?", personID)
	}
	if belongMonth != "" {
		q = q.Where("belong_month = ?", belongMonth)
	}
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetSalaryEventsByPersonAndMonth(personID uint, belongMonth string) ([]models.SalaryEvent, error) {
	var list []models.SalaryEvent
	if err := DB().Where("person_id = ? AND belong_month = ?", personID, belongMonth).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetLatestPerformanceEvent(personID uint, belongMonth string) (*models.SalaryEvent, error) {
	var e models.SalaryEvent
	if err := DB().Where("person_id = ? AND belong_month = ? AND event_type = ?",
		personID, belongMonth, "绩效调整").Order("created_at DESC").First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func GetSalaryEventByID(id uint) (*models.SalaryEvent, error) {
	var e models.SalaryEvent
	if err := DB().First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func CreateSalaryEvent(e *models.SalaryEvent) error {
	return DB().Create(e).Error
}

func UpdateSalaryEvent(e *models.SalaryEvent) error {
	return DB().Save(e).Error
}

func DeleteSalaryEvent(id uint) error {
	return DB().Delete(&models.SalaryEvent{}, id).Error
}

func GetLastSalaryEventTimeInMonth(personID uint, belongMonth string) (*time.Time, error) {
	var e models.SalaryEvent
	if err := DB().Where("person_id = ? AND belong_month = ?", personID, belongMonth).
		Order("updated_at DESC").First(&e).Error; err != nil {
		return nil, err
	}
	return &e.UpdatedAt, nil
}
