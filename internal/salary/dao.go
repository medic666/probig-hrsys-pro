package salary

import "gorm.io/gorm"

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) CreateEvent(event *SalaryEvent) error {
	return d.db.Create(event).Error
}

func (d *DAO) GetEventByID(id uint) (*SalaryEvent, error) {
	var event SalaryEvent
	err := d.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (d *DAO) ListEvents(personID uint, belongMonth string, eventType string, page, pageSize int) ([]SalaryEvent, int64, error) {
	var events []SalaryEvent
	var total int64
	query := d.db.Model(&SalaryEvent{})
	if personID > 0 {
		query = query.Where("person_id = ?", personID)
	}
	if belongMonth != "" {
		query = query.Where("belong_month = ?", belongMonth)
	}
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&events).Error
	return events, total, err
}

func (d *DAO) UpdateEvent(event *SalaryEvent) error {
	return d.db.Save(event).Error
}

func (d *DAO) DeleteEvent(id uint) error {
	return d.db.Delete(&SalaryEvent{}, id).Error
}

func (d *DAO) GetEventsByPersonAndMonth(personID uint, belongMonth string) ([]SalaryEvent, error) {
	var events []SalaryEvent
	err := d.db.Where("person_id = ? AND belong_month = ?", personID, belongMonth).Find(&events).Error
	return events, err
}

func (d *DAO) GetEventsByPersonIDsAndMonth(personIDs []uint, belongMonth string) ([]SalaryEvent, error) {
	var events []SalaryEvent
	err := d.db.Where("person_id IN ? AND belong_month = ?", personIDs, belongMonth).Find(&events).Error
	return events, err
}

func (d *DAO) GetLatestPerformanceEvent(personID uint, belongMonth string) (*SalaryEvent, error) {
	var event SalaryEvent
	err := d.db.Where("person_id = ? AND belong_month = ? AND event_type = ?", personID, belongMonth, "绩效调整").
		Order("created_at DESC").First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (d *DAO) GetLatestEventUpdateTime(personID uint, belongMonth string) (*string, error) {
	var event SalaryEvent
	err := d.db.Where("person_id = ? AND belong_month = ?", personID, belongMonth).
		Order("updated_at DESC").First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event.UpdatedAt, nil
}

func (d *DAO) CreateOrUpdateSummary(summary *SalarySummary) error {
	var existing SalarySummary
	err := d.db.Unscoped().Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).First(&existing).Error
	if err == nil {
		summary.ID = existing.ID
		summary.DeletedAt = gorm.DeletedAt{}
		return d.db.Unscoped().Save(summary).Error
	}
	return d.db.Create(summary).Error
}

func (d *DAO) GetSummaryByPersonMonth(personID uint, belongMonth string) (*SalarySummary, error) {
	var summary SalarySummary
	err := d.db.Where("person_id = ? AND belong_month = ?", personID, belongMonth).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (d *DAO) ListSummaries(personIDs []uint, belongMonth string, isLocked *bool, page, pageSize int) ([]SalarySummary, int64, error) {
	var summaries []SalarySummary
	var total int64
	query := d.db.Model(&SalarySummary{})
	if len(personIDs) > 0 {
		query = query.Where("person_id IN ?", personIDs)
	}
	if belongMonth != "" {
		query = query.Where("belong_month = ?", belongMonth)
	}
	if isLocked != nil {
		query = query.Where("is_locked = ?", *isLocked)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("belong_month DESC, person_id ASC").Offset(offset).Limit(pageSize).Find(&summaries).Error
	return summaries, total, err
}

func (d *DAO) DeleteSummaryByPersonMonth(personID uint, belongMonth string) error {
	return d.db.Where("person_id = ? AND belong_month = ?", personID, belongMonth).Delete(&SalarySummary{}).Error
}

func (d *DAO) LockSummary(personID uint, belongMonth string) error {
	return d.db.Model(&SalarySummary{}).Where("person_id = ? AND belong_month = ?", personID, belongMonth).Update("is_locked", true).Error
}

func (d *DAO) UnlockSummary(personID uint, belongMonth string) error {
	return d.db.Model(&SalarySummary{}).Where("person_id = ? AND belong_month = ?", personID, belongMonth).Update("is_locked", false).Error
}

func (d *DAO) GetAllSummariesByMonth(belongMonth string) ([]SalarySummary, error) {
	var summaries []SalarySummary
	err := d.db.Where("belong_month = ?", belongMonth).Find(&summaries).Error
	return summaries, err
}
