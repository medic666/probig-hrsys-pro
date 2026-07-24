package dao

import (
	"time"
	"probig/internal/models"
)

func GetAttendanceEventList(page, pageSize int, personID uint, eventType, subType, startDate, endDate string) ([]models.AttendanceEvent, int64, error) {
	var list []models.AttendanceEvent
	var total int64
	q := DB().Model(&models.AttendanceEvent{}).Preload("Person")
	if personID > 0 {
		q = q.Where("person_id = ?", personID)
	}
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if subType != "" {
		q = q.Where("sub_type = ?", subType)
	}
	if startDate != "" {
		q = q.Where("event_date >= ?", startDate)
	}
	if endDate != "" {
		q = q.Where("event_date <= ?", endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("event_date DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetAttendanceEventsByPersonAndMonth(personID uint, belongMonth string) ([]models.AttendanceEvent, error) {
	var list []models.AttendanceEvent
	if err := DB().Where("person_id = ? AND event_date >= ? AND event_date <= ?",
		personID, belongMonth+"-01", belongMonth+"-31").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func GetAttendanceEventByID(id uint) (*models.AttendanceEvent, error) {
	var e models.AttendanceEvent
	if err := DB().First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func CreateAttendanceEvent(e *models.AttendanceEvent) error {
	return DB().Create(e).Error
}

func CreateAttendanceEvents(events []models.AttendanceEvent) error {
	return DB().Create(&events).Error
}

func UpdateAttendanceEvent(e *models.AttendanceEvent) error {
	return DB().Save(e).Error
}

func DeleteAttendanceEvent(id uint) error {
	return DB().Delete(&models.AttendanceEvent{}, id).Error
}

func GetLastAttendanceEventTimeInMonth(personID uint, belongMonth string) (*time.Time, error) {
	var e models.AttendanceEvent
	if err := DB().Where("person_id = ? AND event_date >= ? AND event_date <= ?",
		personID, belongMonth+"-01", belongMonth+"-31").
		Order("updated_at DESC").First(&e).Error; err != nil {
		return nil, err
	}
	return &e.UpdatedAt, nil
}
