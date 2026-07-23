package services

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/models"
)

type SnapshotEngine struct {
	db *sqlx.DB
}

func NewSnapshotEngine(db *sqlx.DB) *SnapshotEngine {
	return &SnapshotEngine{db: db}
}

func (e *SnapshotEngine) RebuildPersonSnapshots(tx *sqlx.Tx, personID uint) error {
	var events []models.PersonEvent
	if err := tx.Select(&events,
		"SELECT * FROM person_events WHERE person_id = ? ORDER BY effective_date ASC, id ASC",
		personID,
	); err != nil {
		return err
	}

	_, err := tx.Exec("DELETE FROM person_snapshots WHERE person_id = ?", personID)
	if err != nil {
		return err
	}

	current := models.DefaultPersonSnapshotData()

	for _, evt := range events {
		var eventData models.PersonSnapshotData
		if err := json.Unmarshal([]byte(evt.Payload), &eventData); err != nil {
			return err
		}
		current = mergePersonData(current, eventData)

		snapshotJSON, _ := json.Marshal(current)
		_, err := tx.Exec(
			`INSERT INTO person_snapshots (person_id, event_id, effective_date, snapshot_data, created_at) VALUES (?, ?, ?, ?, ?)`,
			personID, evt.ID, evt.EffectiveDate, string(snapshotJSON), time.Now(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergePersonData(base, overlay models.PersonSnapshotData) models.PersonSnapshotData {
	if overlay.Name != "" {
		base.Name = overlay.Name
	}
	if overlay.AttendanceGroup != "" {
		base.AttendanceGroup = overlay.AttendanceGroup
	}
	if overlay.EntryDate != "" {
		base.EntryDate = overlay.EntryDate
	}
	if overlay.BasicSalary != 0 {
		base.BasicSalary = overlay.BasicSalary
	}
	if overlay.PerformanceSalary != 0 {
		base.PerformanceSalary = overlay.PerformanceSalary
	}
	if overlay.SalaryDays != 0 {
		base.SalaryDays = overlay.SalaryDays
	}
	if overlay.PositionAllowance != 0 {
		base.PositionAllowance = overlay.PositionAllowance
	}
	if overlay.MealSubsidy != 0 {
		base.MealSubsidy = overlay.MealSubsidy
	}
	if overlay.HousingSubsidy != 0 {
		base.HousingSubsidy = overlay.HousingSubsidy
	}
	if overlay.TransportSubsidy != 0 {
		base.TransportSubsidy = overlay.TransportSubsidy
	}
	if overlay.HeatSubsidy != 0 {
		base.HeatSubsidy = overlay.HeatSubsidy
	}
	if overlay.InsuranceCompensation != 0 {
		base.InsuranceCompensation = overlay.InsuranceCompensation
	}
	if overlay.HousingFundCompensation != 0 {
		base.HousingFundCompensation = overlay.HousingFundCompensation
	}
	if overlay.SocialInsuranceDeduct != 0 {
		base.SocialInsuranceDeduct = overlay.SocialInsuranceDeduct
	}
	if overlay.HousingFundDeduct != 0 {
		base.HousingFundDeduct = overlay.HousingFundDeduct
	}
	if len(overlay.Phones) > 0 {
		base.Phones = overlay.Phones
	}
	if len(overlay.Emails) > 0 {
		base.Emails = overlay.Emails
	}
	if overlay.IDNumber != "" {
		base.IDNumber = overlay.IDNumber
	}
	if overlay.Gender != "" {
		base.Gender = overlay.Gender
	}
	if overlay.Birthday != "" {
		base.Birthday = overlay.Birthday
	}
	if overlay.Ethnicity != "" {
		base.Ethnicity = overlay.Ethnicity
	}
	if overlay.NativePlace != "" {
		base.NativePlace = overlay.NativePlace
	}
	if overlay.Address != "" {
		base.Address = overlay.Address
	}
	if len(overlay.BankCards) > 0 {
		base.BankCards = overlay.BankCards
	}
	if overlay.PoliticalStatus != "" {
		base.PoliticalStatus = overlay.PoliticalStatus
	}
	if overlay.MaritalStatus != "" {
		base.MaritalStatus = overlay.MaritalStatus
	}
	if overlay.Alias != "" {
		base.Alias = overlay.Alias
	}
	return base
}

func (e *SnapshotEngine) RebuildOrgSnapshots(tx *sqlx.Tx, orgID uint) error {
	var events []models.OrgEvent
	if err := tx.Select(&events,
		"SELECT * FROM org_events WHERE org_id = ? ORDER BY effective_date ASC, id ASC",
		orgID,
	); err != nil {
		return err
	}

	_, err := tx.Exec("DELETE FROM org_snapshots WHERE org_id = ?", orgID)
	if err != nil {
		return err
	}

	current := models.OrgSnapshotData{}

	for _, evt := range events {
		var eventData models.OrgSnapshotData
		if err := json.Unmarshal([]byte(evt.Payload), &eventData); err != nil {
			return err
		}
		current = mergeOrgData(current, eventData)

		snapshotJSON, _ := json.Marshal(current)
		_, err := tx.Exec(
			`INSERT INTO org_snapshots (org_id, event_id, effective_date, snapshot_data, created_at) VALUES (?, ?, ?, ?, ?)`,
			orgID, evt.ID, evt.EffectiveDate, string(snapshotJSON), time.Now(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeOrgData(base, overlay models.OrgSnapshotData) models.OrgSnapshotData {
	if overlay.CompanyName != "" {
		base.CompanyName = overlay.CompanyName
	}
	if overlay.CreditCode != "" {
		base.CreditCode = overlay.CreditCode
	}
	if overlay.Address != "" {
		base.Address = overlay.Address
	}
	if overlay.ContactPhone != "" {
		base.ContactPhone = overlay.ContactPhone
	}
	if overlay.BankName != "" {
		base.BankName = overlay.BankName
	}
	if overlay.BankAccount != "" {
		base.BankAccount = overlay.BankAccount
	}
	if overlay.BusinessLicenseFileID != nil {
		base.BusinessLicenseFileID = overlay.BusinessLicenseFileID
	}
	if overlay.SealFileID != nil {
		base.SealFileID = overlay.SealFileID
	}
	return base
}

func (e *SnapshotEngine) GetLatestPersonSnapshot(personID uint) (*models.PersonSnapshot, error) {
	var snapshot models.PersonSnapshot
	err := e.db.Get(&snapshot,
		"SELECT * FROM person_snapshots WHERE person_id = ? ORDER BY effective_date DESC, id DESC LIMIT 1",
		personID,
	)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (e *SnapshotEngine) GetLatestOrgSnapshot(orgID uint) (*models.OrgSnapshot, error) {
	var snapshot models.OrgSnapshot
	err := e.db.Get(&snapshot,
		"SELECT * FROM org_snapshots WHERE org_id = ? ORDER BY effective_date DESC, id DESC LIMIT 1",
		orgID,
	)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
