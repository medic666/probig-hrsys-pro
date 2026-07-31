package service

import (
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

var realFarFuture = func() utils.DateOnly {
	t, _ := utils.ParseDate("9999-12-31")
	return utils.DateOnlyFromTime(t)
}()

func RebuildPositionSnapshots(tx *gorm.DB, personID uint) error {
	tx.Where("person_id = ?", personID).Delete(&model.PositionSnapshot{})

	var events []model.PositionEvent
	tx.Where("person_id = ?", personID).Order("effective_date ASC, seq ASC").Find(&events)

	if len(events) == 0 {
		return nil
	}

	type state struct {
		IsActive           bool
		EntryDate          *utils.DateOnly
		LeaveDate          *utils.DateOnly
		AttendanceGroup    string
		HasAnnualLeave     bool
		HasAttendanceBonus bool
		BaseSalary         float64
		PerformanceSalary  float64
		SalaryDays         int
		PostAllowance      float64
		MealAllowance      float64
		HousingAllowance   float64
		TransportAllowance float64
		HighTempAllowance  float64
		InsuranceCompensation float64
		FundCompensation      float64
		SocialSecurityDeduct  float64
		HousingFundDeduct     float64
	}

	apply := func(s *state, e model.PositionEvent) bool {
		changed := false
		if e.EntryDate != nil {
			s.EntryDate = e.EntryDate
			s.IsActive = true
			changed = true
		}
		if e.LeaveDate != nil {
			s.LeaveDate = e.LeaveDate
			s.IsActive = false
			changed = true
		}
		if e.AttendanceGroup != nil {
			if s.AttendanceGroup != *e.AttendanceGroup { changed = true }
			s.AttendanceGroup = *e.AttendanceGroup
		}
		if e.HasAnnualLeave != nil {
			if s.HasAnnualLeave != *e.HasAnnualLeave { changed = true }
			s.HasAnnualLeave = *e.HasAnnualLeave
		}
		if e.HasAttendanceBonus != nil {
			if s.HasAttendanceBonus != *e.HasAttendanceBonus { changed = true }
			s.HasAttendanceBonus = *e.HasAttendanceBonus
		}
		if e.BaseSalary != nil {
			if s.BaseSalary != *e.BaseSalary { changed = true }
			s.BaseSalary = *e.BaseSalary
		}
		if e.PerformanceSalary != nil {
			if s.PerformanceSalary != *e.PerformanceSalary { changed = true }
			s.PerformanceSalary = *e.PerformanceSalary
		}
		if e.SalaryDays != nil {
			if s.SalaryDays != *e.SalaryDays { changed = true }
			s.SalaryDays = *e.SalaryDays
		}
		if e.PostAllowance != nil { applyFloat(&s.PostAllowance, e.PostAllowance, &changed) }
		if e.MealAllowance != nil { applyFloat(&s.MealAllowance, e.MealAllowance, &changed) }
		if e.HousingAllowance != nil { applyFloat(&s.HousingAllowance, e.HousingAllowance, &changed) }
		if e.TransportAllowance != nil { applyFloat(&s.TransportAllowance, e.TransportAllowance, &changed) }
		if e.HighTempAllowance != nil { applyFloat(&s.HighTempAllowance, e.HighTempAllowance, &changed) }
		if e.InsuranceCompensation != nil { applyFloat(&s.InsuranceCompensation, e.InsuranceCompensation, &changed) }
		if e.FundCompensation != nil { applyFloat(&s.FundCompensation, e.FundCompensation, &changed) }
		if e.SocialSecurityDeduct != nil { applyFloat(&s.SocialSecurityDeduct, e.SocialSecurityDeduct, &changed) }
		if e.HousingFundDeduct != nil { applyFloat(&s.HousingFundDeduct, e.HousingFundDeduct, &changed) }
		return changed
	}

	makeSnapshot := func(s state, start, end utils.DateOnly) model.PositionSnapshot {
		return model.PositionSnapshot{
			PersonID:             personID,
			EffectiveStartDate:   start,
			EffectiveEndDate:     end,
			IsActive:             s.IsActive,
			EntryDate:            s.EntryDate,
			LeaveDate:            s.LeaveDate,
			AttendanceGroup:      s.AttendanceGroup,
			HasAnnualLeave:       s.HasAnnualLeave,
			HasAttendanceBonus:   s.HasAttendanceBonus,
			BaseSalary:           s.BaseSalary,
			PerformanceSalary:    s.PerformanceSalary,
			SalaryDays:           s.SalaryDays,
			PostAllowance:        s.PostAllowance,
			MealAllowance:        s.MealAllowance,
			HousingAllowance:     s.HousingAllowance,
			TransportAllowance:   s.TransportAllowance,
			HighTempAllowance:    s.HighTempAllowance,
			InsuranceCompensation: s.InsuranceCompensation,
			FundCompensation:      s.FundCompensation,
			SocialSecurityDeduct:  s.SocialSecurityDeduct,
			HousingFundDeduct:     s.HousingFundDeduct,
			LastCalcAt: time.Now(),
		}
	}

	current := state{}
	var snapshots []model.PositionSnapshot
	var snapshotStart utils.DateOnly
	firstChange := true

	for _, e := range events {
		prev := current
		changed := apply(&current, e)

		if changed || firstChange {
			if firstChange {
				snapshotStart = e.EffectiveDate
				firstChange = false
			} else {
				endDate := e.EffectiveDate.AddDate(0, 0, -1)
				if !endDate.Before(snapshotStart) {
					snapshots = append(snapshots, makeSnapshot(prev, snapshotStart, endDate))
				}
				snapshotStart = e.EffectiveDate
			}
		}
	}

	snapshots = append(snapshots, makeSnapshot(current, snapshotStart, realFarFuture))

	for _, s := range snapshots {
		if err := tx.Create(&s).Error; err != nil {
			return err
		}
	}

	return nil
}

func applyFloat(target *float64, src *float64, changed *bool) {
	if *target != *src {
		*changed = true
		*target = *src
	}
}

func GetCurrentPositionSnapshot(personID uint) (*model.PositionSnapshot, error) {
	var snapshot model.PositionSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date DESC").First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func GetPositionSnapshotHistory(personID uint) ([]model.PositionSnapshot, error) {
	var snapshots []model.PositionSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date ASC").Find(&snapshots).Error
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}
