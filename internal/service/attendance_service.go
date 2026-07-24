package service

import (
	"math"
	"time"

	"gorm.io/gorm"

	"probig/internal/database"
	"probig/internal/models"
)

type AttendanceService struct{}

func NewAttendanceService() *AttendanceService {
	return &AttendanceService{}
}

type AttendanceEventInput struct {
	EntityID      uint    `json:"entity_id"`
	EventCategory string  `json:"event_category"`
	EventSubtype  string  `json:"event_subtype"`
	EventDate     string  `json:"event_date"`
	DurationDays  float64 `json:"duration_days"`
	Description   string  `json:"description"`
}

func (s *AttendanceService) CreateEvent(audit *AuditService, ctx EventContext, input *AttendanceEventInput) (*models.AttendanceEvent, error) {
	event := &models.AttendanceEvent{
		EntityID:      input.EntityID,
		EventCategory: input.EventCategory,
		EventSubtype:  input.EventSubtype,
		EventDate:     input.EventDate,
		DurationDays:  input.DurationDays,
		Description:   input.Description,
		CreatedBy:     ctx.UserID,
	}
	entityName := GetEntityNameMap(input.EntityID)[input.EntityID]

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "create", "attendance_event", event.ID, entityName, event.EventCategory+"-"+event.EventSubtype, models.JSONMap{
			"entity_id":      event.EntityID,
			"entity_name":    entityName,
			"event_category": event.EventCategory,
			"event_subtype":  event.EventSubtype,
			"event_date":     event.EventDate,
			"duration_days":  event.DurationDays,
		})
	})
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *AttendanceService) UpdateEvent(audit *AuditService, ctx EventContext, eventID uint, input *AttendanceEventInput) (*models.AttendanceEvent, error) {
	var event models.AttendanceEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return nil, err
	}
	entityName := GetEntityNameMap(event.EntityID)[event.EntityID]

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&event).Updates(map[string]interface{}{
			"entity_id":      input.EntityID,
			"event_category": input.EventCategory,
			"event_subtype":  input.EventSubtype,
			"event_date":     input.EventDate,
			"duration_days":  input.DurationDays,
			"description":    input.Description,
		}).Error; err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "update", "attendance_event", event.ID, entityName, input.EventCategory+"-"+input.EventSubtype, models.JSONMap{
			"entity_id":      input.EntityID,
			"event_category": input.EventCategory,
			"event_subtype":  input.EventSubtype,
			"event_date":     input.EventDate,
			"duration_days":  input.DurationDays,
		})
	})
	if err != nil {
		return nil, err
	}
	database.DB.First(&event, eventID)
	return &event, nil
}

func (s *AttendanceService) DeleteEvent(audit *AuditService, ctx EventContext, eventID uint) error {
	var event models.AttendanceEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return err
	}
	entityName := GetEntityNameMap(event.EntityID)[event.EntityID]

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "delete", "attendance_event", eventID, entityName, event.EventCategory+"-"+event.EventSubtype, models.JSONMap{
			"entity_id":      event.EntityID,
			"event_category": event.EventCategory,
			"event_subtype":  event.EventSubtype,
			"event_date":     event.EventDate,
		})
	})
}

func (s *AttendanceService) ListEvents(entityID uint, startDate, endDate string, page, pageSize int) ([]models.AttendanceEvent, int64, error) {
	var events []models.AttendanceEvent
	var total int64
	query := database.DB.Model(&models.AttendanceEvent{})
	if entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	if startDate != "" {
		query = query.Where("event_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("event_date <= ?", endDate)
	}
	query.Count(&total)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("event_date DESC, id DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	fillAttendanceEventEntityNames(events)
	return events, total, nil
}

func (s *AttendanceService) Calculate(audit *AuditService, ctx EventContext, periodStart, periodEnd string) ([]models.AttendanceSummary, error) {
	type agg struct {
		EntityID         uint
		NormalDays       float64
		MakeupDays       float64
		LieuDays         float64
		PersonalDays     float64
		SickDays         float64
		AnnualDays       float64
		StatutoryDays    float64
		WelfareDays      float64
		WorkdayOvertime  float64
		HolidayOvertime  float64
		MissingCardCount float64
		LateCount        float64
		EarlyCount       float64
		AnnualAllocated  float64
		AnnualCarriedOver float64
	}

	var events []models.AttendanceEvent
	database.DB.Where("event_date >= ? AND event_date <= ?", periodStart, periodEnd).Find(&events)

	aggMap := make(map[uint]*agg)
	for _, e := range events {
		a, ok := aggMap[e.EntityID]
		if !ok {
			a = &agg{EntityID: e.EntityID}
			aggMap[e.EntityID] = a
		}
		switch e.EventCategory {
		case "attendance":
			switch e.EventSubtype {
			case "normal":
				a.NormalDays += e.DurationDays
			case "makeup":
				a.MakeupDays += e.DurationDays
			}
		case "leave":
			switch e.EventSubtype {
			case "lieu":
				a.LieuDays += e.DurationDays
			case "personal":
				a.PersonalDays += e.DurationDays
			case "sick":
				a.SickDays += e.DurationDays
			case "annual":
				a.AnnualDays += e.DurationDays
			case "statutory":
				a.StatutoryDays += e.DurationDays
			case "welfare":
				a.WelfareDays += e.DurationDays
			}
		case "overtime":
			switch e.EventSubtype {
			case "workday":
				a.WorkdayOvertime += e.DurationDays
			case "holiday":
				a.HolidayOvertime += e.DurationDays
			}
		case "discipline":
			switch e.EventSubtype {
			case "missing_card":
				a.MissingCardCount += e.DurationDays
			case "late":
				a.LateCount += e.DurationDays
			case "early":
				a.EarlyCount += e.DurationDays
			}
		case "annual_adjustment":
			switch e.EventSubtype {
			case "allocation":
				a.AnnualAllocated += e.DurationDays
			case "carryover":
				a.AnnualCarriedOver += e.DurationDays
			}
		}
	}

	var summaries []models.AttendanceSummary
	for _, a := range aggMap {
		summary := models.AttendanceSummary{
			EntityID:          a.EntityID,
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			NormalDays:        math.Round(a.NormalDays*100) / 100,
			MakeupDays:        math.Round(a.MakeupDays*100) / 100,
			LieuDays:          math.Round(a.LieuDays*100) / 100,
			PersonalDays:      math.Round(a.PersonalDays*100) / 100,
			SickDays:          math.Round(a.SickDays*100) / 100,
			AnnualDays:        math.Round(a.AnnualDays*100) / 100,
			StatutoryDays:     math.Round(a.StatutoryDays*100) / 100,
			WelfareDays:       math.Round(a.WelfareDays*100) / 100,
			WorkdayOvertime:   math.Round(a.WorkdayOvertime*100) / 100,
			HolidayOvertime:   math.Round(a.HolidayOvertime*100) / 100,
			MissingCardCount:  math.Round(a.MissingCardCount*100) / 100,
			LateCount:         math.Round(a.LateCount*100) / 100,
			EarlyCount:        math.Round(a.EarlyCount*100) / 100,
			AnnualAllocated:   math.Round(a.AnnualAllocated*100) / 100,
			AnnualCarriedOver: math.Round(a.AnnualCarriedOver*100) / 100,
			CalculatedAt:      time.Now(),
		}

		database.DB.Where("entity_id = ? AND period_start = ? AND period_end = ?",
			a.EntityID, periodStart, periodEnd).Delete(&models.AttendanceSummary{})
		database.DB.Create(&summary)
		summaries = append(summaries, summary)
	}

	audit.Log(ctx.UserID, ctx.Username, "calculate", "attendance_summary", 0, "批量考勤核算", periodStart+"~"+periodEnd, models.JSONMap{
		"period_start": periodStart,
		"period_end":   periodEnd,
		"count":        len(summaries),
	})
	return summaries, nil
}

func (s *AttendanceService) ListSummaries(entityID uint, periodStart, periodEnd string, page, pageSize int) ([]models.AttendanceSummary, int64, error) {
	var summaries []models.AttendanceSummary
	var total int64
	query := database.DB.Model(&models.AttendanceSummary{})
	if entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	if periodStart != "" {
		query = query.Where("period_start >= ?", periodStart)
	}
	if periodEnd != "" {
		query = query.Where("period_end <= ?", periodEnd)
	}
	query.Count(&total)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("period_start DESC, entity_id ASC").Find(&summaries).Error; err != nil {
		return nil, 0, err
	}
	fillAttendanceSummaryEntityNames(summaries)
	return summaries, total, nil
}

func fillAttendanceEventEntityNames(events []models.AttendanceEvent) {
	if len(events) == 0 {
		return
	}
	ids := make([]uint, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.EntityID)
	}
	nameMap := GetEntityNameMap(ids...)
	for i := range events {
		events[i].EntityName = nameMap[events[i].EntityID]
	}
}

func fillAttendanceSummaryEntityNames(summaries []models.AttendanceSummary) {
	if len(summaries) == 0 {
		return
	}
	ids := make([]uint, 0, len(summaries))
	for _, s := range summaries {
		ids = append(ids, s.EntityID)
	}
	nameMap := GetEntityNameMap(ids...)
	for i := range summaries {
		summaries[i].EntityName = nameMap[summaries[i].EntityID]
	}
}
