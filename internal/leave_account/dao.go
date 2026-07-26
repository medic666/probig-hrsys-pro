package leave_account

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

func ListEvents(personID uint, leaveType string, startDate, endDate *time.Time, sourceType string, pageNum, pageSize int) ([]LeaveAccountEventWithName, int64, error) {
	var events []LeaveAccountEventWithName
	var total int64

	query := db().Model(&LeaveAccountEvent{}).Where("leave_account_events.deleted_at IS NULL")

	if personID > 0 {
		query = query.Where("leave_account_events.person_id = ?", personID)
	}
	if leaveType != "" {
		query = query.Where("leave_account_events.leave_type = ?", leaveType)
	}
	if startDate != nil {
		query = query.Where("leave_account_events.effective_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("leave_account_events.effective_date <= ?", *endDate)
	}
	if sourceType != "" {
		query = query.Where("leave_account_events.source_type = ?", sourceType)
	}

	query = query.Joins("LEFT JOIN persons ON persons.id = leave_account_events.person_id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Select("leave_account_events.*, persons.name as person_name").
		Order("leave_account_events.effective_date DESC, leave_account_events.created_at DESC").
		Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func ListTrashEvents(personID uint, leaveType string, pageNum, pageSize int) ([]LeaveAccountEventWithName, int64, error) {
	var events []LeaveAccountEventWithName
	var total int64

	query := db().Model(&LeaveAccountEvent{}).Unscoped().Where("leave_account_events.deleted_at IS NOT NULL")

	if personID > 0 {
		query = query.Where("leave_account_events.person_id = ?", personID)
	}
	if leaveType != "" {
		query = query.Where("leave_account_events.leave_type = ?", leaveType)
	}

	query = query.Joins("LEFT JOIN persons ON persons.id = leave_account_events.person_id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Select("leave_account_events.*, persons.name as person_name").
		Order("leave_account_events.deleted_at DESC").
		Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func GetEventByID(id uint) (*LeaveAccountEvent, error) {
	var event LeaveAccountEvent
	if err := db().First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func CreateEventRecord(event *LeaveAccountEvent) error {
	if event.SourceType == "" {
		event.SourceType = "manual"
	}
	return db().Create(event).Error
}

func UpdateEventRecord(event *LeaveAccountEvent) error {
	return db().Model(&LeaveAccountEvent{}).Where("id = ? AND source_type = ?", event.ID, "manual").Updates(event).Error
}

func DeleteEventRecord(id uint) error {
	return db().Where("id = ? AND source_type = ?", id, "manual").Delete(&LeaveAccountEvent{}).Error
}

func RestoreEvent(id uint) error {
	return db().Unscoped().Model(&LeaveAccountEvent{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func BatchCreateEvents(events []LeaveAccountEvent) error {
	if len(events) == 0 {
		return nil
	}
	return db().Create(&events).Error
}

func GetBalance(personID uint, leaveType string) (*LeaveAccountBalance, error) {
	var balance LeaveAccountBalance
	if err := db().Where("person_id = ? AND leave_type = ?", personID, leaveType).First(&balance).Error; err != nil {
		return nil, err
	}
	return &balance, nil
}

func UpsertBalance(balance *LeaveAccountBalance) error {
	var existing LeaveAccountBalance
	result := db().Where("person_id = ? AND leave_type = ?", balance.PersonID, balance.LeaveType).First(&existing)
	if result.Error != nil {
		return db().Create(balance).Error
	}
	balance.ID = existing.ID
	return db().Save(balance).Error
}

func ListBalances(personID uint, leaveType string) ([]LeaveAccountBalanceWithName, error) {
	var balances []LeaveAccountBalanceWithName

	query := db().Model(&LeaveAccountBalance{}).
		Joins("LEFT JOIN persons ON persons.id = leave_account_balances.person_id")

	if personID > 0 {
		query = query.Where("leave_account_balances.person_id = ?", personID)
	}
	if leaveType != "" {
		query = query.Where("leave_account_balances.leave_type = ?", leaveType)
	}

	if err := query.Select("leave_account_balances.*, persons.name as person_name").Find(&balances).Error; err != nil {
		return nil, err
	}
	return balances, nil
}

func GetEventsByBatchID(batchID uint) ([]LeaveAccountEvent, error) {
	var events []LeaveAccountEvent
	if err := db().Where("batch_id = ?", batchID).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func SoftDeleteEventsByBatchID(batchID uint) error {
	return db().Model(&LeaveAccountEvent{}).Where("batch_id = ?", batchID).Update("deleted_at", time.Now()).Error
}

func GetEventsByPersonAndType(personID uint, leaveType string) ([]LeaveAccountEvent, error) {
	var events []LeaveAccountEvent
	if err := db().Where("person_id = ? AND leave_type = ?", personID, leaveType).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func GetAttendanceEventsByPersonAndSubType(personID uint, subType string) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	if err := db().Where("person_id = ? AND sub_type = ? AND event_type = ?", personID, subType, "休假").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func ListBatches(pageNum, pageSize int) ([]SysBatch, int64, error) {
	var batches []SysBatch
	var total int64

	query := db().Model(&SysBatch{}).Where("business_type = ?", "annual_leave_carryover")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&batches).Error; err != nil {
		return nil, 0, err
	}

	return batches, total, nil
}

func GetBatchByID(batchID uint) (*SysBatch, error) {
	var batch SysBatch
	if err := db().First(&batch, batchID).Error; err != nil {
		return nil, err
	}
	return &batch, nil
}

func GetPersonName(personID uint) string {
	var person Person
	if err := db().First(&person, personID).Error; err != nil {
		return fmt.Sprintf("人员#%d", personID)
	}
	return person.Name
}

func GetCurrentPositionSnapshot(personID uint) (*PositionSnapshot, error) {
	farFutureStr := FarFutureDate.Format("2006-01-02")
	var snapshot PositionSnapshot
	if err := db().Where("person_id = ? AND effective_end_date = ?", personID, farFutureStr).First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}
