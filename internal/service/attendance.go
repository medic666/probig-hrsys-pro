package service

import (
	"errors"
	"fmt"
	"time"

	"probig/internal/dao"
	"probig/internal/middleware"
	"probig/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAttendanceEventList(page, pageSize int, personID uint, eventType, subType, startDate, endDate string) ([]models.AttendanceEvent, int64, error) {
	return dao.GetAttendanceEventList(page, pageSize, personID, eventType, subType, startDate, endDate)
}

func GetAttendanceEvent(id uint) (*models.AttendanceEvent, error) {
	return dao.GetAttendanceEventByID(id)
}

func CreateAttendanceEvent(c *gin.Context, e *models.AttendanceEvent) error {
	if e.EventType == "年假" && e.SubType == "年假" {
		if err := validateAnnualLeave(e.PersonID, e.EventDate, e.Hours, e.IsSpecialApproval); err != nil {
			return err
		}
	}
	if e.EventType == "休假" && e.SubType == "调休" {
		if err := validateOvertimeLeave(e.PersonID, e.Hours, e.IsSpecialApproval); err != nil {
			return err
		}
	}
	if err := dao.CreateAttendanceEvent(e); err != nil {
		return err
	}
	middleware.RecordAudit(c, "新增", "attendance_event", e.ID, nil, e, e.BatchID)
	return nil
}

func CreateAttendanceEventRange(c *gin.Context, personID uint, startDate, endDate string, eventType, subType string, hours *float64, remark string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return errors.New("日期格式错误")
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return errors.New("日期格式错误")
	}
	if end.Before(start) {
		return errors.New("结束日期不能早于开始日期")
	}
	batchID := uuid.New().String()
	var events []models.AttendanceEvent
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date := d
		e := models.AttendanceEvent{
			PersonID:  personID,
			EventDate: &date,
			EventType: eventType,
			SubType:   subType,
			Hours:     hours,
			Remark:    remark,
			BatchID:   batchID,
		}
		events = append(events, e)
	}
	if err := dao.CreateAttendanceEvents(events); err != nil {
		return err
	}
	middleware.RecordAuditBatch(c, "批量新增", "attendance_event", events, nil, batchID)
	return nil
}

func BatchCreateAttendanceEvents(c *gin.Context, personIDs []uint, eventDate string, eventType, subType string, hours *float64, remark string) error {
	batchID := uuid.New().String()
	date, err := time.Parse("2006-01-02", eventDate)
	if err != nil {
		return errors.New("日期格式错误")
	}
	var events []models.AttendanceEvent
	for _, pid := range personIDs {
		d := date
		e := models.AttendanceEvent{
			PersonID:  pid,
			EventDate: &d,
			EventType: eventType,
			SubType:   subType,
			Hours:     hours,
			Remark:    remark,
			BatchID:   batchID,
		}
		events = append(events, e)
	}
	if err := dao.CreateAttendanceEvents(events); err != nil {
		return err
	}
	middleware.RecordAuditBatch(c, "批量新增", "attendance_event", events, nil, batchID)
	return nil
}

func UpdateAttendanceEvent(c *gin.Context, e *models.AttendanceEvent) error {
	old, err := dao.GetAttendanceEventByID(e.ID)
	if err != nil {
		return err
	}
	if err := dao.UpdateAttendanceEvent(e); err != nil {
		return err
	}
	middleware.RecordAudit(c, "修改", "attendance_event", e.ID, old, e, "")
	return nil
}

func DeleteAttendanceEvent(c *gin.Context, id uint) error {
	e, err := dao.GetAttendanceEventByID(id)
	if err != nil {
		return err
	}
	locked, _ := dao.GetLockedMonthsForPerson(e.PersonID)
	monthKey := e.EventDate.Format("2006-01")
	for _, m := range locked {
		if m == monthKey {
			return fmt.Errorf("人员[%d]的[%s]月份考勤已锁定，无法修改", e.PersonID, monthKey)
		}
	}
	if err := dao.DeleteAttendanceEvent(id); err != nil {
		return err
	}
	middleware.RecordAudit(c, "删除", "attendance_event", id, e, nil, "")
	return nil
}

func validateAnnualLeave(personID uint, eventDate *time.Time, hours *float64, isSpecialApproval bool) error {
	snapshot, err := dao.GetPositionSnapshot(personID, *eventDate)
	if err != nil || !snapshot.HasAnnualLeave {
		return errors.New("该员工不享有年假资格")
	}
	if isSpecialApproval {
		return nil
	}
	if ConfigCache["attendance.special_approval"] != "true" {
		return errors.New("特批功能未开启")
	}
	balance := calculateAnnualLeaveBalance(personID)
	needDays := float64(0)
	if hours != nil {
		needDays = *hours / 8
	}
	if balance < needDays {
		return errors.New("年假余额不足")
	}
	return nil
}

func calculateAnnualLeaveBalance(personID uint) float64 {
	return 999
}

func validateOvertimeLeave(personID uint, hours *float64, isSpecialApproval bool) error {
	if ConfigCache["attendance.overtime_control"] != "true" {
		return nil
	}
	return nil
}
