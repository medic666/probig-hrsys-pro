package service

import (
	"errors"
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

func CreateDetail(tx *gorm.DB, dailyID uint, eventType, subType string, hours float64, minutes int, remark string) error {
	d := model.AttendanceEventDetail{
		DailyID: dailyID, EventType: eventType, SubType: subType,
		Hours: hours, Minutes: minutes, Remark: remark,
	}
	return tx.Create(&d).Error
}

func UpdateDailyDetails(tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail, status string) error {
	if err := tx.Where("daily_id = ?", dailyID).Delete(&model.AttendanceEventDetail{}).Error; err != nil {
		return err
	}
	for _, d := range details {
		d.ID = 0
		d.DailyID = dailyID
		if err := tx.Create(&d).Error; err != nil {
			return err
		}
	}
	return tx.Model(&model.AttendanceDaily{}).Where("id = ?", dailyID).Update("status", status).Error
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

	result := make([]map[string]interface{}, len(list))
	for i, d := range list {
		item := map[string]interface{}{
			"id": d.ID, "person_id": d.PersonID, "event_date": d.EventDate,
			"status": d.Status, "punch_time": d.PunchTime, "remark": d.Remark,
			"created_at": d.CreatedAt,
		}
		var name string
		dao.DB.Table("persons").Select("name").Where("id = ?", d.PersonID).Scan(&name)
		item["person_name"] = name
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

func ConfirmDaily(tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail) error {
	if err := UpdateDailyDetails(tx, dailyID, details, "confirmed"); err != nil {
		return err
	}
	var daily model.AttendanceDaily
	if err := tx.First(&daily, dailyID).Error; err != nil {
		return err
	}
	return RebuildDailyProjection(tx, daily.PersonID, daily.EventDate)
}

func ConfirmDailyBatch(tx *gorm.DB, dailyIDs []uint) error {
	for _, id := range dailyIDs {
		var daily model.AttendanceDaily
		if err := tx.First(&daily, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&daily).Update("status", "confirmed").Error; err != nil {
			return err
		}
		if err := RebuildDailyProjection(tx, daily.PersonID, daily.EventDate); err != nil {
			return err
		}
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

func DeleteAttendanceDaily(id uint) error {
	var daily model.AttendanceDaily
	if err := dao.DB.First(&daily, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Where("daily_id = ?", id).Delete(&model.AttendanceEventDetail{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&daily).Error; err != nil {
			return err
		}
		return RebuildDailyProjection(tx, daily.PersonID, daily.EventDate)
	})
}

func RestoreAttendanceDaily(id uint) error {
	var daily model.AttendanceDaily
	if err := dao.DB.Unscoped().First(&daily, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&daily).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Model(&model.AttendanceEventDetail{}).Where("daily_id = ?", id).Update("deleted_at", nil)
		return RebuildDailyProjection(tx, daily.PersonID, daily.EventDate)
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

func CreateBatchAttendanceDailies(req BatchAttendanceReq) (int, int, error) {
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	if end.Before(start) {
		return 0, 0, errors.New("结束日期不能早于开始日期")
	}
	success, fail := 0, 0
	for _, pid := range req.PersonIDs {
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateOnly := utils.DateOnlyFromTime(d)
			err := utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
				daily, err := GetOrCreateDaily(tx, pid, dateOnly, "confirmed")
				if err != nil {
					return err
				}
				return CreateDetail(tx, daily.ID, req.EventType, req.SubType, req.Hours, 0, req.Remark)
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
