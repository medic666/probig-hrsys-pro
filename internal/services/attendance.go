package services

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/models"
)

type AttendanceService struct {
	db    *sqlx.DB
	audit *AuditService
}

func NewAttendanceService(db *sqlx.DB, audit *AuditService) *AttendanceService {
	return &AttendanceService{db: db, audit: audit}
}

func (s *AttendanceService) ListEvents(page, pageSize int, personID uint, period, eventType string) ([]models.AttendanceEvent, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM attendance_events ae WHERE 1=1"
	args := []interface{}{}

	if personID > 0 {
		countSQL += " AND ae.person_id = ?"
		args = append(args, personID)
	}
	if period != "" {
		countSQL += " AND ae.date >= ? AND ae.date <= ?"
		start := period + "-01"
		end := lastDayOfMonth(period)
		args = append(args, start, end)
	}
	if eventType != "" {
		countSQL += " AND ae.event_type = ?"
		args = append(args, eventType)
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := "SELECT ae.*, e.name as person_name FROM attendance_events ae LEFT JOIN entities e ON e.id = ae.person_id WHERE 1=1"
	if personID > 0 {
		querySQL += " AND ae.person_id = ?"
	}
	if period != "" {
		querySQL += " AND ae.date >= ? AND ae.date <= ?"
	}
	if eventType != "" {
		querySQL += " AND ae.event_type = ?"
	}
	querySQL += " ORDER BY ae.date DESC, ae.id DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	var events []models.AttendanceEvent
	if err := s.db.Select(&events, querySQL, queryArgs...); err != nil {
		return nil, 0, err
	}
	if events == nil {
		events = []models.AttendanceEvent{}
	}
	return events, total, nil
}

func (s *AttendanceService) CreateEvent(req models.AttendanceEventRequest, userID uint, ip string) (*models.AttendanceEvent, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if req.Duration <= 0 {
		req.Duration = 1.0
	}

	result, err := tx.Exec(
		`INSERT INTO attendance_events (person_id, date, event_type, duration, remark, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.PersonID, req.Date, req.EventType, req.Duration, req.Remark, userID, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create attendance event: %w", err)
	}
	eventID, _ := result.LastInsertId()

	s.audit.Log(tx, userID, "create", "attendance_event", ptrUint(uint(eventID)), req, ip)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.AttendanceEvent{
		ID:        uint(eventID),
		PersonID:  req.PersonID,
		Date:      req.Date,
		EventType: req.EventType,
		Duration:  req.Duration,
		Remark:    req.Remark,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *AttendanceService) UpdateEvent(eventID uint, req models.AttendanceEventRequest, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE attendance_events SET person_id = ?, date = ?, event_type = ?, duration = ?, remark = ?, updated_at = ? WHERE id = ?`,
		req.PersonID, req.Date, req.EventType, req.Duration, req.Remark, time.Now(), eventID,
	)
	if err != nil {
		return err
	}

	s.audit.Log(tx, userID, "update", "attendance_event", ptrUint(eventID), req, ip)
	return tx.Commit()
}

func (s *AttendanceService) DeleteEvent(eventID uint, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var evt models.AttendanceEvent
	if err := tx.Get(&evt, "SELECT * FROM attendance_events WHERE id = ?", eventID); err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM attendance_events WHERE id = ?", eventID)
	if err != nil {
		return err
	}

	s.audit.Log(tx, userID, "delete", "attendance_event", ptrUint(eventID), evt, ip)
	return tx.Commit()
}

func (s *AttendanceService) Calculate(req models.CalculateRequest, userID uint, ip string) error {
	period := req.Period
	start := period + "-01"
	end := lastDayOfMonth(period)

	var personIDs []uint
	if req.PersonID > 0 {
		personIDs = []uint{req.PersonID}
	} else {
		if err := s.db.Select(&personIDs,
			"SELECT id FROM entities WHERE type = 'person' AND status = 'active'",
		); err != nil {
			return err
		}
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, pid := range personIDs {
		summary, err := s.computeAttendanceSummary(tx, pid, start, end, period)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			`INSERT OR REPLACE INTO attendance_summaries
			 (person_id, period, normal_attendance_days, supplementary_attendance_days,
			  compensatory_leave_days, personal_leave_days, sick_leave_days,
			  annual_leave_days, statutory_leave_days, welfare_leave_days,
			  workday_overtime_days, holiday_overtime_days, missing_clock_count,
			  late_count, early_leave_count, annual_leave_allot, annual_leave_carryover,
			  violation_count, calculated_at)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			pid, period,
			summary["normal_attendance_days"], summary["supplementary_attendance_days"],
			summary["compensatory_leave_days"], summary["personal_leave_days"],
			summary["sick_leave_days"], summary["annual_leave_days"],
			summary["statutory_leave_days"], summary["welfare_leave_days"],
			summary["workday_overtime_days"], summary["holiday_overtime_days"],
			summary["missing_clock_count"], summary["late_count"],
			summary["early_leave_count"], summary["annual_leave_allot"],
			summary["annual_leave_carryover"], summary["violation_count"],
			time.Now(),
		)
		if err != nil {
			return err
		}
	}

	s.audit.Log(tx, userID, "calculate", "attendance_summary", nil, req, ip)
	return tx.Commit()
}

func (s *AttendanceService) computeAttendanceSummary(tx *sqlx.Tx, personID uint, start, end, period string) (map[string]float64, error) {
	summary := map[string]float64{
		"normal_attendance_days":          0,
		"supplementary_attendance_days":   0,
		"compensatory_leave_days":         0,
		"personal_leave_days":             0,
		"sick_leave_days":                 0,
		"annual_leave_days":               0,
		"statutory_leave_days":            0,
		"welfare_leave_days":              0,
		"workday_overtime_days":           0,
		"holiday_overtime_days":           0,
		"missing_clock_count":             0,
		"late_count":                      0,
		"early_leave_count":               0,
		"annual_leave_allot":              0,
		"annual_leave_carryover":          0,
		"violation_count":                 0,
	}

	var events []models.AttendanceEvent
	if err := tx.Select(&events,
		"SELECT * FROM attendance_events WHERE person_id = ? AND date >= ? AND date <= ?",
		personID, start, end,
	); err != nil {
		return nil, err
	}

	for _, e := range events {
		switch e.EventType {
		case "normal_attendance":
			summary["normal_attendance_days"] += e.Duration
		case "supplementary_attendance":
			summary["supplementary_attendance_days"] += e.Duration
		case "compensatory_leave":
			summary["compensatory_leave_days"] += e.Duration
		case "personal_leave":
			summary["personal_leave_days"] += e.Duration
		case "sick_leave":
			summary["sick_leave_days"] += e.Duration
		case "annual_leave":
			summary["annual_leave_days"] += e.Duration
		case "statutory_leave":
			summary["statutory_leave_days"] += e.Duration
		case "welfare_leave":
			summary["welfare_leave_days"] += e.Duration
		case "workday_overtime":
			summary["workday_overtime_days"] += e.Duration
		case "holiday_overtime":
			summary["holiday_overtime_days"] += e.Duration
		case "missing_clock":
			summary["missing_clock_count"] += e.Duration
			summary["violation_count"] += e.Duration
		case "late":
			summary["late_count"] += e.Duration
			summary["violation_count"] += e.Duration
		case "early_leave":
			summary["early_leave_count"] += e.Duration
			summary["violation_count"] += e.Duration
		case "annual_leave_allot":
			summary["annual_leave_allot"] += e.Duration
		case "annual_leave_carryover":
			summary["annual_leave_carryover"] += e.Duration
		}
	}

	return summary, nil
}

func (s *AttendanceService) ListSummaries(page, pageSize int, personID uint, period string) ([]models.AttendanceSummary, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM attendance_summaries as2 WHERE 1=1"
	args := []interface{}{}
	if personID > 0 {
		countSQL += " AND as2.person_id = ?"
		args = append(args, personID)
	}
	if period != "" {
		countSQL += " AND as2.period = ?"
		args = append(args, period)
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := "SELECT as2.*, e.name as person_name FROM attendance_summaries as2 LEFT JOIN entities e ON e.id = as2.person_id WHERE 1=1"
	if personID > 0 {
		querySQL += " AND as2.person_id = ?"
	}
	if period != "" {
		querySQL += " AND as2.period = ?"
	}
	querySQL += " ORDER BY as2.period DESC, as2.person_id ASC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	var summaries []models.AttendanceSummary
	if err := s.db.Select(&summaries, querySQL, queryArgs...); err != nil {
		return nil, 0, err
	}
	if summaries == nil {
		summaries = []models.AttendanceSummary{}
	}
	return summaries, total, nil
}

func lastDayOfMonth(period string) string {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return period + "-31"
	}
	return t.AddDate(0, 1, -1).Format("2006-01-02")
}
