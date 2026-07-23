package attendance

import (
	"fmt"

	"probig/internal/common"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository() *Repository {
	return &Repository{db: common.DB}
}

func (r *Repository) CreateEvent(event *AttendanceEvent) error {
	result, err := r.db.Exec(`INSERT INTO attendance_events (person_id, event_date, event_type, start_time, end_time, duration_hours, description, operator_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		event.PersonID, event.EventDate, event.EventType, event.StartTime, event.EndTime, event.DurationHours, event.Description, event.OperatorID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	event.ID = id
	return nil
}

func (r *Repository) CreateEventTx(tx *sqlx.Tx, event *AttendanceEvent) error {
	result, err := tx.Exec(`INSERT INTO attendance_events (person_id, event_date, event_type, start_time, end_time, duration_hours, description, operator_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		event.PersonID, event.EventDate, event.EventType, event.StartTime, event.EndTime, event.DurationHours, event.Description, event.OperatorID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	event.ID = id
	return nil
}

func (r *Repository) UpdateEvent(id int64, event *AttendanceEvent) error {
	_, err := r.db.Exec(`UPDATE attendance_events SET event_date=?, event_type=?, start_time=?, end_time=?, duration_hours=?, description=?
		WHERE id=?`,
		event.EventDate, event.EventType, event.StartTime, event.EndTime, event.DurationHours, event.Description, id)
	return err
}

func (r *Repository) GetEventByID(id int64) (*AttendanceEvent, error) {
	var e AttendanceEvent
	err := r.db.Get(&e, "SELECT * FROM attendance_events WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) ListEvents(personID int64, yearMonth, eventType string, page, pageSize int) ([]AttendanceEvent, int64, error) {
	where := "1=1"
	args := []interface{}{}

	if personID > 0 {
		where += " AND person_id = ?"
		args = append(args, personID)
	}
	if yearMonth != "" {
		where += " AND event_date LIKE ?"
		args = append(args, yearMonth+"%")
	}
	if eventType != "" {
		where += " AND event_type = ?"
		args = append(args, eventType)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM attendance_events WHERE %s", where)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM attendance_events WHERE %s ORDER BY event_date DESC LIMIT ? OFFSET ?", where)
	queryArgs := append(args, pageSize, offset)

	var events []AttendanceEvent
	err := r.db.Select(&events, query, queryArgs...)
	return events, total, err
}

func (r *Repository) GetEventsByPersonMonth(personID int64, yearMonth string) ([]AttendanceEvent, error) {
	var events []AttendanceEvent
	err := r.db.Select(&events,
		"SELECT * FROM attendance_events WHERE person_id = ? AND event_date LIKE ? ORDER BY event_date",
		personID, yearMonth+"%")
	return events, err
}

func (r *Repository) DeleteEvent(id int64) error {
	_, err := r.db.Exec("DELETE FROM attendance_events WHERE id = ?", id)
	return err
}

func (r *Repository) GetAnnualLeaveBalance(personID int64) (*MonthlyLeaveBalance, error) {
	var balance MonthlyLeaveBalance

	err := r.db.Get(&balance.TotalDays,
		"SELECT COALESCE(SUM(days_granted), 0) FROM annual_leave_grants WHERE person_id = ?", personID)
	if err != nil {
		return nil, err
	}

	var used float64
	err = r.db.Get(&used,
		"SELECT COALESCE(SUM(duration_hours), 0) FROM attendance_events WHERE person_id = ? AND event_type = '年假'", personID)
	if err != nil {
		return nil, err
	}

	balance.UsedDays = used / 8.0
	balance.Remaining = balance.TotalDays - balance.UsedDays
	return &balance, nil
}

func (r *Repository) GrantAnnualLeave(grant *AnnualLeaveGrant) error {
	_, err := r.db.Exec(`INSERT INTO annual_leave_grants (person_id, grant_date, days_granted, days_remaining, year_month)
		VALUES (?, ?, ?, ?, ?)`,
		grant.PersonID, grant.GrantDate, grant.DaysGranted, grant.DaysGranted, grant.YearMonth)
	return err
}

func (r *Repository) GetAllActiveGrants(personID int64) ([]AnnualLeaveGrant, error) {
	var grants []AnnualLeaveGrant
	err := r.db.Select(&grants,
		"SELECT * FROM annual_leave_grants WHERE person_id = ? AND days_remaining > 0 ORDER BY grant_date",
		personID)
	return grants, err
}

func (r *Repository) UpdateGrantRemaining(id int64, daysRemaining float64) error {
	_, err := r.db.Exec("UPDATE annual_leave_grants SET days_remaining = ? WHERE id = ?", daysRemaining, id)
	return err
}

func (r *Repository) CloseMonthGrants(personID int64, yearMonth string, remark string) (float64, error) {
	grants, err := r.GetAllActiveGrants(personID)
	if err != nil {
		return 0, err
	}

	totalRemaining := 0.0
	for _, g := range grants {
		if g.DaysRemaining > 0 {
			totalRemaining += g.DaysRemaining
			if err := r.UpdateGrantRemaining(g.ID, 0); err != nil {
				return 0, err
			}
		}
	}

	return totalRemaining, nil
}
