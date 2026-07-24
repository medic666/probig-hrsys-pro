package salary

import "gorm.io/gorm"

type SalaryEvent struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PersonID    uint           `gorm:"not null;index" json:"personId"`
	BelongMonth string         `gorm:"type:varchar(7);not null" json:"belongMonth"`
	EventType   string         `gorm:"type:varchar(32);not null" json:"eventType"`
	Amount      float64        `gorm:"type:decimal(10,2)" json:"amount"`
	EventName   string         `gorm:"type:varchar(128)" json:"eventName"`
	Remark      string         `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt   string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt   string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalaryEvent) TableName() string { return "salary_events" }

type SalarySummary struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	PersonID          uint           `gorm:"not null;index" json:"personId"`
	BelongMonth       string         `gorm:"type:varchar(7);not null" json:"belongMonth"`
	AttendanceSalary  float64        `gorm:"type:decimal(10,2)" json:"attendanceSalary"`
	OvertimeSalary    float64        `gorm:"type:decimal(10,2)" json:"overtimeSalary"`
	AttendanceBonus   float64        `gorm:"type:decimal(10,2)" json:"attendanceBonus"`
	PerformanceSalary float64        `gorm:"type:decimal(10,2)" json:"performanceSalary"`
	TotalAllowance    float64        `gorm:"type:decimal(10,2)" json:"totalAllowance"`
	TotalAdjustment   float64        `gorm:"type:decimal(10,2)" json:"totalAdjustment"`
	TotalDeduction    float64        `gorm:"type:decimal(10,2)" json:"totalDeduction"`
	FinalSalary       float64        `gorm:"type:decimal(10,2)" json:"finalSalary"`
	LastCalcAt        *string        `gorm:"type:datetime" json:"lastCalcAt"`
	IsLocked          bool           `gorm:"default:false" json:"isLocked"`
	CreatedAt         string         `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt         string         `gorm:"type:datetime" json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SalarySummary) TableName() string { return "salary_summaries" }
