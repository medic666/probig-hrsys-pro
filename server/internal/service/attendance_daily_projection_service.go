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

	var daily model.AttendanceDaily
	if err := tx.Where("person_id = ? AND event_date = ?", personID, workDate).First(&daily).Error; err != nil {
		return nil
	}

	proj := model.AttendanceDailyProjection{
		PersonID: personID,
		WorkDate: workDate,
		Status:   daily.Status,
		PunchTime: daily.PunchTime,
		Remark:   daily.Remark,
		LastCalcAt: time.Now(),
	}

	if daily.Status == "pending" {
		return tx.Create(&proj).Error
	}

	var details []model.AttendanceEventDetail
	details, err := GetDetailsByDailyID(tx, daily.ID)
	if err != nil {
		return err
	}

	var workHours, overtimeWorkday, overtimeHoliday float64
	var hasPersonalLeave bool
	var violationCount int

	for _, e := range details {
		switch e.EventType {
		case "出勤":
			workHours += e.Hours
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

	proj.WorkHours = workHours
	proj.OvertimeWorkdayHours = overtimeWorkday
	proj.OvertimeHolidayHours = overtimeHoliday
	proj.HasPersonalLeave = hasPersonalLeave
	proj.ViolationCount = violationCount
	return tx.Create(&proj).Error
}

func getSickLeaveRatio() float64 {
	v := GetConfigValueOrDefault("attendance.sick_leave_ratio", "0.8")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

// DailyProjectionListQuery 日记工时投影列表查询（列表与导出共用）
type DailyProjectionListQuery struct {
	PageNum   int
	PageSize  int
	PersonID  uint
	DateStart string
	DateEnd   string
}

func GetDailyProjections(q DailyProjectionListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.AttendanceDailyProjection{})
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}
	if q.DateStart != "" {
		tx = tx.Where("work_date >= ?", q.DateStart)
	}
	if q.DateEnd != "" {
		tx = tx.Where("work_date <= ?", q.DateEnd)
	}
	var total int64
	tx.Count(&total)
	var list []model.AttendanceDailyProjection
	offset := (q.PageNum - 1) * q.PageSize
	err := tx.Order("work_date ASC").Offset(offset).Limit(q.PageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	ids := make([]uint, len(list))
	for i, p := range list {
		ids[i] = p.PersonID
	}
	nameMap := PersonNameMap(ids)

	result := make([]map[string]interface{}, len(list))
	for i, p := range list {
		item := map[string]interface{}{
			"id":                     p.ID,
			"person_id":              p.PersonID,
			"person_name":            nameMap[p.PersonID],
			"work_date":              p.WorkDate,
			"punch_time":             p.PunchTime,
			"work_hours":             p.WorkHours,
			"overtime_workday_hours": p.OvertimeWorkdayHours,
			"overtime_holiday_hours": p.OvertimeHolidayHours,
			"has_personal_leave":     p.HasPersonalLeave,
			"violation_count":        p.ViolationCount,
			"remark":                 p.Remark,
			"status":                 p.Status,
			"last_calc_at":           p.LastCalcAt,
		}
		result[i] = item
	}
	return result, total, nil
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
	v := GetConfigValueOrDefault("attendance.work_hours_per_day", "8")
	f, _ := strconv.ParseFloat(v, 64)
	return f
}
