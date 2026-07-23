package salary

import "time"

type SalaryRecord struct {
	ID                 int64     `json:"id" db:"id"`
	PersonID           int64     `json:"personId" db:"person_id"`
	YearMonth          string    `json:"yearMonth" db:"year_month"`
	BaseSalary         float64   `json:"baseSalary" db:"base_salary"`
	AttendanceSalary   float64   `json:"attendanceSalary" db:"attendance_salary"`
	PerformanceSalary  float64   `json:"performanceSalary" db:"performance_salary"`
	TotalAllowances    float64   `json:"totalAllowances" db:"total_allowances"`
	TotalDeductions    float64   `json:"totalDeductions" db:"total_deductions"`
	NetSalary          float64   `json:"netSalary" db:"net_salary"`
	Detail             string    `json:"detail" db:"detail"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}

type SalaryAdjustment struct {
	ID             int64     `json:"id" db:"id"`
	PersonID       int64     `json:"personId" db:"person_id"`
	YearMonth      string    `json:"yearMonth" db:"year_month"`
	AdjustmentType string    `json:"adjustmentType" db:"adjustment_type"`
	Amount         float64   `json:"amount" db:"amount"`
	Description    string    `json:"description" db:"description"`
	OperatorID     int64     `json:"operatorId" db:"operator_id"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}

type SalaryDetail struct {
	BaseSalary        float64               `json:"baseSalary"`
	AttendanceDays    int                   `json:"attendanceDays"`
	LeaveDeductions   map[string]float64    `json:"leaveDeductions"`
	AttendanceSalary  float64               `json:"attendanceSalary"`
	PerformanceSalary float64               `json:"performanceSalary"`
	Allowances        map[string]float64    `json:"allowances"`
	Deductions        map[string]float64    `json:"deductions"`
	Adjustments       []SalaryAdjustment    `json:"adjustments"`
	NetSalary         float64               `json:"netSalary"`
}
