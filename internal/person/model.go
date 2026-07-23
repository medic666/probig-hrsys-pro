package person

import "time"

type Person struct {
	ID                   int64     `json:"id" db:"id"`
	Name                 string    `json:"name" db:"name"`
	AttendanceGroup      string    `json:"attendanceGroup" db:"attendance_group"`
	HireDate             string    `json:"hireDate" db:"hire_date"`
	BaseSalary           float64   `json:"baseSalary" db:"base_salary"`
	PerformanceSalary    float64   `json:"performanceSalary" db:"performance_salary"`
	SalaryDays           float64   `json:"salaryDays" db:"salary_days"`
	PositionAllowance    float64   `json:"positionAllowance" db:"position_allowance"`
	MealSubsidy          float64   `json:"mealSubsidy" db:"meal_subsidy"`
	HousingSubsidy       float64   `json:"housingSubsidy" db:"housing_subsidy"`
	TransportSubsidy     float64   `json:"transportSubsidy" db:"transport_subsidy"`
	HeatSubsidy          float64   `json:"heatSubsidy" db:"heat_subsidy"`
	InsuranceSubsidy     float64   `json:"insuranceSubsidy" db:"insurance_subsidy"`
	HousingFundSubsidy   float64   `json:"housingFundSubsidy" db:"housing_fund_subsidy"`
	SocialInsuranceDeduct float64  `json:"socialInsuranceDeduct" db:"social_insurance_deduct"`
	HousingFundDeduct    float64   `json:"housingFundDeduct" db:"housing_fund_deduct"`
	TaxDeduct            float64   `json:"taxDeduct" db:"tax_deduct"`
	Phones               string    `json:"phones" db:"phones"`
	Emails               string    `json:"emails" db:"emails"`
	IDNumber             string    `json:"idNumber" db:"id_number"`
	Gender               string    `json:"gender" db:"gender"`
	Birthday             string    `json:"birthday" db:"birthday"`
	Ethnicity            string    `json:"ethnicity" db:"ethnicity"`
	NativePlace          string    `json:"nativePlace" db:"native_place"`
	Address              string    `json:"address" db:"address"`
	BankCards            string    `json:"bankCards" db:"bank_cards"`
	PoliticalStatus      string    `json:"politicalStatus" db:"political_status"`
	MaritalStatus        string    `json:"maritalStatus" db:"marital_status"`
	Alias                string    `json:"alias" db:"alias"`
	Resume               string    `json:"resume" db:"resume"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time `json:"updatedAt" db:"updated_at"`
}

type PhoneEntry struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type EmailEntry struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type BankCardEntry struct {
	BankName string `json:"bankName"`
	Account  string `json:"account"`
}

type ResumeEntry struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Company   string `json:"company"`
	Position  string `json:"position"`
	Description string `json:"description"`
}
