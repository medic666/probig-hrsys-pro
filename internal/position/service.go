package position

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/database"
)

var OnSnapshotsRebuilt func(personID uint)

type positionFieldState struct {
	AttendanceGroup       string
	HasAnnualLeave        *bool
	HasAttendanceBonus    *bool
	BaseSalary            *float64
	PerformanceSalary     *float64
	SalaryDays            *int
	PostAllowance         *float64
	MealAllowance         *float64
	HousingAllowance      *float64
	TransportAllowance    *float64
	HighTempAllowance     *float64
	InsuranceCompensation *float64
	FundCompensation      *float64
	SocialSecurityDeduct  *float64
	HousingFundDeduct     *float64
}

func (s positionFieldState) isEmpty() bool {
	return s.AttendanceGroup == "" &&
		s.HasAnnualLeave == nil &&
		s.HasAttendanceBonus == nil &&
		s.BaseSalary == nil &&
		s.PerformanceSalary == nil &&
		s.SalaryDays == nil &&
		s.PostAllowance == nil &&
		s.MealAllowance == nil &&
		s.HousingAllowance == nil &&
		s.TransportAllowance == nil &&
		s.HighTempAllowance == nil &&
		s.InsuranceCompensation == nil &&
		s.FundCompensation == nil &&
		s.SocialSecurityDeduct == nil &&
		s.HousingFundDeduct == nil
}

func cloneState(s positionFieldState) positionFieldState {
	return s
}

func stateChanged(a, b positionFieldState) bool {
	if a.AttendanceGroup != b.AttendanceGroup {
		return true
	}
	if (a.HasAnnualLeave == nil) != (b.HasAnnualLeave == nil) {
		return true
	}
	if a.HasAnnualLeave != nil && b.HasAnnualLeave != nil && *a.HasAnnualLeave != *b.HasAnnualLeave {
		return true
	}
	if (a.HasAttendanceBonus == nil) != (b.HasAttendanceBonus == nil) {
		return true
	}
	if a.HasAttendanceBonus != nil && b.HasAttendanceBonus != nil && *a.HasAttendanceBonus != *b.HasAttendanceBonus {
		return true
	}
	if (a.BaseSalary == nil) != (b.BaseSalary == nil) {
		return true
	}
	if a.BaseSalary != nil && b.BaseSalary != nil && *a.BaseSalary != *b.BaseSalary {
		return true
	}
	if (a.PerformanceSalary == nil) != (b.PerformanceSalary == nil) {
		return true
	}
	if a.PerformanceSalary != nil && b.PerformanceSalary != nil && *a.PerformanceSalary != *b.PerformanceSalary {
		return true
	}
	if (a.SalaryDays == nil) != (b.SalaryDays == nil) {
		return true
	}
	if a.SalaryDays != nil && b.SalaryDays != nil && *a.SalaryDays != *b.SalaryDays {
		return true
	}
	if (a.PostAllowance == nil) != (b.PostAllowance == nil) {
		return true
	}
	if a.PostAllowance != nil && b.PostAllowance != nil && *a.PostAllowance != *b.PostAllowance {
		return true
	}
	if (a.MealAllowance == nil) != (b.MealAllowance == nil) {
		return true
	}
	if a.MealAllowance != nil && b.MealAllowance != nil && *a.MealAllowance != *b.MealAllowance {
		return true
	}
	if (a.HousingAllowance == nil) != (b.HousingAllowance == nil) {
		return true
	}
	if a.HousingAllowance != nil && b.HousingAllowance != nil && *a.HousingAllowance != *b.HousingAllowance {
		return true
	}
	if (a.TransportAllowance == nil) != (b.TransportAllowance == nil) {
		return true
	}
	if a.TransportAllowance != nil && b.TransportAllowance != nil && *a.TransportAllowance != *b.TransportAllowance {
		return true
	}
	if (a.HighTempAllowance == nil) != (b.HighTempAllowance == nil) {
		return true
	}
	if a.HighTempAllowance != nil && b.HighTempAllowance != nil && *a.HighTempAllowance != *b.HighTempAllowance {
		return true
	}
	if (a.InsuranceCompensation == nil) != (b.InsuranceCompensation == nil) {
		return true
	}
	if a.InsuranceCompensation != nil && b.InsuranceCompensation != nil && *a.InsuranceCompensation != *b.InsuranceCompensation {
		return true
	}
	if (a.FundCompensation == nil) != (b.FundCompensation == nil) {
		return true
	}
	if a.FundCompensation != nil && b.FundCompensation != nil && *a.FundCompensation != *b.FundCompensation {
		return true
	}
	if (a.SocialSecurityDeduct == nil) != (b.SocialSecurityDeduct == nil) {
		return true
	}
	if a.SocialSecurityDeduct != nil && b.SocialSecurityDeduct != nil && *a.SocialSecurityDeduct != *b.SocialSecurityDeduct {
		return true
	}
	if (a.HousingFundDeduct == nil) != (b.HousingFundDeduct == nil) {
		return true
	}
	if a.HousingFundDeduct != nil && b.HousingFundDeduct != nil && *a.HousingFundDeduct != *b.HousingFundDeduct {
		return true
	}
	return false
}

func applyEvent(state *positionFieldState, event *PositionEvent) {
	if event.AttendanceGroup != nil {
		state.AttendanceGroup = *event.AttendanceGroup
	}
	if event.HasAnnualLeave != nil {
		state.HasAnnualLeave = boolPtr(*event.HasAnnualLeave)
	}
	if event.HasAttendanceBonus != nil {
		state.HasAttendanceBonus = boolPtr(*event.HasAttendanceBonus)
	}
	if event.BaseSalary != nil {
		state.BaseSalary = floatPtr(*event.BaseSalary)
	}
	if event.PerformanceSalary != nil {
		state.PerformanceSalary = floatPtr(*event.PerformanceSalary)
	}
	if event.SalaryDays != nil {
		state.SalaryDays = intPtr(*event.SalaryDays)
	}
	if event.PostAllowance != nil {
		state.PostAllowance = floatPtr(*event.PostAllowance)
	}
	if event.MealAllowance != nil {
		state.MealAllowance = floatPtr(*event.MealAllowance)
	}
	if event.HousingAllowance != nil {
		state.HousingAllowance = floatPtr(*event.HousingAllowance)
	}
	if event.TransportAllowance != nil {
		state.TransportAllowance = floatPtr(*event.TransportAllowance)
	}
	if event.HighTempAllowance != nil {
		state.HighTempAllowance = floatPtr(*event.HighTempAllowance)
	}
	if event.InsuranceCompensation != nil {
		state.InsuranceCompensation = floatPtr(*event.InsuranceCompensation)
	}
	if event.FundCompensation != nil {
		state.FundCompensation = floatPtr(*event.FundCompensation)
	}
	if event.SocialSecurityDeduct != nil {
		state.SocialSecurityDeduct = floatPtr(*event.SocialSecurityDeduct)
	}
	if event.HousingFundDeduct != nil {
		state.HousingFundDeduct = floatPtr(*event.HousingFundDeduct)
	}
}

func boolPtr(v bool) *bool {
	b := v
	return &b
}

func floatPtr(v float64) *float64 {
	f := v
	return &f
}

func intPtr(v int) *int {
	i := v
	return &i
}

func stateToSnapshot(personID uint, startDate time.Time, endDate time.Time, state positionFieldState, entryDate, leaveDate *time.Time) PositionSnapshot {
	now := time.Now()
	s := PositionSnapshot{
		PersonID:           personID,
		AttendanceGroup:    state.AttendanceGroup,
		HasAnnualLeave:     derefBool(state.HasAnnualLeave, false),
		HasAttendanceBonus: derefBool(state.HasAttendanceBonus, false),
		BaseSalary:         derefFloat(state.BaseSalary, 0),
		PerformanceSalary:  derefFloat(state.PerformanceSalary, 0),
		SalaryDays:         derefInt(state.SalaryDays, 0),
		PostAllowance:      derefFloat(state.PostAllowance, 0),
		MealAllowance:      derefFloat(state.MealAllowance, 0),
		HousingAllowance:   derefFloat(state.HousingAllowance, 0),
		TransportAllowance: derefFloat(state.TransportAllowance, 0),
		HighTempAllowance:  derefFloat(state.HighTempAllowance, 0),
		InsuranceCompensation: derefFloat(state.InsuranceCompensation, 0),
		FundCompensation:      derefFloat(state.FundCompensation, 0),
		SocialSecurityDeduct:  derefFloat(state.SocialSecurityDeduct, 0),
		HousingFundDeduct:     derefFloat(state.HousingFundDeduct, 0),
		EntryDate:          entryDate,
		LeaveDate:          leaveDate,
		LastCalcAt:         &now,
	}

	effStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	effEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
	s.EffectiveStartDate = &effStart
	s.EffectiveEndDate = &effEnd

	return s
}

func derefBool(p *bool, defaultVal bool) bool {
	if p == nil {
		return defaultVal
	}
	return *p
}

func derefFloat(p *float64, defaultVal float64) float64 {
	if p == nil {
		return defaultVal
	}
	return *p
}

func derefInt(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}

func RebuildSnapshots(personID uint) error {
	events, err := FindEventsByPerson(personID)
	if err != nil {
		return fmt.Errorf("failed to find events: %w", err)
	}

	if err := DeleteAllSnapshotsByPerson(personID); err != nil {
		return fmt.Errorf("failed to delete existing snapshots: %w", err)
	}

	if len(events) == 0 {
		if OnSnapshotsRebuilt != nil {
			OnSnapshotsRebuilt(personID)
		}
		return nil
	}

	if events[0].EffectiveDate == nil {
		return fmt.Errorf("event with nil effective_date")
	}

	var snapshots []PositionSnapshot
	var currentState positionFieldState
	var entryDate *time.Time
	var leaveDate *time.Time
	var periodStart time.Time
	periodStartSet := false
	farFuture := FarFutureDate

	i := 0
	for i < len(events) {
		date := *events[i].EffectiveDate

		j := i
		for j < len(events) && events[j].EffectiveDate != nil && sameDay(*events[j].EffectiveDate, date) {
			j++
		}

		stateBefore := cloneState(currentState)

		for k := i; k < j; k++ {
			e := events[k]
			applyEvent(&currentState, &e)

			if e.EventName == "入职" {
				if leaveDate != nil && entryDate != nil {
					entryDate = e.EffectiveDate
					leaveDate = nil
				} else if entryDate == nil {
					entryDate = e.EffectiveDate
				}
			}
			if e.EventName == "离职" {
				leaveDate = e.EffectiveDate
			}
		}

		if !periodStartSet {
			periodStart = date
			periodStartSet = true
		} else if stateChanged(stateBefore, currentState) && !stateBefore.isEmpty() {
			endDate := date.Add(-24 * time.Hour)
			if endDate.Before(periodStart) {
				endDate = periodStart
			}
			snapshots = append(snapshots, stateToSnapshot(personID, periodStart, endDate, stateBefore, entryDate, nil))
			periodStart = date
		}

		i = j
	}

	if periodStartSet {
		finalLeaveDate := leaveDate
		snapshots = append(snapshots, stateToSnapshot(personID, periodStart, farFuture, currentState, entryDate, finalLeaveDate))
	}

	for idx := range snapshots {
		if entryDate != nil {
			ed := time.Date(entryDate.Year(), entryDate.Month(), entryDate.Day(), 0, 0, 0, 0, time.UTC)
			snapshots[idx].EntryDate = &ed
		}
	}

	if err := BatchCreateSnapshots(snapshots); err != nil {
		return fmt.Errorf("failed to create snapshots: %w", err)
	}

	if OnSnapshotsRebuilt != nil {
		OnSnapshotsRebuilt(personID)
	}

	return nil
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func CreateEvent(event *PositionEvent, operatorID uint, operatorName string, ip string) error {
	if event.PersonID == 0 {
		return fmt.Errorf("person_id is required")
	}
	if event.EffectiveDate == nil {
		return fmt.Errorf("effective_date is required")
	}
	if event.EventName == "" {
		return fmt.Errorf("event_name is required")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		personName := ""
		var p Person
		if err := tx.First(&p, event.PersonID).Error; err == nil {
			personName = p.Name
		}

		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "position_event", event.ID, personName, "新增", nil, event, ip); err != nil {
			return err
		}

		return RebuildSnapshots(event.PersonID)
	})
}

func UpdateEvent(event *PositionEvent, operatorID uint, operatorName string, ip string) error {
	if event.ID == 0 {
		return fmt.Errorf("id is required")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var oldEvent PositionEvent
		if err := tx.First(&oldEvent, event.ID).Error; err != nil {
			return err
		}

		if err := tx.Save(event).Error; err != nil {
			return err
		}

		personName := ""
		var p Person
		if err := tx.First(&p, event.PersonID).Error; err == nil {
			personName = p.Name
		}

		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "position_event", event.ID, personName, "修改", oldEvent, event, ip); err != nil {
			return err
		}

		return RebuildSnapshots(event.PersonID)
	})
}

func DeleteEventByID(id uint, operatorID uint, operatorName string, ip string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event PositionEvent
		if err := tx.First(&event, id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&event).Error; err != nil {
			return err
		}

		personName := ""
		var p Person
		if err := tx.First(&p, event.PersonID).Error; err == nil {
			personName = p.Name
		}

		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "position_event", event.ID, personName, "删除", event, nil, ip); err != nil {
			return err
		}

		return RebuildSnapshots(event.PersonID)
	})
}

func RestoreEventByID(id uint, operatorID uint, operatorName string, ip string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event PositionEvent
		if err := tx.Unscoped().First(&event, id).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}

		personName := ""
		var p Person
		if err := tx.First(&p, event.PersonID).Error; err == nil {
			personName = p.Name
		}

		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "position_event", event.ID, personName, "恢复", nil, event, ip); err != nil {
			return err
		}

		return RebuildSnapshots(event.PersonID)
	})
}

func GetCurrentSnapshot(personID uint) (*PositionSnapshot, error) {
	snapshots, err := GetSnapshotsByPerson(personID)
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, nil
	}
	latest := snapshots[len(snapshots)-1]
	return &latest, nil
}

func GetEmploymentStatus(personID uint) string {
	snap, err := GetCurrentSnapshot(personID)
	if err != nil || snap == nil {
		return "未知"
	}
	if snap.LeaveDate != nil {
		return "离职"
	}
	return "在职"
}

func MarshalSnapshotJSON(s PositionSnapshot) (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ListEventsWithFilter(personID uint, startDate, endDate *time.Time, eventName string, pageNum, pageSize int) ([]PositionEventWithName, int64, error) {
	return ListEvents(personID, startDate, endDate, eventName, pageNum, pageSize)
}

func ListSnapshotsWithFilter(q SnapshotQuery) ([]PositionSnapshotWithName, int64, error) {
	return ListSnapshots(q)
}
