package salary

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

func (r *Repository) UpsertSalary(record *SalaryRecord) error {
	_, err := r.db.Exec(`INSERT INTO salary_records (person_id, year_month, base_salary, attendance_salary, performance_salary, total_allowances, total_deductions, net_salary, detail)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(person_id, year_month) DO UPDATE SET
		base_salary=excluded.base_salary, attendance_salary=excluded.attendance_salary,
		performance_salary=excluded.performance_salary, total_allowances=excluded.total_allowances,
		total_deductions=excluded.total_deductions, net_salary=excluded.net_salary,
		detail=excluded.detail, updated_at=CURRENT_TIMESTAMP`,
		record.PersonID, record.YearMonth, record.BaseSalary, record.AttendanceSalary,
		record.PerformanceSalary, record.TotalAllowances, record.TotalDeductions,
		record.NetSalary, record.Detail)
	return err
}

func (r *Repository) ListRecords(personID int64, yearMonth string, page, pageSize int) ([]SalaryRecord, int64, error) {
	where := "1=1"
	args := []interface{}{}

	if personID > 0 {
		where += " AND person_id = ?"
		args = append(args, personID)
	}
	if yearMonth != "" {
		where += " AND year_month = ?"
		args = append(args, yearMonth)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM salary_records WHERE %s", where)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM salary_records WHERE %s ORDER BY year_month DESC LIMIT ? OFFSET ?", where)
	queryArgs := append(args, pageSize, offset)

	var records []SalaryRecord
	err := r.db.Select(&records, query, queryArgs...)
	return records, total, err
}

func (r *Repository) GetRecord(personID int64, yearMonth string) (*SalaryRecord, error) {
	var record SalaryRecord
	err := r.db.Get(&record, "SELECT * FROM salary_records WHERE person_id = ? AND year_month = ?", personID, yearMonth)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) CreateAdjustment(adj *SalaryAdjustment) error {
	result, err := r.db.Exec(`INSERT INTO salary_adjustments (person_id, year_month, adjustment_type, amount, description, operator_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		adj.PersonID, adj.YearMonth, adj.AdjustmentType, adj.Amount, adj.Description, adj.OperatorID)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	adj.ID = id
	return nil
}

func (r *Repository) GetAdjustments(personID int64, yearMonth string) ([]SalaryAdjustment, error) {
	var adjs []SalaryAdjustment
	err := r.db.Select(&adjs,
		"SELECT * FROM salary_adjustments WHERE person_id = ? AND year_month = ? ORDER BY created_at",
		personID, yearMonth)
	return adjs, err
}

func (r *Repository) DeleteAdjustment(id int64) error {
	_, err := r.db.Exec("DELETE FROM salary_adjustments WHERE id = ?", id)
	return err
}

func (r *Repository) ListByMonth(yearMonth string) ([]SalaryRecord, error) {
	var records []SalaryRecord
	err := r.db.Select(&records, "SELECT * FROM salary_records WHERE year_month = ?", yearMonth)
	return records, err
}
