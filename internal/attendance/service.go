package attendance

import (
	"fmt"
	"strconv"
	"time"

	"probig/internal/pkg/config"
	"probig/internal/pkg/utils"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) ListEvents(pageNum, pageSize int, personID uint, startDate, endDate, eventType, subType string) ([]AttendanceEvent, int64, error) {
	var list []AttendanceEvent
	var total int64
	db := s.DB.Model(&AttendanceEvent{})
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if startDate != "" {
		db = db.Where("event_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("event_date <= ?", endDate)
	}
	if eventType != "" {
		db = db.Where("event_type = ?", eventType)
	}
	if subType != "" {
		db = db.Where("sub_type = ?", subType)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("event_date desc").Find(&list).Error
	return list, total, err
}

func (s *Service) CreateEvent(req map[string]interface{}) (uint, error) {
	event := AttendanceEvent{
		PersonID:    uint(getF(req, "person_id")),
		EventDate:   getS(req, "event_date"),
		PunchTime:   getS(req, "punch_time"),
		EventType:   getS(req, "event_type"),
		SubType:     getS(req, "sub_type"),
		Hours:       getF(req, "hours"),
		LateMinutes: int(getF(req, "late_minutes")),
		Remark:      getS(req, "remark"),
	}
	if v, ok := req["is_special_approval"]; ok {
		if b, ok := v.(bool); ok {
			event.IsSpecialApproval = b
		}
	}
	if err := s.DB.Create(&event).Error; err != nil {
		return 0, err
	}
	go s.RebuildDaily(event.PersonID, event.EventDate)

	if event.SubType == "年假" || event.SubType == "调休" {
		go s.triggerLeaveBalanceRebuild(event.PersonID, event.SubType)
	}

	return event.ID, nil
}

func (s *Service) CreateEventsBatch(personIDs []uint, req map[string]interface{}) (int, []string) {
	successCount := 0
	var errors []string
	for _, pid := range personIDs {
		req["person_id"] = float64(pid)
		_, err := s.CreateEvent(req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("person_id=%d: %v", pid, err))
		} else {
			successCount++
		}
	}
	return successCount, errors
}

func (s *Service) CreateCrossDayEvent(personID uint, startDate, endDate, eventType, subType string, hoursPerDay float64, remark string) error {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	for d := start; !d.After(end); d = d.Add(24 * time.Hour) {
		event := AttendanceEvent{
			PersonID:         personID,
			EventDate:        d.Format("2006-01-02"),
			EventType:        eventType,
			SubType:          subType,
			Hours:            hoursPerDay,
			IsSpecialApproval: false,
			Remark:           remark,
		}
		if err := s.DB.Create(&event).Error; err != nil {
			return err
		}
		go s.RebuildDaily(event.PersonID, event.EventDate)
	}
	if subType == "年假" || subType == "调休" {
		go s.triggerLeaveBalanceRebuild(personID, subType)
	}
	return nil
}

func (s *Service) UpdateEvent(id uint, req map[string]interface{}) error {
	var event AttendanceEvent
	if err := s.DB.First(&event, id).Error; err != nil {
		return err
	}
	oldSubType := event.SubType
	oldEventDate := event.EventDate

	updates := map[string]interface{}{}
	for _, k := range []string{"event_date", "punch_time", "event_type", "sub_type", "remark"} {
		if v := getS(req, k); v != "" {
			updates[k] = v
		}
	}
	if v, ok := req["hours"]; ok {
		updates["hours"] = v
	}
	if v, ok := req["late_minutes"]; ok {
		updates["late_minutes"] = v
	}
	if v, ok := req["is_special_approval"]; ok {
		updates["is_special_approval"] = v
	}
	if err := s.DB.Model(&event).Updates(updates).Error; err != nil {
		return err
	}
	go s.RebuildDaily(event.PersonID, event.EventDate)
	if newDate, ok := updates["event_date"]; ok {
		go s.RebuildDaily(event.PersonID, oldEventDate)
		_ = newDate
	}
	newSubType := event.SubType
	if v, ok := updates["sub_type"]; ok {
		newSubType, _ = v.(string)
	}
	if oldSubType == "年假" || oldSubType == "调休" || newSubType == "年假" || newSubType == "调休" {
		if oldSubType == "年假" || oldSubType == "调休" {
			go s.triggerLeaveBalanceRebuild(event.PersonID, oldSubType)
		}
		if newSubType == "年假" || newSubType == "调休" {
			go s.triggerLeaveBalanceRebuild(event.PersonID, newSubType)
		}
	}
	return nil
}

func (s *Service) DeleteEvent(id uint) error {
	var event AttendanceEvent
	if err := s.DB.First(&event, id).Error; err != nil {
		return err
	}
	if err := s.DB.Delete(&event).Error; err != nil {
		return err
	}
	go s.RebuildDaily(event.PersonID, event.EventDate)
	if event.SubType == "年假" || event.SubType == "调休" {
		go s.triggerLeaveBalanceRebuild(event.PersonID, event.SubType)
	}
	return nil
}

func (s *Service) RebuildDaily(personID uint, date string) error {
	var events []AttendanceEvent
	s.DB.Where("person_id = ? AND event_date = ?", personID, date).Find(&events)

	s.DB.Where("person_id = ? AND work_date = ?", personID, date).Delete(&AttendanceDaily{})

	if len(events) == 0 {
		return nil
	}

	now := time.Now()
	daily := AttendanceDaily{
		PersonID:   personID,
		WorkDate:   date,
		LastCalcAt: now,
	}

	sickRatio := 0.6
	if v := config.GetConfig("attendance.sick_leave_ratio"); v != "" {
		sickRatio = utils.ParseFloat(v)
	}

	for _, e := range events {
		daily.PunchTime = e.PunchTime
		daily.Remark = e.Remark
		switch {
		case e.EventType == EventTypeAttendance && (e.SubType == "普通出勤" || e.SubType == "补班出勤" || e.SubType == "外勤出勤"):
			daily.WorkHours += e.Hours
		case e.EventType == EventTypeLeave:
			switch e.SubType {
			case "病假":
				daily.WorkHours += e.Hours * sickRatio
			case "事假":
				daily.HasPersonalLeave = true
			case "年假", "调休", "法定假", "福利假":
				daily.WorkHours += e.Hours
			}
		case e.EventType == EventTypeOvertime:
			switch e.SubType {
			case "工作日加班":
				daily.OvertimeWorkdayHours += e.Hours
			case "节假日加班":
				daily.OvertimeHolidayHours += e.Hours
			}
		case e.EventType == EventTypeViolation:
			daily.ViolationCount++
		}
	}

	return s.DB.Create(&daily).Error
}

func (s *Service) GetDailyList(pageNum, pageSize int, personID uint, startDate, endDate string) ([]AttendanceDaily, int64, error) {
	var list []AttendanceDaily
	var total int64
	db := s.DB.Model(&AttendanceDaily{})
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if startDate != "" {
		db = db.Where("work_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("work_date <= ?", endDate)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("work_date desc").Find(&list).Error
	return list, total, err
}

func (s *Service) GetDailyEvents(personID uint, date string) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	err := s.DB.Where("person_id = ? AND event_date = ?", personID, date).Find(&events).Error
	return events, err
}

func (s *Service) CalculateMonthlySalary(personID uint, belongMonth string) error {
	s.DB.Where("person_id = ? AND belong_month = ?", personID, belongMonth).Delete(&AttendanceSalary{})

	monthStart := belongMonth + "-01"
	y, _ := strconv.Atoi(belongMonth[:4])
	m, _ := strconv.Atoi(belongMonth[5:7])
	monthEnd := time.Date(y, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	var dailies []AttendanceDaily
	s.DB.Where("person_id = ? AND work_date >= ? AND work_date <= ?", personID, monthStart, monthEnd).Find(&dailies)
	if len(dailies) == 0 {
		return nil
	}

	snapshots, err := s.getSnapshotsInRange(personID, monthStart, monthEnd)
	if err != nil || len(snapshots) == 0 {
		return nil
	}

	weightedBase, weightedMeal, weightedSalaryDays := s.calcWeighted(snapshots, monthStart, monthEnd)
	salaryDays := weightedSalaryDays

	var totalWorkHours, totalOTWorkday, totalOTHoliday float64
	hasPersonalLeave := false
	totalViolations := 0

	for _, d := range dailies {
		totalWorkHours += d.WorkHours
		totalOTWorkday += d.OvertimeWorkdayHours
		totalOTHoliday += d.OvertimeHolidayHours
		if d.HasPersonalLeave {
			hasPersonalLeave = true
		}
		totalViolations += d.ViolationCount
	}

	hoursPerDay := 8.0
	if v := config.GetConfig("system.work_hours_per_day"); v != "" {
		hoursPerDay = utils.ParseFloat(v)
	}
	otWorkdayRatio := 1.5
	if v := config.GetConfig("attendance.overtime_workday_ratio"); v != "" {
		otWorkdayRatio = utils.ParseFloat(v)
	}
	otHolidayRatio := 2.0
	if v := config.GetConfig("attendance.overtime_holiday_ratio"); v != "" {
		otHolidayRatio = utils.ParseFloat(v)
	}
	bonusDaily := 50.0
	if v := config.GetConfig("attendance.bonus_daily"); v != "" {
		bonusDaily = utils.ParseFloat(v)
	}

	if salaryDays == 0 {
		salaryDays = 21
	}

	hourlyRate := weightedBase / float64(salaryDays) / hoursPerDay
	attendanceSalary := totalWorkHours * hourlyRate
	otWorkdaySalary := totalOTWorkday * (weightedBase + weightedMeal) / float64(salaryDays) / hoursPerDay * otWorkdayRatio
	otHolidaySalary := totalOTHoliday * (weightedBase + weightedMeal) / float64(salaryDays) / hoursPerDay * otHolidayRatio

	attendanceBonus := 0.0
	hasNoBonusSnap := false
	for _, snap := range snapshots {
		if !snap.HasAttendanceBonus {
			hasNoBonusSnap = true
			break
		}
	}
	if !hasNoBonusSnap && !hasPersonalLeave {
		days := totalWorkHours / hoursPerDay
		attendanceBonus = utils.MaxFloat(days-float64(totalViolations), 0) * bonusDaily
	}

	now := time.Now()
	as := AttendanceSalary{
		PersonID:              personID,
		BelongMonth:           belongMonth,
		SalaryDays:            salaryDays,
		WeightedBaseSalary:    weightedBase,
		WeightedMealAllowance: weightedMeal,
		TotalWorkHours:        totalWorkHours,
		TotalOvertimeWorkdayH: totalOTWorkday,
		TotalOvertimeHolidayH: totalOTHoliday,
		AttendanceSalary:      attendanceSalary,
		OvertimeWorkdaySalary: otWorkdaySalary,
		OvertimeHolidaySalary: otHolidaySalary,
		HasPersonalLeaveMonth: hasPersonalLeave,
		TotalViolationCount:   totalViolations,
		AttendanceBonus:       attendanceBonus,
		LastCalcAt:            &now,
	}
	return s.DB.Create(&as).Error
}

func (s *Service) GetMonthlySalaryList(belongMonth string, personID uint, pageNum, pageSize int) ([]AttendanceSalary, int64, error) {
	var list []AttendanceSalary
	var total int64
	db := s.DB.Model(&AttendanceSalary{})
	if belongMonth != "" {
		db = db.Where("belong_month = ?", belongMonth)
	}
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("belong_month desc, person_id asc").Find(&list).Error
	return list, total, err
}

func (s *Service) HasMonthlySalary(personID uint, belongMonth string) bool {
	var count int64
	s.DB.Model(&AttendanceSalary{}).Where("person_id = ? AND belong_month = ?", personID, belongMonth).Count(&count)
	return count > 0
}

func (s *Service) getSnapshotsInRange(personID uint, start, end string) ([]PositionSnapshotInfo, error) {
	var snaps []PositionSnapshotInfo
	rows, err := s.DB.Raw(`
		SELECT base_salary, meal_allowance, salary_days, has_attendance_bonus,
			effective_start_date, effective_end_date
		FROM position_snapshot
		WHERE person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?
		ORDER BY effective_start_date`, personID, end, start).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s PositionSnapshotInfo
		rows.Scan(&s.BaseSalary, &s.MealAllowance, &s.SalaryDays, &s.HasAttendanceBonus,
			&s.EffectiveStartDate, &s.EffectiveEndDate)
		snaps = append(snaps, s)
	}
	return snaps, nil
}

type PositionSnapshotInfo struct {
	BaseSalary         float64
	MealAllowance      float64
	SalaryDays         int
	HasAttendanceBonus bool
	EffectiveStartDate string
	EffectiveEndDate   string
}

func (s *Service) calcWeighted(snaps []PositionSnapshotInfo, monthStart, monthEnd string) (float64, float64, int) {
	start, _ := time.Parse("2006-01-02", monthStart)
	end, _ := time.Parse("2006-01-02", monthEnd)
	totalDays := end.Sub(start).Hours()/24 + 1

	var weightedBase, weightedMeal float64
	var weightedSalaryDays float64
	for _, snap := range snaps {
		snapStart, _ := time.Parse("2006-01-02", snap.EffectiveStartDate)
		snapEnd, _ := time.Parse("2006-01-02", snap.EffectiveEndDate)
		if snapStart.Before(start) {
			snapStart = start
		}
		if snapEnd.After(end) {
			snapEnd = end
		}
		days := snapEnd.Sub(snapStart).Hours()/24 + 1
		if days > 0 && snap.SalaryDays > 0 {
			ratio := days / totalDays
			weightedBase += snap.BaseSalary * ratio
			weightedMeal += snap.MealAllowance * ratio
			weightedSalaryDays += float64(snap.SalaryDays) * ratio
		}
	}
	if weightedSalaryDays < 1 {
		weightedSalaryDays = 21
	}
	return weightedBase, weightedMeal, int(weightedSalaryDays + 0.5)
}

func (s *Service) GetMaxDailyUpdatedAt(personID uint, start, end string) *time.Time {
	var maxT time.Time
	s.DB.Model(&AttendanceDaily{}).Where("person_id = ? AND work_date >= ? AND work_date <= ?", personID, start, end).
		Select("MAX(updated_at)").Scan(&maxT)
	if maxT.IsZero() {
		return nil
	}
	return &maxT
}

func (s *Service) GetMaxEventUpdatedAt(personID uint, start, end string) *time.Time {
	var maxT time.Time
	s.DB.Model(&AttendanceEvent{}).Unscoped().Where("person_id = ? AND event_date >= ? AND event_date <= ?", personID, start, end).
		Select("MAX(updated_at)").Scan(&maxT)
	var maxD time.Time
	s.DB.Model(&AttendanceEvent{}).Unscoped().Where("person_id = ? AND event_date >= ? AND event_date <= ? AND deleted_at IS NOT NULL", personID, start, end).
		Select("MAX(deleted_at)").Scan(&maxD)
	if maxT.After(maxD) {
		return &maxT
	}
	if !maxD.IsZero() {
		return &maxD
	}
	return nil
}

var RebuildLeaveBalanceCallback func(personID uint, leaveType string) error

func (s *Service) triggerLeaveBalanceRebuild(personID uint, leaveType string) {
	if RebuildLeaveBalanceCallback != nil {
		RebuildLeaveBalanceCallback(personID, leaveType)
	}
}

func getS(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		s, _ := v.(string)
		return s
	}
	return ""
}

func getF(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return val
		case int:
		return float64(val)
		case string:
			return utils.ParseFloat(val)
		}
	}
	return 0
}
