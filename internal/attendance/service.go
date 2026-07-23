package attendance

import (
	"encoding/json"
	"fmt"
	"time"

	"probig/internal/common"
	"probig/internal/event"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo         *Repository
	eventService *event.Service
}

func NewService(repo *Repository, eventService *event.Service) *Service {
	return &Service{repo: repo, eventService: eventService}
}

func (s *Service) CreateEvent(attEvent *AttendanceEvent, operatorID int64, remark string) error {
	tx, err := common.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.CreateEventTx(tx, attEvent); err != nil {
		return err
	}

	payload, _ := json.Marshal(attEvent)
	if err := s.eventService.RecordEventTx(tx, "attendance", attEvent.ID, "create", string(payload), operatorID, remark); err != nil {
		return err
	}

	if attEvent.EventType == "年假" {
		if err := s.consumeAnnualLeaveTx(tx, attEvent); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) consumeAnnualLeaveTx(tx *sqlx.Tx, attEvent *AttendanceEvent) error {
	daysToConsume := attEvent.DurationHours / 8.0

	rows, err := tx.Query(`SELECT id, days_remaining FROM annual_leave_grants
		WHERE person_id = ? AND days_remaining > 0 ORDER BY grant_date`, attEvent.PersonID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var remaining float64
		if err := rows.Scan(&id, &remaining); err != nil {
			return err
		}

		if daysToConsume <= 0 {
			break
		}

		if remaining >= daysToConsume {
			newRemaining := remaining - daysToConsume
			if _, err := tx.Exec("UPDATE annual_leave_grants SET days_remaining = ? WHERE id = ?", newRemaining, id); err != nil {
				return err
			}
			daysToConsume = 0
		} else {
			if _, err := tx.Exec("UPDATE annual_leave_grants SET days_remaining = 0 WHERE id = ?", id); err != nil {
				return err
			}
			daysToConsume -= remaining
		}
	}

	return nil
}

func (s *Service) UpdateEvent(id int64, attEvent *AttendanceEvent, operatorID int64, remark string) error {
	existing, err := s.repo.GetEventByID(id)
	if err != nil {
		return common.ErrNotFound
	}
	_ = existing

	if err := s.repo.UpdateEvent(id, attEvent); err != nil {
		return err
	}

	payload, _ := json.Marshal(attEvent)
	return s.eventService.RecordEvent("attendance", id, "update", string(payload), operatorID, remark)
}

func (s *Service) DeleteEvent(id int64, operatorID int64, remark string) error {
	existing, err := s.repo.GetEventByID(id)
	if err != nil {
		return common.ErrNotFound
	}

	if err := s.repo.DeleteEvent(id); err != nil {
		return err
	}

	payload, _ := json.Marshal(existing)
	return s.eventService.RecordEvent("attendance", id, "delete", string(payload), operatorID, remark)
}

func (s *Service) ListEvents(personID int64, yearMonth, eventType string, page, pageSize int) ([]AttendanceEvent, int64, error) {
	return s.repo.ListEvents(personID, yearMonth, eventType, page, pageSize)
}

func (s *Service) GetLeaveBalance(personID int64) (*MonthlyLeaveBalance, error) {
	return s.repo.GetAnnualLeaveBalance(personID)
}

func (s *Service) GrantAnnualLeave(personID int64, operatorID int64) error {
	hireDate := ""
	err := common.DB.Get(&hireDate, "SELECT hire_date FROM persons WHERE id = ?", personID)
	if err != nil {
		return fmt.Errorf("person not found: %w", err)
	}

	if hireDate == "" {
		return fmt.Errorf("person has no hire date")
	}

	hireTime, err := time.Parse("2006-01-02", hireDate)
	if err != nil {
		return fmt.Errorf("invalid hire date: %w", err)
	}

	now := time.Now()
	yearsOfService := now.Year() - hireTime.Year()
	if now.Month() < hireTime.Month() || (now.Month() == hireTime.Month() && now.Day() < hireTime.Day()) {
		yearsOfService--
	}

	if yearsOfService < 1 {
		return fmt.Errorf("employee has less than 1 year of service")
	}

	daysGranted := 5.0
	if yearsOfService >= 20 {
		daysGranted = 15.0
	} else if yearsOfService >= 10 {
		daysGranted = 10.0
	}

	grant := &AnnualLeaveGrant{
		PersonID:    personID,
		GrantDate:   now.Format("2006-01-02"),
		DaysGranted: daysGranted,
		YearMonth:   now.Format("2006-01"),
	}

	if err := s.repo.GrantAnnualLeave(grant); err != nil {
		return err
	}

	payload, _ := json.Marshal(grant)
	return s.eventService.RecordEvent("annual_leave_grant", grant.ID, "grant", string(payload), operatorID, "发放年假")
}

func (s *Service) CloseMonth(personID int64, yearMonth string, operatorID int64) error {
	remaining, err := s.repo.CloseMonthGrants(personID, yearMonth, "月结结转")
	if err != nil {
		return err
	}

	if remaining > 0 {
		carryEvent := &AttendanceEvent{
			PersonID:      personID,
			EventDate:     yearMonth + "-28",
			EventType:     "节假日出勤",
			DurationHours: remaining * 8,
			Description:   fmt.Sprintf("年假结转: %.1f天转为休息日加班", remaining),
			OperatorID:    operatorID,
		}

		if err := s.repo.CreateEvent(carryEvent); err != nil {
			return err
		}

		payload, _ := json.Marshal(carryEvent)
		return s.eventService.RecordEvent("attendance", carryEvent.ID, "carry_over", string(payload), operatorID, "年假结转")
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"personId":   personID,
		"yearMonth":  yearMonth,
		"carriedDays": remaining,
	})
	return s.eventService.RecordEvent("attendance_close", personID, "close_month", string(payload), operatorID, "考勤月结")
}

func (s *Service) GetEventsByPersonMonth(personID int64, yearMonth string) ([]AttendanceEvent, error) {
	return s.repo.GetEventsByPersonMonth(personID, yearMonth)
}
