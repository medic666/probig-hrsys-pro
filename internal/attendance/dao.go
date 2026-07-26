package attendance

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

func ListEvents(personID uint, startDate, endDate *time.Time, eventType, subType string, pageNum, pageSize int) ([]AttendanceEventWithName, int64, error) {
	var events []AttendanceEventWithName
	var total int64

	query := db().Model(&AttendanceEvent{}).Where("attendance_events.deleted_at IS NULL")

	if personID > 0 {
		query = query.Where("attendance_events.person_id = ?", personID)
	}
	if startDate != nil {
		query = query.Where("attendance_events.event_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("attendance_events.event_date <= ?", *endDate)
	}
	if eventType != "" {
		query = query.Where("attendance_events.event_type = ?", eventType)
	}
	if subType != "" {
		query = query.Where("attendance_events.sub_type = ?", subType)
	}

	query = query.Joins("LEFT JOIN persons ON persons.id = attendance_events.person_id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Select("attendance_events.*, persons.name as person_name").
		Order("attendance_events.event_date DESC, attendance_events.created_at DESC").
		Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func ListTrashEvents(personID uint, startDate, endDate *time.Time, eventType, subType string, pageNum, pageSize int) ([]AttendanceEventWithName, int64, error) {
	var events []AttendanceEventWithName
	var total int64

	query := db().Model(&AttendanceEvent{}).Unscoped().Where("attendance_events.deleted_at IS NOT NULL")

	if personID > 0 {
		query = query.Where("attendance_events.person_id = ?", personID)
	}
	if startDate != nil {
		query = query.Where("attendance_events.event_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("attendance_events.event_date <= ?", *endDate)
	}
	if eventType != "" {
		query = query.Where("attendance_events.event_type = ?", eventType)
	}
	if subType != "" {
		query = query.Where("attendance_events.sub_type = ?", subType)
	}

	query = query.Joins("LEFT JOIN persons ON persons.id = attendance_events.person_id")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Select("attendance_events.*, persons.name as person_name").
		Order("attendance_events.deleted_at DESC").
		Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func GetEventByID(id uint) (*AttendanceEvent, error) {
	var event AttendanceEvent
	if err := db().First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func CreateEvent(event *AttendanceEvent) error {
	return db().Create(event).Error
}

func UpdateEvent(event *AttendanceEvent) error {
	return db().Save(event).Error
}

func DeleteEvent(id uint) error {
	return db().Delete(&AttendanceEvent{}, id).Error
}

func RestoreEvent(id uint) error {
	return db().Unscoped().Model(&AttendanceEvent{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func BatchCreate(events []AttendanceEvent) error {
	if len(events) == 0 {
		return nil
	}
	return db().Create(&events).Error
}

func DeleteDailyByPersonAndDate(personID uint, workDate time.Time) error {
	dateStr := workDate.Format("2006-01-02")
	return db().Where("person_id = ? AND work_date = ?", personID, dateStr).Delete(&AttendanceDailyProjection{}).Error
}

func UpsertDailyProjection(proj *AttendanceDailyProjection) error {
	var existing AttendanceDailyProjection
	dateStr := proj.WorkDate.Format("2006-01-02")
	result := db().Where("person_id = ? AND work_date = ?", proj.PersonID, dateStr).First(&existing)
	if result.Error != nil {
		return db().Create(proj).Error
	}
	proj.ID = existing.ID
	return db().Save(proj).Error
}

func ListDailyProjections(personID uint, startDate, endDate *time.Time, pageNum, pageSize int) ([]DailyProjectionWithName, int64, error) {
	var projections []DailyProjectionWithName
	var total int64

	query := db().Model(&AttendanceDailyProjection{}).
		Joins("LEFT JOIN persons ON persons.id = attendance_daily_projections.person_id")

	if personID > 0 {
		query = query.Where("attendance_daily_projections.person_id = ?", personID)
	}
	if startDate != nil {
		query = query.Where("attendance_daily_projections.work_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("attendance_daily_projections.work_date <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Select("attendance_daily_projections.*, persons.name as person_name").
		Order("attendance_daily_projections.work_date DESC, attendance_daily_projections.person_id ASC").
		Offset(offset).Limit(pageSize).Find(&projections).Error; err != nil {
		return nil, 0, err
	}

	return projections, total, nil
}

func GetDailyByPersonAndDate(personID uint, workDate time.Time) (*AttendanceDailyProjection, error) {
	var proj AttendanceDailyProjection
	dateStr := workDate.Format("2006-01-02")
	if err := db().Where("person_id = ? AND work_date = ?", personID, dateStr).First(&proj).Error; err != nil {
		return nil, err
	}
	return &proj, nil
}

func GetDailyByPersonAndDateRange(personID uint, startDate, endDate time.Time) ([]AttendanceDailyProjection, error) {
	var projections []AttendanceDailyProjection
	if err := db().Where("person_id = ? AND work_date >= ? AND work_date <= ?",
		personID, startDate, endDate).
		Order("work_date ASC").
		Find(&projections).Error; err != nil {
		return nil, err
	}
	return projections, nil
}

func DeleteMonthlyByPersonAndMonth(personID uint, belongMonth string) error {
	return db().Where("person_id = ? AND belong_month = ?", personID, belongMonth).Delete(&AttendanceSalaryMonthly{}).Error
}

func UpsertMonthlySalary(ms *AttendanceSalaryMonthly) error {
	var existing AttendanceSalaryMonthly
	result := db().Where("person_id = ? AND belong_month = ?", ms.PersonID, ms.BelongMonth).First(&existing)
	if result.Error != nil {
		return db().Create(ms).Error
	}
	ms.ID = existing.ID
	return db().Save(ms).Error
}

func DeleteAllByMonth(belongMonth string) error {
	return db().Where("belong_month = ?", belongMonth).Delete(&AttendanceSalaryMonthly{}).Error
}

type MonthlySalaryQuery struct {
	PersonID     uint
	BelongMonth  string
	StartMonth   string
	EndMonth     string
	AttendanceGroup string
	PageNum      int
	PageSize     int
}

func ListMonthlySalary(q MonthlySalaryQuery) ([]MonthlySalaryWithName, int64, error) {
	var results []MonthlySalaryWithName
	var total int64

	query := db().Model(&AttendanceSalaryMonthly{}).
		Joins("LEFT JOIN persons ON persons.id = attendance_salary_monthlies.person_id")

	if q.PersonID > 0 {
		query = query.Where("attendance_salary_monthlies.person_id = ?", q.PersonID)
	}
	if q.BelongMonth != "" {
		query = query.Where("attendance_salary_monthlies.belong_month = ?", q.BelongMonth)
	}
	if q.StartMonth != "" {
		query = query.Where("attendance_salary_monthlies.belong_month >= ?", q.StartMonth)
	}
	if q.EndMonth != "" {
		query = query.Where("attendance_salary_monthlies.belong_month <= ?", q.EndMonth)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (q.PageNum - 1) * q.PageSize
	if err := query.Select("attendance_salary_monthlies.*, persons.name as person_name").
		Order("attendance_salary_monthlies.belong_month DESC, attendance_salary_monthlies.person_id ASC").
		Offset(offset).Limit(q.PageSize).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func GetMonthlyByPersonAndMonth(personID uint, belongMonth string) (*AttendanceSalaryMonthly, error) {
	var ms AttendanceSalaryMonthly
	if err := db().Where("person_id = ? AND belong_month = ?", personID, belongMonth).First(&ms).Error; err != nil {
		return nil, err
	}
	return &ms, nil
}

func GetMonthlyByMonth(belongMonth string) ([]AttendanceSalaryMonthly, error) {
	var results []AttendanceSalaryMonthly
	if err := db().Where("belong_month = ?", belongMonth).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func FindEventsByPersonAndDate(personID uint, eventDate time.Time) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	dateStr := eventDate.Format("2006-01-02")
	if err := db().Where("person_id = ? AND event_date = ?", personID, dateStr).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func FindEventsByPersonAndDateRange(personID uint, startDate, endDate time.Time) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	if err := db().Where("person_id = ? AND event_date >= ? AND event_date <= ?",
		personID, startDate, endDate).
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func GetPersonName(personID uint) string {
	var person Person
	if err := db().First(&person, personID).Error; err != nil {
		return fmt.Sprintf("人员#%d", personID)
	}
	return person.Name
}
