package services

import (
	"fmt"
	"sort"
	"time"

	"probig/database"
	"probig/models"

	"gorm.io/gorm"
)

func ListPositionEvents(personID uint, offset, limit int) ([]models.PositionEvent, int64, error) {
	var events []models.PositionEvent
	var total int64
	db := database.DB.Model(&models.PositionEvent{}).Preload("Person")
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("effective_date DESC, created_at DESC").Find(&events).Error
	return events, total, err
}

func GetPositionEvent(id uint) (*models.PositionEvent, error) {
	var event models.PositionEvent
	err := database.DB.Preload("Person").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func CreatePositionEvent(event *models.PositionEvent) error {
	if event.EffectiveDate.IsZero() {
		return fmt.Errorf("生效日期不能为空")
	}
	err := database.DB.Create(event).Error
	if err != nil {
		return err
	}
	go RebuildSnapshots(event.PersonID)
	return nil
}

func UpdatePositionEvent(id uint, updates map[string]interface{}) error {
	var event models.PositionEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		return err
	}
	personID := event.PersonID
	if err := database.DB.Model(&event).Updates(updates).Error; err != nil {
		return err
	}
	go RebuildSnapshots(personID)
	return nil
}

func DeletePositionEvent(id uint) error {
	var event models.PositionEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		return err
	}
	personID := event.PersonID
	if err := database.DB.Delete(&event).Error; err != nil {
		return err
	}
	go RebuildSnapshots(personID)
	return nil
}

func RebuildSnapshots(personID uint) {
	database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("person_id = ?", personID).Delete(&models.PositionSnapshot{})

		var events []models.PositionEvent
		tx.Where("person_id = ?", personID).Order("effective_date ASC, created_at ASC").Find(&events)

		if len(events) == 0 {
			return nil
		}

		type dayMap = map[string]interface{}
		byDate := make(map[string]dayMap)

		for _, e := range events {
			dateStr := e.EffectiveDate.Format("2006-01-02")
			if _, ok := byDate[dateStr]; !ok {
				byDate[dateStr] = make(dayMap)
			}
			if e.AttendanceGroup != nil {
				byDate[dateStr]["attendance_group"] = *e.AttendanceGroup
			}
			if e.EntryDate != nil {
				byDate[dateStr]["entry_date"] = *e.EntryDate
			}
			if e.LeaveDate != nil {
				byDate[dateStr]["leave_date"] = *e.LeaveDate
			}
			if e.HasAnnualLeave != nil {
				byDate[dateStr]["has_annual_leave"] = *e.HasAnnualLeave
			}
			if e.HasAttendanceBonus != nil {
				byDate[dateStr]["has_attendance_bonus"] = *e.HasAttendanceBonus
			}
			if e.BaseSalary != nil {
				byDate[dateStr]["base_salary"] = *e.BaseSalary
			}
			if e.PerformanceSalary != nil {
				byDate[dateStr]["performance_salary"] = *e.PerformanceSalary
			}
			if e.SalaryDays != nil {
				byDate[dateStr]["salary_days"] = *e.SalaryDays
			}
			if e.PostAllowance != nil {
				byDate[dateStr]["post_allowance"] = *e.PostAllowance
			}
			if e.MealAllowance != nil {
				byDate[dateStr]["meal_allowance"] = *e.MealAllowance
			}
			if e.HousingAllowance != nil {
				byDate[dateStr]["housing_allowance"] = *e.HousingAllowance
			}
			if e.TransportAllowance != nil {
				byDate[dateStr]["transport_allowance"] = *e.TransportAllowance
			}
			if e.HighTempAllowance != nil {
				byDate[dateStr]["high_temp_allowance"] = *e.HighTempAllowance
			}
			if e.InsuranceCompensation != nil {
				byDate[dateStr]["insurance_compensation"] = *e.InsuranceCompensation
			}
			if e.FundCompensation != nil {
				byDate[dateStr]["fund_compensation"] = *e.FundCompensation
			}
			if e.SocialSecurityDeduct != nil {
				byDate[dateStr]["social_security_deduct"] = *e.SocialSecurityDeduct
			}
			if e.HousingFundDeduct != nil {
				byDate[dateStr]["housing_fund_deduct"] = *e.HousingFundDeduct
			}
		}

		var dates []string
		for d := range byDate {
			dates = append(dates, d)
		}
		sort.Strings(dates)

		current := make(map[string]interface{})
		// apply events day by day, carrying forward
		for _, d := range dates {
			m := byDate[d]
			for k, v := range m {
				current[k] = v
			}
			byDate[d] = current
		}

		// expand to all days from min to max
		minDate, _ := time.Parse("2006-01-02", dates[0])
		maxDate, _ := time.Parse("2006-01-02", dates[len(dates)-1])

		currentState := make(map[string]interface{})
		for d := minDate; !d.After(maxDate); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			if dm, ok := byDate[dateStr]; ok {
				for k, v := range dm {
					currentState[k] = v
				}
			}

			if len(currentState) == 0 {
				continue
			}

			snap := models.PositionSnapshot{
				PersonID:     personID,
				SnapshotDate: d,
			}
			applySnapshotFields(&snap, currentState)
			tx.Create(&snap)
		}
		return nil
	})
}

func applySnapshotFields(snap *models.PositionSnapshot, state map[string]interface{}) {
	if v, ok := state["attendance_group"].(string); ok {
		snap.AttendanceGroup = v
	}
	if v, ok := state["entry_date"].(time.Time); ok {
		snap.EntryDate = &v
	}
	if v, ok := state["leave_date"].(time.Time); ok {
		snap.LeaveDate = &v
	}
	if v, ok := state["has_annual_leave"].(bool); ok {
		snap.HasAnnualLeave = v
	}
	if v, ok := state["has_attendance_bonus"].(bool); ok {
		snap.HasAttendanceBonus = v
	}
	if v, ok := state["base_salary"].(float64); ok {
		snap.BaseSalary = v
	}
	if v, ok := state["performance_salary"].(float64); ok {
		snap.PerformanceSalary = v
	}
	if v, ok := state["salary_days"].(int); ok {
		snap.SalaryDays = v
	}
	if v, ok := state["post_allowance"].(float64); ok {
		snap.PostAllowance = v
	}
	if v, ok := state["meal_allowance"].(float64); ok {
		snap.MealAllowance = v
	}
	if v, ok := state["housing_allowance"].(float64); ok {
		snap.HousingAllowance = v
	}
	if v, ok := state["transport_allowance"].(float64); ok {
		snap.TransportAllowance = v
	}
	if v, ok := state["high_temp_allowance"].(float64); ok {
		snap.HighTempAllowance = v
	}
	if v, ok := state["insurance_compensation"].(float64); ok {
		snap.InsuranceCompensation = v
	}
	if v, ok := state["fund_compensation"].(float64); ok {
		snap.FundCompensation = v
	}
	if v, ok := state["social_security_deduct"].(float64); ok {
		snap.SocialSecurityDeduct = v
	}
	if v, ok := state["housing_fund_deduct"].(float64); ok {
		snap.HousingFundDeduct = v
	}
}

func GetPositionSnapshot(personID uint, dateStr string) (*models.PositionSnapshot, error) {
	var snap models.PositionSnapshot
	date, _ := time.Parse("2006-01-02", dateStr)
	err := database.DB.Where("person_id = ? AND snapshot_date <= ?", personID, date).
		Order("snapshot_date DESC").First(&snap).Error
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func ListPositionSnapshots(personID uint, offset, limit int) ([]models.PositionSnapshot, int64, error) {
	var snaps []models.PositionSnapshot
	var total int64
	db := database.DB.Model(&models.PositionSnapshot{}).Where("person_id = ?", personID)
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("snapshot_date DESC").Find(&snaps).Error
	return snaps, total, err
}

func GetPersonLatestSnapshot(personID uint) (*models.PositionSnapshot, error) {
	var snap models.PositionSnapshot
	err := database.DB.Where("person_id = ?", personID).Order("snapshot_date DESC").First(&snap).Error
	if err != nil {
		return nil, err
	}
	return &snap, nil
}
