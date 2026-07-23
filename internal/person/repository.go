package person

import (
	"fmt"
	"strings"

	"probig/internal/common"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository() *Repository {
	return &Repository{db: common.DB}
}

func (r *Repository) Create(person *Person) error {
	_, err := r.db.Exec(`INSERT INTO persons (
		name, attendance_group, hire_date, base_salary, performance_salary, salary_days,
		position_allowance, meal_subsidy, housing_subsidy, transport_subsidy, heat_subsidy,
		insurance_subsidy, housing_fund_subsidy, social_insurance_deduct, housing_fund_deduct, tax_deduct,
		phones, emails, id_number, gender, birthday, ethnicity, native_place, address,
		bank_cards, political_status, marital_status, alias, resume
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		person.Name, person.AttendanceGroup, person.HireDate, person.BaseSalary, person.PerformanceSalary, person.SalaryDays,
		person.PositionAllowance, person.MealSubsidy, person.HousingSubsidy, person.TransportSubsidy, person.HeatSubsidy,
		person.InsuranceSubsidy, person.HousingFundSubsidy, person.SocialInsuranceDeduct, person.HousingFundDeduct, person.TaxDeduct,
		person.Phones, person.Emails, person.IDNumber, person.Gender, person.Birthday, person.Ethnicity, person.NativePlace, person.Address,
		person.BankCards, person.PoliticalStatus, person.MaritalStatus, person.Alias, person.Resume,
	)
	return err
}

func (r *Repository) CreateTx(tx *sqlx.Tx, person *Person) error {
	result, err := tx.Exec(`INSERT INTO persons (
		name, attendance_group, hire_date, base_salary, performance_salary, salary_days,
		position_allowance, meal_subsidy, housing_subsidy, transport_subsidy, heat_subsidy,
		insurance_subsidy, housing_fund_subsidy, social_insurance_deduct, housing_fund_deduct, tax_deduct,
		phones, emails, id_number, gender, birthday, ethnicity, native_place, address,
		bank_cards, political_status, marital_status, alias, resume
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		person.Name, person.AttendanceGroup, person.HireDate, person.BaseSalary, person.PerformanceSalary, person.SalaryDays,
		person.PositionAllowance, person.MealSubsidy, person.HousingSubsidy, person.TransportSubsidy, person.HeatSubsidy,
		person.InsuranceSubsidy, person.HousingFundSubsidy, person.SocialInsuranceDeduct, person.HousingFundDeduct, person.TaxDeduct,
		person.Phones, person.Emails, person.IDNumber, person.Gender, person.Birthday, person.Ethnicity, person.NativePlace, person.Address,
		person.BankCards, person.PoliticalStatus, person.MaritalStatus, person.Alias, person.Resume,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	person.ID = id
	return nil
}

func (r *Repository) Update(person *Person) error {
	_, err := r.db.Exec(`UPDATE persons SET
		name=?, attendance_group=?, hire_date=?, base_salary=?, performance_salary=?, salary_days=?,
		position_allowance=?, meal_subsidy=?, housing_subsidy=?, transport_subsidy=?, heat_subsidy=?,
		insurance_subsidy=?, housing_fund_subsidy=?, social_insurance_deduct=?, housing_fund_deduct=?, tax_deduct=?,
		phones=?, emails=?, id_number=?, gender=?, birthday=?, ethnicity=?, native_place=?, address=?,
		bank_cards=?, political_status=?, marital_status=?, alias=?, resume=?,
		updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		person.Name, person.AttendanceGroup, person.HireDate, person.BaseSalary, person.PerformanceSalary, person.SalaryDays,
		person.PositionAllowance, person.MealSubsidy, person.HousingSubsidy, person.TransportSubsidy, person.HeatSubsidy,
		person.InsuranceSubsidy, person.HousingFundSubsidy, person.SocialInsuranceDeduct, person.HousingFundDeduct, person.TaxDeduct,
		person.Phones, person.Emails, person.IDNumber, person.Gender, person.Birthday, person.Ethnicity, person.NativePlace, person.Address,
		person.BankCards, person.PoliticalStatus, person.MaritalStatus, person.Alias, person.Resume,
		person.ID,
	)
	return err
}

func (r *Repository) UpdateTx(tx *sqlx.Tx, person *Person) error {
	_, err := tx.Exec(`UPDATE persons SET
		name=?, attendance_group=?, hire_date=?, base_salary=?, performance_salary=?, salary_days=?,
		position_allowance=?, meal_subsidy=?, housing_subsidy=?, transport_subsidy=?, heat_subsidy=?,
		insurance_subsidy=?, housing_fund_subsidy=?, social_insurance_deduct=?, housing_fund_deduct=?, tax_deduct=?,
		phones=?, emails=?, id_number=?, gender=?, birthday=?, ethnicity=?, native_place=?, address=?,
		bank_cards=?, political_status=?, marital_status=?, alias=?, resume=?,
		updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		person.Name, person.AttendanceGroup, person.HireDate, person.BaseSalary, person.PerformanceSalary, person.SalaryDays,
		person.PositionAllowance, person.MealSubsidy, person.HousingSubsidy, person.TransportSubsidy, person.HeatSubsidy,
		person.InsuranceSubsidy, person.HousingFundSubsidy, person.SocialInsuranceDeduct, person.HousingFundDeduct, person.TaxDeduct,
		person.Phones, person.Emails, person.IDNumber, person.Gender, person.Birthday, person.Ethnicity, person.NativePlace, person.Address,
		person.BankCards, person.PoliticalStatus, person.MaritalStatus, person.Alias, person.Resume,
		person.ID,
	)
	return err
}

func (r *Repository) GetByID(id int64) (*Person, error) {
	var p Person
	err := r.db.Get(&p, "SELECT * FROM persons WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) List(search string, page, pageSize int) ([]Person, int64, error) {
	where := "1=1"
	args := []interface{}{}

	if search != "" {
		where = "(name LIKE ? OR id_number LIKE ? OR phones LIKE ?)"
		s := "%" + search + "%"
		args = append(args, s, s, s)
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM persons WHERE %s", where)
	if err := r.db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM persons WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", where)
	queryArgs := append(args, pageSize, offset)

	var persons []Person
	err := r.db.Select(&persons, query, queryArgs...)
	return persons, total, err
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM persons WHERE id = ?", id)
	return err
}

func (r *Repository) DeleteTx(tx *sqlx.Tx, id int64) error {
	_, err := tx.Exec("DELETE FROM persons WHERE id = ?", id)
	return err
}

func (r *Repository) All() ([]Person, error) {
	var persons []Person
	err := r.db.Select(&persons, "SELECT * FROM persons ORDER BY id")
	return persons, err
}

func (r *Repository) BuildSetClause(person *Person) (string, []interface{}) {
	var sets []string
	var args []interface{}

	add := func(col string, val interface{}) {
		sets = append(sets, fmt.Sprintf("%s=?", col))
		args = append(args, val)
	}

	if person.Name != "" {
		add("name", person.Name)
	}
	add("attendance_group", person.AttendanceGroup)
	add("hire_date", person.HireDate)
	add("base_salary", person.BaseSalary)
	add("performance_salary", person.PerformanceSalary)
	add("salary_days", person.SalaryDays)
	add("position_allowance", person.PositionAllowance)
	add("meal_subsidy", person.MealSubsidy)
	add("housing_subsidy", person.HousingSubsidy)
	add("transport_subsidy", person.TransportSubsidy)
	add("heat_subsidy", person.HeatSubsidy)
	add("insurance_subsidy", person.InsuranceSubsidy)
	add("housing_fund_subsidy", person.HousingFundSubsidy)
	add("social_insurance_deduct", person.SocialInsuranceDeduct)
	add("housing_fund_deduct", person.HousingFundDeduct)
	add("tax_deduct", person.TaxDeduct)
	add("phones", person.Phones)
	add("emails", person.Emails)
	add("id_number", person.IDNumber)
	add("gender", person.Gender)
	add("birthday", person.Birthday)
	add("ethnicity", person.Ethnicity)
	add("native_place", person.NativePlace)
	add("address", person.Address)
	add("bank_cards", person.BankCards)
	add("political_status", person.PoliticalStatus)
	add("marital_status", person.MaritalStatus)
	add("alias", person.Alias)
	add("resume", person.Resume)

	sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
	return strings.Join(sets, ", "), args
}
