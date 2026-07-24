package services

import (
	"math"
	"strings"
	"time"

	"probig/database"
	"probig/middleware"
	"probig/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListSalaryEvents(personID uint, belongMonth, eventType string, offset, limit int) ([]models.SalaryEvent, int64, error) {
	var events []models.SalaryEvent
	var total int64
	db := database.DB.Model(&models.SalaryEvent{}).Preload("Person")
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if belongMonth != "" {
		db = db.Where("belong_month = ?", belongMonth)
	}
	if eventType != "" {
		db = db.Where("event_type = ?", eventType)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&events).Error
	return events, total, err
}

func CreateSalaryEvent(event *models.SalaryEvent) error {
	month := event.BelongMonth
	_ = month
	return database.DB.Create(event).Error
}

func UpdateSalaryEvent(id uint, updates map[string]interface{}) error {
	var event models.SalaryEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		return err
	}
	if err := middleware.CheckConfigLocked(event.BelongMonth, "salary"); err != nil {
		return err
	}
	return database.DB.Model(&event).Updates(updates).Error
}

func DeleteSalaryEvent(id uint) error {
	var event models.SalaryEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		return err
	}
	if err := middleware.CheckConfigLocked(event.BelongMonth, "salary"); err != nil {
		return err
	}
	return database.DB.Delete(&event).Error
}

func CalcSalary(c *gin.Context, belongMonth string, personIDs []uint) ([]models.SalarySummary, error) {
	startDate, _ := time.Parse("2006-01-02", belongMonth+"-01")
	endDate := startDate.AddDate(0, 1, -1)
	_ = c

	var summaries []models.SalarySummary

	database.DB.Transaction(func(tx *gorm.DB) error {
		var persons []models.Person
		if len(personIDs) > 0 {
			tx.Where("id IN ?", personIDs).Find(&persons)
		} else {
			tx.Find(&persons)
		}

		sickRatio := GetConfigValueFloat("salary.sick_leave_ratio")
		if sickRatio == 0 {
			sickRatio = 0.6
		}
		workdayOvertimeRatio := GetConfigValueFloat("salary.workday_overtime_ratio")
		if workdayOvertimeRatio == 0 {
			workdayOvertimeRatio = 1.5
		}
		holidayOvertimeRatio := GetConfigValueFloat("salary.holiday_overtime_ratio")
		if holidayOvertimeRatio == 0 {
			holidayOvertimeRatio = 2.0
		}
		attendanceBonusRatio := GetConfigValueFloat("salary.attendance_bonus_ratio")
		if attendanceBonusRatio == 0 {
			attendanceBonusRatio = 50
		}
		highTempMonthsStr := GetConfigValue("salary.high_temp_months")
		highTempMonths := parseHighTempMonths(highTempMonthsStr)
		currentMonth := strings.TrimLeft(belongMonth[5:], "0")

		for _, p := range persons {
			var attEvents []models.AttendanceEvent
			tx.Where("person_id = ? AND event_date >= ? AND event_date <= ?", p.ID, startDate, endDate).
				Find(&attEvents)

			var salEvents []models.SalaryEvent
			tx.Where("person_id = ? AND belong_month = ?", p.ID, belongMonth).
				Find(&salEvents)

			// Get monthly snapshot (use mid-month snapshot)
			midDate := startDate.AddDate(0, 0, 15)
			if midDate.After(endDate) {
				midDate = endDate
			}
			var snap models.PositionSnapshot
			snapErr := tx.Where("person_id = ? AND snapshot_date <= ?", p.ID, midDate).
				Order("snapshot_date DESC").First(&snap).Error

			if snapErr != nil || snap.BaseSalary == 0 {
				// no snapshot, skip
				continue
			}

			// Formula 1: daily salary
			salaryDays := float64(snap.SalaryDays)
			if salaryDays <= 0 {
				salaryDays = 21.75
			}
			dailySalary := snap.BaseSalary / salaryDays

			// Count attendance days
			workDays := 0.0
			makeUpDays := 0.0
			sickLeaveDays := 0.0
			personalLeaveDays := 0.0
			paidLeaveDays := 0.0 // annual + statutory + welfare
			otWorkdayHours := 0.0
			otHolidayHours := 0.0
			violationCount := 0

			for _, e := range attEvents {
				hours := 0.0
				if e.Hours != nil {
					hours = *e.Hours
				}
				days := hours / 8.0

				switch e.EventType {
				case "出勤":
					switch e.SubType {
					case "普通出勤":
						workDays += days
					case "补班出勤":
						makeUpDays += days
					}
				case "休假":
					switch e.SubType {
					case "病假":
						sickLeaveDays += days
					case "事假":
						personalLeaveDays += days
					case "年假", "法定假", "福利假":
						paidLeaveDays += days
					}
				case "加班":
					switch e.SubType {
					case "工作日加班":
						otWorkdayHours += hours
					case "节假日加班":
						otHolidayHours += hours
					}
				case "违纪":
					violationCount++
				}
			}

			// Formula 3: attendance salary
			attendanceSalary := (workDays + makeUpDays + sickLeaveDays*sickRatio + paidLeaveDays) * dailySalary

			// Formula 5: overtime salary
			otBase := snap.BaseSalary + snap.MealAllowance
			overtimeSalary := (otBase/salaryDays/8)*otWorkdayHours*workdayOvertimeRatio +
				(otBase/salaryDays/8)*otHolidayHours*holidayOvertimeRatio

			// Formula 4: attendance bonus
			var attendanceBonus float64
			if !snap.HasAttendanceBonus {
				attendanceBonus = 0
			} else if personalLeaveDays > 0 {
				attendanceBonus = 0
			} else {
				bonusBase := workDays + makeUpDays + sickLeaveDays + paidLeaveDays
				bonusBase -= float64(violationCount)
				if bonusBase < 0 {
					bonusBase = 0
				}
				attendanceBonus = bonusBase * attendanceBonusRatio
			}

			// Formula 6: performance salary
			var perfCoeff float64 = 1.0
			for _, e := range salEvents {
				if e.EventType == "绩效调整" {
					perfCoeff = e.Amount
				}
			}
			performanceSalary := snap.PerformanceSalary * perfCoeff

			// Formula 7: total allowance
			totalAllowance := snap.PostAllowance + snap.MealAllowance + snap.HousingAllowance + snap.TransportAllowance +
				snap.InsuranceCompensation + snap.FundCompensation
			if containsString(highTempMonths, currentMonth) {
				totalAllowance += snap.HighTempAllowance
			}

			// Formula 8: total adjustment (non-performance salary events)
			var totalAdjustment float64
			for _, e := range salEvents {
				if e.EventType != "绩效调整" {
					totalAdjustment += e.Amount
				}
			}

			// Formula 9: total deduction
			totalDeduction := snap.SocialSecurityDeduct + snap.HousingFundDeduct

			// Formula 10: final salary
			finalSalary := attendanceSalary + overtimeSalary + attendanceBonus +
				performanceSalary + totalAllowance + totalAdjustment - totalDeduction

			finalSalary = math.Round(finalSalary*100) / 100

			now := time.Now()
			s := models.SalarySummary{
				PersonID:          p.ID,
				BelongMonth:       belongMonth,
				AttendanceSalary:  math.Round(attendanceSalary*100) / 100,
				OvertimeSalary:    math.Round(overtimeSalary*100) / 100,
				AttendanceBonus:   math.Round(attendanceBonus*100) / 100,
				PerformanceSalary: math.Round(performanceSalary*100) / 100,
				TotalAllowance:    math.Round(totalAllowance*100) / 100,
				TotalAdjustment:   math.Round(totalAdjustment*100) / 100,
				TotalDeduction:    math.Round(totalDeduction*100) / 100,
				FinalSalary:       finalSalary,
				LastCalcAt:        &now,
			}

			var existing models.SalarySummary
			result := tx.Where("person_id = ? AND belong_month = ?", p.ID, belongMonth).First(&existing)
			if result.Error == nil {
				tx.Model(&existing).Updates(map[string]interface{}{
					"attendance_salary":   s.AttendanceSalary,
					"overtime_salary":     s.OvertimeSalary,
					"attendance_bonus":    s.AttendanceBonus,
					"performance_salary":  s.PerformanceSalary,
					"total_allowance":     s.TotalAllowance,
					"total_adjustment":    s.TotalAdjustment,
					"total_deduction":     s.TotalDeduction,
					"final_salary":        s.FinalSalary,
					"last_calc_at":        &now,
				})
			} else {
				tx.Create(&s)
			}
		}
		return nil
	})

	database.DB.Where("belong_month = ?", belongMonth).Preload("Person").Find(&summaries)
	return summaries, nil
}

func ListSalarySummaries(belongMonth string, offset, limit int) ([]models.SalarySummary, int64, error) {
	var summaries []models.SalarySummary
	var total int64
	db := database.DB.Model(&models.SalarySummary{}).Preload("Person")
	if belongMonth != "" {
		db = db.Where("belong_month = ?", belongMonth)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("person_id ASC").Find(&summaries).Error
	return summaries, total, err
}

func LockSalarySummary(id uint, lock bool) error {
	return database.DB.Model(&models.SalarySummary{}).Where("id = ?", id).Update("is_locked", lock).Error
}

func parseHighTempMonths(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "[]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var months []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		months = append(months, p)
	}
	return months
}

func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}
