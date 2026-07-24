package attendance

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"probig/internal/person"
	"probig/internal/pkg/audit"
	"probig/internal/position"
)

type Service struct {
	dao             *DAO
	personService   *person.Service
	positionService *position.Service
}

var globalService *Service

func NewService(db *gorm.DB) *Service {
	svc := &Service{
		dao:             NewDAO(db),
		personService:   person.GetService(),
		positionService: position.GetService(),
	}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) CreateEvent(event *AttendanceEvent, operatorID uint, operatorName string) error {
	if err := s.dao.CreateEvent(event); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "attendance_event", event.ID, "create", "", event)
	return nil
}

func (s *Service) GetEventsByPersonAndMonth(personID uint, yearMonth string) ([]AttendanceEvent, error) {
	return s.dao.GetEventsByPersonAndMonth(personID, yearMonth)
}

func (s *Service) UpdateEvent(event *AttendanceEvent, operatorID uint, operatorName string) error {
	old, err := s.dao.GetEventByID(event.ID)
	if err != nil {
		return err
	}
	if err := s.dao.UpdateEvent(event); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "attendance_event", event.ID, "update", old, event)
	return nil
}

func (s *Service) DeleteEvent(id uint, operatorID uint, operatorName string) error {
	event, err := s.dao.GetEventByID(id)
	if err != nil {
		return err
	}
	if err := s.dao.DeleteEvent(id); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "attendance_event", id, "delete", event, nil)
	return nil
}

func (s *Service) CreateCrossDayEvents(personID uint, startDate, endDate, eventType, subType string, hours float64, remark string, operatorID uint, operatorName string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return errors.New("invalid start date")
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return errors.New("invalid end date")
	}
	if end.Before(start) {
		return errors.New("end date before start date")
	}

	batchID := fmt.Sprintf("CROSS-%d", time.Now().UnixNano())
	var events []AttendanceEvent
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		h := hours
		events = append(events, AttendanceEvent{
			PersonID:  personID,
			EventDate: d.Format("2006-01-02"),
			EventType: eventType,
			SubType:   subType,
			Hours:     &h,
			Remark:    remark,
			BatchID:   batchID,
		})
	}

	if err := s.dao.BatchCreateEvents(events); err != nil {
		return err
	}

	for i := range events {
		s.logAudit(operatorID, operatorName, "attendance_event", events[i].ID, "create", "", &events[i])
	}
	return nil
}

func (s *Service) BatchCreateEvents(personIDs []uint, eventDate, eventType, subType string, hours float64, remark string, operatorID uint, operatorName string) error {
	batchID := fmt.Sprintf("BATCH-%d", time.Now().UnixNano())
	var h *float64
	if eventType == "出勤" || eventType == "休假" || eventType == "加班" {
		h = &hours
	}
	var events []AttendanceEvent
	for _, pid := range personIDs {
		events = append(events, AttendanceEvent{
			PersonID:  pid,
			EventDate: eventDate,
			EventType: eventType,
			SubType:   subType,
			Hours:     h,
			Remark:    remark,
			BatchID:   batchID,
		})
	}

	if err := s.dao.BatchCreateEvents(events); err != nil {
		return err
	}

	for i := range events {
		s.logAudit(operatorID, operatorName, "attendance_event", events[i].ID, "create", "", &events[i])
	}
	return nil
}

func (s *Service) ListEvents(params AttendanceListParams) ([]AttendanceEvent, int64, error) {
	return s.dao.ListEvents(params)
}

func (s *Service) IsMonthLocked(personID uint, yearMonth string) bool {
	summary, err := s.dao.GetSummaryByPersonMonth(personID, yearMonth)
	if err != nil {
		return false
	}
	return summary.IsLocked
}

func (s *Service) GetAnnualLeaveBalance(personID uint) (float64, error) {
	now := time.Now()
	currentYear := now.Year()

	startDate := fmt.Sprintf("%d-01-01", currentYear)
	endDate := fmt.Sprintf("%d-12-31", currentYear)

	var events []AttendanceEvent
	var adjustEvents []AttendanceEvent

	allEvents, err := s.getEventsByPersonAndDateRange(personID, startDate, endDate)
	if err != nil {
		return 0, err
	}

	for _, e := range allEvents {
		if e.EventType == "年假调整" {
			adjustEvents = append(adjustEvents, e)
		}
		if e.EventType == "休假" && e.SubType == "年假" {
			events = append(events, e)
		}
	}

	var totalAdjust float64
	for _, e := range adjustEvents {
		if e.LeaveAdjustAmount != nil {
			totalAdjust += *e.LeaveAdjustAmount
		}
	}

	var totalUsed float64
	for _, e := range events {
		if e.Hours != nil {
			totalUsed += *e.Hours
		}
	}

	return totalAdjust - totalUsed/8, nil
}

func (s *Service) ValidateAnnualLeave(personID uint, days float64) error {
	snapshotDate := time.Now().Format("2006-01-02")
	snapshot, err := s.positionService.GetSnapshotForDate(personID, snapshotDate)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("未找到职务信息快照，无法验证年假资格")
		}
		return err
	}

	if !snapshot.HasAnnualLeave {
		return errors.New("该人员不享有年假资格")
	}

	balance, err := s.GetAnnualLeaveBalance(personID)
	if err != nil {
		return err
	}

	if balance < days {
		return fmt.Errorf("年假余额不足，当前余额 %.1f 天，申请 %.1f 天", balance, days)
	}

	return nil
}

func (s *Service) CalculateSummary(personID uint, yearMonth string) (*AttendanceSummary, error) {
	events, err := s.dao.GetEventsByPersonAndMonth(personID, yearMonth)
	if err != nil {
		return nil, err
	}

	summary := &AttendanceSummary{
		PersonID:    personID,
		BelongMonth: yearMonth,
	}

	for _, e := range events {
		hours := floatPtrVal(e.Hours)
		switch e.EventType {
		case "出勤":
			switch e.SubType {
			case "普通出勤":
				summary.WorkDays += hours
			case "补班出勤":
				summary.MakeUpDays += hours
			}
		case "休假":
			switch e.SubType {
			case "病假":
				summary.SickLeaveDays += hours
			case "事假":
				summary.PersonalLeaveDays += hours
			case "年假":
				summary.AnnualLeaveDays += hours
			case "法定假":
				summary.StatutoryLeaveDays += hours
			case "福利假":
				summary.WelfareLeaveDays += hours
			}
		case "加班":
			switch e.SubType {
			case "工作日加班":
				summary.OvertimeWorkdayHours += hours
			case "节假日加班":
				summary.OvertimeHolidayHours += hours
			}
		case "违纪":
			summary.ViolationCount++
		}
	}

	summary.WorkDays = summary.WorkDays / 8
	summary.MakeUpDays = summary.MakeUpDays / 8
	summary.SickLeaveDays = summary.SickLeaveDays / 8
	summary.PersonalLeaveDays = summary.PersonalLeaveDays / 8
	summary.AnnualLeaveDays = summary.AnnualLeaveDays / 8
	summary.StatutoryLeaveDays = summary.StatutoryLeaveDays / 8
	summary.WelfareLeaveDays = summary.WelfareLeaveDays / 8

	now := time.Now().Format("2006-01-02 15:04:05")
	summary.LastCalcAt = &now

	if err := s.dao.CreateOrUpdateSummary(summary); err != nil {
		return nil, err
	}

	return summary, nil
}

func (s *Service) CalculateSummaries(personIDs []uint, yearMonth string) ([]AttendanceSummary, error) {
	var summaries []AttendanceSummary
	for _, pid := range personIDs {
		summary, err := s.CalculateSummary(pid, yearMonth)
		if err != nil {
			return nil, fmt.Errorf("计算人员ID %d 的汇总失败: %w", pid, err)
		}
		summaries = append(summaries, *summary)
	}
	return summaries, nil
}

func (s *Service) LockSummary(personID uint, yearMonth string, operatorID uint, operatorName string) error {
	if err := s.dao.LockSummary(personID, yearMonth); err != nil {
		return err
	}
	s.logAuditAction(operatorID, operatorName, "attendance_summary", 0, "lock", yearMonth)
	return nil
}

func (s *Service) UnlockSummary(personID uint, yearMonth string, operatorID uint, operatorName string) error {
	if err := s.dao.UnlockSummary(personID, yearMonth); err != nil {
		return err
	}
	s.logAuditAction(operatorID, operatorName, "attendance_summary", 0, "unlock", yearMonth)
	return nil
}

func (s *Service) ListSummaries(personIDs []uint, yearMonth string, isLocked *bool, page, pageSize int) ([]AttendanceSummary, int64, error) {
	return s.dao.ListSummaries(personIDs, yearMonth, isLocked, page, pageSize)
}

func (s *Service) GetSummary(personID uint, yearMonth string) (*AttendanceSummary, error) {
	return s.dao.GetSummaryByPersonMonth(personID, yearMonth)
}

func (s *Service) getEventsByPersonAndDateRange(personID uint, startDate, endDate string) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	params := AttendanceListParams{
		PersonIDs:      []uint{personID},
		EventDateStart: startDate,
		EventDateEnd:   endDate,
		Page:           1,
		PageSize:       10000,
	}
	var err error
	events, _, err = s.dao.ListEvents(params)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Service) logAudit(operatorID uint, operatorName, targetType string, targetID uint, action string, before, after interface{}) {
	if audit.GlobalAuditService == nil {
		return
	}
	var beforeJSON, afterJSON string
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			beforeJSON = string(b)
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			afterJSON = string(b)
		}
	}
	audit.GlobalAuditService.Log(operatorID, operatorName, targetType, targetID, action, beforeJSON, afterJSON, "", "")
}

func (s *Service) logAuditAction(operatorID uint, operatorName, targetType string, targetID uint, action, yearMonth string) {
	if audit.GlobalAuditService == nil {
		return
	}
	beforeJSON, _ := json.Marshal(map[string]string{"belongMonth": yearMonth})
	audit.GlobalAuditService.Log(operatorID, operatorName, targetType, targetID, action, string(beforeJSON), string(beforeJSON), "", "")
}

func floatPtrVal(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
