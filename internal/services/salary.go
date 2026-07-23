package services

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/models"
)

type SalaryService struct {
	db    *sqlx.DB
	audit *AuditService
}

func NewSalaryService(db *sqlx.DB, audit *AuditService) *SalaryService {
	return &SalaryService{db: db, audit: audit}
}

func (s *SalaryService) ListEvents(page, pageSize int, personID uint, period, eventType string) ([]models.SalaryEvent, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM salary_events se WHERE 1=1"
	args := []interface{}{}
	if personID > 0 {
		countSQL += " AND se.person_id = ?"
		args = append(args, personID)
	}
	if period != "" {
		countSQL += " AND se.period = ?"
		args = append(args, period)
	}
	if eventType != "" {
		countSQL += " AND se.event_type = ?"
		args = append(args, eventType)
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := "SELECT se.*, e.name as person_name FROM salary_events se LEFT JOIN entities e ON e.id = se.person_id WHERE 1=1"
	if personID > 0 {
		querySQL += " AND se.person_id = ?"
	}
	if period != "" {
		querySQL += " AND se.period = ?"
	}
	if eventType != "" {
		querySQL += " AND se.event_type = ?"
	}
	querySQL += " ORDER BY se.period DESC, se.id DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	var events []models.SalaryEvent
	if err := s.db.Select(&events, querySQL, queryArgs...); err != nil {
		return nil, 0, err
	}
	if events == nil {
		events = []models.SalaryEvent{}
	}
	return events, total, nil
}

func (s *SalaryService) CreateEvent(req models.SalaryEventRequest, userID uint, ip string) (*models.SalaryEvent, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO salary_events (person_id, period, event_type, amount, detail, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.PersonID, req.Period, req.EventType, req.Amount, req.Detail, userID, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create salary event: %w", err)
	}
	eventID, _ := result.LastInsertId()

	s.audit.Log(tx, userID, "create", "salary_event", ptrUint(uint(eventID)), req, ip)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.SalaryEvent{
		ID:        uint(eventID),
		PersonID:  req.PersonID,
		Period:    req.Period,
		EventType: req.EventType,
		Amount:    req.Amount,
		Detail:    req.Detail,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *SalaryService) UpdateEvent(eventID uint, req models.SalaryEventRequest, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE salary_events SET person_id = ?, period = ?, event_type = ?, amount = ?, detail = ?, updated_at = ? WHERE id = ?`,
		req.PersonID, req.Period, req.EventType, req.Amount, req.Detail, time.Now(), eventID,
	)
	if err != nil {
		return err
	}

	s.audit.Log(tx, userID, "update", "salary_event", ptrUint(eventID), req, ip)
	return tx.Commit()
}

func (s *SalaryService) DeleteEvent(eventID uint, userID uint, ip string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var evt models.SalaryEvent
	if err := tx.Get(&evt, "SELECT * FROM salary_events WHERE id = ?", eventID); err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM salary_events WHERE id = ?", eventID)
	if err != nil {
		return err
	}

	s.audit.Log(tx, userID, "delete", "salary_event", ptrUint(eventID), evt, ip)
	return tx.Commit()
}

func (s *SalaryService) ListSummaries(page, pageSize int, personID uint, period string) ([]models.SalarySummary, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM salary_summaries ss WHERE 1=1"
	args := []interface{}{}
	if personID > 0 {
		countSQL += " AND ss.person_id = ?"
		args = append(args, personID)
	}
	if period != "" {
		countSQL += " AND ss.period = ?"
		args = append(args, period)
	}
	if err := s.db.Get(&total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	querySQL := "SELECT ss.*, e.name as person_name FROM salary_summaries ss LEFT JOIN entities e ON e.id = ss.person_id WHERE 1=1"
	if personID > 0 {
		querySQL += " AND ss.person_id = ?"
	}
	if period != "" {
		querySQL += " AND ss.period = ?"
	}
	querySQL += " ORDER BY ss.period DESC, ss.person_id ASC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	var summaries []models.SalarySummary
	if err := s.db.Select(&summaries, querySQL, queryArgs...); err != nil {
		return nil, 0, err
	}
	if summaries == nil {
		summaries = []models.SalarySummary{}
	}
	return summaries, total, nil
}

func (s *SalaryService) Calculate(req models.CalculateRequest, userID uint, ip string) error {
	period := req.Period

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
		summary, err := s.computeSalaryForPerson(tx, pid, period)
		if err != nil {
			return err
		}

		detailJSON, _ := json.Marshal(summary["detail"])

		totalSalary := summary["attendance_salary"].(float64) +
			summary["full_attendance_bonus"].(float64) +
			summary["overtime_salary"].(float64) +
			summary["performance_salary"].(float64) +
			summary["position_allowance"].(float64) +
			summary["meal_subsidy"].(float64) +
			summary["housing_subsidy"].(float64) +
			summary["transport_subsidy"].(float64) +
			summary["heat_subsidy"].(float64) +
			summary["insurance_compensation"].(float64) +
			summary["housing_fund_compensation"].(float64) -
			summary["social_insurance_deduct"].(float64) -
			summary["housing_fund_deduct"].(float64) -
			summary["tax_deduct"].(float64) -
			summary["loan_deduct"].(float64) +
			summary["reward_punish"].(float64) +
			summary["other_adjustments"].(float64)

		totalSalary = math.Round(totalSalary*100) / 100

		_, err = tx.Exec(
			`INSERT OR REPLACE INTO salary_summaries
			 (person_id, period, attendance_salary, full_attendance_bonus, overtime_salary,
			  performance_salary, position_allowance, meal_subsidy, housing_subsidy,
			  transport_subsidy, heat_subsidy, insurance_compensation, housing_fund_compensation,
			  social_insurance_deduct, housing_fund_deduct, tax_deduct, loan_deduct,
			  reward_punish, other_adjustments, total_salary, detail_data, calculated_at)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			pid, period,
			summary["attendance_salary"], summary["full_attendance_bonus"], summary["overtime_salary"],
			summary["performance_salary"], summary["position_allowance"], summary["meal_subsidy"],
			summary["housing_subsidy"], summary["transport_subsidy"], summary["heat_subsidy"],
			summary["insurance_compensation"], summary["housing_fund_compensation"],
			summary["social_insurance_deduct"], summary["housing_fund_deduct"],
			summary["tax_deduct"], summary["loan_deduct"],
			summary["reward_punish"], summary["other_adjustments"],
			totalSalary, string(detailJSON), time.Now(),
		)
		if err != nil {
			return err
		}
	}

	s.audit.Log(tx, userID, "calculate", "salary_summary", nil, req, ip)
	return tx.Commit()
}

func (s *SalaryService) computeSalaryForPerson(tx *sqlx.Tx, personID uint, period string) (map[string]interface{}, error) {
	summary := map[string]interface{}{
		"attendance_salary":        0.0,
		"full_attendance_bonus":    0.0,
		"overtime_salary":          0.0,
		"performance_salary":       0.0,
		"position_allowance":       0.0,
		"meal_subsidy":             0.0,
		"housing_subsidy":          0.0,
		"transport_subsidy":        0.0,
		"heat_subsidy":             0.0,
		"insurance_compensation":   0.0,
		"housing_fund_compensation": 0.0,
		"social_insurance_deduct":  0.0,
		"housing_fund_deduct":      0.0,
		"tax_deduct":               0.0,
		"loan_deduct":              0.0,
		"reward_punish":            0.0,
		"other_adjustments":        0.0,
		"detail":                   make(map[string]interface{}),
	}

	var personData models.PersonSnapshotData
	var snapshot models.PersonSnapshot
	err := tx.Get(&snapshot,
		"SELECT * FROM person_snapshots WHERE person_id = ? ORDER BY effective_date DESC, id DESC LIMIT 1",
		personID,
	)
	if err == nil {
		json.Unmarshal([]byte(snapshot.SnapshotData), &personData)
	} else {
		personData = models.DefaultPersonSnapshotData()
	}

	var attSummary models.AttendanceSummary
	err = tx.Get(&attSummary,
		"SELECT * FROM attendance_summaries WHERE person_id = ? AND period = ?",
		personID, period,
	)
	if err != nil {
		attSummary = models.AttendanceSummary{}
	}

	basicSalary := personData.BasicSalary
	salaryDays := personData.SalaryDays
	mealSubsidy := personData.MealSubsidy
	performanceSalary := personData.PerformanceSalary

	if salaryDays <= 0 {
		salaryDays = 21.75
	}

	dailyRate := basicSalary / salaryDays

	attendanceDays := attSummary.NormalAttendanceDays + attSummary.SupplementaryAttendanceDays
	paidLeaveDays := attSummary.AnnualLeaveDays + attSummary.StatutoryLeaveDays + attSummary.WelfareLeaveDays
	sickLeaveWeighted := attSummary.SickLeaveDays * 0.6

	attendanceSalary := (attendanceDays*1.0 + paidLeaveDays*1.0 + sickLeaveWeighted) * dailyRate
	attendanceSalary = math.Round(attendanceSalary*100) / 100
	summary["attendance_salary"] = attendanceSalary

	if attSummary.PersonalLeaveDays == 0 {
		attendCoeff := attendanceDays*1.0 + paidLeaveDays*1.0 + sickLeaveWeighted - float64(attSummary.ViolationCount)
		if attendCoeff < 0 {
			attendCoeff = 0
		}
		fullBonus := attendCoeff * 10
		fullBonus = math.Round(fullBonus*100) / 100
		summary["full_attendance_bonus"] = fullBonus
	}

	overtimeDailyBase := (basicSalary + mealSubsidy) / salaryDays
	overtimeSalary := overtimeDailyBase*attSummary.WorkdayOvertimeDays*1.5 +
		overtimeDailyBase*(attSummary.HolidayOvertimeDays+attSummary.AnnualLeaveCarryover)*2.0
	overtimeSalary = math.Round(overtimeSalary*100) / 100
	summary["overtime_salary"] = overtimeSalary

	summary["performance_salary"] = performanceSalary
	summary["position_allowance"] = personData.PositionAllowance
	summary["meal_subsidy"] = mealSubsidy
	summary["housing_subsidy"] = personData.HousingSubsidy
	summary["transport_subsidy"] = personData.TransportSubsidy
	summary["heat_subsidy"] = personData.HeatSubsidy
	summary["insurance_compensation"] = personData.InsuranceCompensation
	summary["housing_fund_compensation"] = personData.HousingFundCompensation
	summary["social_insurance_deduct"] = personData.SocialInsuranceDeduct
	summary["housing_fund_deduct"] = personData.HousingFundDeduct

	var salaryEvents []models.SalaryEvent
	if err := tx.Select(&salaryEvents,
		"SELECT * FROM salary_events WHERE person_id = ? AND period = ?",
		personID, period,
	); err == nil {
		for _, evt := range salaryEvents {
			switch evt.EventType {
			case "performance":
				summary["performance_salary"] = evt.Amount
			case "reward_punish":
				summary["reward_punish"] = summary["reward_punish"].(float64) + evt.Amount
			case "loan_deduct":
				summary["loan_deduct"] = summary["loan_deduct"].(float64) + evt.Amount
			case "tax_deduct":
				summary["tax_deduct"] = summary["tax_deduct"].(float64) + evt.Amount
			case "other":
				summary["other_adjustments"] = summary["other_adjustments"].(float64) + evt.Amount
			}
		}
	}

	detail := map[string]interface{}{
		"person_name":          personData.Name,
		"basic_salary":         basicSalary,
		"salary_days":          salaryDays,
		"daily_rate":           dailyRate,
		"attendance_days":      attendanceDays,
		"paid_leave_days":      paidLeaveDays,
		"sick_leave_days":      attSummary.SickLeaveDays,
		"personal_leave_days":  attSummary.PersonalLeaveDays,
		"workday_overtime":     attSummary.WorkdayOvertimeDays,
		"holiday_overtime":     attSummary.HolidayOvertimeDays,
		"violation_count":      attSummary.ViolationCount,
		"attendance_salary":    attendanceSalary,
		"full_attendance_bonus": summary["full_attendance_bonus"],
		"overtime_salary":      overtimeSalary,
		"formula_daily_rate":   fmt.Sprintf("%.2f / %.2f = %.2f", basicSalary, salaryDays, dailyRate),
	}
	summary["detail"] = detail

	return summary, nil
}
