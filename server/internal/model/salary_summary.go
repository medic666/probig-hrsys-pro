package model

import "probig/server/internal/utils"

type SalarySummary struct {
	ID                            uint           `gorm:"primarykey" json:"id"`
	PersonID                      uint           `gorm:"not null;index" json:"person_id"`
	BelongMonth                   string         `gorm:"type:varchar(7);not null;uniqueIndex:idx_person_month" json:"belong_month"`
	SalaryDays                    int            `gorm:"default:0" json:"salary_days"`
	WeightedBaseSalary            float64        `gorm:"type:decimal(10,2);default:0" json:"weighted_base_salary"`
	TotalWorkHours                float64        `gorm:"type:decimal(5,1);default:0" json:"total_work_hours"`
	TotalOvertimeWorkdayHours     float64        `gorm:"type:decimal(5,1);default:0" json:"total_overtime_workday_hours"`
	TotalOvertimeHolidayHours     float64        `gorm:"type:decimal(5,1);default:0" json:"total_overtime_holiday_hours"`
	AttendanceSalary              float64        `gorm:"type:decimal(10,2);default:0" json:"attendance_salary"`
	OvertimeWorkdaySalary         float64        `gorm:"type:decimal(10,2);default:0" json:"overtime_workday_salary"`
	OvertimeHolidaySalary         float64        `gorm:"type:decimal(10,2);default:0" json:"overtime_holiday_salary"`
	AnnualLeaveCarryoverDeduct    float64        `gorm:"type:decimal(5,1);default:0" json:"annual_leave_carryover_deduct"`
	AnnualLeaveCarryoverSalary    float64        `gorm:"type:decimal(10,2);default:0" json:"annual_leave_carryover_salary"`
	AttendanceBonus               float64        `gorm:"type:decimal(10,2);default:0" json:"attendance_bonus"`
	PerformanceSalary             float64        `gorm:"type:decimal(10,2);default:0" json:"performance_salary"`
	PostAllowance                 float64        `gorm:"type:decimal(10,2);default:0" json:"post_allowance"`
	MealAllowance                 float64        `gorm:"type:decimal(10,2);default:0" json:"meal_allowance"`
	HousingAllowance              float64        `gorm:"type:decimal(10,2);default:0" json:"housing_allowance"`
	TransportAllowance            float64        `gorm:"type:decimal(10,2);default:0" json:"transport_allowance"`
	HighTempAllowance             float64        `gorm:"type:decimal(10,2);default:0" json:"high_temp_allowance"`
	InsuranceCompensation         float64        `gorm:"type:decimal(10,2);default:0" json:"insurance_compensation"`
	FundCompensation              float64        `gorm:"type:decimal(10,2);default:0" json:"fund_compensation"`
	SalesCommission               float64        `gorm:"type:decimal(10,2);default:0" json:"sales_commission"`
	RewardPunishment              float64        `gorm:"type:decimal(10,2);default:0" json:"reward_punishment"`
	BorrowingRepayment            float64        `gorm:"type:decimal(10,2);default:0" json:"borrowing_repayment"`
	SocialSecurityDeduct          float64        `gorm:"type:decimal(10,2);default:0" json:"social_security_deduct"`
	HousingFundDeduct             float64        `gorm:"type:decimal(10,2);default:0" json:"housing_fund_deduct"`
	TaxDeduct                     float64        `gorm:"type:decimal(10,2);default:0" json:"tax_deduct"`
	FinalSalary                   float64        `gorm:"type:decimal(10,2);default:0" json:"final_salary"`
	LastCalcAt                    utils.DateOnly `json:"last_calc_at"`
	CreatedAt                     utils.DateOnly `json:"created_at"`
	UpdatedAt                     utils.DateOnly `json:"updated_at"`
}

func (SalarySummary) TableName() string { return "salary_summaries" }
