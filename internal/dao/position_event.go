package dao

import (
	"sort"
	"time"

	"probig/internal/models"
)

func GetPositionEventsByPersonID(personID uint) ([]models.PositionEvent, error) {
	var events []models.PositionEvent
	if err := DB().Where("person_id = ?", personID).Order("effective_date ASC, created_at DESC").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func GetPositionEventByID(id uint) (*models.PositionEvent, error) {
	var e models.PositionEvent
	if err := DB().First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func CreatePositionEvent(e *models.PositionEvent) error {
	return DB().Create(e).Error
}

func UpdatePositionEvent(e *models.PositionEvent) error {
	return DB().Save(e).Error
}

func DeletePositionEvent(id uint) error {
	return DB().Delete(&models.PositionEvent{}, id).Error
}

func GetPositionEventsDateRange(personID uint) (minDate, maxDate *time.Time, err error) {
	var events []models.PositionEvent
	if err := DB().Where("person_id = ?", personID).Find(&events).Error; err != nil {
		return nil, nil, err
	}
	if len(events) == 0 {
		return nil, nil, nil
	}
	dates := make([]time.Time, 0)
	for _, e := range events {
		if e.EffectiveDate != nil {
			dates = append(dates, *e.EffectiveDate)
		}
	}
	if len(dates) == 0 {
		return nil, nil, nil
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	first := dates[0]
	last := dates[len(dates)-1]
	return &first, &last, nil
}
