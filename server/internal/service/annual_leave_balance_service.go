package service

import (
	"sort"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func RebuildAnnualLeaveBalance(tx *gorm.DB, personID uint) error {
	tx.Where("person_id = ?", personID).Delete(&model.AnnualLeaveBalanceSnapshot{})

	var accountEvents []model.AnnualLeaveAccountEvent
	tx.Where("person_id = ?", personID).Order("effective_date ASC, seq ASC").Find(&accountEvents)

	var attendEvents []model.AttendanceEventDetail
	var attendDates []utils.DateOnly
	rows2, _ := tx.Table("attendance_event_details").
		Select("attendance_event_details.hours, attendance_daily.event_date").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id").
		Where("attendance_daily.person_id = ? AND attendance_event_details.event_type = ? AND attendance_event_details.sub_type = ?",
			personID, "休假", "年假").
		Order("attendance_daily.event_date ASC").
		Rows()
	if rows2 != nil {
		for rows2.Next() {
			var hours float64
			var eventDate utils.DateOnly
			rows2.Scan(&hours, &eventDate)
			attendEvents = append(attendEvents, model.AttendanceEventDetail{Hours: hours})
			attendDates = append(attendDates, eventDate)
		}
		rows2.Close()
	}

	type ch struct {
		date  utils.DateOnly
		hours float64
	}

	var changes []ch
	for _, e := range accountEvents {
		h := e.Hours
		if e.EventType == "carryover_deduct" {
			h = -h
		}
		changes = append(changes, ch{e.EffectiveDate, h})
	}
	for i, e := range attendEvents {
		changes = append(changes, ch{attendDates[i], -e.Hours})
	}

	if len(changes) == 0 {
		return nil
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].date.Time().Before(changes[j].date.Time())
	})

	var snapshots []model.AnnualLeaveBalanceSnapshot
	var runningBalance float64
	var snapshotStart utils.DateOnly
	firstChange := true

	for _, ch := range changes {
		if !firstChange {
			endDate := ch.date.AddDate(0, 0, -1)
			if !endDate.Before(snapshotStart) {
				snapshots = append(snapshots, makeALSnapshot(personID, snapshotStart, endDate, runningBalance))
			}
			snapshotStart = ch.date
		} else {
			snapshotStart = ch.date
			firstChange = false
		}
		runningBalance += ch.hours
	}

	t, _ := utils.ParseDate("9999-12-31")
	snapshots = append(snapshots, makeALSnapshot(personID, snapshotStart, utils.DateOnlyFromTime(t), runningBalance))

	for _, s := range snapshots {
		if err := tx.Create(&s).Error; err != nil {
			return err
		}
	}
	return nil
}

func makeALSnapshot(personID uint, start, end utils.DateOnly, balance float64) model.AnnualLeaveBalanceSnapshot {
	return model.AnnualLeaveBalanceSnapshot{
		PersonID:           personID,
		EffectiveStartDate: start,
		EffectiveEndDate:   end,
		BalanceHours:       balance,
		LastCalcAt:         utils.DateOnlyFromTime(time.Now()),
	}
}

func GetCurrentAnnualLeaveBalance(personID uint) (*model.AnnualLeaveBalanceSnapshot, error) {
	var snapshot model.AnnualLeaveBalanceSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date DESC").First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func GetAnnualLeaveBalanceHistory(personID uint) ([]model.AnnualLeaveBalanceSnapshot, error) {
	var snapshots []model.AnnualLeaveBalanceSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date ASC").Find(&snapshots).Error
	return snapshots, err
}

func RebuildLeaveInLieuBalance(tx *gorm.DB, personID uint) error {
	tx.Where("person_id = ?", personID).Delete(&model.LeaveInLieuBalanceSnapshot{})

	var events []model.AttendanceEventDetail
	var eventDates []utils.DateOnly
	rows, _ := tx.Table("attendance_event_details").
		Select("attendance_event_details.hours, attendance_event_details.sub_type, attendance_daily.event_date").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id").
		Where("attendance_daily.person_id = ? AND attendance_event_details.sub_type IN ?",
			personID, []string{"补班出勤", "调休"}).
		Order("attendance_daily.event_date ASC").
		Rows()
	if rows != nil {
		for rows.Next() {
			var hours float64
			var subType string
			var eventDate utils.DateOnly
			rows.Scan(&hours, &subType, &eventDate)
			events = append(events, model.AttendanceEventDetail{Hours: hours, SubType: subType})
			eventDates = append(eventDates, eventDate)
		}
		rows.Close()
	}

	if len(events) == 0 {
		return nil
	}

	type ch struct {
		date  utils.DateOnly
		hours float64
	}

	var changes []ch
	for i, e := range events {
		h := e.Hours
		if e.SubType == "调休" {
			h = -h
		}
		changes = append(changes, ch{eventDates[i], h})
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].date.Time().Before(changes[j].date.Time())
	})

	var snapshots []model.LeaveInLieuBalanceSnapshot
	var runningBalance float64
	var snapshotStart utils.DateOnly
	firstChange := true

	for _, ch := range changes {
		if !firstChange {
			endDate := ch.date.AddDate(0, 0, -1)
			if !endDate.Before(snapshotStart) {
				snapshots = append(snapshots, makeLILSnapshot(personID, snapshotStart, endDate, runningBalance))
			}
			snapshotStart = ch.date
		} else {
			snapshotStart = ch.date
			firstChange = false
		}
		runningBalance += ch.hours
	}

	t, _ := utils.ParseDate("9999-12-31")
	snapshots = append(snapshots, makeLILSnapshot(personID, snapshotStart, utils.DateOnlyFromTime(t), runningBalance))

	for _, s := range snapshots {
		if err := tx.Create(&s).Error; err != nil {
			return err
		}
	}
	return nil
}

func makeLILSnapshot(personID uint, start, end utils.DateOnly, balance float64) model.LeaveInLieuBalanceSnapshot {
	return model.LeaveInLieuBalanceSnapshot{
		PersonID:           personID,
		EffectiveStartDate: start,
		EffectiveEndDate:   end,
		BalanceHours:       balance,
		LastCalcAt:         utils.DateOnlyFromTime(time.Now()),
	}
}

func GetCurrentLILBalance(personID uint) (*model.LeaveInLieuBalanceSnapshot, error) {
	var snapshot model.LeaveInLieuBalanceSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date DESC").First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func GetLILBalanceHistory(personID uint) ([]model.LeaveInLieuBalanceSnapshot, error) {
	var snapshots []model.LeaveInLieuBalanceSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date ASC").Find(&snapshots).Error
	return snapshots, err
}

type BalanceListItem struct {
	PersonID     uint    `json:"person_id"`
	PersonName   string  `json:"person_name"`
	BalanceHours float64 `json:"balance_hours"`
	LastCalcAt   string  `json:"last_calc_at"`
}

func GetAllALBalances(pageNum, pageSize int, personID uint) ([]BalanceListItem, int64, error) {
	baseTx := dao.DB.Model(&model.AnnualLeaveBalanceSnapshot{}).
		Select("annual_leave_balance_snapshots.person_id, persons.name, annual_leave_balance_snapshots.balance_hours, annual_leave_balance_snapshots.last_calc_at").
		Joins("LEFT JOIN persons ON persons.id = annual_leave_balance_snapshots.person_id").
		Where("annual_leave_balance_snapshots.effective_end_date = ? AND persons.deleted_at IS NULL", realFarFuture)
	if personID > 0 {
		baseTx = baseTx.Where("annual_leave_balance_snapshots.person_id = ?", personID)
	}
	var total int64
	baseTx.Count(&total)
	offset := (pageNum - 1) * pageSize
	var list []BalanceListItem
	baseTx.Offset(offset).Limit(pageSize).Order("annual_leave_balance_snapshots.person_id ASC").Scan(&list)
	return list, total, nil
}

func GetAllLILBalances(pageNum, pageSize int, personID uint) ([]BalanceListItem, int64, error) {
	baseTx := dao.DB.Model(&model.LeaveInLieuBalanceSnapshot{}).
		Select("leave_in_lieu_balance_snapshots.person_id, persons.name, leave_in_lieu_balance_snapshots.balance_hours, leave_in_lieu_balance_snapshots.last_calc_at").
		Joins("LEFT JOIN persons ON persons.id = leave_in_lieu_balance_snapshots.person_id").
		Where("leave_in_lieu_balance_snapshots.effective_end_date = ? AND persons.deleted_at IS NULL", realFarFuture)
	if personID > 0 {
		baseTx = baseTx.Where("leave_in_lieu_balance_snapshots.person_id = ?", personID)
	}
	var total int64
	baseTx.Count(&total)
	offset := (pageNum - 1) * pageSize
	var list []BalanceListItem
	baseTx.Offset(offset).Limit(pageSize).Order("leave_in_lieu_balance_snapshots.person_id ASC").Scan(&list)
	return list, total, nil
}

type ALBalanceDetail struct {
	Grant    float64 `json:"grant"`
	Consumed float64 `json:"consumed"`
	Adjust   float64 `json:"adjust"`
	Carryover float64 `json:"carryover"`
	Balance  float64 `json:"balance"`
}

func GetAnnualLeaveBalanceDetail(personID uint) (*ALBalanceDetail, error) {
	var accountEvents []model.AnnualLeaveAccountEvent
	dao.DB.Where("person_id = ?", personID).Find(&accountEvents)

	var attendEvents []model.AttendanceEventDetail
	dao.DB.Table("attendance_event_details").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id").
		Where("attendance_daily.person_id = ? AND attendance_event_details.event_type = ? AND attendance_event_details.sub_type = ?", personID, "休假", "年假").
		Select("attendance_event_details.hours").
		Scan(&attendEvents)

	detail := &ALBalanceDetail{}
	for _, e := range accountEvents {
		switch e.EventType {
		case "grant":
			detail.Grant += e.Hours
		case "adjust":
			detail.Adjust += e.Hours
		case "carryover_deduct":
			detail.Carryover += e.Hours
		}
	}
	for _, e := range attendEvents {
		detail.Consumed += e.Hours
	}
	detail.Balance = detail.Grant + detail.Adjust - detail.Carryover - detail.Consumed
	return detail, nil
}

type LILBalanceDetail struct {
	Makeup  float64 `json:"makeup"`
	Consumed float64 `json:"consumed"`
	Balance float64 `json:"balance"`
}

func GetLILBalanceDetail(personID uint) (*LILBalanceDetail, error) {
	var events []model.AttendanceEventDetail
	dao.DB.Table("attendance_event_details").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id").
		Where("attendance_daily.person_id = ? AND attendance_event_details.sub_type IN ?", personID, []string{"补班出勤", "调休"}).
		Select("attendance_event_details.hours, attendance_event_details.sub_type").
		Scan(&events)

	detail := &LILBalanceDetail{}
	for _, e := range events {
		if e.SubType == "补班出勤" {
			detail.Makeup += e.Hours
		} else {
			detail.Consumed += e.Hours
		}
	}
	detail.Balance = detail.Makeup - detail.Consumed
	return detail, nil
}
