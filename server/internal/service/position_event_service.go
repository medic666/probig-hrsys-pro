package service

import (
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreatePositionEvent(event *model.PositionEvent) error {
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		var maxSeq int
		tx.Unscoped().Model(&model.PositionEvent{}).Where("person_id = ?", event.PersonID).
			Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
		event.Seq = maxSeq + 1

		if err := tx.Create(event).Error; err != nil {
			return err
		}
		return RebuildPositionSnapshots(tx, event.PersonID)
	})
}

func UpdatePositionEvent(id uint, event *model.PositionEvent) error {
	var existing model.PositionEvent
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("职务事件不存在")
	}

	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"event_type":      event.EventType,
			"remark":          event.Remark,
			"effective_date":  event.EffectiveDate,
			"entry_date":      event.EntryDate,
			"leave_date":      event.LeaveDate,
			"attendance_group":    event.AttendanceGroup,
			"has_annual_leave":    event.HasAnnualLeave,
			"has_attendance_bonus": event.HasAttendanceBonus,
			"base_salary":        event.BaseSalary,
			"performance_salary": event.PerformanceSalary,
			"salary_days":        event.SalaryDays,
			"post_allowance":      event.PostAllowance,
			"meal_allowance":      event.MealAllowance,
			"housing_allowance":   event.HousingAllowance,
			"transport_allowance": event.TransportAllowance,
			"high_temp_allowance": event.HighTempAllowance,
			"insurance_compensation": event.InsuranceCompensation,
			"fund_compensation":      event.FundCompensation,
			"social_security_deduct": event.SocialSecurityDeduct,
			"housing_fund_deduct":    event.HousingFundDeduct,
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		return RebuildPositionSnapshots(tx, existing.PersonID)
	})
}

func DeletePositionEvent(id uint) error {
	var event model.PositionEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		return RebuildPositionSnapshots(tx, event.PersonID)
	})
}

func RestorePositionEvent(id uint) error {
	var event model.PositionEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return RebuildPositionSnapshots(tx, event.PersonID)
	})
}

func GetPositionEvent(id uint) (*model.PositionEvent, error) {
	var event model.PositionEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func GetPositionEventList(pageNum, pageSize int, personID uint, startDate, endDate, eventType string) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.PositionEvent{})
	if personID > 0 {
		tx = tx.Where("person_id = ?", personID)
	}
	if startDate != "" {
		tx = tx.Where("effective_date >= ?", startDate)
	}
	if endDate != "" {
		tx = tx.Where("effective_date <= ?", endDate)
	}
	if eventType != "" {
		tx = tx.Where("event_type = ?", eventType)
	}

	var total int64
	tx.Count(&total)

	var events []model.PositionEvent
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("person_id ASC, effective_date DESC, seq DESC").Find(&events)

	var result []map[string]interface{}
	for _, e := range events {
		item := map[string]interface{}{
			"id":             e.ID,
			"person_id":      e.PersonID,
			"seq":            e.Seq,
			"event_type":     e.EventType,
			"remark":         e.Remark,
			"effective_date": e.EffectiveDate,
			"created_at":     e.CreatedAt,
		}
		var personName string
		dao.DB.Table("persons").Select("name").Where("id = ?", e.PersonID).Scan(&personName)
		item["person_name"] = personName

		changedFields := collectChangedFields(e)
		item["changed_fields"] = changedFields

		result = append(result, item)
	}

	return result, total, nil
}

func collectChangedFields(e model.PositionEvent) []string {
	var fields []string
	if e.EntryDate != nil { fields = append(fields, "入职日期") }
	if e.LeaveDate != nil { fields = append(fields, "离职日期") }
	if e.AttendanceGroup != nil { fields = append(fields, "考勤组") }
	if e.HasAnnualLeave != nil { fields = append(fields, "年假标识") }
	if e.HasAttendanceBonus != nil { fields = append(fields, "全勤奖标识") }
	if e.BaseSalary != nil { fields = append(fields, "基本工资") }
	if e.PerformanceSalary != nil { fields = append(fields, "绩效工资基数") }
	if e.SalaryDays != nil { fields = append(fields, "计薪天数") }
	if e.PostAllowance != nil { fields = append(fields, "职位津贴") }
	if e.MealAllowance != nil { fields = append(fields, "餐补") }
	if e.HousingAllowance != nil { fields = append(fields, "房补") }
	if e.TransportAllowance != nil { fields = append(fields, "交通补贴") }
	if e.HighTempAllowance != nil { fields = append(fields, "高温补贴") }
	if e.InsuranceCompensation != nil { fields = append(fields, "保险补偿") }
	if e.FundCompensation != nil { fields = append(fields, "公积金补偿") }
	if e.SocialSecurityDeduct != nil { fields = append(fields, "社保代扣") }
	if e.HousingFundDeduct != nil { fields = append(fields, "公积金代扣") }
	return fields
}

func GetDeletedPositionEvents(pageNum, pageSize int) ([]model.PositionEvent, int64, error) {
	var list []model.PositionEvent
	var total int64
	tx := dao.DB.Unscoped().Model(&model.PositionEvent{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}
