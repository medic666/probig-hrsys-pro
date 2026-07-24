package service

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"probig/internal/database"
	"probig/internal/models"
)

type PersonnelService struct{}

func NewPersonnelService() *PersonnelService {
	return &PersonnelService{}
}

type PersonnelEventInput struct {
	EntityID      uint           `json:"entity_id"`
	EventType     string         `json:"event_type"`
	EventName     string         `json:"event_name"`
	EffectiveDate string         `json:"effective_date"`
	Name          string         `json:"name"`
	ChangedFields models.JSONMap `json:"changed_fields"`
}

type EventContext struct {
	UserID   uint
	Username string
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *PersonnelService) CreateEvent(audit *AuditService, ctx EventContext, input *PersonnelEventInput) (*models.PersonnelEvent, error) {
	effectiveDate, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("无效的日期格式: %s", input.EffectiveDate)
	}

	event := s.buildEventFromInput(input, effectiveDate)
	event.CreatedBy = ctx.UserID
	event.ChangedFields = input.ChangedFields

	isCreate := input.EventType == "create" && event.EntityID == 0

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if isCreate {
			entity := &models.Entity{Type: "person", Name: event.Name, Status: "active"}
			if err := tx.Create(entity).Error; err != nil {
				return err
			}
			event.EntityID = entity.ID
		}
		if err := tx.Create(event).Error; err != nil {
			return err
		}
		if err := s.rebuildSnapshotsTx(tx, event.EntityID); err != nil {
			return err
		}
		if input.EventType == "delete" {
			tx.Model(&models.Entity{}).Where("id = ?", event.EntityID).Update("status", "inactive")
		}

		targetSummary := input.EventName
		if targetSummary == "" {
			targetSummary = input.Name + "-" + input.EventType
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "create", "personnel_event", event.ID, event.Name, targetSummary, models.JSONMap{
			"entity_id":      event.EntityID,
			"name":           event.Name,
			"effective_date": input.EffectiveDate,
			"event_type":     event.EventType,
			"event_name":     input.EventName,
			"changed_fields": input.ChangedFields,
		})
	})
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *PersonnelService) UpdateEvent(audit *AuditService, ctx EventContext, eventID uint, input *PersonnelEventInput) (*models.PersonnelEvent, error) {
	var event models.PersonnelEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return nil, err
	}

	effectiveDate, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("无效的日期格式: %s", input.EffectiveDate)
	}

	entityID := event.EntityID

	newEvent := s.buildEventFromInput(input, effectiveDate)
	newEvent.EntityID = entityID
	newEvent.CreatedBy = event.CreatedBy
	newEvent.ChangedFields = input.ChangedFields

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&event).Updates(map[string]interface{}{
			"effective_date":  effectiveDate,
			"event_name":      input.EventName,
			"name":            newEvent.Name,
			"changed_fields":  input.ChangedFields,
			"attendance_group":          newEvent.AttendanceGroup,
			"hire_date":                 newEvent.HireDate,
			"base_salary":               newEvent.BaseSalary,
			"performance_salary":        newEvent.PerformanceSalary,
			"pay_days":                  newEvent.PayDays,
			"position_allowance":        newEvent.PositionAllowance,
			"meal_subsidy":              newEvent.MealSubsidy,
			"housing_subsidy":           newEvent.HousingSubsidy,
			"transport_subsidy":         newEvent.TransportSubsidy,
			"heat_subsidy":              newEvent.HeatSubsidy,
			"insurance_compensation":    newEvent.InsuranceCompensation,
			"housing_fund_compensation": newEvent.HousingFundCompensation,
			"social_insurance_deduct":   newEvent.SocialInsuranceDeduct,
			"housing_fund_deduct":       newEvent.HousingFundDeduct,
			"extended_info":             newEvent.ExtendedInfo,
		}).Error; err != nil {
			return err
		}
		if err := s.rebuildSnapshotsTx(tx, entityID); err != nil {
			return err
		}

		targetSummary := input.EventName
		if targetSummary == "" {
			targetSummary = event.Name + "-update"
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "update", "personnel_event", event.ID, event.Name, targetSummary, models.JSONMap{
			"entity_id":      entityID,
			"name":           event.Name,
			"effective_date": input.EffectiveDate,
			"event_name":     input.EventName,
			"changed_fields": input.ChangedFields,
		})
	})
	if err != nil {
		return nil, err
	}

	database.DB.First(&event, eventID)
	return &event, nil
}

func (s *PersonnelService) DeleteEvent(audit *AuditService, ctx EventContext, eventID uint) error {
	var event models.PersonnelEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return err
	}
	entityID := event.EntityID
	eventType := event.EventType
	eventName := event.Name

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		if eventType == "create" {
			var count int64
			tx.Model(&models.PersonnelEvent{}).Where("entity_id = ?", entityID).Count(&count)
			if count == 0 {
				tx.Model(&models.Entity{}).Where("id = ?", entityID).Update("status", "deleted")
			}
		}
		if err := s.rebuildSnapshotsTx(tx, entityID); err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "delete", "personnel_event", eventID, eventName, "删除事件", models.JSONMap{
			"entity_id": entityID,
			"name":      eventName,
		})
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *PersonnelService) buildEventFromInput(input *PersonnelEventInput, effectiveDate time.Time) *models.PersonnelEvent {
	event := &models.PersonnelEvent{
		EntityID:      input.EntityID,
		EventType:     input.EventType,
		EventName:     input.EventName,
		EffectiveDate: effectiveDate,
		Name:          input.Name,
	}

	cf := input.ChangedFields
	if cf == nil {
		cf = models.JSONMap{}
	}

	if v, ok := cf["attendance_group"]; ok {
		event.AttendanceGroup, _ = v.(string)
	}
	if v, ok := cf["hire_date"]; ok {
		if s, isStr := v.(string); isStr && s != "" {
			event.HireDate = stringPtr(s)
		}
	}
	if v, ok := cf["base_salary"]; ok {
		event.BaseSalary = toFloat64(v)
	}
	if v, ok := cf["performance_salary"]; ok {
		event.PerformanceSalary = toFloat64(v)
	}
	if v, ok := cf["pay_days"]; ok {
		event.PayDays = toFloat64(v)
	}
	if v, ok := cf["position_allowance"]; ok {
		event.PositionAllowance = toFloat64(v)
	}
	if v, ok := cf["meal_subsidy"]; ok {
		event.MealSubsidy = toFloat64(v)
	}
	if v, ok := cf["housing_subsidy"]; ok {
		event.HousingSubsidy = toFloat64(v)
	}
	if v, ok := cf["transport_subsidy"]; ok {
		event.TransportSubsidy = toFloat64(v)
	}
	if v, ok := cf["heat_subsidy"]; ok {
		event.HeatSubsidy = toFloat64(v)
	}
	if v, ok := cf["insurance_compensation"]; ok {
		event.InsuranceCompensation = toFloat64(v)
	}
	if v, ok := cf["housing_fund_compensation"]; ok {
		event.HousingFundCompensation = toFloat64(v)
	}
	if v, ok := cf["social_insurance_deduct"]; ok {
		event.SocialInsuranceDeduct = toFloat64(v)
	}
	if v, ok := cf["housing_fund_deduct"]; ok {
		event.HousingFundDeduct = toFloat64(v)
	}
	if v, ok := cf["extended_info"]; ok {
		if m, isMap := v.(map[string]interface{}); isMap {
			event.ExtendedInfo = models.JSONMap(m)
		}
	}

	return event
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

func (s *PersonnelService) rebuildSnapshotsTx(tx *gorm.DB, entityID uint) error {
	var events []models.PersonnelEvent
	if err := tx.Where("entity_id = ?", entityID).Order("effective_date ASC, created_at ASC").Find(&events).Error; err != nil {
		return err
	}

	tx.Where("entity_id = ?", entityID).Delete(&models.PersonnelSnapshot{})

	var prev models.PersonnelSnapshot
	firstEvent := true

	for i, event := range events {
		var snapshot models.PersonnelSnapshot

		if i == 0 {
			snapshot = s.snapshotFromEvent(&event)
		} else {
			snapshot = prev
			snapshot.ID = 0
			snapshot.EventID = event.ID
			snapshot.EffectiveDate = event.EffectiveDate
			s.mergeFields(&snapshot, &event)
		}

		snapshot.EntityID = entityID
		snapshot.IsLatest = i == len(events)-1
		if firstEvent {
			firstEvent = false
		}

		if err := tx.Create(&snapshot).Error; err != nil {
			return err
		}
		prev = snapshot
	}
	return nil
}

func (s *PersonnelService) snapshotFromEvent(event *models.PersonnelEvent) models.PersonnelSnapshot {
	return models.PersonnelSnapshot{
		EntityID:                event.EntityID,
		EventID:                 event.ID,
		EffectiveDate:           event.EffectiveDate,
		Name:                    event.Name,
		AttendanceGroup:         event.AttendanceGroup,
		HireDate:                event.HireDate,
		BaseSalary:              event.BaseSalary,
		PerformanceSalary:       event.PerformanceSalary,
		PayDays:                 event.PayDays,
		PositionAllowance:       event.PositionAllowance,
		MealSubsidy:             event.MealSubsidy,
		HousingSubsidy:          event.HousingSubsidy,
		TransportSubsidy:        event.TransportSubsidy,
		HeatSubsidy:             event.HeatSubsidy,
		InsuranceCompensation:   event.InsuranceCompensation,
		HousingFundCompensation: event.HousingFundCompensation,
		SocialInsuranceDeduct:   event.SocialInsuranceDeduct,
		HousingFundDeduct:       event.HousingFundDeduct,
		ExtendedInfo:            event.ExtendedInfo,
	}
}

func (s *PersonnelService) mergeFields(snap *models.PersonnelSnapshot, event *models.PersonnelEvent) {
	if event.AttendanceGroup != "" {
		snap.AttendanceGroup = event.AttendanceGroup
	}
	if event.HireDate != nil {
		snap.HireDate = event.HireDate
	}
	cf := event.ChangedFields
	if cf == nil {
		return
	}
	if _, ok := cf["name"]; ok {
		snap.Name = event.Name
	}
	if _, ok := cf["attendance_group"]; ok {
		snap.AttendanceGroup = event.AttendanceGroup
	}
	if _, ok := cf["hire_date"]; ok {
		snap.HireDate = event.HireDate
	}
	if _, ok := cf["base_salary"]; ok {
		snap.BaseSalary = event.BaseSalary
	}
	if _, ok := cf["performance_salary"]; ok {
		snap.PerformanceSalary = event.PerformanceSalary
	}
	if _, ok := cf["pay_days"]; ok {
		snap.PayDays = event.PayDays
	}
	if _, ok := cf["position_allowance"]; ok {
		snap.PositionAllowance = event.PositionAllowance
	}
	if _, ok := cf["meal_subsidy"]; ok {
		snap.MealSubsidy = event.MealSubsidy
	}
	if _, ok := cf["housing_subsidy"]; ok {
		snap.HousingSubsidy = event.HousingSubsidy
	}
	if _, ok := cf["transport_subsidy"]; ok {
		snap.TransportSubsidy = event.TransportSubsidy
	}
	if _, ok := cf["heat_subsidy"]; ok {
		snap.HeatSubsidy = event.HeatSubsidy
	}
	if _, ok := cf["insurance_compensation"]; ok {
		snap.InsuranceCompensation = event.InsuranceCompensation
	}
	if _, ok := cf["housing_fund_compensation"]; ok {
		snap.HousingFundCompensation = event.HousingFundCompensation
	}
	if _, ok := cf["social_insurance_deduct"]; ok {
		snap.SocialInsuranceDeduct = event.SocialInsuranceDeduct
	}
	if _, ok := cf["housing_fund_deduct"]; ok {
		snap.HousingFundDeduct = event.HousingFundDeduct
	}
	if _, ok := cf["extended_info"]; ok {
		snap.ExtendedInfo = event.ExtendedInfo
	}
}

func (s *PersonnelService) ListSnapshots(search string, page, pageSize int) ([]models.PersonnelSnapshot, int64, error) {
	var total int64
	subQuery := database.DB.Table("personnel_snapshots").
		Select("MAX(id) as id").
		Where("is_latest = ?", true).
		Group("entity_id")

	query := database.DB.Table("personnel_snapshots").Where("id IN (?)", subQuery)
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}
	query.Count(&total)

	var snapshots []models.PersonnelSnapshot
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&snapshots).Error; err != nil {
		return nil, 0, err
	}
	return snapshots, total, nil
}

func (s *PersonnelService) GetSnapshot(entityID uint) (*models.PersonnelSnapshot, error) {
	var snapshot models.PersonnelSnapshot
	if err := database.DB.Where("entity_id = ? AND is_latest = ?", entityID, true).First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *PersonnelService) GetHistory(entityID uint) ([]models.PersonnelSnapshot, error) {
	var snapshots []models.PersonnelSnapshot
	if err := database.DB.Where("entity_id = ?", entityID).Order("effective_date ASC").Find(&snapshots).Error; err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (s *PersonnelService) ListEvents(entityID uint, page, pageSize int) ([]models.PersonnelEvent, int64, error) {
	var events []models.PersonnelEvent
	var total int64
	query := database.DB.Model(&models.PersonnelEvent{})
	if entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	query.Count(&total)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("effective_date DESC, id DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (s *PersonnelService) GetSnapshotForDate(entityID uint, date time.Time) (*models.PersonnelSnapshot, error) {
	var snapshot models.PersonnelSnapshot
	if err := database.DB.Where("entity_id = ? AND effective_date <= ?", entityID, date).
		Order("effective_date DESC").First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *PersonnelService) CreateEntity(name string) (*models.Entity, error) {
	entity := &models.Entity{Type: "person", Name: name, Status: "active"}
	if err := database.DB.Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func GetEntityNameMap(entityIDs ...uint) map[uint]string {
	if len(entityIDs) == 0 {
		return map[uint]string{}
	}
	var entities []models.Entity
	database.DB.Where("id IN ?", entityIDs).Find(&entities)
	m := make(map[uint]string, len(entities))
	for _, e := range entities {
		m[e.ID] = e.Name
	}
	return m
}
