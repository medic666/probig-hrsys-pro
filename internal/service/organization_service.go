package service

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"probig/internal/database"
	"probig/internal/models"
)

type OrganizationService struct{}

func NewOrganizationService() *OrganizationService {
	return &OrganizationService{}
}

type OrganizationEventInput struct {
	EntityID      uint           `json:"entity_id"`
	EventType     string         `json:"event_type"`
	EventName     string         `json:"event_name"`
	EffectiveDate string         `json:"effective_date"`
	CompanyName   string         `json:"company_name"`
	ChangedFields models.JSONMap `json:"changed_fields"`
}

func (s *OrganizationService) CreateEvent(audit *AuditService, ctx EventContext, input *OrganizationEventInput) (*models.OrganizationEvent, error) {
	effectiveDate, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("无效的日期格式")
	}

	event := s.buildEventFromInput(input, effectiveDate)
	event.CreatedBy = ctx.UserID
	event.ChangedFields = input.ChangedFields

	isCreate := input.EventType == "create" && event.EntityID == 0

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if isCreate {
			entity := &models.Entity{Type: "organization", Name: event.CompanyName, Status: "active"}
			if err := tx.Create(entity).Error; err != nil {
				return err
			}
			event.EntityID = entity.ID
		}
		if err := tx.Create(event).Error; err != nil {
			return err
		}
		if input.EventType == "delete" {
			tx.Model(&models.Entity{}).Where("id = ?", event.EntityID).Update("status", "inactive")
		}
		if err := s.rebuildSnapshotsTx(tx, event.EntityID); err != nil {
			return err
		}

		targetSummary := input.EventName
		if targetSummary == "" {
			targetSummary = input.CompanyName + "-" + input.EventType
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "create", "organization_event", event.ID, event.CompanyName, targetSummary, models.JSONMap{
			"entity_id":      event.EntityID,
			"company_name":   event.CompanyName,
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

func (s *OrganizationService) UpdateEvent(audit *AuditService, ctx EventContext, eventID uint, input *OrganizationEventInput) (*models.OrganizationEvent, error) {
	var event models.OrganizationEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return nil, err
	}

	effectiveDate, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("无效的日期格式")
	}

	entityID := event.EntityID

	newEvent := s.buildEventFromInput(input, effectiveDate)
	newEvent.EntityID = entityID
	newEvent.CreatedBy = event.CreatedBy
	newEvent.ChangedFields = input.ChangedFields

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&event).Updates(map[string]interface{}{
			"effective_date":           effectiveDate,
			"event_name":               input.EventName,
			"company_name":             newEvent.CompanyName,
			"changed_fields":           input.ChangedFields,
			"credit_code":              newEvent.CreditCode,
			"address":                  newEvent.Address,
			"phone":                    newEvent.Phone,
			"bank_name":                newEvent.BankName,
			"bank_account":             newEvent.BankAccount,
			"business_license_file_id": newEvent.BusinessLicenseFileID,
			"official_seal_file_id":    newEvent.OfficialSealFileID,
		}).Error; err != nil {
			return err
		}
		if err := s.rebuildSnapshotsTx(tx, entityID); err != nil {
			return err
		}

		targetSummary := input.EventName
		if targetSummary == "" {
			targetSummary = event.CompanyName + "-update"
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "update", "organization_event", event.ID, event.CompanyName, targetSummary, models.JSONMap{
			"entity_id":      entityID,
			"company_name":   event.CompanyName,
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

func (s *OrganizationService) DeleteEvent(audit *AuditService, ctx EventContext, eventID uint) error {
	var event models.OrganizationEvent
	if err := database.DB.First(&event, eventID).Error; err != nil {
		return err
	}
	entityID := event.EntityID
	eventType := event.EventType
	companyName := event.CompanyName

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		if eventType == "create" {
			var count int64
			tx.Model(&models.OrganizationEvent{}).Where("entity_id = ?", entityID).Count(&count)
			if count == 0 {
				tx.Model(&models.Entity{}).Where("id = ?", entityID).Update("status", "deleted")
			}
		}
		if err := s.rebuildSnapshotsTx(tx, entityID); err != nil {
			return err
		}
		return audit.LogTx(tx, ctx.UserID, ctx.Username, "delete", "organization_event", eventID, companyName, "删除事件", models.JSONMap{
			"entity_id":    entityID,
			"company_name": companyName,
		})
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *OrganizationService) buildEventFromInput(input *OrganizationEventInput, effectiveDate time.Time) *models.OrganizationEvent {
	event := &models.OrganizationEvent{
		EntityID:      input.EntityID,
		EventType:     input.EventType,
		EventName:     input.EventName,
		EffectiveDate: effectiveDate,
		CompanyName:   input.CompanyName,
	}

	cf := input.ChangedFields
	if cf == nil {
		cf = models.JSONMap{}
	}

	if v, ok := cf["credit_code"]; ok {
		event.CreditCode, _ = v.(string)
	}
	if v, ok := cf["address"]; ok {
		event.Address, _ = v.(string)
	}
	if v, ok := cf["phone"]; ok {
		event.Phone, _ = v.(string)
	}
	if v, ok := cf["bank_name"]; ok {
		event.BankName, _ = v.(string)
	}
	if v, ok := cf["bank_account"]; ok {
		event.BankAccount, _ = v.(string)
	}
	if v, ok := cf["business_license_file_id"]; ok {
		event.BusinessLicenseFileID = uintPtr(v)
	}
	if v, ok := cf["official_seal_file_id"]; ok {
		event.OfficialSealFileID = uintPtr(v)
	}

	return event
}

func uintPtr(v interface{}) *uint {
	switch val := v.(type) {
	case float64:
		if val > 0 {
			u := uint(val)
			return &u
		}
	case json.Number:
		if i, err := val.Int64(); err == nil && i > 0 {
			u := uint(i)
			return &u
		}
	}
	return nil
}

func (s *OrganizationService) rebuildSnapshotsTx(tx *gorm.DB, entityID uint) error {
	var events []models.OrganizationEvent
	if err := tx.Where("entity_id = ?", entityID).Order("effective_date ASC, created_at ASC").Find(&events).Error; err != nil {
		return err
	}

	tx.Where("entity_id = ?", entityID).Delete(&models.OrganizationSnapshot{})

	var prev models.OrganizationSnapshot

	for i, event := range events {
		var snapshot models.OrganizationSnapshot

		if i == 0 {
			snapshot = s.snapshotFromEvent(&event)
		} else {
			snapshot = prev
			snapshot.ID = 0
			snapshot.EventID = event.ID
			snapshot.EffectiveDate = event.EffectiveDate
			s.mergeOrgFields(&snapshot, &event)
		}

		snapshot.EntityID = entityID
		snapshot.IsLatest = i == len(events)-1

		if err := tx.Create(&snapshot).Error; err != nil {
			return err
		}
		prev = snapshot
	}
	return nil
}

func (s *OrganizationService) snapshotFromEvent(event *models.OrganizationEvent) models.OrganizationSnapshot {
	return models.OrganizationSnapshot{
		EntityID:              event.EntityID,
		EventID:               event.ID,
		EffectiveDate:         event.EffectiveDate,
		CompanyName:           event.CompanyName,
		CreditCode:            event.CreditCode,
		Address:               event.Address,
		Phone:                 event.Phone,
		BankName:              event.BankName,
		BankAccount:           event.BankAccount,
		BusinessLicenseFileID: event.BusinessLicenseFileID,
		OfficialSealFileID:    event.OfficialSealFileID,
	}
}

func (s *OrganizationService) mergeOrgFields(snap *models.OrganizationSnapshot, event *models.OrganizationEvent) {
	cf := event.ChangedFields
	if cf == nil {
		return
	}
	if _, ok := cf["company_name"]; ok {
		snap.CompanyName = event.CompanyName
	}
	if _, ok := cf["credit_code"]; ok {
		snap.CreditCode = event.CreditCode
	}
	if _, ok := cf["address"]; ok {
		snap.Address = event.Address
	}
	if _, ok := cf["phone"]; ok {
		snap.Phone = event.Phone
	}
	if _, ok := cf["bank_name"]; ok {
		snap.BankName = event.BankName
	}
	if _, ok := cf["bank_account"]; ok {
		snap.BankAccount = event.BankAccount
	}
	if _, ok := cf["business_license_file_id"]; ok {
		snap.BusinessLicenseFileID = event.BusinessLicenseFileID
	}
	if _, ok := cf["official_seal_file_id"]; ok {
		snap.OfficialSealFileID = event.OfficialSealFileID
	}
}

func (s *OrganizationService) ListSnapshots(search string, page, pageSize int) ([]models.OrganizationSnapshot, int64, error) {
	var total int64
	subQuery := database.DB.Table("organization_snapshots").
		Select("MAX(id) as id").
		Where("is_latest = ?", true).
		Group("entity_id")

	query := database.DB.Table("organization_snapshots").Where("id IN (?)", subQuery)
	if search != "" {
		query = query.Where("company_name LIKE ?", "%"+search+"%")
	}
	query.Count(&total)

	var snapshots []models.OrganizationSnapshot
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&snapshots).Error; err != nil {
		return nil, 0, err
	}
	return snapshots, total, nil
}

func (s *OrganizationService) GetSnapshot(entityID uint) (*models.OrganizationSnapshot, error) {
	var snapshot models.OrganizationSnapshot
	if err := database.DB.Where("entity_id = ? AND is_latest = ?", entityID, true).First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *OrganizationService) GetHistory(entityID uint) ([]models.OrganizationSnapshot, error) {
	var snapshots []models.OrganizationSnapshot
	if err := database.DB.Where("entity_id = ?", entityID).Order("effective_date ASC").Find(&snapshots).Error; err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (s *OrganizationService) ListEvents(entityID uint, page, pageSize int) ([]models.OrganizationEvent, int64, error) {
	var events []models.OrganizationEvent
	var total int64
	query := database.DB.Model(&models.OrganizationEvent{})
	if entityID > 0 {
		query = query.Where("entity_id = ?", entityID)
	}
	query.Count(&total)
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("effective_date DESC, id DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (s *OrganizationService) CreateEntity(name string) (*models.Entity, error) {
	entity := &models.Entity{Type: "organization", Name: name, Status: "active"}
	if err := database.DB.Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}
