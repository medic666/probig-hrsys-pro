package position

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

func db() *gorm.DB {
	if DB != nil {
		return DB
	}
	return database.DB
}

func ListEvents(personID uint, startDate, endDate *time.Time, eventName string, pageNum, pageSize int) ([]PositionEventWithName, int64, error) {
	var events []PositionEventWithName
	var total int64

	query := db().Model(&PositionEvent{}).Where("position_events.deleted_at IS NULL")

	if personID > 0 {
		query = query.Where("position_events.person_id = ?", personID)
	}
	if startDate != nil {
		query = query.Where("position_events.effective_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("position_events.effective_date <= ?", *endDate)
	}
	if eventName != "" {
		query = query.Where("position_events.event_name LIKE ?", "%"+eventName+"%")
	}

	query = query.Joins("LEFT JOIN persons ON persons.id = position_events.person_id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Select("position_events.*, persons.name as person_name").
		Order("position_events.effective_date DESC, position_events.created_at DESC").
		Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func ListTrashEvents(personID uint, startDate, endDate *time.Time, eventName string, pageNum, pageSize int) ([]PositionEventWithName, int64, error) {
	var events []PositionEventWithName
	var total int64

	query := db().Model(&PositionEvent{}).Unscoped().Where("position_events.deleted_at IS NOT NULL")

	if personID > 0 {
		query = query.Where("position_events.person_id = ?", personID)
	}
	if startDate != nil {
		query = query.Where("position_events.effective_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("position_events.effective_date <= ?", *endDate)
	}
	if eventName != "" {
		query = query.Where("position_events.event_name LIKE ?", "%"+eventName+"%")
	}

	query = query.Joins("LEFT JOIN persons ON persons.id = position_events.person_id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Select("position_events.*, persons.name as person_name").
		Order("position_events.deleted_at DESC").
		Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func GetEventByID(id uint) (*PositionEvent, error) {
	var event PositionEvent
	if err := db().First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func createEventRecord(event *PositionEvent) error {
	return db().Create(event).Error
}

func updateEventRecord(event *PositionEvent) error {
	return db().Save(event).Error
}

func DeleteEvent(id uint) error {
	return db().Delete(&PositionEvent{}, id).Error
}

func RestoreEvent(id uint) error {
	return db().Unscoped().Model(&PositionEvent{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func FindEventsByPerson(personID uint) ([]PositionEvent, error) {
	var events []PositionEvent
	if err := db().Where("person_id = ?", personID).
		Order("effective_date ASC, created_at DESC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func DeleteAllSnapshotsByPerson(personID uint) error {
	return db().Where("person_id = ?", personID).Delete(&PositionSnapshot{}).Error
}

func BatchCreateSnapshots(snapshots []PositionSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return db().Create(&snapshots).Error
}

func GetSnapshotsByPerson(personID uint) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	if err := db().Where("person_id = ?", personID).
		Order("effective_start_date ASC").
		Find(&snapshots).Error; err != nil {
		return nil, err
	}
	return snapshots, nil
}

func GetSnapshotsByPersonAndDate(personID uint, date time.Time) ([]PositionSnapshot, error) {
	dateStr := date.Format("2006-01-02")
	var snapshots []PositionSnapshot
	if err := db().Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
		personID, dateStr, dateStr).
		Find(&snapshots).Error; err != nil {
		return nil, err
	}
	return snapshots, nil
}

func GetSnapshotByID(id uint) (*PositionSnapshot, error) {
	var snapshot PositionSnapshot
	if err := db().First(&snapshot, id).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

type SnapshotQuery struct {
	PersonID  uint
	StartDate *time.Time
	EndDate   *time.Time
	PageNum   int
	PageSize  int
}

func ListSnapshots(q SnapshotQuery) ([]PositionSnapshotWithName, int64, error) {
	var snapshots []PositionSnapshotWithName
	var total int64

	query := db().Model(&PositionSnapshot{}).Joins("LEFT JOIN persons ON persons.id = position_snapshots.person_id")

	if q.PersonID > 0 {
		query = query.Where("position_snapshots.person_id = ?", q.PersonID)
	}
	if q.StartDate != nil {
		query = query.Where("position_snapshots.effective_start_date >= ?", q.StartDate)
	}
	if q.EndDate != nil {
		query = query.Where("position_snapshots.effective_start_date <= ?", q.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (q.PageNum - 1) * q.PageSize
	if err := query.Select("position_snapshots.*, persons.name as person_name").
		Order("position_snapshots.person_id ASC, position_snapshots.effective_start_date ASC").
		Offset(offset).Limit(q.PageSize).Find(&snapshots).Error; err != nil {
		return nil, 0, err
	}

	return snapshots, total, nil
}

func GetLatestSnapshotsForAllPersons() ([]PositionSnapshotWithName, error) {
	farFutureStr := FarFutureDate.Format("2006-01-02")
	var snapshots []PositionSnapshotWithName
	if err := db().Model(&PositionSnapshot{}).
		Joins("LEFT JOIN persons ON persons.id = position_snapshots.person_id").
		Where("position_snapshots.effective_end_date = ?", farFutureStr).
		Select("position_snapshots.*, persons.name as person_name").
		Find(&snapshots).Error; err != nil {
		return nil, err
	}
	return snapshots, nil
}

func GetPersonName(personID uint) string {
	var person Person
	if err := db().First(&person, personID).Error; err != nil {
		return fmt.Sprintf("人员#%d", personID)
	}
	return person.Name
}
