package service

import (
	"errors"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreateAttendanceEvent(event *model.AttendanceEvent) error {
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		var maxSeq int
		tx.Unscoped().Model(&model.AttendanceEvent{}).Where("person_id = ?", event.PersonID).
			Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
		event.Seq = maxSeq + 1

		if err := tx.Create(event).Error; err != nil {
			return err
		}
		if err := RebuildDailyProjection(tx, event.PersonID, event.EventDate); err != nil {
			return err
		}
		return triggerLeaveRebuild(tx, event)
	})
}

func UpdateAttendanceEvent(id uint, event *model.AttendanceEvent) error {
	var existing model.AttendanceEvent
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("考勤事件不存在")
	}
	oldDate := existing.EventDate
	oldType := existing.EventType
	oldSubType := existing.SubType
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"event_date": event.EventDate,
			"punch_time": event.PunchTime,
			"event_type": event.EventType,
			"sub_type":   event.SubType,
			"hours":      event.Hours,
			"remark":     event.Remark,
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		if err := RebuildDailyProjection(tx, existing.PersonID, oldDate); err != nil {
			return err
		}
		if event.EventDate != oldDate {
			if err := RebuildDailyProjection(tx, existing.PersonID, event.EventDate); err != nil {
				return err
			}
		}
		oldEvent := model.AttendanceEvent{EventType: oldType, SubType: oldSubType, PersonID: existing.PersonID}
		if err := triggerLeaveRebuild(tx, &oldEvent); err != nil {
			return err
		}
		return triggerLeaveRebuild(tx, event)
	})
}

func DeleteAttendanceEvent(id uint) error {
	var event model.AttendanceEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		if err := RebuildDailyProjection(tx, event.PersonID, event.EventDate); err != nil {
			return err
		}
		return triggerLeaveRebuild(tx, &event)
	})
}

func RestoreAttendanceEvent(id uint) error {
	var event model.AttendanceEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := RebuildDailyProjection(tx, event.PersonID, event.EventDate); err != nil {
			return err
		}
		return triggerLeaveRebuild(tx, &event)
	})
}

func triggerLeaveRebuild(tx *gorm.DB, event *model.AttendanceEvent) error {
	if event.EventType == "休假" && event.SubType == "年假" {
		if err := RebuildAnnualLeaveBalance(tx, event.PersonID); err != nil {
			return err
		}
	}
	if event.SubType == "补班出勤" || event.SubType == "调休" {
		if err := RebuildLeaveInLieuBalance(tx, event.PersonID); err != nil {
			return err
		}
	}
	return nil
}

func GetAttendanceEvent(id uint) (*model.AttendanceEvent, error) {
	var event model.AttendanceEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func GetAttendanceEventList(pageNum, pageSize int, personID uint, dateStart, dateEnd, eventType, subType string) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.AttendanceEvent{})
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	if dateStart != "" {
		tx = tx.Where("event_date >= ?", dateStart)
	}
	if dateEnd != "" {
		tx = tx.Where("event_date <= ?", dateEnd)
	}
	if eventType != "" {
		tx = tx.Where("event_type = ?", eventType)
	}
	if subType != "" {
		tx = tx.Where("sub_type = ?", subType)
	}

	var total int64
	tx.Count(&total)

	var events []model.AttendanceEvent
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("person_id ASC, event_date DESC, seq DESC").Find(&events)

	var result []map[string]interface{}
	for _, e := range events {
		item := map[string]interface{}{
			"id":         e.ID,
			"person_id":  e.PersonID,
			"seq":        e.Seq,
			"event_date": e.EventDate,
			"punch_time": e.PunchTime,
			"event_type": e.EventType,
			"sub_type":   e.SubType,
			"hours":      e.Hours,
			"remark":     e.Remark,
			"created_at": e.CreatedAt,
		}
		var personName string
		dao.DB.Table("persons").Select("name").Where("id = ?", e.PersonID).Scan(&personName)
		item["person_name"] = personName
		result = append(result, item)
	}
	return result, total, nil
}

func GetAttendanceEventsByPersonDate(personID uint, date string) ([]model.AttendanceEvent, error) {
	var events []model.AttendanceEvent
	err := dao.DB.Where("person_id = ? AND event_date = ?", personID, date).Find(&events).Error
	return events, err
}

type BatchAttendanceReq struct {
	PersonIDs  []uint `json:"person_ids"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	EventType  string `json:"event_type"`
	SubType    string `json:"sub_type"`
	Hours      float64 `json:"hours"`
	PunchTime  string `json:"punch_time"`
	Remark     string `json:"remark"`
}

func CreateBatchAttendanceEvents(req BatchAttendanceReq) (int, int, error) {
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	if end.Before(start) {
		return 0, 0, errors.New("结束日期不能早于开始日期")
	}

	success := 0
	fail := 0

	for _, pid := range req.PersonIDs {
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			eventDate := utils.DateOnlyFromTime(d)
			event := &model.AttendanceEvent{
				PersonID:  pid,
				EventDate: eventDate,
				EventType: req.EventType,
				SubType:   req.SubType,
				Hours:     req.Hours,
				PunchTime: req.PunchTime,
				Remark:    req.Remark,
			}
			if err := CreateAttendanceEvent(event); err != nil {
				fail++
			} else {
				success++
			}
		}
	}

	return success, fail, nil
}

func GetDeletedAttendanceEvents(pageNum, pageSize int) ([]model.AttendanceEvent, int64, error) {
	var list []model.AttendanceEvent
	var total int64
	tx := dao.DB.Unscoped().Model(&model.AttendanceEvent{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}

func Atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
