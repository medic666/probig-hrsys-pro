package position

import (
	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) CreateEvent(event *PositionEvent) error {
	return d.db.Create(event).Error
}

func (d *DAO) GetEventByID(id uint) (*PositionEvent, error) {
	var event PositionEvent
	err := d.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (d *DAO) ListEvents(personID uint, page, pageSize int) ([]PositionEvent, int64, error) {
	var events []PositionEvent
	var total int64
	query := d.db.Model(&PositionEvent{})
	if personID > 0 {
		query = query.Where("person_id = ?", personID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("effective_date DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&events).Error
	return events, total, err
}

func (d *DAO) GetAllEventsByPersonID(personID uint) ([]PositionEvent, error) {
	var events []PositionEvent
	err := d.db.Where("person_id = ?", personID).Order("effective_date ASC, created_at ASC").Find(&events).Error
	return events, err
}

func (d *DAO) UpdateEvent(event *PositionEvent) error {
	return d.db.Save(event).Error
}

func (d *DAO) DeleteEvent(id uint) error {
	return d.db.Delete(&PositionEvent{}, id).Error
}

func (d *DAO) GetLatestEventBeforeDate(personID uint, date string) (*PositionEvent, error) {
	var event PositionEvent
	err := d.db.Where("person_id = ? AND effective_date <= ?", personID, date).
		Order("effective_date DESC, created_at DESC").First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (d *DAO) HasEventsAfterDate(personID uint, date string) (bool, error) {
	var count int64
	err := d.db.Model(&PositionEvent{}).Where("person_id = ? AND effective_date > ?", personID, date).Count(&count).Error
	return count > 0, err
}

func (d *DAO) GetAllPersonIDs() ([]uint, error) {
	var ids []uint
	err := d.db.Model(&PositionEvent{}).Distinct("person_id").Pluck("person_id", &ids).Error
	return ids, err
}

func (d *DAO) DeleteSnapshotsByPersonID(personID uint) error {
	return d.db.Unscoped().Where("person_id = ?", personID).Delete(&PositionSnapshot{}).Error
}

func (d *DAO) BatchCreateSnapshots(snapshots []PositionSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return d.db.CreateInBatches(snapshots, 100).Error
}

func (d *DAO) GetSnapshotByPersonIDAndDate(personID uint, date string) (*PositionSnapshot, error) {
	var snapshot PositionSnapshot
	err := d.db.Where("person_id = ? AND snapshot_date = ?", personID, date).First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (d *DAO) GetSnapshotsByPersonAndMonthRange(personID uint, startDate, endDate string) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	err := d.db.Where("person_id = ? AND snapshot_date >= ? AND snapshot_date <= ?", personID, startDate, endDate).
		Order("snapshot_date ASC").Find(&snapshots).Error
	return snapshots, err
}

func (d *DAO) GetSnapshotsByMonth(personIDs []uint, yearMonth string) ([]PositionSnapshot, error) {
	var snapshots []PositionSnapshot
	startDate := yearMonth + "-01"
	endDate := yearMonth + "-31"
	err := d.db.Where("person_id IN ? AND snapshot_date >= ? AND snapshot_date <= ?", personIDs, startDate, endDate).
		Order("snapshot_date ASC").Find(&snapshots).Error
	return snapshots, err
}
