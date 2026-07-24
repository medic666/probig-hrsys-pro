package attendance

import (
	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

type AttendanceListParams struct {
	PersonIDs      []uint
	EventDateStart string
	EventDateEnd   string
	EventType      string
	SubType        string
	Page           int
	PageSize       int
}

func (d *DAO) CreateEvent(event *AttendanceEvent) error {
	return d.db.Create(event).Error
}

func (d *DAO) GetEventByID(id uint) (*AttendanceEvent, error) {
	var event AttendanceEvent
	err := d.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (d *DAO) ListEvents(params AttendanceListParams) ([]AttendanceEvent, int64, error) {
	var events []AttendanceEvent
	var total int64
	query := d.db.Model(&AttendanceEvent{})
	if len(params.PersonIDs) > 0 {
		query = query.Where("person_id IN ?", params.PersonIDs)
	}
	if params.EventDateStart != "" {
		query = query.Where("event_date >= ?", params.EventDateStart)
	}
	if params.EventDateEnd != "" {
		query = query.Where("event_date <= ?", params.EventDateEnd)
	}
	if params.EventType != "" {
		query = query.Where("event_type = ?", params.EventType)
	}
	if params.SubType != "" {
		query = query.Where("sub_type = ?", params.SubType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (params.Page - 1) * params.PageSize
	err := query.Order("event_date DESC, created_at DESC").Offset(offset).Limit(params.PageSize).Find(&events).Error
	return events, total, err
}

func (d *DAO) UpdateEvent(event *AttendanceEvent) error {
	return d.db.Save(event).Error
}

func (d *DAO) DeleteEvent(id uint) error {
	return d.db.Delete(&AttendanceEvent{}, id).Error
}

func (d *DAO) DeleteEventsByPersonMonth(personID uint, yearMonth string) error {
	return d.db.Where("person_id = ? AND event_date LIKE ?", personID, yearMonth+"%").Delete(&AttendanceEvent{}).Error
}

func (d *DAO) GetEventsByPersonAndMonth(personID uint, yearMonth string) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	err := d.db.Where("person_id = ? AND event_date LIKE ?", personID, yearMonth+"%").
		Order("event_date ASC").Find(&events).Error
	return events, err
}

func (d *DAO) GetEventsByPersonIDsAndMonth(personIDs []uint, yearMonth string) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	err := d.db.Where("person_id IN ? AND event_date LIKE ?", personIDs, yearMonth+"%").
		Order("event_date ASC").Find(&events).Error
	return events, err
}

func (d *DAO) GetLatestEventUpdateTime(personID uint, yearMonth string) (*string, error) {
	var updatedAt string
	err := d.db.Model(&AttendanceEvent{}).
		Where("person_id = ? AND event_date LIKE ?", personID, yearMonth+"%").
		Select("MAX(updated_at)").
		Scan(&updatedAt).Error
	if err != nil {
		return nil, err
	}
	if updatedAt == "" {
		return nil, nil
	}
	return &updatedAt, nil
}

func (d *DAO) BatchCreateEvents(events []AttendanceEvent) error {
	if len(events) == 0 {
		return nil
	}
	return d.db.Create(&events).Error
}

func (d *DAO) GetEventTypesByPersonMonth(personID uint, yearMonth string) ([]string, error) {
	var types []string
	err := d.db.Model(&AttendanceEvent{}).
		Where("person_id = ? AND event_date LIKE ?", personID, yearMonth+"%").
		Distinct("event_type").
		Pluck("event_type", &types).Error
	return types, err
}

func (d *DAO) CreateOrUpdateSummary(summary *AttendanceSummary) error {
	var existing AttendanceSummary
	err := d.db.Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).First(&existing).Error
	if err == nil {
		summary.ID = existing.ID
		return d.db.Save(summary).Error
	}
	if err == gorm.ErrRecordNotFound {
		return d.db.Create(summary).Error
	}
	return err
}

func (d *DAO) GetSummaryByPersonMonth(personID uint, yearMonth string) (*AttendanceSummary, error) {
	var summary AttendanceSummary
	err := d.db.Where("person_id = ? AND belong_month = ?", personID, yearMonth).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (d *DAO) ListSummaries(personIDs []uint, yearMonth string, isLocked *bool, page, pageSize int) ([]AttendanceSummary, int64, error) {
	var summaries []AttendanceSummary
	var total int64
	query := d.db.Model(&AttendanceSummary{})
	if len(personIDs) > 0 {
		query = query.Where("person_id IN ?", personIDs)
	}
	if yearMonth != "" {
		query = query.Where("belong_month = ?", yearMonth)
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

func (d *DAO) DeleteSummaryByPersonMonth(personID uint, yearMonth string) error {
	return d.db.Where("person_id = ? AND belong_month = ?", personID, yearMonth).Delete(&AttendanceSummary{}).Error
}

func (d *DAO) LockSummary(personID uint, yearMonth string) error {
	return d.db.Model(&AttendanceSummary{}).
		Where("person_id = ? AND belong_month = ?", personID, yearMonth).
		Update("is_locked", true).Error
}

func (d *DAO) UnlockSummary(personID uint, yearMonth string) error {
	return d.db.Model(&AttendanceSummary{}).
		Where("person_id = ? AND belong_month = ?", personID, yearMonth).
		Update("is_locked", false).Error
}

func (d *DAO) GetAllSummariesByMonth(yearMonth string) ([]AttendanceSummary, error) {
	var summaries []AttendanceSummary
	err := d.db.Where("belong_month = ?", yearMonth).Order("person_id ASC").Find(&summaries).Error
	return summaries, err
}
