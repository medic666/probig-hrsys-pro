package service

import (
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func RebuildDailyProjection(tx *gorm.DB, personID uint, workDate utils.DateOnly) error {
	tx.Where("person_id = ? AND work_date = ?", personID, workDate).Delete(&model.AttendanceDailyProjection{})

	var events []model.AttendanceEvent
	tx.Where("person_id = ? AND event_date = ?", personID, workDate).Find(&events)

	if len(events) == 0 {
		return nil
	}

	var workHours, overtimeWorkday, overtimeHoliday float64
	var hasPersonalLeave bool
	var violationCount int
	var punchTime, remark string

	for _, e := range events {
		punchTime = e.PunchTime
		remark = e.Remark

		switch e.EventType {
		case "出勤":
			switch e.SubType {
			case "普通出勤", "补班出勤", "外勤出勤":
				workHours += e.Hours
			case "工作日加班":
				overtimeWorkday += e.Hours
			case "节假日加班":
				overtimeHoliday += e.Hours
			default:
				workHours += e.Hours
			}
		case "休假":
			switch e.SubType {
			case "事假":
				hasPersonalLeave = true
			case "病假":
				workHours += e.Hours * getSickLeaveRatio()
			case "年假", "调休", "法定假", "福利假":
				workHours += e.Hours
			}
		case "加班":
			switch e.SubType {
			case "工作日加班":
				overtimeWorkday += e.Hours
			case "节假日加班":
				overtimeHoliday += e.Hours
			}
		case "违纪":
			violationCount++
		}
	}

	proj := model.AttendanceDailyProjection{
		PersonID:             personID,
		WorkDate:             workDate,
		PunchTime:            punchTime,
		WorkHours:            workHours,
		OvertimeWorkdayHours: overtimeWorkday,
		OvertimeHolidayHours: overtimeHoliday,
		HasPersonalLeave:     hasPersonalLeave,
		ViolationCount:       violationCount,
		Remark:               remark,
		LastCalcAt:           utils.DateOnlyFromTime(time.Now()),
	}
	return tx.Create(&proj).Error
}

func getSickLeaveRatio() float64 {
	v := GetConfigValueOrDefault("attendance.sick_leave_ratio", "0.8")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func GetDailyProjections(personID uint, dateStart, dateEnd string) ([]model.AttendanceDailyProjection, error) {
	tx := dao.DB.Model(&model.AttendanceDailyProjection{})
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	if dateStart != "" {
		tx = tx.Where("work_date >= ?", dateStart)
	}
	if dateEnd != "" {
		tx = tx.Where("work_date <= ?", dateEnd)
	}
	var list []model.AttendanceDailyProjection
	err := tx.Order("work_date ASC").Find(&list).Error
	return list, err
}

func getOvertimeWorkdayRatio() float64 {
	v := GetConfigValueOrDefault("attendance.overtime_workday_ratio", "1.5")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func getOvertimeHolidayRatio() float64 {
	v := GetConfigValueOrDefault("attendance.overtime_holiday_ratio", "2.0")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func getAttendanceBonusDaily() float64 {
	v := GetConfigValueOrDefault("attendance.full_attendance_bonus", "50")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func getWorkHoursPerDay() float64 {
	v := GetConfigValueOrDefault("system.work_hours_per_day", "8")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}
