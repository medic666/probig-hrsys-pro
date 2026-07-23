package models

import "time"

type Entity struct {
	ID        uint       `db:"id" json:"id"`
	Type      string     `db:"type" json:"type"`
	Name      string     `db:"name" json:"name"`
	Status    string     `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

type User struct {
	ID           uint      `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	RoleID       uint      `db:"role_id" json:"role_id"`
	EntityID     *uint     `db:"entity_id" json:"entity_id"`
	Status       string    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type Role struct {
	ID          uint      `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type RolePermission struct {
	ID     uint   `db:"id" json:"id"`
	RoleID uint   `db:"role_id" json:"role_id"`
	Module string `db:"module" json:"module"`
	Action string `db:"action" json:"action"`
}

type File struct {
	ID           uint       `db:"id" json:"id"`
	Filename     string     `db:"filename" json:"filename"`
	OriginalName string     `db:"original_name" json:"original_name"`
	Path         string     `db:"path" json:"path"`
	Size         int64      `db:"size" json:"size"`
	MimeType     string     `db:"mime_type" json:"mime_type"`
	UploadedBy   uint       `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at"`
}

type FileAssociation struct {
	ID         uint      `db:"id" json:"id"`
	FileID     uint      `db:"file_id" json:"file_id"`
	TargetType string    `db:"target_type" json:"target_type"`
	TargetID   uint      `db:"target_id" json:"target_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type PersonEvent struct {
	ID            uint      `db:"id" json:"id"`
	PersonID      uint      `db:"person_id" json:"person_id"`
	EffectiveDate string    `db:"effective_date" json:"effective_date"`
	EventType     string    `db:"event_type" json:"event_type"`
	Payload       string    `db:"payload" json:"payload"`
	CreatedBy     uint      `db:"created_by" json:"created_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type OrgEvent struct {
	ID            uint      `db:"id" json:"id"`
	OrgID         uint      `db:"org_id" json:"org_id"`
	EffectiveDate string    `db:"effective_date" json:"effective_date"`
	EventType     string    `db:"event_type" json:"event_type"`
	Payload       string    `db:"payload" json:"payload"`
	CreatedBy     uint      `db:"created_by" json:"created_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type AttendanceEvent struct {
	ID         uint      `db:"id" json:"id"`
	PersonID   uint      `db:"person_id" json:"person_id"`
	PersonName string    `db:"person_name" json:"person_name"`
	Date       string    `db:"date" json:"date"`
	EventType  string    `db:"event_type" json:"event_type"`
	Duration   float64   `db:"duration" json:"duration"`
	Remark     string    `db:"remark" json:"remark"`
	CreatedBy  uint      `db:"created_by" json:"created_by"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type SalaryEvent struct {
	ID         uint      `db:"id" json:"id"`
	PersonID   uint      `db:"person_id" json:"person_id"`
	PersonName string    `db:"person_name" json:"person_name"`
	Period     string    `db:"period" json:"period"`
	EventType  string    `db:"event_type" json:"event_type"`
	Amount     float64   `db:"amount" json:"amount"`
	Detail     string    `db:"detail" json:"detail"`
	CreatedBy  uint      `db:"created_by" json:"created_by"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type PersonSnapshot struct {
	ID            uint      `db:"id" json:"id"`
	PersonID      uint      `db:"person_id" json:"person_id"`
	EventID       uint      `db:"event_id" json:"event_id"`
	EffectiveDate string    `db:"effective_date" json:"effective_date"`
	SnapshotData  string    `db:"snapshot_data" json:"snapshot_data"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type OrgSnapshot struct {
	ID            uint      `db:"id" json:"id"`
	OrgID         uint      `db:"org_id" json:"org_id"`
	EventID       uint      `db:"event_id" json:"event_id"`
	EffectiveDate string    `db:"effective_date" json:"effective_date"`
	SnapshotData  string    `db:"snapshot_data" json:"snapshot_data"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type AttendanceSummary struct {
	ID                         uint      `db:"id" json:"id"`
	PersonID                   uint      `db:"person_id" json:"person_id"`
	PersonName                 string    `db:"person_name" json:"person_name"`
	Period                     string    `db:"period" json:"period"`
	NormalAttendanceDays      float64   `db:"normal_attendance_days" json:"normal_attendance_days"`
	SupplementaryAttendanceDays float64  `db:"supplementary_attendance_days" json:"supplementary_attendance_days"`
	CompensatoryLeaveDays     float64   `db:"compensatory_leave_days" json:"compensatory_leave_days"`
	PersonalLeaveDays         float64   `db:"personal_leave_days" json:"personal_leave_days"`
	SickLeaveDays             float64   `db:"sick_leave_days" json:"sick_leave_days"`
	AnnualLeaveDays           float64   `db:"annual_leave_days" json:"annual_leave_days"`
	StatutoryLeaveDays        float64   `db:"statutory_leave_days" json:"statutory_leave_days"`
	WelfareLeaveDays          float64   `db:"welfare_leave_days" json:"welfare_leave_days"`
	WorkdayOvertimeDays       float64   `db:"workday_overtime_days" json:"workday_overtime_days"`
	HolidayOvertimeDays       float64   `db:"holiday_overtime_days" json:"holiday_overtime_days"`
	MissingClockCount         int       `db:"missing_clock_count" json:"missing_clock_count"`
	LateCount                 int       `db:"late_count" json:"late_count"`
	EarlyLeaveCount           int       `db:"early_leave_count" json:"early_leave_count"`
	AnnualLeaveAllot          float64   `db:"annual_leave_allot" json:"annual_leave_allot"`
	AnnualLeaveCarryover      float64   `db:"annual_leave_carryover" json:"annual_leave_carryover"`
	ViolationCount            int       `db:"violation_count" json:"violation_count"`
	CalculatedAt              time.Time `db:"calculated_at" json:"calculated_at"`
}

type SalarySummary struct {
	ID                      uint      `db:"id" json:"id"`
	PersonID                uint      `db:"person_id" json:"person_id"`
	PersonName              string    `db:"person_name" json:"person_name"`
	Period                  string    `db:"period" json:"period"`
	AttendanceSalary        float64   `db:"attendance_salary" json:"attendance_salary"`
	FullAttendanceBonus     float64   `db:"full_attendance_bonus" json:"full_attendance_bonus"`
	OvertimeSalary          float64   `db:"overtime_salary" json:"overtime_salary"`
	PerformanceSalary       float64   `db:"performance_salary" json:"performance_salary"`
	PositionAllowance       float64   `db:"position_allowance" json:"position_allowance"`
	MealSubsidy             float64   `db:"meal_subsidy" json:"meal_subsidy"`
	HousingSubsidy          float64   `db:"housing_subsidy" json:"housing_subsidy"`
	TransportSubsidy        float64   `db:"transport_subsidy" json:"transport_subsidy"`
	HeatSubsidy             float64   `db:"heat_subsidy" json:"heat_subsidy"`
	InsuranceCompensation   float64   `db:"insurance_compensation" json:"insurance_compensation"`
	HousingFundCompensation float64   `db:"housing_fund_compensation" json:"housing_fund_compensation"`
	SocialInsuranceDeduct   float64   `db:"social_insurance_deduct" json:"social_insurance_deduct"`
	HousingFundDeduct       float64   `db:"housing_fund_deduct" json:"housing_fund_deduct"`
	TaxDeduct               float64   `db:"tax_deduct" json:"tax_deduct"`
	LoanDeduct              float64   `db:"loan_deduct" json:"loan_deduct"`
	RewardPunish            float64   `db:"reward_punish" json:"reward_punish"`
	OtherAdjustments        float64   `db:"other_adjustments" json:"other_adjustments"`
	TotalSalary             float64   `db:"total_salary" json:"total_salary"`
	DetailData              string    `db:"detail_data" json:"detail_data"`
	CalculatedAt            time.Time `db:"calculated_at" json:"calculated_at"`
}

type AuditLog struct {
	ID         uint      `db:"id" json:"id"`
	UserID     uint      `db:"user_id" json:"user_id"`
	Username   string    `db:"username" json:"username"`
	Action     string    `db:"action" json:"action"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   *uint     `db:"entity_id" json:"entity_id"`
	Payload    string    `db:"payload" json:"payload"`
	IPAddress  string    `db:"ip_address" json:"ip_address"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type PersonSnapshotData struct {
	Name                   string   `json:"name"`
	AttendanceGroup        string   `json:"attendance_group"`
	EntryDate              string   `json:"entry_date"`
	BasicSalary            float64  `json:"basic_salary"`
	PerformanceSalary      float64  `json:"performance_salary"`
	SalaryDays             float64  `json:"salary_days"`
	PositionAllowance      float64  `json:"position_allowance"`
	MealSubsidy            float64  `json:"meal_subsidy"`
	HousingSubsidy         float64  `json:"housing_subsidy"`
	TransportSubsidy       float64  `json:"transport_subsidy"`
	HeatSubsidy            float64  `json:"heat_subsidy"`
	InsuranceCompensation  float64  `json:"insurance_compensation"`
	HousingFundCompensation float64 `json:"housing_fund_compensation"`
	SocialInsuranceDeduct  float64  `json:"social_insurance_deduct"`
	HousingFundDeduct      float64  `json:"housing_fund_deduct"`
	Phones                 []string `json:"phones"`
	Emails                 []string `json:"emails"`
	IDNumber               string   `json:"id_number"`
	Gender                 string   `json:"gender"`
	Birthday               string   `json:"birthday"`
	Ethnicity              string   `json:"ethnicity"`
	NativePlace            string   `json:"native_place"`
	Address                string   `json:"address"`
	BankCards              []string `json:"bank_cards"`
	PoliticalStatus        string   `json:"political_status"`
	MaritalStatus          string   `json:"marital_status"`
	Alias                  string   `json:"alias"`
}

func DefaultPersonSnapshotData() PersonSnapshotData {
	return PersonSnapshotData{
		SalaryDays: 21.75,
	}
}

type OrgSnapshotData struct {
	CompanyName           string `json:"company_name"`
	CreditCode            string `json:"credit_code"`
	Address               string `json:"address"`
	ContactPhone          string `json:"contact_phone"`
	BankName              string `json:"bank_name"`
	BankAccount           string `json:"bank_account"`
	BusinessLicenseFileID *uint  `json:"business_license_file_id"`
	SealFileID            *uint  `json:"seal_file_id"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
	Perms []string `json:"perms"`
}

type PersonEventRequest struct {
	PersonID      uint                   `json:"person_id"`
	EffectiveDate string                 `json:"effective_date" binding:"required"`
	EventType     string                 `json:"event_type" binding:"required"`
	Data          PersonSnapshotData     `json:"data" binding:"required"`
}

type OrgEventRequest struct {
	OrgID         uint               `json:"org_id"`
	EffectiveDate string             `json:"effective_date" binding:"required"`
	EventType     string             `json:"event_type" binding:"required"`
	Data          OrgSnapshotData    `json:"data" binding:"required"`
}

type AttendanceEventRequest struct {
	PersonID  uint    `json:"person_id" binding:"required"`
	Date      string  `json:"date" binding:"required"`
	EventType string  `json:"event_type" binding:"required"`
	Duration  float64 `json:"duration"`
	Remark    string  `json:"remark"`
}

type SalaryEventRequest struct {
	PersonID  uint    `json:"person_id" binding:"required"`
	Period    string  `json:"period" binding:"required"`
	EventType string  `json:"event_type" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Detail    string  `json:"detail"`
}

type CalculateRequest struct {
	Period   string `json:"period" binding:"required"`
	PersonID uint   `json:"person_id"`
}

type FileAssociationRequest struct {
	FileID     uint   `json:"file_id" binding:"required"`
	TargetType string `json:"target_type" binding:"required"`
	TargetID   uint   `json:"target_id" binding:"required"`
}
