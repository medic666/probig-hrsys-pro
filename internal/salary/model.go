package salary

import (
	"time"

	"probig/internal/pkg/database"
)

type SalaryEvent = database.SalaryEvent
type SalarySummary = database.SalarySummary

type SalaryEventFilter struct {
	PersonID    *uint   `form:"person_id"`
	BelongMonth string  `form:"belong_month"`
	EventType   string  `form:"event_type"`
	PersonName  string  `form:"person_name"`
	PageNum     int     `form:"page_num"`
	PageSize    int     `form:"page_size"`
}

type SalaryEventCreateRequest struct {
	PersonID    uint    `json:"person_id" binding:"required"`
	BelongMonth string  `json:"belong_month" binding:"required"`
	EventType   string  `json:"event_type" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	EventName   string  `json:"event_name" binding:"required"`
	Remark      string  `json:"remark"`
}

type SalaryEventUpdateRequest struct {
	PersonID    *uint    `json:"person_id"`
	BelongMonth *string  `json:"belong_month"`
	EventType   *string  `json:"event_type"`
	Amount      *float64 `json:"amount"`
	EventName   *string  `json:"event_name"`
	Remark      *string  `json:"remark"`
}

type SalarySummaryFilter struct {
	PersonID    *uint  `form:"person_id"`
	BelongMonth string `form:"belong_month"`
	PersonName  string `form:"person_name"`
	AttendanceGroup string `form:"attendance_group"`
	PageNum     int    `form:"page_num"`
	PageSize    int    `form:"page_size"`
}

type CalcRequest struct {
	BelongMonth string `json:"belong_month" binding:"required"`
	PersonIDs   []uint `json:"person_ids"`
}

type SalarySummaryVO struct {
	ID                          uint       `json:"id"`
	PersonID                    uint       `json:"person_id"`
	PersonName                  string     `json:"person_name"`
	BelongMonth                 string     `json:"belong_month"`
	SalaryDays                  int        `json:"salary_days"`
	WeightedBaseSalary          float64    `json:"weighted_base_salary"`
	TotalWorkHours              float64    `json:"total_work_hours"`
	TotalOvertimeWorkdayHours   float64    `json:"total_overtime_workday_hours"`
	TotalOvertimeHolidayHours   float64    `json:"total_overtime_holiday_hours"`
	AttendanceSalary            float64    `json:"attendance_salary"`
	OvertimeWorkdaySalary       float64    `json:"overtime_workday_salary"`
	OvertimeHolidaySalary       float64    `json:"overtime_holiday_salary"`
	AnnualLeaveCarryoverSalary  float64    `json:"annual_leave_carryover_salary"`
	AttendanceBonus             float64    `json:"attendance_bonus"`
	PerformanceSalary           float64    `json:"performance_salary"`
	PostAllowance               float64    `json:"post_allowance"`
	MealAllowance               float64    `json:"meal_allowance"`
	HousingAllowance            float64    `json:"housing_allowance"`
	TransportAllowance          float64    `json:"transport_allowance"`
	HighTempAllowance           float64    `json:"high_temp_allowance"`
	InsuranceCompensation       float64    `json:"insurance_compensation"`
	FundCompensation            float64    `json:"fund_compensation"`
	TotalAdjustment             float64    `json:"total_adjustment"`
	SocialSecurityDeduct        float64    `json:"social_security_deduct"`
	HousingFundDeduct           float64    `json:"housing_fund_deduct"`
	TaxDeduct                   float64    `json:"tax_deduct"`
	FinalSalary                 float64    `json:"final_salary"`
	Status                      string     `json:"status"`
	LastCalcAt                  *time.Time `json:"last_calc_at"`
	CreatedAt                   time.Time  `json:"created_at"`
	UpdatedAt                   time.Time  `json:"updated_at"`
}
