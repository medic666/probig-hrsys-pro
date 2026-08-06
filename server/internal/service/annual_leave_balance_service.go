package service

import (
	"sort"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// 年假余额变更的过账类别优先级（财会"过账分类"思想的轻量落地）：
// 同日多变更的次序由类别优先级决定，而非依赖 seq 跨表计数——
//   0 系统配发（周年月首日，先于当月考勤，当天休假用新额度）
//   1 人工调整（manual grant/adjust）
//   2 考勤年假消费（日常业务）
//   3 系统结转 deduct（结算月末，后于当月考勤，先扣消费再结余）
const (
	alPriorityGrant   = 0
	alPriorityManual  = 1
	alPriorityConsume = 2
	alPriorityDeduct  = 3
)

type alChange struct {
	date     utils.DateOnly
	priority int
	seq      int
	hours    float64
}

func RebuildAnnualLeaveBalance(tx *gorm.DB, personID uint) error {
	tx.Where("person_id = ?", personID).Delete(&model.AnnualLeaveBalanceSnapshot{})

	var accountEvents []model.AnnualLeaveAccountEvent
	tx.Where("person_id = ?", personID).Order("effective_date ASC, seq ASC").Find(&accountEvents)

	var attendEvents []model.AttendanceEventDetail
	var attendDates []utils.DateOnly
	rows2, _ := tx.Table("attendance_event_details").
		Select("attendance_event_details.hours, attendance_daily.event_date").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id AND attendance_daily.deleted_at IS NULL AND attendance_daily.status = 'confirmed'").
		Where("attendance_event_details.deleted_at IS NULL AND attendance_daily.person_id = ? AND attendance_event_details.event_type = ? AND attendance_event_details.sub_type = ?",
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

	var changes []alChange
	for _, e := range accountEvents {
		h := e.Hours
		priority := alPriorityManual
		switch e.EventType {
		case "carryover_deduct":
			h = -h
			priority = alPriorityDeduct
		case "grant":
			if e.SourceType == "system_period" {
				priority = alPriorityGrant
			}
		}
		changes = append(changes, alChange{e.EffectiveDate, priority, e.Seq, h})
	}
	for i, e := range attendEvents {
		changes = append(changes, alChange{attendDates[i], alPriorityConsume, 0, -e.Hours})
	}

	if len(changes) == 0 {
		return nil
	}

	sort.SliceStable(changes, func(i, j int) bool {
		if !changes[i].date.Equal(changes[j].date) {
			return changes[i].date.Time().Before(changes[j].date.Time())
		}
		if changes[i].priority != changes[j].priority {
			return changes[i].priority < changes[j].priority
		}
		return changes[i].seq < changes[j].seq
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
		LastCalcAt: time.Now(),
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

// GetAnnualLeaveBalanceAt 查询指定日期末的年假余额（快照段覆盖判定）。
// 结转结算/徽章判定共用：结算月最后一日末余额 = 该日所在快照段的余额。
func GetAnnualLeaveBalanceAt(tx *gorm.DB, personID uint, date utils.DateOnly) (float64, bool) {
	var snap model.AnnualLeaveBalanceSnapshot
	err := tx.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
		personID, date, date).First(&snap).Error
	if err != nil {
		return 0, false
	}
	return snap.BalanceHours, true
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
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id AND attendance_daily.deleted_at IS NULL AND attendance_daily.status = 'confirmed'").
		Where("attendance_event_details.deleted_at IS NULL AND attendance_daily.person_id = ? AND attendance_event_details.sub_type IN ?",
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

	// 调休无系统事件，同日期按序累加（稳定排序保序，语义无歧义）
	sort.SliceStable(changes, func(i, j int) bool {
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
		LastCalcAt: time.Now(),
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
	baseTx := dao.DB.Table("persons").
		Select(`persons.id AS person_id, persons.name AS person_name,
			COALESCE(s.balance_hours, 0) AS balance_hours, s.last_calc_at`).
		Joins(`LEFT JOIN annual_leave_balance_snapshots s
			ON s.person_id = persons.id AND s.effective_end_date = ?`, realFarFuture).
		Where("persons.deleted_at IS NULL")
	if personID > 0 {
		baseTx = baseTx.Where("persons.id = ?", personID)
	}
	var total int64
	baseTx.Count(&total)
	offset := (pageNum - 1) * pageSize
	var list []BalanceListItem
	baseTx.Offset(offset).Limit(pageSize).Order("persons.name").Scan(&list)
	return list, total, nil
}

func GetAllLILBalances(pageNum, pageSize int, personID uint) ([]BalanceListItem, int64, error) {
	baseTx := dao.DB.Table("persons").
		Select(`persons.id AS person_id, persons.name AS person_name,
			COALESCE(s.balance_hours, 0) AS balance_hours, s.last_calc_at`).
		Joins(`LEFT JOIN leave_in_lieu_balance_snapshots s
			ON s.person_id = persons.id AND s.effective_end_date = ?`, realFarFuture).
		Where("persons.deleted_at IS NULL")
	if personID > 0 {
		baseTx = baseTx.Where("persons.id = ?", personID)
	}
	var total int64
	baseTx.Count(&total)
	offset := (pageNum - 1) * pageSize
	var list []BalanceListItem
	baseTx.Offset(offset).Limit(pageSize).Order("persons.name").Scan(&list)
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
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id AND attendance_daily.deleted_at IS NULL AND attendance_daily.status = 'confirmed'").
		Where("attendance_event_details.deleted_at IS NULL AND attendance_daily.person_id = ? AND attendance_event_details.event_type = ? AND attendance_event_details.sub_type = ?", personID, "休假", "年假").
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

// LILEventListItem 调休事件行（补班出勤/调休）
type LILEventListItem struct {
	ID         uint           `json:"id"`
	DailyID    uint           `json:"daily_id"`
	PersonID   uint           `json:"person_id"`
	PersonName string         `json:"person_name"`
	EventDate  utils.DateOnly `json:"event_date"`
	EventType  string         `json:"event_type"`
	SubType    string         `json:"sub_type"`
	Hours      float64        `json:"hours"`
	Remark     string         `json:"remark"`
}

// GetLILEventList 调休事件明细级分页查询（补班出勤/调休）：
// 按「已确认考勤组」的明细行过滤（同日多版本仅最新确认组参与），先过滤后分页，
// 避免"先对考勤日分页、再页内过滤明细"导致列表缺失。
func GetLILEventList(q AttendanceDailyListQuery) ([]LILEventListItem, int64, error) {
	tx := dao.DB.Table("attendance_event_details d").
		Joins("JOIN attendance_daily a ON a.id = d.daily_id AND a.deleted_at IS NULL AND a.status = 'confirmed'").
		Where("d.deleted_at IS NULL AND d.sub_type IN ?", []string{"补班出勤", "调休"})
	if q.PersonID > 0 {
		tx = tx.Where("a.person_id = ?", q.PersonID)
	}
	if q.DateStart != "" {
		tx = tx.Where("a.event_date >= ?", q.DateStart)
	}
	if q.DateEnd != "" {
		tx = tx.Where("a.event_date <= ?", q.DateEnd)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []struct {
		ID        uint
		DailyID   uint
		PersonID  uint
		EventDate utils.DateOnly
		EventType string
		SubType   string
		Hours     float64
		Remark    string
	}
	offset := (q.PageNum - 1) * q.PageSize
	if err := tx.Select("d.id, d.daily_id, a.person_id, a.event_date, d.event_type, d.sub_type, d.hours, d.remark").
		Order("a.event_date DESC, d.id DESC").
		Offset(offset).Limit(q.PageSize).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	ids := make([]uint, len(rows))
	for i, r := range rows {
		ids[i] = r.PersonID
	}
	nameMap := PersonNameMap(ids)

	result := make([]LILEventListItem, len(rows))
	for i, r := range rows {
		result[i] = LILEventListItem{
			ID:         r.ID,
			DailyID:    r.DailyID,
			PersonID:   r.PersonID,
			PersonName: nameMap[r.PersonID],
			EventDate:  r.EventDate,
			EventType:  r.EventType,
			SubType:    r.SubType,
			Hours:      r.Hours,
			Remark:     r.Remark,
		}
	}
	return result, total, nil
}

func GetLILBalanceDetail(personID uint) (*LILBalanceDetail, error) {
	var events []model.AttendanceEventDetail
	dao.DB.Table("attendance_event_details").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id AND attendance_daily.deleted_at IS NULL AND attendance_daily.status = 'confirmed'").
		Where("attendance_event_details.deleted_at IS NULL AND attendance_daily.person_id = ? AND attendance_event_details.sub_type IN ?", personID, []string{"补班出勤", "调休"}).
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
