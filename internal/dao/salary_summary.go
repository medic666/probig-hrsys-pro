package dao

import (
	"probig/internal/models"
)

func GetSalarySummaryList(page, pageSize int, personID uint, belongMonth string) ([]models.SalarySummary, int64, error) {
	var list []models.SalarySummary
	var total int64
	q := DB().Model(&models.SalarySummary{})
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

func UpsertSalarySummary(s *models.SalarySummary) error {
	var existing models.SalarySummary
	err := DB().Where("person_id = ? AND belong_month = ?", s.PersonID, s.BelongMonth).First(&existing).Error
	if err != nil {
		return DB().Create(s).Error
	}
	s.ID = existing.ID
	return DB().Save(s).Error
}

func LockSalarySummary(personID uint, belongMonth string, locked bool) error {
	return DB().Model(&models.SalarySummary{}).
		Where("person_id = ? AND belong_month = ?", personID, belongMonth).
		Update("is_locked", locked).Error
}
