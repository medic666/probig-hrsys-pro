package position

import "gorm.io/gorm"

type PositionEvent struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"not null;index" json:"personId"`
	EventName             string         `gorm:"type:varchar(128)" json:"eventName"`
	EffectiveDate         string         `gorm:"type:date;not null" json:"effectiveDate"`
	AttendanceGroup       *string        `gorm:"type:varchar(64)" json:"attendanceGroup"`
	EntryDate             *string        `gorm:"type:date" json:"entryDate"`
	LeaveDate             *string        `gorm:"type:date" json:"leaveDate"`
	HasAnnualLeave        *bool          `json:"hasAnnualLeave"`
	HasAttendanceBonus    *bool          `json:"hasAttendanceBonus"`
	BaseSalary            *float64       `gorm:"type:decimal(10,2)" json:"baseSalary"`
	PerformanceSalary     *float64       `gorm:"type:decimal(10,2)" json:"performanceSalary"`
	SalaryDays            *int           `json:"salaryDays"`
	PostAllowance         *float64       `gorm:"type:decimal(10,2)" json:"postAllowance"`
	MealAllowance         *float64       `gorm:"type:decimal(10,2)" json:"mealAllowance"`
	HousingAllowance      *float64       `gorm:"type:decimal(10,2)" json:"housingAllowance"`
	TransportAllowance    *float64       `gorm:"type:decimal(10,2)" json:"transportAllowance"`
	HighTempAllowance     *float64       `gorm:"type:decimal(10,2)" json:"highTempAllowance"`
	InsuranceCompensation *float64       `gorm:"type:decimal(10,2)" json:"insuranceCompensation"`
	FundCompensation      *float64       `gorm:"type:decimal(10,2)" json:"fundCompensation"`
	SocialSecurityDeduct  *float64       `gorm:"type:decimal(10,2)" json:"socialSecurityDeduct"`
	HousingFundDeduct     *float64       `gorm:"type:decimal(10,2)" json:"housingFundDeduct"`
	CreatedAt             string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt             string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PositionEvent) TableName() string { return "position_events" }

type PositionSnapshot struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	PersonID              uint           `gorm:"not null;index" json:"personId"`
	SnapshotDate          string         `gorm:"type:date;not null" json:"snapshotDate"`
	AttendanceGroup       string         `gorm:"type:varchar(64)" json:"attendanceGroup"`
	EntryDate             *string        `gorm:"type:date" json:"entryDate"`
	LeaveDate             *string        `gorm:"type:date" json:"leaveDate"`
	HasAnnualLeave        bool           `json:"hasAnnualLeave"`
	HasAttendanceBonus    bool           `json:"hasAttendanceBonus"`
	BaseSalary            float64        `gorm:"type:decimal(10,2)" json:"baseSalary"`
	PerformanceSalary     float64        `gorm:"type:decimal(10,2)" json:"performanceSalary"`
	SalaryDays            int            `json:"salaryDays"`
	PostAllowance         float64        `gorm:"type:decimal(10,2)" json:"postAllowance"`
	MealAllowance         float64        `gorm:"type:decimal(10,2)" json:"mealAllowance"`
	HousingAllowance      float64        `gorm:"type:decimal(10,2)" json:"housingAllowance"`
	TransportAllowance    float64        `gorm:"type:decimal(10,2)" json:"transportAllowance"`
	HighTempAllowance     float64        `gorm:"type:decimal(10,2)" json:"highTempAllowance"`
	InsuranceCompensation float64        `gorm:"type:decimal(10,2)" json:"insuranceCompensation"`
	FundCompensation      float64        `gorm:"type:decimal(10,2)" json:"fundCompensation"`
	SocialSecurityDeduct  float64        `gorm:"type:decimal(10,2)" json:"socialSecurityDeduct"`
	HousingFundDeduct     float64        `gorm:"type:decimal(10,2)" json:"housingFundDeduct"`
	CreatedAt             string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt             string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PositionSnapshot) TableName() string { return "position_snapshots" }
