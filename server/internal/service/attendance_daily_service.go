package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func GetOrCreateDaily(tx *gorm.DB, personID uint, eventDate utils.DateOnly, status string) (*model.AttendanceDaily, error) {
	var existing model.AttendanceDaily
	err := tx.Where("person_id = ? AND event_date = ?", personID, eventDate).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	daily := model.AttendanceDaily{PersonID: personID, EventDate: eventDate, Status: status}
	if err := tx.Create(&daily).Error; err != nil {
		return nil, err
	}
	return &daily, nil
}

func GetDetailsByDailyID(tx *gorm.DB, dailyID uint) ([]model.AttendanceEventDetail, error) {
	var details []model.AttendanceEventDetail
	if err := tx.Where("daily_id = ?", dailyID).Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

// RebuildProjectionsAfterAttendanceChange 考勤事件变动后的统一投影重算入口。
// 日记工时必然重建；年假/调休余额按涉及的事件子类型精细触发全量重建。
func RebuildProjectionsAfterAttendanceChange(tx *gorm.DB, personID uint, workDate utils.DateOnly, involved []model.AttendanceEventDetail) error {
	if err := RebuildDailyProjection(tx, personID, workDate); err != nil {
		return err
	}
	var needAL, needLIL bool
	for _, d := range involved {
		switch d.SubType {
		case "年假":
			needAL = true
		case "补班出勤", "调休":
			needLIL = true
		}
	}
	if needAL {
		if err := RebuildAnnualLeaveBalance(tx, personID); err != nil {
			return err
		}
	}
	if needLIL {
		if err := RebuildLeaveInLieuBalance(tx, personID); err != nil {
			return err
		}
	}
	return nil
}

func CreateDetail(tx *gorm.DB, dailyID uint, eventType, subType string, hours float64, minutes int, remark string) error {
	d := model.AttendanceEventDetail{
		DailyID: dailyID, EventType: eventType, SubType: subType,
		Hours: hours, Minutes: minutes, Remark: remark,
	}
	return tx.Create(&d).Error
}

func UpdateDailyDetails(tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail, status string) error {
	// 同步（UPSERT）模式：未变化明细零操作零审计；变化更新、新增创建、缺失软删除
	if err := SyncChildRecords(tx, "daily_id", dailyID, details,
		func(d model.AttendanceEventDetail) uint { return d.ID },
		func(a, b model.AttendanceEventDetail) bool {
			return a.EventType == b.EventType && a.SubType == b.SubType &&
				a.Hours == b.Hours && a.Minutes == b.Minutes && a.Remark == b.Remark
		},
		func(d *model.AttendanceEventDetail) { d.DailyID = dailyID },
	); err != nil {
		return err
	}
	return tx.Model(&model.AttendanceDaily{}).Where("id = ?", dailyID).Update("status", status).Error
}

// SaveDailyDetailsKeepStatus 暂存保存当日事件明细：更新 details 但保持原状态不变（供编辑后暂存）
func SaveDailyDetailsKeepStatus(tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail) error {
	return SyncChildRecords(tx, "daily_id", dailyID, details,
		func(d model.AttendanceEventDetail) uint { return d.ID },
		func(a, b model.AttendanceEventDetail) bool {
			return a.EventType == b.EventType && a.SubType == b.SubType &&
				a.Hours == b.Hours && a.Minutes == b.Minutes && a.Remark == b.Remark
		},
		func(d *model.AttendanceEventDetail) { d.DailyID = dailyID },
	)
}

func GetAttendanceDailyList(personID uint, dateStart, dateEnd string, status string, pageNum, pageSize int) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.AttendanceDaily{}).Preload("Details")
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	if dateStart != "" {
		tx = tx.Where("event_date >= ?", dateStart)
	}
	if dateEnd != "" {
		tx = tx.Where("event_date <= ?", dateEnd)
	}
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	var total int64
	tx.Count(&total)

	var list []model.AttendanceDaily
	offset := (pageNum - 1) * pageSize
	tx.Order("event_date DESC, person_id ASC").Offset(offset).Limit(pageSize).Find(&list)

	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.PersonID
	}
	nameMap := PersonNameMap(ids)

	result := make([]map[string]interface{}, len(list))
	for i, d := range list {
		item := map[string]interface{}{
			"id": d.ID, "person_id": d.PersonID, "event_date": d.EventDate,
			"status": d.Status, "punch_time": d.PunchTime, "remark": d.Remark,
			"created_at": d.CreatedAt,
		}
		item["person_name"] = nameMap[d.PersonID]
		detailList := make([]map[string]interface{}, len(d.Details))
		for j, dt := range d.Details {
			detailList[j] = map[string]interface{}{
				"id": dt.ID, "event_type": dt.EventType, "sub_type": dt.SubType,
				"hours": dt.Hours, "minutes": dt.Minutes, "remark": dt.Remark,
			}
		}
		item["details"] = detailList
		result[i] = item
	}
	return result, total, nil
}

func GetPendingDailyList(pageNum, pageSize int, personID uint) ([]map[string]interface{}, int64, error) {
	return GetAttendanceDailyList(personID, "", "", "pending", pageNum, pageSize)
}

// detailSnapshots 将考勤事件明细转成审计快照（不含技术字段）
func detailSnapshots(details []model.AttendanceEventDetail) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(details))
	for _, d := range details {
		result = append(result, map[string]interface{}{
			"event_type": d.EventType, "sub_type": d.SubType,
			"hours": d.Hours, "minutes": d.Minutes, "remark": d.Remark,
		})
	}
	return result
}

func writeConfirmAudit(ctx context.Context, tx *gorm.DB, daily model.AttendanceDaily, oldDetails, newDetails []model.AttendanceEventDetail) {
	var personName string
	tx.Table("persons").Select("name").Where("id = ?", daily.PersonID).Scan(&personName)
	before, _ := json.Marshal(detailSnapshots(oldDetails))
	after, _ := json.Marshal(detailSnapshots(newDetails))
	// 复用业务事务连接写入，随事务提交/回滚（审计=实际发生的操作）
	info := dao.AuditFromContext(ctx)
	tx.Create(&model.AuditLog{
		OperatorID:     info.OperatorID,
		OperatorName:   info.OperatorName,
		TargetType:     "attendance_daily",
		TargetID:       daily.ID,
		TargetName:     fmt.Sprintf("%s %s", personName, daily.EventDate.String()),
		Action:         "确认",
		BeforeSnapshot: string(before),
		AfterSnapshot:  string(after),
		IP:             info.IP,
	})
}

func ConfirmDaily(ctx context.Context, tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail) error {
	var daily model.AttendanceDaily
	if err := tx.First(&daily, dailyID).Error; err != nil {
		return err
	}
	oldDetails, err := GetDetailsByDailyID(tx, dailyID)
	if err != nil {
		return err
	}
	if err := UpdateDailyDetails(tx, dailyID, details, "confirmed"); err != nil {
		return err
	}
	if err := RebuildProjectionsAfterAttendanceChange(tx, daily.PersonID, daily.EventDate, append(oldDetails, details...)); err != nil {
		return err
	}
	writeConfirmAudit(ctx, tx, daily, oldDetails, details)
	return nil
}

func ConfirmDailyBatch(ctx context.Context, tx *gorm.DB, dailyIDs []uint) error {
	for _, id := range dailyIDs {
		var daily model.AttendanceDaily
		if err := tx.First(&daily, id).Error; err != nil {
			return err
		}
		details, err := GetDetailsByDailyID(tx, id)
		if err != nil {
			return err
		}
		if err := tx.Model(&daily).Update("status", "confirmed").Error; err != nil {
			return err
		}
		if err := RebuildProjectionsAfterAttendanceChange(tx, daily.PersonID, daily.EventDate, details); err != nil {
			return err
		}
		writeConfirmAudit(ctx, tx, daily, details, details)
	}
	return nil
}

func GetDeletedAttendanceDailies(pageNum, pageSize int) ([]model.AttendanceDaily, int64, error) {
	var list []model.AttendanceDaily
	var total int64
	tx := dao.DB.Unscoped().Model(&model.AttendanceDaily{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}

func DeleteAttendanceDaily(ctx context.Context, id uint) error {
	var daily model.AttendanceDaily
	if err := dao.DB.First(&daily, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		details, err := GetDetailsByDailyID(tx, id)
		if err != nil {
			return err
		}
		if err := tx.Where("daily_id = ?", id).Delete(&model.AttendanceEventDetail{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&daily).Error; err != nil {
			return err
		}
		return RebuildProjectionsAfterAttendanceChange(tx, daily.PersonID, daily.EventDate, details)
	})
}

func RestoreAttendanceDaily(ctx context.Context, id uint) error {
	var daily model.AttendanceDaily
	if err := dao.DB.Unscoped().First(&daily, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&daily).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&model.AttendanceEventDetail{}).Where("daily_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		details, err := GetDetailsByDailyID(tx, id)
		if err != nil {
			return err
		}
		return RebuildProjectionsAfterAttendanceChange(tx, daily.PersonID, daily.EventDate, details)
	})
}

type BatchAttendanceReq struct {
	PersonIDs []uint  `json:"person_ids"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	EventType string  `json:"event_type"`
	SubType   string  `json:"sub_type"`
	Hours     float64 `json:"hours"`
	PunchTime string  `json:"punch_time"`
	Remark    string  `json:"remark"`
}

func CreateBatchAttendanceDailies(ctx context.Context, req BatchAttendanceReq) (int, int, error) {
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	if end.Before(start) {
		return 0, 0, errors.New("结束日期不能早于开始日期")
	}
	involved := []model.AttendanceEventDetail{{EventType: req.EventType, SubType: req.SubType}}
	success, fail := 0, 0
	for _, pid := range req.PersonIDs {
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateOnly := utils.DateOnlyFromTime(d)
			err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
				daily, err := GetOrCreateDaily(tx, pid, dateOnly, "confirmed")
				if err != nil {
					return err
				}
				if err := CreateDetail(tx, daily.ID, req.EventType, req.SubType, req.Hours, 0, req.Remark); err != nil {
					return err
				}
				return RebuildProjectionsAfterAttendanceChange(tx, pid, dateOnly, involved)
			})
			if err != nil {
				fail++
			} else {
				success++
			}
		}
	}
	return success, fail, nil
}

func GetAttendanceDailyByID(id uint) (*model.AttendanceDaily, error) {
	var daily model.AttendanceDaily
	if err := dao.DB.First(&daily, id).Error; err != nil {
		return nil, err
	}
	return &daily, nil
}
