package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreateAnnualLeaveEvent(ctx context.Context, event *model.AnnualLeaveAccountEvent) error {
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		var maxSeq int
		tx.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ?", event.PersonID).
			Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
		event.Seq = maxSeq + 1
		event.SourceType = "manual"

		if err := tx.Create(event).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, event.PersonID)
	})
}

func UpdateAnnualLeaveEvent(ctx context.Context, id uint, event *model.AnnualLeaveAccountEvent) error {
	var existing model.AnnualLeaveAccountEvent
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("年假权益事件不存在")
	}
	if existing.SourceType == "system_period" {
		return errors.New("系统周期事件不可编辑")
	}
	if err := EnsureOwnPerson(ctx, existing.PersonID); err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"event_type":     event.EventType,
			"hours":          event.Hours,
			"effective_date": event.EffectiveDate,
			"remark":         event.Remark,
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, existing.PersonID)
	})
}

func DeleteAnnualLeaveEvent(ctx context.Context, id uint) error {
	var event model.AnnualLeaveAccountEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	if event.SourceType == "system_period" {
		return errors.New("系统周期事件不可删除")
	}
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, event.PersonID)
	})
}

func RestoreAnnualLeaveEvent(ctx context.Context, id uint) error {
	var event model.AnnualLeaveAccountEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	if event.SourceType == "system_period" {
		return errors.New("系统周期事件不可人工恢复")
	}
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return RebuildAnnualLeaveBalance(tx, event.PersonID)
	})
}

func GetAnnualLeaveEvent(ctx context.Context, id uint) (*model.AnnualLeaveAccountEvent, error) {
	var event model.AnnualLeaveAccountEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return nil, err
	}
	return &event, nil
}

// AnnualLeaveListQuery 年假事件列表查询（列表与导出共用）
type AnnualLeaveListQuery struct {
	PageNum   int
	PageSize  int
	PersonID  uint
	DateStart string
	DateEnd   string
	EventType string
}

func GetAnnualLeaveEventList(ctx context.Context, q AnnualLeaveListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DBFromContext(ctx).Model(&model.AnnualLeaveAccountEvent{})
	tx = OwnFilter(ctx, tx, "person_id")
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}
	if q.DateStart != "" {
		tx = tx.Where("effective_date >= ?", q.DateStart)
	}
	if q.DateEnd != "" {
		tx = tx.Where("effective_date <= ?", q.DateEnd)
	}
	if q.EventType != "" {
		tx = tx.Where("event_type = ?", q.EventType)
	}

	var events []model.AnnualLeaveAccountEvent
	tx.Order("effective_date DESC, seq DESC").Find(&events)

	var attendanceLeaves []map[string]interface{}
	if q.EventType == "" || q.EventType == "休假" {
		attendanceLeaves = getAnnualLeaveAttendanceEvents(ctx, q.PersonID, q.DateStart, q.DateEnd)
	}

	ids := make([]uint, 0, len(events)+len(attendanceLeaves))
	for _, e := range events {
		ids = append(ids, e.PersonID)
	}
	for _, a := range attendanceLeaves {
		ids = append(ids, a["person_id"].(uint))
	}
	nameMap := PersonNameMap(ids)

	type item struct {
		date utils.DateOnly
		seq  int
		data map[string]interface{}
	}
	var items []item
	for _, e := range events {
		items = append(items, item{e.EffectiveDate, e.Seq, annualLeaveEventToMap(e, nameMap)})
	}
	for _, a := range attendanceLeaves {
		items = append(items, item{a["effective_date"].(utils.DateOnly), 0, a})
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].date.Equal(items[j].date) {
			return items[i].date.Time().After(items[j].date.Time())
		}
		return items[i].seq > items[j].seq
	})

	total := int64(len(items))
	start := (q.PageNum - 1) * q.PageSize
	end := start + q.PageSize
	if start > len(items) {
		start = len(items)
	}
	if end > len(items) {
		end = len(items)
	}
	var result []map[string]interface{}
	for _, it := range items[start:end] {
		result = append(result, it.data)
	}
	return result, total, nil
}

// getAnnualLeaveAttendanceEvents 考勤事件中的「休假-年假」确认记录，映射为年假事件同构项
func getAnnualLeaveAttendanceEvents(ctx context.Context, personID uint, dateStart, dateEnd string) []map[string]interface{} {
	type attRow struct {
		DetailID  uint
		DailyID   uint
		PersonID  uint
		Hours     float64
		EventDate utils.DateOnly
		Remark    string
		CreatedAt time.Time
	}
	var rows []attRow
	q := dao.DBFromContext(ctx).Table("attendance_event_details").
		Select("attendance_event_details.id AS detail_id, attendance_event_details.daily_id, attendance_daily.person_id, attendance_event_details.hours, attendance_daily.event_date, attendance_event_details.remark, attendance_event_details.created_at").
		Joins("JOIN attendance_daily ON attendance_daily.id = attendance_event_details.daily_id AND attendance_daily.deleted_at IS NULL AND attendance_daily.status = 'confirmed'").
		Where("attendance_event_details.deleted_at IS NULL AND attendance_event_details.event_type = ? AND attendance_event_details.sub_type = ?", "休假", "年假")
	if pid, ok := dao.OwnPersonID(ctx); ok {
		q = q.Where("attendance_daily.person_id = ?", pid)
	}
	if personID > 0 {
		q = q.Where("attendance_daily.person_id = ?", personID)
	}
	if dateStart != "" {
		q = q.Where("attendance_daily.event_date >= ?", dateStart)
	}
	if dateEnd != "" {
		q = q.Where("attendance_daily.event_date <= ?", dateEnd)
	}
	q.Scan(&rows)

	var result []map[string]interface{}
	ids := make([]uint, len(rows))
	for i, r := range rows {
		ids[i] = r.PersonID
	}
	nameMap := PersonNameMap(ids)

	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id": r.DetailID, "daily_id": r.DailyID, "person_id": r.PersonID, "seq": 0,
			"event_type": "休假", "sub_type": "年假", "source_type": "attendance",
			"batch_id": nil, "hours": r.Hours, "effective_date": r.EventDate,
			"remark": r.Remark, "created_at": r.CreatedAt, "person_name": nameMap[r.PersonID],
		})
	}
	return result
}

func annualLeaveEventToMap(e model.AnnualLeaveAccountEvent, nameMap map[uint]string) map[string]interface{} {
	item := map[string]interface{}{
		"id":             e.ID,
		"person_id":      e.PersonID,
		"seq":            e.Seq,
		"event_type":     e.EventType,
		"source_type":    e.SourceType,
		"batch_id":       e.BatchID,
		"hours":          e.Hours,
		"effective_date": e.EffectiveDate,
		"remark":         e.Remark,
		"created_at":     e.CreatedAt,
	}
	item["person_name"] = nameMap[e.PersonID]
	return item
}

func GetDeletedAnnualLeaveEvents(ctx context.Context, pageNum, pageSize int) ([]model.AnnualLeaveAccountEvent, int64, error) {
	var list []model.AnnualLeaveAccountEvent
	var total int64
	tx := dao.DBFromContext(ctx).Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("deleted_at IS NOT NULL")
	tx = OwnFilter(ctx, tx, "person_id")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}
