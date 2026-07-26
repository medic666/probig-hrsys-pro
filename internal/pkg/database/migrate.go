package database

import (
	"time"

	"gorm.io/gorm"
)

type Person struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:64;not null" json:"name"`
	IDCard          string         `gorm:"uniqueIndex;size:32" json:"id_card" encrypt:"true"`
	Gender          int8           `json:"gender"`
	Birthday        *time.Time     `json:"birthday"`
	Nation          string         `gorm:"size:32" json:"nation"`
	NativePlace     string         `gorm:"size:128" json:"native_place"`
	Address         string         `gorm:"size:256" json:"address"`
	PoliticalStatus string         `gorm:"size:32" json:"political_status"`
	MaritalStatus   int8           `json:"marital_status"`
	Alias           string         `gorm:"size:64" json:"alias"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type PersonPhone struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"index;not null" json:"person_id"`
	Phone     string         `gorm:"size:32" json:"phone" encrypt:"true"`
	PhoneType string         `gorm:"size:32" json:"phone_type"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type PersonEmail struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"index;not null" json:"person_id"`
	Email     string         `gorm:"size:128" json:"email" encrypt:"true"`
	EmailType string         `gorm:"size:32" json:"email_type"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type PersonBankCard struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"index;not null" json:"person_id"`
	BankName  string         `gorm:"size:64" json:"bank_name"`
	CardNo    string         `gorm:"size:64" json:"card_no" encrypt:"true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Company struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	CreditCode   string         `gorm:"uniqueIndex;size:64" json:"credit_code"`
	Address      string         `gorm:"size:256" json:"address"`
	ContactPhone string         `gorm:"size:32" json:"contact_phone"`
	BankName     string         `gorm:"size:64" json:"bank_name"`
	BankAccount  string         `gorm:"size:64" json:"bank_account" encrypt:"true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type PositionEvent struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"index;not null" json:"person_id"`
	EventName             string         `gorm:"size:128" json:"event_name"`
	EffectiveDate         *time.Time     `json:"effective_date"`
	AttendanceGroup       *string        `gorm:"size:64" json:"attendance_group"`
	HasAnnualLeave        *bool          `json:"has_annual_leave"`
	HasAttendanceBonus    *bool          `json:"has_attendance_bonus"`
	BaseSalary            *float64       `gorm:"type:decimal(10,2)" json:"base_salary"`
	PerformanceSalary     *float64       `gorm:"type:decimal(10,2)" json:"performance_salary"`
	SalaryDays            *int           `json:"salary_days"`
	PostAllowance         *float64       `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance         *float64       `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance      *float64       `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance    *float64       `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance     *float64       `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceCompensation *float64       `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundCompensation      *float64       `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	SocialSecurityDeduct  *float64       `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct     *float64       `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type PositionSnapshot struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	PersonID              uint       `gorm:"index;not null" json:"person_id"`
	EffectiveStartDate    *time.Time `json:"effective_start_date"`
	EffectiveEndDate      *time.Time `json:"effective_end_date"`
	EntryDate             *time.Time `json:"entry_date"`
	LeaveDate             *time.Time `json:"leave_date"`
	AttendanceGroup       string     `gorm:"size:64" json:"attendance_group"`
	HasAnnualLeave        bool       `json:"has_annual_leave"`
	HasAttendanceBonus    bool       `json:"has_attendance_bonus"`
	BaseSalary            float64    `gorm:"type:decimal(10,2)" json:"base_salary"`
	PerformanceSalary     float64    `gorm:"type:decimal(10,2)" json:"performance_salary"`
	SalaryDays            int        `json:"salary_days"`
	PostAllowance         float64    `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance         float64    `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance      float64    `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance    float64    `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance     float64    `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceCompensation float64    `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundCompensation      float64    `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	SocialSecurityDeduct  float64    `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct     float64    `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	LastCalcAt            *time.Time `json:"last_calc_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type AttendanceEvent struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	PersonID           uint           `gorm:"index;not null" json:"person_id"`
	EventDate          *time.Time     `json:"event_date"`
	PunchTime          string         `gorm:"size:32" json:"punch_time"`
	EventType          string         `gorm:"size:32" json:"event_type"`
	SubType            string         `gorm:"size:32" json:"sub_type"`
	Hours              float64        `gorm:"type:decimal(4,1)" json:"hours"`
	LateMinutes        int            `json:"late_minutes"`
	IsSpecialApproval  bool           `json:"is_special_approval"`
	Remark             string         `gorm:"size:256" json:"remark"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type AttendanceDailyProjection struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	PersonID               uint       `gorm:"index;not null" json:"person_id"`
	WorkDate               *time.Time `json:"work_date"`
	PunchTime              string     `gorm:"size:32" json:"punch_time"`
	WorkHours              float64    `gorm:"type:decimal(4,1)" json:"work_hours"`
	OvertimeWorkdayHours   float64    `gorm:"type:decimal(4,1)" json:"overtime_workday_hours"`
	OvertimeHolidayHours   float64    `gorm:"type:decimal(4,1)" json:"overtime_holiday_hours"`
	HasPersonalLeave       bool       `json:"has_personal_leave"`
	ViolationCount         int        `json:"violation_count"`
	Remark                 string     `gorm:"size:256" json:"remark"`
	LastCalcAt             *time.Time `json:"last_calc_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type AttendanceSalaryMonthly struct {
	ID                       uint       `gorm:"primaryKey" json:"id"`
	PersonID                 uint       `gorm:"index;not null" json:"person_id"`
	BelongMonth              string     `gorm:"size:7" json:"belong_month"`
	SalaryDays               int        `json:"salary_days"`
	WeightedBaseSalary       float64    `gorm:"type:decimal(10,2)" json:"weighted_base_salary"`
	WeightedMealAllowance    float64    `gorm:"type:decimal(10,2)" json:"weighted_meal_allowance"`
	TotalWorkHours           float64    `gorm:"type:decimal(5,1)" json:"total_work_hours"`
	TotalOvertimeWorkdayHours float64   `gorm:"type:decimal(5,1)" json:"total_overtime_workday_hours"`
	TotalOvertimeHolidayHours float64   `gorm:"type:decimal(5,1)" json:"total_overtime_holiday_hours"`
	AttendanceSalary         float64    `gorm:"type:decimal(10,2)" json:"attendance_salary"`
	OvertimeWorkdaySalary    float64    `gorm:"type:decimal(10,2)" json:"overtime_workday_salary"`
	OvertimeHolidaySalary    float64    `gorm:"type:decimal(10,2)" json:"overtime_holiday_salary"`
	HasPersonalLeaveMonth    bool       `json:"has_personal_leave_month"`
	TotalViolationCount      int        `json:"total_violation_count"`
	AttendanceBonus          float64    `gorm:"type:decimal(10,2)" json:"attendance_bonus"`
	LastCalcAt               *time.Time `json:"last_calc_at"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type LeaveAccountEvent struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	PersonID      uint           `gorm:"index;not null" json:"person_id"`
	LeaveType     string         `gorm:"size:32" json:"leave_type"`
	EventType     string         `gorm:"size:32" json:"event_type"`
	SourceType    string         `gorm:"size:16" json:"source_type"`
	BatchID       *uint          `json:"batch_id"`
	Hours         float64        `gorm:"type:decimal(5,1)" json:"hours"`
	EffectiveDate *time.Time     `json:"effective_date"`
	Remark        string         `gorm:"size:256" json:"remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type LeaveAccountBalance struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	PersonID     uint       `gorm:"index;not null" json:"person_id"`
	LeaveType    string     `gorm:"size:32" json:"leave_type"`
	BalanceHours float64    `gorm:"type:decimal(5,1)" json:"balance_hours"`
	LastCalcAt   *time.Time `json:"last_calc_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SalaryEvent struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PersonID    uint           `gorm:"index;not null" json:"person_id"`
	BelongMonth string         `gorm:"size:7" json:"belong_month"`
	EventType   string         `gorm:"size:32" json:"event_type"`
	Amount      float64        `gorm:"type:decimal(10,2)" json:"amount"`
	EventName   string         `gorm:"size:128" json:"event_name"`
	Remark      string         `gorm:"size:256" json:"remark"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type SalarySummary struct {
	ID                          uint       `gorm:"primaryKey" json:"id"`
	PersonID                    uint       `gorm:"index;not null" json:"person_id"`
	BelongMonth                 string     `gorm:"size:7" json:"belong_month"`
	SalaryDays                  int        `json:"salary_days"`
	WeightedBaseSalary          float64    `gorm:"type:decimal(10,2)" json:"weighted_base_salary"`
	TotalWorkHours              float64    `gorm:"type:decimal(5,1)" json:"total_work_hours"`
	TotalOvertimeWorkdayHours   float64    `gorm:"type:decimal(5,1)" json:"total_overtime_workday_hours"`
	TotalOvertimeHolidayHours   float64    `gorm:"type:decimal(5,1)" json:"total_overtime_holiday_hours"`
	AttendanceSalary            float64    `gorm:"type:decimal(10,2)" json:"attendance_salary"`
	OvertimeWorkdaySalary       float64    `gorm:"type:decimal(10,2)" json:"overtime_workday_salary"`
	OvertimeHolidaySalary       float64    `gorm:"type:decimal(10,2)" json:"overtime_holiday_salary"`
	AnnualLeaveCarryoverSalary  float64    `gorm:"type:decimal(10,2)" json:"annual_leave_carryover_salary"`
	AttendanceBonus             float64    `gorm:"type:decimal(10,2)" json:"attendance_bonus"`
	PerformanceSalary           float64    `gorm:"type:decimal(10,2)" json:"performance_salary"`
	PostAllowance               float64    `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance               float64    `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance            float64    `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance          float64    `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance           float64    `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceCompensation       float64    `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundCompensation            float64    `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	TotalAdjustment             float64    `gorm:"type:decimal(10,2)" json:"total_adjustment"`
	SocialSecurityDeduct        float64    `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct           float64    `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	TaxDeduct                   float64    `gorm:"type:decimal(10,2)" json:"tax_deduct"`
	FinalSalary                 float64    `gorm:"type:decimal(10,2)" json:"final_salary"`
	LastCalcAt                  *time.Time `json:"last_calc_at"`
	CreatedAt                   time.Time  `json:"created_at"`
	UpdatedAt                   time.Time  `json:"updated_at"`
}

type File struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileName   string         `gorm:"size:256;not null" json:"file_name"`
	FileType   string         `gorm:"size:64" json:"file_type"`
	FileSize   int64          `json:"file_size"`
	FilePath   string         `gorm:"size:512" json:"file_path"`
	UploaderID uint           `json:"uploader_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type FileRelation struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileID     uint           `gorm:"index;not null" json:"file_id"`
	TargetType string         `gorm:"size:64;not null" json:"target_type"`
	TargetID   uint           `gorm:"index;not null" json:"target_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type AuditLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OperatorID     uint      `json:"operator_id"`
	OperatorName   string    `gorm:"size:64" json:"operator_name"`
	TargetType     string    `gorm:"size:64" json:"target_type"`
	TargetID       uint      `json:"target_id"`
	TargetName     string    `gorm:"size:128" json:"target_name"`
	Action         string    `gorm:"size:32" json:"action"`
	BeforeSnapshot string    `gorm:"type:text" json:"before_snapshot"`
	AfterSnapshot  string    `gorm:"type:text" json:"after_snapshot"`
	IP             string    `gorm:"size:64" json:"ip"`
	CreatedAt      time.Time `json:"created_at"`
}

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Username      string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password      string         `gorm:"size:256;not null" json:"-"`
	PersonID      *uint          `json:"person_id"`
	IsFirstLogin  bool           `gorm:"default:true" json:"is_first_login"`
	Status        int8           `gorm:"default:1" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Role struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	Remark    string         `gorm:"size:256" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Permission struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Module     string    `gorm:"size:64;not null" json:"module"`
	Action     string    `gorm:"size:32;not null" json:"action"`
	PermKey    string    `gorm:"uniqueIndex;size:128;not null" json:"perm_key"`
	PermName   string    `gorm:"size:128" json:"perm_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UserRole struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	RoleID    uint      `gorm:"index;not null" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

type RolePermission struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	RoleID       uint      `gorm:"index;not null" json:"role_id"`
	PermissionID uint      `gorm:"index;not null" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type SysBatch struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	BatchNo        string     `gorm:"uniqueIndex;size:64;not null" json:"batch_no"`
	BusinessType   string     `gorm:"size:32;not null" json:"business_type"`
	BusinessPeriod string     `gorm:"size:32" json:"business_period"`
	OperatorID     uint       `json:"operator_id"`
	OperatorName   string     `gorm:"size:64" json:"operator_name"`
	Status         int8       `gorm:"default:1" json:"status"`
	TotalCount     int        `json:"total_count"`
	Remark         string     `gorm:"size:256" json:"remark"`
	CreatedAt      time.Time  `json:"created_at"`
	ExecutedAt     *time.Time `json:"executed_at"`
	CanceledAt     *time.Time `json:"canceled_at"`
}

func AutoMigrateAll(db *gorm.DB) {
	db.AutoMigrate(
		&SysConfig{},
		&User{},
		&Role{},
		&Permission{},
		&UserRole{},
		&RolePermission{},
		&Person{},
		&PersonPhone{},
		&PersonEmail{},
		&PersonBankCard{},
		&Company{},
		&PositionEvent{},
		&PositionSnapshot{},
		&AttendanceEvent{},
		&AttendanceDailyProjection{},
		&AttendanceSalaryMonthly{},
		&LeaveAccountEvent{},
		&LeaveAccountBalance{},
		&SalaryEvent{},
		&SalarySummary{},
		&File{},
		&FileRelation{},
		&AuditLog{},
		&SysBatch{},
	)
}
