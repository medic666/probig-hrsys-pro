package service

import (
	"math"
	"time"

	"gorm.io/gorm"

	"probig/internal/database"
	"probig/internal/models"
)

type SalaryService struct{}

func NewSalaryService() *SalaryService {
	return &SalaryService{}
}

type SalaryEventInput struct {
	EntityID    uint    `json:"entity_id"`
	EventType   string  `json:"event_type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   string  `json:"period_end"`
}

func (s *SalaryService) CreateEvent(audit *AuditService, ctx EventContext, input *SalaryEventInput) (*models.SalaryEvent, error) {
	event := &models.SalaryEvent{
		EntityID:    input.EntityID,
		EventType:   input.EventType,
		Amount:      input.Amount,
		Description: input.Description,
		PeriodStart: input.PeriodStart,
		PeriodEnd:   input.PeriodEnd,
		CreatedBy:   ctx.UserID,
	}
	entityName := GetEntityNameMap(input.EntityID)[input.EntityID]

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "create", "salary_event", event.ID, entityName, event.EventType, models.JSONMap{
			"entity_id":  event.EntityID,
			"entity_name": entityName,
			"event_type": event.EventType,
			"amount":     event.Amount,
			"period_start": event.PeriodStart,
			"period_end":   event.PeriodEnd,
		})
	})
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *SalaryService) UpdateEvent(audit *AuditService, ctx EventContext, eventID uint, input *SalaryEventInput) (*models.SalaryEvent, error) {
	var event models.SalaryEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return nil, err
	}
	entityName := GetEntityNameMap(event.EntityID)[event.EntityID]

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&event).Updates(map[string]interface{}{
			"entity_id":    input.EntityID,
			"event_type":   input.EventType,
			"amount":       input.Amount,
			"description":  input.Description,
			"period_start": input.PeriodStart,
			"period_end":   input.PeriodEnd,
		}).Error; err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "update", "salary_event", event.ID, entityName, input.EventType, models.JSONMap{
			"entity_id":    input.EntityID,
			"event_type":   input.EventType,
			"amount":       input.Amount,
			"period_start": input.PeriodStart,
			"period_end":   input.PeriodEnd,
		})
	})
	if err != nil {
		return nil, err
	}
	database.DB.First(&event, eventID)
	return &event, nil
}

func (s *SalaryService) DeleteEvent(audit *AuditService, ctx EventContext, eventID uint) error {
	var event models.SalaryEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return err
	}
	entityName := GetEntityNameMap(event.EntityID)[event.EntityID]

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "delete", "salary_event", eventID, entityName, event.EventType, models.JSONMap{
			"entity_id":  event.EntityID,
			"event_type": event.EventType,
			"amount":     event.Amount,
		})
	})
}

func (s *SalaryService) ListEvents(entityID uint, page, pageSize int) ([]models.SalaryEvent, int64, error) {
	var events []models.SalaryEvent
	var total int64
	query := database.DB.Model(&models.SalaryEvent{})
	if entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	query.Count(&total)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("period_start DESC, id DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	fillSalaryEventEntityNames(events)
	return events, total, nil
}

func (s *SalaryService) Calculate(audit *AuditService, ctx EventContext, periodStart, periodEnd string) ([]models.SalarySummary, error) {
	startDate, _ := time.Parse("2006-01-02", periodStart)
	endDate, _ := time.Parse("2006-01-02", periodEnd)

	personnelSvc := NewPersonnelService()

	var entities []models.Entity
	database.DB.Where("type = ? AND status = ?", "person", "active").Find(&entities)

	var summaries []models.SalarySummary
	for _, entity := range entities {
		snap, err := personnelSvc.GetSnapshotForDate(entity.ID, startDate)
		if err != nil {
			continue
		}

		var attSummary models.AttendanceSummary
		attResult := database.DB.Where("entity_id = ? AND period_start = ? AND period_end = ?",
			entity.ID, periodStart, periodEnd).First(&attSummary)
		if attResult.Error != nil {
			continue
		}

		var salaryEvents []models.SalaryEvent
		database.DB.Where("entity_id = ? AND period_start >= ? AND period_start <= ?",
			entity.ID, periodStart, endDate.Format("2006-01-02")).Find(&salaryEvents)

		summary := s.calculateSalary(snap, &attSummary, salaryEvents, periodStart, periodEnd)

		database.DB.Where("entity_id = ? AND period_start = ? AND period_end = ?",
			entity.ID, periodStart, periodEnd).Delete(&models.SalarySummary{})
		database.DB.Create(&summary)
		summaries = append(summaries, summary)
	}

	audit.Log(ctx.UserID, ctx.Username, "calculate", "salary_summary", 0, "批量工资核算", periodStart+"~"+periodEnd, models.JSONMap{
		"period_start": periodStart,
		"period_end":   periodEnd,
		"count":        len(summaries),
	})
	return summaries, nil
}

func (s *SalaryService) calculateSalary(snap *models.PersonnelSnapshot, att *models.AttendanceSummary, salaryEvents []models.SalaryEvent, periodStart, periodEnd string) models.SalarySummary {
	r := func(v float64) float64 { return math.Round(v*100) / 100 }

	payDays := snap.PayDays
	if payDays <= 0 {
		payDays = 21.75
	}

	dailySalary := r(snap.BaseSalary / payDays)

	attendanceWage := r((att.NormalDays+att.MakeupDays)*1*dailySalary +
		att.SickDays*0.6*dailySalary +
		(att.AnnualDays+att.StatutoryDays+att.WelfareDays)*1*dailySalary)

	fullAttendanceBonus := 0.0
	if att.PersonalDays == 0 {
		disciplineCount := att.MissingCardCount + att.LateCount + att.EarlyCount
		bonusDays := (att.NormalDays+att.MakeupDays)*1 +
			att.SickDays*0.6 +
			(att.AnnualDays+att.StatutoryDays+att.WelfareDays)*1 -
			disciplineCount
		fullAttendanceBonus = r(math.Max(0, bonusDays) * 10)
	}

	overtimeRate := r((snap.BaseSalary + snap.MealSubsidy) / payDays)
	overtimeWage := r(att.WorkdayOvertime*overtimeRate*1.5 +
		(att.HolidayOvertime+att.AnnualCarriedOver)*overtimeRate*2)

	performanceSalary := snap.PerformanceSalary

	allowances := snap.PositionAllowance + snap.MealSubsidy + snap.HousingSubsidy +
		snap.TransportSubsidy + snap.HeatSubsidy + snap.InsuranceCompensation + snap.HousingFundCompensation

	performanceAdjustment := 0.0
	rewardPunishment := 0.0
	loanDeduction := 0.0
	taxDeduction := 0.0

	for _, e := range salaryEvents {
		switch e.EventType {
		case "performance":
			performanceAdjustment += e.Amount
		case "reward_punishment":
			rewardPunishment += e.Amount
		case "loan_deduction":
			loanDeduction += e.Amount
		case "tax_deduction":
			taxDeduction += e.Amount
		default:
			rewardPunishment += e.Amount
		}
	}

	grossPay := r(attendanceWage + fullAttendanceBonus + overtimeWage + performanceSalary +
		allowances + performanceAdjustment + rewardPunishment)

	deductions := r(snap.SocialInsuranceDeduct + snap.HousingFundDeduct + loanDeduction + taxDeduction)

	netPay := r(grossPay - deductions)

	return models.SalarySummary{
		EntityID:                snap.EntityID,
		PeriodStart:             periodStart,
		PeriodEnd:               periodEnd,
		BaseSalary:              r(snap.BaseSalary),
		DailySalary:             dailySalary,
		AttendanceWage:          attendanceWage,
		FullAttendanceBonus:     r(fullAttendanceBonus),
		OvertimeWage:            r(overtimeWage),
		PerformanceSalary:       r(performanceSalary),
		PositionAllowance:       r(snap.PositionAllowance),
		MealSubsidy:             r(snap.MealSubsidy),
		HousingSubsidy:          r(snap.HousingSubsidy),
		TransportSubsidy:        r(snap.TransportSubsidy),
		HeatSubsidy:             r(snap.HeatSubsidy),
		InsuranceCompensation:   r(snap.InsuranceCompensation),
		HousingFundCompensation: r(snap.HousingFundCompensation),
		PerformanceAdjustment:   r(performanceAdjustment),
		RewardPunishment:        r(rewardPunishment),
		LoanDeduction:           r(loanDeduction),
		SocialInsuranceDeduct:   r(snap.SocialInsuranceDeduct),
		HousingFundDeduct:       r(snap.HousingFundDeduct),
		TaxDeduction:            r(taxDeduction),
		GrossPay:                grossPay,
		NetPay:                  netPay,
		CalculatedAt:            time.Now(),
	}
}

func (s *SalaryService) ListSummaries(entityID uint, periodStart, periodEnd string, page, pageSize int) ([]models.SalarySummary, int64, error) {
	var summaries []models.SalarySummary
	var total int64
	query := database.DB.Model(&models.SalarySummary{})
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
	fillSalarySummaryEntityNames(summaries)
	return summaries, total, nil
}

func fillSalaryEventEntityNames(events []models.SalaryEvent) {
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

func fillSalarySummaryEntityNames(summaries []models.SalarySummary) {
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
