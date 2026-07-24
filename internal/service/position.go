package service

import (
	"database/sql"
	"sort"
	"time"

	"probig/internal/dao"
	"probig/internal/middleware"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
)

func GetPositionSnapshots(personID uint) ([]models.PositionSnapshot, error) {
	return dao.GetPositionSnapshotsByPersonID(personID)
}

func GetPositionEvents(personID uint) ([]models.PositionEvent, error) {
	return dao.GetPositionEventsByPersonID(personID)
}

func GetPositionEvent(id uint) (*models.PositionEvent, error) {
	return dao.GetPositionEventByID(id)
}

func CreatePositionEvent(c *gin.Context, e *models.PositionEvent) error {
	if err := dao.CreatePositionEvent(e); err != nil {
		return err
	}
	middleware.RecordAudit(c, "新增", "position_event", e.ID, nil, e, "")
	go RebuildPositionSnapshots(e.PersonID)
	return nil
}

func UpdatePositionEvent(c *gin.Context, e *models.PositionEvent) error {
	old, err := dao.GetPositionEventByID(e.ID)
	if err != nil {
		return err
	}
	if err := dao.UpdatePositionEvent(e); err != nil {
		return err
	}
	middleware.RecordAudit(c, "修改", "position_event", e.ID, old, e, "")
	go RebuildPositionSnapshots(e.PersonID)
	return nil
}

func DeletePositionEvent(c *gin.Context, id uint) error {
	e, err := dao.GetPositionEventByID(id)
	if err != nil {
		return err
	}
	personID := e.PersonID
	if err := dao.DeletePositionEvent(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "position_event", id, e, nil, "")
	go RebuildPositionSnapshots(personID)
	return nil
}

func RebuildPositionSnapshots(personID uint) error {
	if err := dao.DeletePositionSnapshotsByPersonID(personID); err != nil {
		return err
	}
	events, err := dao.GetPositionEventsByPersonID(personID)
	if err != nil || len(events) == 0 {
		return err
	}
	var earliestDate, latestDate time.Time
	for _, e := range events {
		if e.EffectiveDate != nil {
			if earliestDate.IsZero() || e.EffectiveDate.Before(earliestDate) {
				earliestDate = *e.EffectiveDate
			}
			if latestDate.IsZero() || e.EffectiveDate.After(latestDate) {
				latestDate = *e.EffectiveDate
			}
		}
	}
	if earliestDate.IsZero() {
		return nil
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].EffectiveDate == nil || events[j].EffectiveDate == nil {
			return false
		}
		if events[i].EffectiveDate.Equal(*events[j].EffectiveDate) {
			return events[i].CreatedAt.After(events[j].CreatedAt)
		}
		return events[i].EffectiveDate.Before(*events[j].EffectiveDate)
	})

	var snapshots []models.PositionSnapshot
	current := currentDateState()
	for d := earliestDate; !d.After(latestDate); d = d.AddDate(0, 0, 1) {
		current = applyEventsForDate(current, events, d)
		snapshot := models.PositionSnapshot{
			PersonID:             personID,
			SnapshotDate:         copyTimePtr(&d),
			AttendanceGroup:      stringPtrVal(current.AttendanceGroup),
			EntryDate:            current.EntryDate,
			LeaveDate:            current.LeaveDate,
			HasAnnualLeave:       boolVal(current.HasAnnualLeave),
			HasAttendanceBonus:   boolVal(current.HasAttendanceBonus),
			BaseSalary:           float64Val(current.BaseSalary),
			PerformanceSalary:    float64Val(current.PerformanceSalary),
			SalaryDays:           intVal(current.SalaryDays),
			PostAllowance:        float64Val(current.PostAllowance),
			MealAllowance:        float64Val(current.MealAllowance),
			HousingAllowance:     float64Val(current.HousingAllowance),
			TransportAllowance:   float64Val(current.TransportAllowance),
			HighTempAllowance:    float64Val(current.HighTempAllowance),
			InsuranceCompensation: float64Val(current.InsuranceCompensation),
			FundCompensation:     float64Val(current.FundCompensation),
			SocialSecurityDeduct: float64Val(current.SocialSecurityDeduct),
			HousingFundDeduct:    float64Val(current.HousingFundDeduct),
		}
		snapshots = append(snapshots, snapshot)
	}
	return dao.CreatePositionSnapshots(snapshots)
}

type dateState struct {
	AttendanceGroup      *string
	EntryDate            *time.Time
	LeaveDate            *time.Time
	HasAnnualLeave       *bool
	HasAttendanceBonus   *bool
	BaseSalary           sql.NullFloat64
	PerformanceSalary    sql.NullFloat64
	SalaryDays           *int
	PostAllowance        sql.NullFloat64
	MealAllowance        sql.NullFloat64
	HousingAllowance     sql.NullFloat64
	TransportAllowance   sql.NullFloat64
	HighTempAllowance    sql.NullFloat64
	InsuranceCompensation sql.NullFloat64
	FundCompensation     sql.NullFloat64
	SocialSecurityDeduct sql.NullFloat64
	HousingFundDeduct    sql.NullFloat64
}

func currentDateState() dateState {
	return dateState{}
}

func applyEventsForDate(s dateState, events []models.PositionEvent, date time.Time) dateState {
	for _, e := range events {
		if e.EffectiveDate == nil || e.EffectiveDate.After(date) {
			continue
		}
		if e.AttendanceGroup != nil {
			s.AttendanceGroup = e.AttendanceGroup
		}
		if e.EntryDate != nil {
			s.EntryDate = e.EntryDate
		}
		if e.LeaveDate != nil {
			s.LeaveDate = e.LeaveDate
		}
		if e.HasAnnualLeave != nil {
			s.HasAnnualLeave = e.HasAnnualLeave
		}
		if e.HasAttendanceBonus != nil {
			s.HasAttendanceBonus = e.HasAttendanceBonus
		}
		if e.BaseSalary.Valid {
			s.BaseSalary = e.BaseSalary
		}
		if e.PerformanceSalary.Valid {
			s.PerformanceSalary = e.PerformanceSalary
		}
		if e.SalaryDays != nil {
			s.SalaryDays = e.SalaryDays
		}
		if e.PostAllowance.Valid {
			s.PostAllowance = e.PostAllowance
		}
		if e.MealAllowance.Valid {
			s.MealAllowance = e.MealAllowance
		}
		if e.HousingAllowance.Valid {
			s.HousingAllowance = e.HousingAllowance
		}
		if e.TransportAllowance.Valid {
			s.TransportAllowance = e.TransportAllowance
		}
		if e.HighTempAllowance.Valid {
			s.HighTempAllowance = e.HighTempAllowance
		}
		if e.InsuranceCompensation.Valid {
			s.InsuranceCompensation = e.InsuranceCompensation
		}
		if e.FundCompensation.Valid {
			s.FundCompensation = e.FundCompensation
		}
		if e.SocialSecurityDeduct.Valid {
			s.SocialSecurityDeduct = e.SocialSecurityDeduct
		}
		if e.HousingFundDeduct.Valid {
			s.HousingFundDeduct = e.HousingFundDeduct
		}
	}
	return s
}

func stringPtrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func intVal(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func float64Val(f sql.NullFloat64) float64 {
	if !f.Valid {
		return 0
	}
	return f.Float64
}

func copyTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	c := *t
	return &c
}
