package dao

import (
	"probig/internal/models"
)

func GetAttendanceSummaryList(page, pageSize int, personID uint, belongMonth string) ([]models.AttendanceSummary, int64, error) {
	var list []models.AttendanceSummary
	var total int64
	q := DB().Model(&models.AttendanceSummary{})
	if personID > 0 {
		q = q.Where("person_id = ?", personID)
	}
	if belongMonth != "" {
		q = q.Where("belong_month = ?", belongMonth)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetAttendanceSummaryByPersonAndMonth(personID uint, belongMonth string) (*models.AttendanceSummary, error) {
	var s models.AttendanceSummary
	if err := DB().Where("person_id = ? AND belong_month = ?", personID, belongMonth).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func UpsertAttendanceSummary(s *models.AttendanceSummary) error {
	var existing models.AttendanceSummary
	err := DB().Where("person_id = ? AND belong_month = ?", s.PersonID, s.BelongMonth).First(&existing).Error
	if err != nil {
		return DB().Create(s).Error
	}
	s.ID = existing.ID
	return DB().Save(s).Error
}

func LockAttendanceSummary(personID uint, belongMonth string, locked bool) error {
	return DB().Model(&models.AttendanceSummary{}).
		Where("person_id = ? AND belong_month = ?", personID, belongMonth).
		Update("is_locked", locked).Error
}

func GetLockedMonthsForPerson(personID uint) ([]string, error) {
	var months []string
	if err := DB().Model(&models.AttendanceSummary{}).
		Where("person_id = ? AND is_locked = ?", personID, true).
		Pluck("belong_month", &months).Error; err != nil {
		return nil, err
	}
	return months, nil
}
