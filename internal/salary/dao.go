package salary

import (
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/database"
)

func db() *gorm.DB {
	return database.DB
}

func CreateSalaryEvent(event *SalaryEvent) error {
	return db().Create(event).Error
}

func GetSalaryEventByID(id uint) (*SalaryEvent, error) {
	var event SalaryEvent
	err := db().Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func UpdateSalaryEvent(event *SalaryEvent) error {
	return db().Save(event).Error
}

func DeleteSalaryEvent(id uint) error {
	return db().Delete(&SalaryEvent{}, id).Error
}

func RestoreSalaryEvent(id uint) error {
	return db().Unscoped().Model(&SalaryEvent{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func ListSalaryEvents(filter SalaryEventFilter) ([]SalaryEvent, int64, error) {
	query := db().Model(&SalaryEvent{})

	if filter.PersonID != nil {
		query = query.Where("person_id = ?", *filter.PersonID)
	}
	if filter.BelongMonth != "" {
		query = query.Where("belong_month = ?", filter.BelongMonth)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}
	if filter.PersonName != "" {
		query = query.Joins("JOIN persons ON persons.id = salary_events.person_id").
			Where("persons.name LIKE ?", "%"+filter.PersonName+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []SalaryEvent
	pageNum := filter.PageNum
	pageSize := filter.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	if err := query.Order("salary_events.created_at DESC").Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func ListTrashSalaryEvents(filter SalaryEventFilter) ([]SalaryEvent, int64, error) {
	query := db().Unscoped().Model(&SalaryEvent{}).Where("deleted_at IS NOT NULL")

	if filter.PersonID != nil {
		query = query.Where("person_id = ?", *filter.PersonID)
	}
	if filter.BelongMonth != "" {
		query = query.Where("belong_month = ?", filter.BelongMonth)
	}
	if filter.EventType != "" {
		query = query.Where("event_type = ?", filter.EventType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []SalaryEvent
	pageNum := filter.PageNum
	pageSize := filter.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	if err := query.Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func GetSalaryEventsByPersonAndMonth(personID uint, belongMonth string) ([]SalaryEvent, error) {
	var events []SalaryEvent
	err := db().Where("person_id = ? AND belong_month = ?", personID, belongMonth).Find(&events).Error
	return events, err
}

func GetLatestPerformanceCoefficientEvent(personID uint, belongMonth string) (*SalaryEvent, error) {
	var event SalaryEvent
	err := db().Where("person_id = ? AND belong_month = ? AND event_type = ?", personID, belongMonth, "绩效系数").
		Order("created_at DESC").First(&event).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func GetSalarySummaryByPersonAndMonth(personID uint, belongMonth string) (*SalarySummary, error) {
	var summary SalarySummary
	err := db().Where("person_id = ? AND belong_month = ?", personID, belongMonth).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func DeleteSalarySummaryByMonth(belongMonth string) error {
	return db().Where("belong_month = ?", belongMonth).Delete(&SalarySummary{}).Error
}

func UpsertSalarySummary(summary *SalarySummary) error {
	var existing SalarySummary
	err := db().Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return db().Create(summary).Error
		}
		return err
	}
	summary.ID = existing.ID
	summary.CreatedAt = existing.CreatedAt
	return db().Save(summary).Error
}

func ListSalarySummaries(filter SalarySummaryFilter) ([]SalarySummaryVO, int64, error) {
	query := db().Model(&SalarySummary{})

	if filter.PersonID != nil {
		query = query.Where("salary_summary.person_id = ?", *filter.PersonID)
	}
	if filter.BelongMonth != "" {
		query = query.Where("salary_summary.belong_month = ?", filter.BelongMonth)
	}
	if filter.PersonName != "" {
		query = query.Joins("JOIN persons ON persons.id = salary_summary.person_id").
			Where("persons.name LIKE ?", "%"+filter.PersonName+"%")
	}
	if filter.AttendanceGroup != "" {
		query = query.Joins("JOIN position_snapshots ON position_snapshots.person_id = salary_summary.person_id").
			Where("position_snapshots.attendance_group = ? AND position_snapshots.effective_start_date <= salary_summary.belong_month || '-01' AND position_snapshots.effective_end_date >= salary_summary.belong_month || '-01'",
				filter.AttendanceGroup)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var summaries []SalarySummary
	pageNum := filter.PageNum
	pageSize := filter.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (pageNum - 1) * pageSize

	if err := query.Order("salary_summary.belong_month DESC, salary_summary.person_id ASC").
		Offset(offset).Limit(pageSize).Find(&summaries).Error; err != nil {
		return nil, 0, err
	}

	result := make([]SalarySummaryVO, 0, len(summaries))
	for _, s := range summaries {
		personName := getPersonName(s.PersonID)
		status := getSummaryStatus(&s)
		result = append(result, SalarySummaryVO{
			ID:                          s.ID,
			PersonID:                    s.PersonID,
			PersonName:                  personName,
			BelongMonth:                 s.BelongMonth,
			SalaryDays:                  s.SalaryDays,
			WeightedBaseSalary:          s.WeightedBaseSalary,
			TotalWorkHours:              s.TotalWorkHours,
			TotalOvertimeWorkdayHours:   s.TotalOvertimeWorkdayHours,
			TotalOvertimeHolidayHours:   s.TotalOvertimeHolidayHours,
			AttendanceSalary:            s.AttendanceSalary,
			OvertimeWorkdaySalary:       s.OvertimeWorkdaySalary,
			OvertimeHolidaySalary:       s.OvertimeHolidaySalary,
			AnnualLeaveCarryoverSalary:  s.AnnualLeaveCarryoverSalary,
			AttendanceBonus:             s.AttendanceBonus,
			PerformanceSalary:           s.PerformanceSalary,
			PostAllowance:               s.PostAllowance,
			MealAllowance:               s.MealAllowance,
			HousingAllowance:            s.HousingAllowance,
			TransportAllowance:          s.TransportAllowance,
			HighTempAllowance:           s.HighTempAllowance,
			InsuranceCompensation:       s.InsuranceCompensation,
			FundCompensation:            s.FundCompensation,
			TotalAdjustment:             s.TotalAdjustment,
			SocialSecurityDeduct:        s.SocialSecurityDeduct,
			HousingFundDeduct:           s.HousingFundDeduct,
			TaxDeduct:                   s.TaxDeduct,
			FinalSalary:                 s.FinalSalary,
			Status:                      status,
			LastCalcAt:                  s.LastCalcAt,
			CreatedAt:                   s.CreatedAt,
			UpdatedAt:                   s.UpdatedAt,
		})
	}

	return result, total, nil
}

func GetAllSalarySummaries(filter SalarySummaryFilter) ([]SalarySummaryVO, error) {
	query := db().Model(&SalarySummary{})

	if filter.PersonID != nil {
		query = query.Where("salary_summary.person_id = ?", *filter.PersonID)
	}
	if filter.BelongMonth != "" {
		query = query.Where("salary_summary.belong_month = ?", filter.BelongMonth)
	}
	if filter.PersonName != "" {
		query = query.Joins("JOIN persons ON persons.id = salary_summary.person_id").
			Where("persons.name LIKE ?", "%"+filter.PersonName+"%")
	}

	var summaries []SalarySummary
	if err := query.Order("salary_summary.belong_month DESC, salary_summary.person_id ASC").Find(&summaries).Error; err != nil {
		return nil, err
	}

	result := make([]SalarySummaryVO, 0, len(summaries))
	for _, s := range summaries {
		personName := getPersonName(s.PersonID)
		status := getSummaryStatus(&s)
		result = append(result, SalarySummaryVO{
			ID:                          s.ID,
			PersonID:                    s.PersonID,
			PersonName:                  personName,
			BelongMonth:                 s.BelongMonth,
			SalaryDays:                  s.SalaryDays,
			WeightedBaseSalary:          s.WeightedBaseSalary,
			TotalWorkHours:              s.TotalWorkHours,
			TotalOvertimeWorkdayHours:   s.TotalOvertimeWorkdayHours,
			TotalOvertimeHolidayHours:   s.TotalOvertimeHolidayHours,
			AttendanceSalary:            s.AttendanceSalary,
			OvertimeWorkdaySalary:       s.OvertimeWorkdaySalary,
			OvertimeHolidaySalary:       s.OvertimeHolidaySalary,
			AnnualLeaveCarryoverSalary:  s.AnnualLeaveCarryoverSalary,
			AttendanceBonus:             s.AttendanceBonus,
			PerformanceSalary:           s.PerformanceSalary,
			PostAllowance:               s.PostAllowance,
			MealAllowance:               s.MealAllowance,
			HousingAllowance:            s.HousingAllowance,
			TransportAllowance:          s.TransportAllowance,
			HighTempAllowance:           s.HighTempAllowance,
			InsuranceCompensation:       s.InsuranceCompensation,
			FundCompensation:            s.FundCompensation,
			TotalAdjustment:             s.TotalAdjustment,
			SocialSecurityDeduct:        s.SocialSecurityDeduct,
			HousingFundDeduct:           s.HousingFundDeduct,
			TaxDeduct:                   s.TaxDeduct,
			FinalSalary:                 s.FinalSalary,
			Status:                      status,
			LastCalcAt:                  s.LastCalcAt,
			CreatedAt:                   s.CreatedAt,
			UpdatedAt:                   s.UpdatedAt,
		})
	}

	return result, nil
}

func getPersonName(personID uint) string {
	var person database.Person
	if err := db().Unscoped().Where("id = ?", personID).First(&person).Error; err != nil {
		return ""
	}
	return person.Name
}

func getSummaryStatus(summary *SalarySummary) string {
	if summary.LastCalcAt == nil {
		return "未核算"
	}

	var attendanceUpdated *time.Time
	var attendanceSalary database.AttendanceSalaryMonthly
	if err := db().Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).
		First(&attendanceSalary).Error; err == nil && attendanceSalary.LastCalcAt != nil {
		attendanceUpdated = attendanceSalary.LastCalcAt
	}

	var maxSalaryEventUpdated time.Time
	db().Model(&SalaryEvent{}).Unscoped().
		Select("MAX(COALESCE(updated_at, created_at))").
		Where("person_id = ? AND belong_month = ?", summary.PersonID, summary.BelongMonth).
		Scan(&maxSalaryEventUpdated)

	var maxPositionEventUpdated time.Time
	belongStart := summary.BelongMonth + "-01"
	db().Model(&database.PositionEvent{}).Unscoped().
		Select("MAX(COALESCE(updated_at, created_at))").
		Where("person_id = ? AND effective_date LIKE ?", summary.PersonID, summary.BelongMonth+"%").
		Scan(&maxPositionEventUpdated)

	var maxLeaveEventUpdated time.Time
	db().Model(&database.LeaveAccountEvent{}).Unscoped().
		Select("MAX(COALESCE(updated_at, created_at))").
		Where("person_id = ? AND effective_date LIKE ?", summary.PersonID, summary.BelongMonth+"%").
		Scan(&maxLeaveEventUpdated)

	_ = belongStart
	if attendanceUpdated != nil && attendanceUpdated.After(*summary.LastCalcAt) {
		return "数据已变动"
	}
	if !maxSalaryEventUpdated.IsZero() && maxSalaryEventUpdated.After(*summary.LastCalcAt) {
		return "数据已变动"
	}
	if !maxPositionEventUpdated.IsZero() && maxPositionEventUpdated.After(*summary.LastCalcAt) {
		return "数据已变动"
	}
	if !maxLeaveEventUpdated.IsZero() && maxLeaveEventUpdated.After(*summary.LastCalcAt) {
		return "数据已变动"
	}

	return "已核算"
}

func GetSummaryByID(id uint) (*SalarySummary, error) {
	var summary SalarySummary
	err := db().Where("id = ?", id).First(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}
