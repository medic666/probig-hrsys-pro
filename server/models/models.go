package models

import (
	"time"

	"gorm.io/gorm"
)

type Person struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(64);not null" json:"name"`
	IdCard          string         `gorm:"type:varchar(128);uniqueIndex;encrypted" json:"id_card"`
	Gender          *int           `gorm:"type:tinyint" json:"gender"`
	Birthday        *time.Time     `gorm:"type:date" json:"birthday"`
	Nation          string         `gorm:"type:varchar(32)" json:"nation"`
	NativePlace     string         `gorm:"type:varchar(128)" json:"native_place"`
	Address         string         `gorm:"type:varchar(256)" json:"address"`
	PoliticalStatus string         `gorm:"type:varchar(32)" json:"political_status"`
	MaritalStatus   *int           `gorm:"type:tinyint" json:"marital_status"`
	Alias           string         `gorm:"type:varchar(64)" json:"alias"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Phones          []PersonPhone  `gorm:"foreignKey:PersonID" json:"phones,omitempty"`
	Emails          []PersonEmail  `gorm:"foreignKey:PersonID" json:"emails,omitempty"`
	BankCards       []PersonBankCard `gorm:"foreignKey:PersonID" json:"bank_cards,omitempty"`
}

type PersonPhone struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"person_id"`
	Phone     string         `gorm:"type:varchar(128);encrypted" json:"phone"`
	Remark    string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PersonEmail struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"person_id"`
	Email     string         `gorm:"type:varchar(256);encrypted" json:"email"`
	Remark    string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PersonBankCard struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PersonID  uint           `gorm:"not null;index" json:"person_id"`
	BankCard  string         `gorm:"type:varchar(128);encrypted" json:"bank_card"`
	BankName  string         `gorm:"type:varchar(64)" json:"bank_name"`
	Remark    string         `gorm:"type:varchar(64)" json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Company struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(128);not null" json:"name"`
	CreditCode   string         `gorm:"type:varchar(64);uniqueIndex" json:"credit_code"`
	Address      string         `gorm:"type:varchar(256)" json:"address"`
	ContactPhone string         `gorm:"type:varchar(32)" json:"contact_phone"`
	BankName     string         `gorm:"type:varchar(64)" json:"bank_name"`
	BankAccount  string         `gorm:"type:varchar(128);encrypted" json:"bank_account"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type File struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"type:varchar(128)" json:"name"`
	Size       int64          `json:"size"`
	MimeType   string         `gorm:"type:varchar(64)" json:"mime_type"`
	Content    []byte         `gorm:"type:blob" json:"-"`
	UploaderID uint           `json:"uploader_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type FileRelation struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FileID     uint           `gorm:"not null;index" json:"file_id"`
	TargetType string         `gorm:"type:varchar(32);not null;index" json:"target_type"`
	TargetID   uint           `gorm:"not null;index" json:"target_id"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type PositionEvent struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"not null;index" json:"person_id"`
	EventName             string         `gorm:"type:varchar(128)" json:"event_name"`
	EffectiveDate         time.Time      `gorm:"type:date;not null" json:"effective_date"`
	AttendanceGroup       *string        `gorm:"type:varchar(64)" json:"attendance_group"`
	EntryDate             *time.Time     `gorm:"type:date" json:"entry_date"`
	LeaveDate             *time.Time     `gorm:"type:date" json:"leave_date"`
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
	Person                *Person        `gorm:"foreignKey:PersonID" json:"person,omitempty"`
}

type AttendanceEvent struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	PersonID           uint           `gorm:"not null;index" json:"person_id"`
	EventDate          time.Time      `gorm:"type:date;not null" json:"event_date"`
	EventType          string         `gorm:"type:varchar(32);not null" json:"event_type"`
	SubType            string         `gorm:"type:varchar(32);not null" json:"sub_type"`
	Hours              *float64       `gorm:"type:decimal(4,1)" json:"hours"`
	LateMinutes        *int           `json:"late_minutes"`
	LeaveAdjustAmount  *float64       `gorm:"type:decimal(4,1)" json:"leave_adjust_amount"`
	IsSpecialApproval  *bool          `json:"is_special_approval"`
	Remark             string         `gorm:"type:varchar(256)" json:"remark"`
	BatchID            string         `gorm:"type:varchar(64)" json:"batch_id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Person             *Person        `gorm:"foreignKey:PersonID" json:"person,omitempty"`
}

type SalaryEvent struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PersonID    uint           `gorm:"not null;index" json:"person_id"`
	BelongMonth string         `gorm:"type:varchar(7);not null" json:"belong_month"`
	EventType   string         `gorm:"type:varchar(32);not null" json:"event_type"`
	Amount      float64        `gorm:"type:decimal(10,2)" json:"amount"`
	EventName   string         `gorm:"type:varchar(128)" json:"event_name"`
	Remark      string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Person      *Person        `gorm:"foreignKey:PersonID" json:"person,omitempty"`
}

type PositionSnapshot struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"not null;index:idx_person_date" json:"person_id"`
	SnapshotDate          time.Time      `gorm:"type:date;not null;index:idx_person_date" json:"snapshot_date"`
	AttendanceGroup       string         `gorm:"type:varchar(64)" json:"attendance_group"`
	EntryDate             *time.Time     `gorm:"type:date" json:"entry_date"`
	LeaveDate             *time.Time     `gorm:"type:date" json:"leave_date"`
	HasAnnualLeave        bool           `json:"has_annual_leave"`
	HasAttendanceBonus    bool           `json:"has_attendance_bonus"`
	BaseSalary            float64        `gorm:"type:decimal(10,2)" json:"base_salary"`
	PerformanceSalary     float64        `gorm:"type:decimal(10,2)" json:"performance_salary"`
	SalaryDays            int            `json:"salary_days"`
	PostAllowance         float64        `gorm:"type:decimal(10,2)" json:"post_allowance"`
	MealAllowance         float64        `gorm:"type:decimal(10,2)" json:"meal_allowance"`
	HousingAllowance      float64        `gorm:"type:decimal(10,2)" json:"housing_allowance"`
	TransportAllowance    float64        `gorm:"type:decimal(10,2)" json:"transport_allowance"`
	HighTempAllowance     float64        `gorm:"type:decimal(10,2)" json:"high_temp_allowance"`
	InsuranceCompensation float64        `gorm:"type:decimal(10,2)" json:"insurance_compensation"`
	FundCompensation      float64        `gorm:"type:decimal(10,2)" json:"fund_compensation"`
	SocialSecurityDeduct  float64        `gorm:"type:decimal(10,2)" json:"social_security_deduct"`
	HousingFundDeduct     float64        `gorm:"type:decimal(10,2)" json:"housing_fund_deduct"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type AttendanceSummary struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"not null;index" json:"person_id"`
	BelongMonth           string         `gorm:"type:varchar(7);not null" json:"belong_month"`
	WorkDays              float64        `gorm:"type:decimal(4,1)" json:"work_days"`
	MakeUpDays            float64        `gorm:"type:decimal(4,1)" json:"make_up_days"`
	SickLeaveDays         float64        `gorm:"type:decimal(4,1)" json:"sick_leave_days"`
	PersonalLeaveDays     float64        `gorm:"type:decimal(4,1)" json:"personal_leave_days"`
	AnnualLeaveDays       float64        `gorm:"type:decimal(4,1)" json:"annual_leave_days"`
	StatutoryLeaveDays    float64        `gorm:"type:decimal(4,1)" json:"statutory_leave_days"`
	WelfareLeaveDays      float64        `gorm:"type:decimal(4,1)" json:"welfare_leave_days"`
	OvertimeWorkdayHours  float64        `gorm:"type:decimal(5,1)" json:"overtime_workday_hours"`
	OvertimeHolidayHours  float64        `gorm:"type:decimal(5,1)" json:"overtime_holiday_hours"`
	ViolationCount        int            `json:"violation_count"`
	LastCalcAt            *time.Time     `json:"last_calc_at"`
	IsLocked              bool           `json:"is_locked"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Person                *Person        `gorm:"foreignKey:PersonID" json:"person,omitempty"`
}

type SalarySummary struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	PersonID           uint           `gorm:"not null;index" json:"person_id"`
	BelongMonth        string         `gorm:"type:varchar(7);not null" json:"belong_month"`
	AttendanceSalary   float64        `gorm:"type:decimal(10,2)" json:"attendance_salary"`
	OvertimeSalary     float64        `gorm:"type:decimal(10,2)" json:"overtime_salary"`
	AttendanceBonus    float64        `gorm:"type:decimal(10,2)" json:"attendance_bonus"`
	PerformanceSalary  float64        `gorm:"type:decimal(10,2)" json:"performance_salary"`
	TotalAllowance     float64        `gorm:"type:decimal(10,2)" json:"total_allowance"`
	TotalAdjustment    float64        `gorm:"type:decimal(10,2)" json:"total_adjustment"`
	TotalDeduction     float64        `gorm:"type:decimal(10,2)" json:"total_deduction"`
	FinalSalary        float64        `gorm:"type:decimal(10,2)" json:"final_salary"`
	LastCalcAt         *time.Time     `json:"last_calc_at"`
	IsLocked           bool           `json:"is_locked"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Person             *Person        `gorm:"foreignKey:PersonID" json:"person,omitempty"`
}

type AuditLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OperatorID     uint      `json:"operator_id"`
	OperatorName   string    `gorm:"type:varchar(64)" json:"operator_name"`
	TargetType     string    `gorm:"type:varchar(64)" json:"target_type"`
	TargetID       uint      `json:"target_id"`
	Action         string    `gorm:"type:varchar(32)" json:"action"`
	BeforeSnapshot string    `gorm:"type:text" json:"before_snapshot"`
	AfterSnapshot  string    `gorm:"type:text" json:"after_snapshot"`
	BatchID        string    `gorm:"type:varchar(64)" json:"batch_id"`
	IP             string    `gorm:"type:varchar(64)" json:"ip"`
	CreatedAt      time.Time `json:"created_at"`
}

type SysConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ConfigKey    string    `gorm:"type:varchar(64);uniqueIndex" json:"config_key"`
	ConfigValue  string    `gorm:"type:text" json:"config_value"`
	ConfigName   string    `gorm:"type:varchar(128)" json:"config_name"`
	ConfigDesc   string    `gorm:"type:varchar(256)" json:"config_desc"`
	ValueType    string    `gorm:"type:varchar(16)" json:"value_type"`
	OptionValues string    `gorm:"type:text" json:"option_values"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex" json:"username"`
	PasswordHash string         `gorm:"type:varchar(128)" json:"-"`
	PersonID     *uint          `json:"person_id"`
	IsFirstLogin bool           `json:"is_first_login"`
	Status       int            `gorm:"type:tinyint;default:1" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Person       *Person        `gorm:"foreignKey:PersonID" json:"person,omitempty"`
	Roles        []Role         `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(64);not null" json:"name"`
	Remark      string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Permissions []Permission   `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

type Permission struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PermissionKey string    `gorm:"type:varchar(64);uniqueIndex" json:"permission_key"`
	Name          string    `gorm:"type:varchar(64)" json:"name"`
	Module        string    `gorm:"type:varchar(32)" json:"module"`
	Action        string    `gorm:"type:varchar(32)" json:"action"`
}

type UserRole struct {
	UserID    uint           `gorm:"primaryKey" json:"user_id"`
	RoleID    uint           `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type RolePermission struct {
	RoleID       uint           `gorm:"primaryKey" json:"role_id"`
	PermissionID uint           `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
