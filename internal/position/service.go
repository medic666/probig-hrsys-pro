package position

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"probig/internal/person"
	"probig/internal/pkg/audit"
)

type Service struct {
	dao           *DAO
	personService *person.Service
}

var globalService *Service

func NewService(db *gorm.DB, personSvc *person.Service) *Service {
	svc := &Service{dao: NewDAO(db), personService: personSvc}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) CreateEvent(event *PositionEvent, operatorID uint, operatorName string) error {
	if err := s.dao.CreateEvent(event); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "position_event", event.ID, "create", "", event)
	s.RebuildSnapshotsForPerson(event.PersonID)
	return nil
}

func (s *Service) GetEventByID(id uint) (*PositionEvent, error) {
	return s.dao.GetEventByID(id)
}

func (s *Service) ListEvents(personID uint, page, pageSize int) ([]PositionEvent, int64, error) {
	return s.dao.ListEvents(personID, page, pageSize)
}

func (s *Service) UpdateEvent(event *PositionEvent, operatorID uint, operatorName string) error {
	old, err := s.dao.GetEventByID(event.ID)
	if err != nil {
		return err
	}
	if err := s.dao.UpdateEvent(event); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "position_event", event.ID, "update", old, event)
	s.RebuildSnapshotsForPerson(event.PersonID)
	return nil
}

func (s *Service) DeleteEvent(id uint, operatorID uint, operatorName string) error {
	event, err := s.dao.GetEventByID(id)
	if err != nil {
		return err
	}
	personID := event.PersonID
	if err := s.dao.DeleteEvent(id); err != nil {
		return err
	}
	s.logAudit(operatorID, operatorName, "position_event", id, "delete", event, nil)
	s.RebuildSnapshotsForPerson(personID)
	return nil
}

func (s *Service) RebuildSnapshotsForPerson(personID uint) error {
	if err := s.dao.DeleteSnapshotsByPersonID(personID); err != nil {
		return err
	}

	events, err := s.dao.GetAllEventsByPersonID(personID)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	for i := range events {
		events[i].EffectiveDate = truncateDate(events[i].EffectiveDate)
		if events[i].EntryDate != nil {
			v := truncateDate(*events[i].EntryDate)
			events[i].EntryDate = &v
		}
		if events[i].LeaveDate != nil {
			v := truncateDate(*events[i].LeaveDate)
			events[i].LeaveDate = &v
		}
	}

	earliestDate, err := time.Parse("2006-01-02", events[0].EffectiveDate)
	if err != nil {
		return err
	}

	now := time.Now()
	endOfMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	var snapshots []PositionSnapshot
	for d := earliestDate; !d.After(endOfMonth); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		snapshot := s.buildSnapshot(events, dateStr)
		if snapshot != nil {
			snapshot.PersonID = personID
			snapshot.SnapshotDate = dateStr
			snapshots = append(snapshots, *snapshot)
		}
	}

	if len(snapshots) > 0 {
		return s.dao.BatchCreateSnapshots(snapshots)
	}
	return nil
}

func (s *Service) buildSnapshot(events []PositionEvent, date string) *PositionSnapshot {
	var snapshot PositionSnapshot
	var leaveDate *string

	for i := range events {
		if events[i].EffectiveDate > date {
			continue
		}

		if events[i].EntryDate != nil {
			snapshot.EntryDate = events[i].EntryDate
		}
		if events[i].LeaveDate != nil {
			snapshot.LeaveDate = events[i].LeaveDate
			leaveDate = events[i].LeaveDate
		}
		if events[i].AttendanceGroup != nil {
			snapshot.AttendanceGroup = *events[i].AttendanceGroup
		}
		if events[i].HasAnnualLeave != nil {
			snapshot.HasAnnualLeave = *events[i].HasAnnualLeave
		}
		if events[i].HasAttendanceBonus != nil {
			snapshot.HasAttendanceBonus = *events[i].HasAttendanceBonus
		}
		if events[i].BaseSalary != nil {
			snapshot.BaseSalary = *events[i].BaseSalary
		}
		if events[i].PerformanceSalary != nil {
			snapshot.PerformanceSalary = *events[i].PerformanceSalary
		}
		if events[i].SalaryDays != nil {
			snapshot.SalaryDays = *events[i].SalaryDays
		}
		if events[i].PostAllowance != nil {
			snapshot.PostAllowance = *events[i].PostAllowance
		}
		if events[i].MealAllowance != nil {
			snapshot.MealAllowance = *events[i].MealAllowance
		}
		if events[i].HousingAllowance != nil {
			snapshot.HousingAllowance = *events[i].HousingAllowance
		}
		if events[i].TransportAllowance != nil {
			snapshot.TransportAllowance = *events[i].TransportAllowance
		}
		if events[i].HighTempAllowance != nil {
			snapshot.HighTempAllowance = *events[i].HighTempAllowance
		}
		if events[i].InsuranceCompensation != nil {
			snapshot.InsuranceCompensation = *events[i].InsuranceCompensation
		}
		if events[i].FundCompensation != nil {
			snapshot.FundCompensation = *events[i].FundCompensation
		}
		if events[i].SocialSecurityDeduct != nil {
			snapshot.SocialSecurityDeduct = *events[i].SocialSecurityDeduct
		}
		if events[i].HousingFundDeduct != nil {
			snapshot.HousingFundDeduct = *events[i].HousingFundDeduct
		}
	}

	if leaveDate != nil {
		leaveParsed, err := time.Parse("2006-01-02", truncateDate(*leaveDate))
		if err != nil {
			return nil
		}
		dateParsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil
		}
		if dateParsed.After(leaveParsed) {
			return nil
		}
	}

	return &snapshot
}

func truncateDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func (s *Service) GetSnapshotByPersonAndDate(personID uint, date string) (*PositionSnapshot, error) {
	return s.dao.GetSnapshotByPersonIDAndDate(personID, date)
}

func (s *Service) GetSnapshotsByMonth(personIDs []uint, yearMonth string) ([]PositionSnapshot, error) {
	return s.dao.GetSnapshotsByMonth(personIDs, yearMonth)
}

func (s *Service) GetSnapshotForDate(personID uint, date string) (*PositionSnapshot, error) {
	snapshot, err := s.dao.GetSnapshotByPersonIDAndDate(personID, date)
	if err != nil {
		return nil, err
	}
	if snapshot.LeaveDate != nil {
		leaveParsed, err := time.Parse("2006-01-02", *snapshot.LeaveDate)
		if err != nil {
			return nil, err
		}
		dateParsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}
		if dateParsed.After(leaveParsed) {
			return nil, gorm.ErrRecordNotFound
		}
	}
	return snapshot, nil
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
