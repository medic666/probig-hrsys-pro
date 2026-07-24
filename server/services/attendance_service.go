package services

import (
	"fmt"
	"strings"
	"time"

	"probig/database"
	"probig/middleware"
	"probig/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListAttendanceEvents(personID uint, eventType, subType, startDate, endDate string, offset, limit int) ([]models.AttendanceEvent, int64, error) {
	var events []models.AttendanceEvent
	var total int64
	db := database.DB.Model(&models.AttendanceEvent{}).Preload("Person")
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if eventType != "" {
		db = db.Where("event_type = ?", eventType)
	}
	if subType != "" {
		db = db.Where("sub_type = ?", subType)
	}
	if startDate != "" {
		db = db.Where("event_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("event_date <= ?", endDate)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("event_date DESC, id DESC").Find(&events).Error
	return events, total, err
}

func CreateAttendanceEvent(event *models.AttendanceEvent) error {
	if event.EventDate.IsZero() {
		return fmt.Errorf("事件日期不能为空")
	}
	return database.DB.Create(event).Error
}

func BatchCreateAttendanceEvents(c *gin.Context, personIDs []uint, startDate, endDate time.Time, eventType, subType string, hours *float64, lateMin *int, leaveAdjAmount *float64, isSpecial *bool, remark string) (string, error) {
	batchID := fmt.Sprintf("BATCH_%d", time.Now().UnixNano())
	userID := middleware.GetUserID(c)
	_ = userID

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, pid := range personIDs {
			d := startDate
			for !d.After(endDate) {
				evt := models.AttendanceEvent{
					PersonID:          pid,
					EventDate:         d,
					EventType:         eventType,
					SubType:           subType,
					Hours:             hours,
					LateMinutes:       lateMin,
					LeaveAdjustAmount: leaveAdjAmount,
					IsSpecialApproval: isSpecial,
					Remark:            remark,
					BatchID:           batchID,
				}
				if err := tx.Create(&evt).Error; err != nil {
					return err
				}
				d = d.AddDate(0, 0, 1)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return batchID, nil
}

func UpdateAttendanceEvent(id uint, updates map[string]interface{}) error {
	var event models.AttendanceEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		return err
	}
	month := event.EventDate.Format("2006-01")
	if err := middleware.CheckConfigLocked(month, "attendance"); err != nil {
		return err
	}
	return database.DB.Model(&event).Updates(updates).Error
}

func DeleteAttendanceEvent(id uint) error {
	var event models.AttendanceEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		return err
	}
	month := event.EventDate.Format("2006-01")
	if err := middleware.CheckConfigLocked(month, "attendance"); err != nil {
		return err
	}
	return database.DB.Delete(&event).Error
}

func CalcAttendanceSummary(c *gin.Context, belongMonth string, personIDs []uint) ([]models.AttendanceSummary, error) {
	var summaries []models.AttendanceSummary

	startDate, _ := time.Parse("2006-01-02", belongMonth+"-01")
	endDate := startDate.AddDate(0, 1, -1)

	database.DB.Transaction(func(tx *gorm.DB) error {
		var persons []models.Person
		if len(personIDs) > 0 {
			tx.Where("id IN ?", personIDs).Find(&persons)
		} else {
			tx.Find(&persons)
		}

		for _, p := range persons {
			var events []models.AttendanceEvent
			tx.Where("person_id = ? AND event_date >= ? AND event_date <= ?", p.ID, startDate, endDate).
				Find(&events)

			now := time.Now()
			s := models.AttendanceSummary{
				PersonID:    p.ID,
				BelongMonth: belongMonth,
				LastCalcAt:  &now,
			}

			for _, e := range events {
				hours := 0.0
				if e.Hours != nil {
					hours = *e.Hours
				}
				days := hours / 8.0

				switch e.EventType {
				case "出勤":
					switch e.SubType {
					case "普通出勤":
						s.WorkDays += days
					case "补班出勤":
						s.MakeUpDays += days
					}
				case "休假":
					switch e.SubType {
					case "病假":
						s.SickLeaveDays += days
					case "事假":
						s.PersonalLeaveDays += days
					case "年假":
						s.AnnualLeaveDays += days
					case "法定假":
						s.StatutoryLeaveDays += days
					case "福利假":
						s.WelfareLeaveDays += days
					}
				case "加班":
					switch e.SubType {
					case "工作日加班":
						s.OvertimeWorkdayHours += hours
					case "节假日加班":
						s.OvertimeHolidayHours += hours
					}
				case "违纪":
					s.ViolationCount++
				}
			}

			var existing models.AttendanceSummary
			result := tx.Where("person_id = ? AND belong_month = ?", p.ID, belongMonth).First(&existing)
			if result.Error == nil {
				tx.Model(&existing).Updates(map[string]interface{}{
					"work_days":               s.WorkDays,
					"make_up_days":            s.MakeUpDays,
					"sick_leave_days":         s.SickLeaveDays,
					"personal_leave_days":     s.PersonalLeaveDays,
					"annual_leave_days":       s.AnnualLeaveDays,
					"statutory_leave_days":    s.StatutoryLeaveDays,
					"welfare_leave_days":      s.WelfareLeaveDays,
					"overtime_workday_hours":  s.OvertimeWorkdayHours,
					"overtime_holiday_hours":  s.OvertimeHolidayHours,
					"violation_count":         s.ViolationCount,
					"last_calc_at":            &now,
				})
			} else {
				tx.Create(&s)
			}
		}
		return nil
	})

	database.DB.Where("belong_month = ?", belongMonth).Preload("Person").Find(&summaries, "person_id IN ?", personIDs)
	return summaries, nil
}

func ListAttendanceSummaries(belongMonth string, offset, limit int) ([]models.AttendanceSummary, int64, error) {
	var summaries []models.AttendanceSummary
	var total int64
	db := database.DB.Model(&models.AttendanceSummary{}).Preload("Person")
	if belongMonth != "" {
		db = db.Where("belong_month = ?", belongMonth)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("person_id ASC").Find(&summaries).Error
	return summaries, total, err
}

func LockAttendanceSummary(id uint, lock bool) error {
	return database.DB.Model(&models.AttendanceSummary{}).Where("id = ?", id).Update("is_locked", lock).Error
}

func CalcAnnualLeaveBalance(personID uint) (float64, error) {
	var events []models.AttendanceEvent
	database.DB.Where("person_id = ?", personID).Find(&events)

	var allocated, carried, used float64
	for _, e := range events {
		if e.EventType == "年假调整" {
			if e.LeaveAdjustAmount != nil {
				switch e.SubType {
				case "年假配发":
					allocated += *e.LeaveAdjustAmount
				case "年假结转":
					carried += *e.LeaveAdjustAmount
				}
			}
		}
		if e.EventType == "休假" && e.SubType == "年假" {
			if e.Hours != nil {
				used += *e.Hours / 8.0
			}
		}
	}
	return allocated + carried - used, nil
}

func AnnualLeaveAnniversary() ([]uint, error) {
	var personIDs []uint
	var persons []models.Person
	database.DB.Find(&persons)

	for _, p := range persons {
		snap, err := GetPersonLatestSnapshot(p.ID)
		if err != nil || snap.EntryDate == nil {
			continue
		}
		if !snap.HasAnnualLeave {
			continue
		}
		entryDate := *snap.EntryDate
		today := time.Now()
		monthsPassed := (today.Year()-entryDate.Year())*12 + int(today.Month()) - int(entryDate.Month())
		if monthsPassed > 0 && monthsPassed%12 == 0 {
			personIDs = append(personIDs, p.ID)
		}
	}
	return personIDs, nil
}

func GetAttendanceEventsByMonth(personID uint, month string) ([]models.AttendanceEvent, error) {
	startDate, _ := time.Parse("2006-01-02", month+"-01")
	endDate := startDate.AddDate(0, 1, -1)
	var events []models.AttendanceEvent
	err := database.DB.Where("person_id = ? AND event_date >= ? AND event_date <= ?", personID, startDate, endDate).
		Order("event_date ASC").Find(&events).Error
	return events, err
}

func CheckAttendanceLocked(month string) bool {
	var count int64
	database.DB.Model(&models.AttendanceSummary{}).Where("belong_month = ? AND is_locked = ?", month, true).Count(&count)
	return count > 0
}

func CheckSalaryLocked(month string) bool {
	var count int64
	database.DB.Model(&models.SalarySummary{}).Where("belong_month = ? AND is_locked = ?", month, true).Count(&count)
	return count > 0
}

func GetConfigValue(key string) string {
	var config models.SysConfig
	if err := database.DB.Where("config_key = ?", key).First(&config).Error; err != nil {
		return ""
	}
	return config.ConfigValue
}

func GetConfigValueBool(key string) bool {
	v := GetConfigValue(key)
	return strings.ToLower(v) == "true"
}

func GetConfigValueFloat(key string) float64 {
	v := GetConfigValue(key)
	var f float64
	fmt.Sscanf(v, "%f", &f)
	return f
}

func GetConfigValueInt(key string) int {
	v := GetConfigValue(key)
	var i int
	fmt.Sscanf(v, "%d", &i)
	return i
}
