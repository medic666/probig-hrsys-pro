package salary

import (
	"encoding/json"
	"fmt"
	"probig/internal/common"
	"probig/internal/event"
)

type Service struct {
	repo         *Repository
	eventService *event.Service
}

func NewService(repo *Repository, eventService *event.Service) *Service {
	return &Service{repo: repo, eventService: eventService}
}

func (s *Service) CalculateSalary(personID int64, yearMonth string, operatorID int64) (*SalaryRecord, error) {
	var person struct {
		BaseSalary        float64 `db:"base_salary"`
		PerformanceSalary float64 `db:"performance_salary"`
		SalaryDays        float64 `db:"salary_days"`
		PositionAllowance float64 `db:"position_allowance"`
		MealSubsidy       float64 `db:"meal_subsidy"`
		HousingSubsidy    float64 `db:"housing_subsidy"`
		TransportSubsidy  float64 `db:"transport_subsidy"`
		HeatSubsidy       float64 `db:"heat_subsidy"`
		InsuranceSubsidy  float64 `db:"insurance_subsidy"`
		HousingFundSubsidy float64 `db:"housing_fund_subsidy"`
		SocialInsuranceDeduct float64 `db:"social_insurance_deduct"`
		HousingFundDeduct float64 `db:"housing_fund_deduct"`
		TaxDeduct         float64 `db:"tax_deduct"`
	}

	err := common.DB.Get(&person, `SELECT base_salary, performance_salary, salary_days, position_allowance,
		meal_subsidy, housing_subsidy, transport_subsidy, heat_subsidy, insurance_subsidy,
		housing_fund_subsidy, social_insurance_deduct, housing_fund_deduct, tax_deduct
		FROM persons WHERE id = ?`, personID)
	if err != nil {
		return nil, fmt.Errorf("人员不存在: %w", err)
	}

	dayRate := person.BaseSalary / person.SalaryDays

	var attendanceEvents []struct {
		EventType     string  `db:"event_type"`
		DurationHours float64 `db:"duration_hours"`
	}
	err = common.DB.Select(&attendanceEvents,
		"SELECT event_type, duration_hours FROM attendance_events WHERE person_id = ? AND event_date LIKE ?",
		personID, yearMonth+"%")
	if err != nil {
		return nil, err
	}

	leaveDeductions := make(map[string]float64)

	for _, e := range attendanceEvents {
		days := e.DurationHours / 8.0
		switch e.EventType {
		case "事假":
			leaveDeductions["事假"] += days * dayRate
		case "病假":
			leaveDeductions["病假"] += days * dayRate * 0.3
		case "迟到":
			leaveDeductions["迟到"] += days * dayRate * 0.25
		case "早退":
			leaveDeductions["早退"] += days * dayRate * 0.25
		case "缺卡":
			leaveDeductions["缺卡"] += days * dayRate * 0.5
		}
	}

	totalLeaveDeduction := 0.0
	for _, v := range leaveDeductions {
		totalLeaveDeduction += v
	}

	attendanceSalary := person.BaseSalary - totalLeaveDeduction
	if attendanceSalary < 0 {
		attendanceSalary = 0
	}

	workDays := person.SalaryDays

	adjustments, err := s.repo.GetAdjustments(personID, yearMonth)
	if err != nil {
		return nil, err
	}

	adjustmentTotal := 0.0
	for _, a := range adjustments {
		adjustmentTotal += a.Amount
	}

	allowances := map[string]float64{
		"职位津贴": person.PositionAllowance,
		"餐补":   person.MealSubsidy,
		"房补":   person.HousingSubsidy,
		"交通补贴": person.TransportSubsidy,
		"高温补贴": person.HeatSubsidy,
		"保险补贴": person.InsuranceSubsidy,
		"公积金补偿": person.HousingFundSubsidy,
	}

	totalAllowances := 0.0
	for _, v := range allowances {
		totalAllowances += v
	}

	deductions := map[string]float64{
		"社保代扣": person.SocialInsuranceDeduct,
		"公积金代扣": person.HousingFundDeduct,
		"个税代扣": person.TaxDeduct,
	}

	totalDeductions := 0.0
	for _, v := range deductions {
		totalDeductions += v
	}

	netSalary := attendanceSalary + person.PerformanceSalary + totalAllowances + adjustmentTotal - totalDeductions
	if netSalary < 0 {
		netSalary = 0
	}

	detail := SalaryDetail{
		BaseSalary:        person.BaseSalary,
		AttendanceDays:    int(workDays),
		LeaveDeductions:   leaveDeductions,
		AttendanceSalary:  attendanceSalary,
		PerformanceSalary: person.PerformanceSalary,
		Allowances:        allowances,
		Deductions:        deductions,
		Adjustments:       adjustments,
		NetSalary:          netSalary,
	}

	detailJSON, _ := json.Marshal(detail)

	record := &SalaryRecord{
		PersonID:          personID,
		YearMonth:         yearMonth,
		BaseSalary:        person.BaseSalary,
		AttendanceSalary:  attendanceSalary,
		PerformanceSalary: person.PerformanceSalary,
		TotalAllowances:   totalAllowances,
		TotalDeductions:   totalDeductions,
		NetSalary:         netSalary,
		Detail:            string(detailJSON),
	}

	if err := s.repo.UpsertSalary(record); err != nil {
		return nil, err
	}

	payload, _ := json.Marshal(record)
	s.eventService.RecordEvent("salary", record.ID, "calculate", string(payload), operatorID, "工资计算")

	return record, nil
}

func (s *Service) AddAdjustment(adj *SalaryAdjustment, operatorID int64, remark string) error {
	adj.OperatorID = operatorID

	if err := s.repo.CreateAdjustment(adj); err != nil {
		return err
	}

	payload, _ := json.Marshal(adj)
	return s.eventService.RecordEvent("salary_adjustment", adj.ID, "create", string(payload), operatorID, remark)
}

func (s *Service) ListRecords(personID int64, yearMonth string, page, pageSize int) ([]SalaryRecord, int64, error) {
	return s.repo.ListRecords(personID, yearMonth, page, pageSize)
}

func (s *Service) GetRecord(personID int64, yearMonth string) (*SalaryRecord, error) {
	return s.repo.GetRecord(personID, yearMonth)
}

func (s *Service) GetAdjustments(personID int64, yearMonth string) ([]SalaryAdjustment, error) {
	return s.repo.GetAdjustments(personID, yearMonth)
}

func (s *Service) DeleteAdjustment(id int64, operatorID int64, remark string) error {
	adj, err := s.getAdjustmentByID(id)
	if err != nil {
		return common.ErrNotFound
	}

	if err := s.repo.DeleteAdjustment(id); err != nil {
		return err
	}

	payload, _ := json.Marshal(adj)
	return s.eventService.RecordEvent("salary_adjustment", id, "delete", string(payload), operatorID, remark)
}

func (s *Service) getAdjustmentByID(id int64) (*SalaryAdjustment, error) {
	// simplified: find from db
	var adj SalaryAdjustment
	err := common.DB.Get(&adj, "SELECT * FROM salary_adjustments WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &adj, nil
}

func (s *Service) ListByMonth(yearMonth string) ([]SalaryRecord, error) {
	return s.repo.ListByMonth(yearMonth)
}
