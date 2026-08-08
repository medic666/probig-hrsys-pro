package service

import (
	"context"
	"errors"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreatePositionEvent(ctx context.Context, event *model.PositionEvent) error {
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		var maxSeq int
		tx.Unscoped().Model(&model.PositionEvent{}).Where("person_id = ?", event.PersonID).
			Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
		event.Seq = maxSeq + 1

		if err := tx.Create(event).Error; err != nil {
			return err
		}
		// 新事件从生效日起才可能影响状态：仅重建生效日及之后的快照段
		return RebuildPositionSnapshotsFrom(tx, event.PersonID, event.EffectiveDate)
	})
}

func UpdatePositionEvent(ctx context.Context, id uint, event *model.PositionEvent) error {
	var existing model.PositionEvent
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("职务事件不存在")
	}
	if err := EnsureOwnPerson(ctx, existing.PersonID); err != nil {
		return err
	}

	// 切点取新旧生效日的较早者：生效日改晚时（如 5/1→6/1），旧日期至新日期之间的
	// 状态会发生回退，必须一并纳入重建，否则中间区间残留旧效果且时间戳不刷新
	cut := existing.EffectiveDate
	if event.EffectiveDate.Before(cut) {
		cut = event.EffectiveDate
	}

	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Model(&existing).Updates(buildPositionEventUpdates(event)).Error; err != nil {
			return err
		}
		return RebuildPositionSnapshotsFrom(tx, existing.PersonID, cut)
	})
}

// buildPositionEventUpdates 由事件模型推导颗粒化更新字段（与 Create 的 reqToModel 同构）：
// 始终覆盖 event_type/remark/effective_date；指针字段非 nil 覆盖（nil 保持原值）
func buildPositionEventUpdates(e *model.PositionEvent) map[string]interface{} {
	updates := map[string]interface{}{
		"event_type":     e.EventType,
		"remark":         e.Remark,
		"effective_date": e.EffectiveDate,
	}
	if e.EntryDate != nil { updates["entry_date"] = e.EntryDate }
	if e.LeaveDate != nil { updates["leave_date"] = e.LeaveDate }
	if e.AttendanceGroup != nil { updates["attendance_group"] = *e.AttendanceGroup }
	if e.HasAnnualLeave != nil { updates["has_annual_leave"] = *e.HasAnnualLeave }
	if e.HasAttendanceBonus != nil { updates["has_attendance_bonus"] = *e.HasAttendanceBonus }
	if e.CompanyID != nil { updates["company_id"] = *e.CompanyID }
	if e.Department != nil { updates["department"] = *e.Department }
	if e.Position != nil { updates["position"] = *e.Position }
	if e.BaseSalary != nil { updates["base_salary"] = *e.BaseSalary }
	if e.PerformanceSalary != nil { updates["performance_salary"] = *e.PerformanceSalary }
	if e.SalaryDays != nil { updates["salary_days"] = *e.SalaryDays }
	if e.PostAllowance != nil { updates["post_allowance"] = *e.PostAllowance }
	if e.MealAllowance != nil { updates["meal_allowance"] = *e.MealAllowance }
	if e.HousingAllowance != nil { updates["housing_allowance"] = *e.HousingAllowance }
	if e.TransportAllowance != nil { updates["transport_allowance"] = *e.TransportAllowance }
	if e.HighTempAllowance != nil { updates["high_temp_allowance"] = *e.HighTempAllowance }
	if e.InsuranceCompensation != nil { updates["insurance_compensation"] = *e.InsuranceCompensation }
	if e.FundCompensation != nil { updates["fund_compensation"] = *e.FundCompensation }
	if e.SocialSecurityDeduct != nil { updates["social_security_deduct"] = *e.SocialSecurityDeduct }
	if e.HousingFundDeduct != nil { updates["housing_fund_deduct"] = *e.HousingFundDeduct }
	return updates
}

func DeletePositionEvent(ctx context.Context, id uint) error {
	var event model.PositionEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		// 删除事件从生效日起状态回退：仅重建生效日及之后的快照段
		return RebuildPositionSnapshotsFrom(tx, event.PersonID, event.EffectiveDate)
	})
}

func RestorePositionEvent(ctx context.Context, id uint) error {
	var event model.PositionEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// 恢复 = 事件重新生效：仅重建生效日及之后的快照段
		return RebuildPositionSnapshotsFrom(tx, event.PersonID, event.EffectiveDate)
	})
}

func GetPositionEvent(ctx context.Context, id uint) (*model.PositionEvent, error) {
	var event model.PositionEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return nil, err
	}
	if err := EnsureOwnPerson(ctx, event.PersonID); err != nil {
		return nil, err
	}
	return &event, nil
}

// PositionEventListQuery 职务事件列表查询（列表与导出共用）
type PositionEventListQuery struct {
	PageNum    int
	PageSize   int
	PersonID   uint
	StartDate  string
	EndDate    string
	EventType  string
}

func GetPositionEventList(ctx context.Context, q PositionEventListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DBFromContext(ctx).Model(&model.PositionEvent{})
	tx = OwnFilter(ctx, tx, "person_id")
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}
	if q.StartDate != "" {
		tx = tx.Where("effective_date >= ?", q.StartDate)
	}
	if q.EndDate != "" {
		tx = tx.Where("effective_date <= ?", q.EndDate)
	}
	if q.EventType != "" {
		tx = tx.Where("event_type = ?", q.EventType)
	}

	var total int64
	tx.Count(&total)

	var events []model.PositionEvent
	offset := (q.PageNum - 1) * q.PageSize
	tx.Offset(offset).Limit(q.PageSize).Order("person_id ASC, effective_date DESC, seq DESC").Find(&events)

	ids := make([]uint, len(events))
	for i, e := range events {
		ids[i] = e.PersonID
	}
	nameMap := PersonNameMap(ids)

	companyIDs := make([]uint, 0, len(events))
	for _, e := range events {
		if e.CompanyID != nil {
			companyIDs = append(companyIDs, *e.CompanyID)
		}
	}
	companyNameMap := CompanyNameMap(companyIDs)

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
			"company_id":     e.CompanyID,
			"department":     e.Department,
			"position":       e.Position,
		}
		item["person_name"] = nameMap[e.PersonID]
		if e.CompanyID != nil {
			item["company_name"] = companyNameMap[*e.CompanyID]
		}

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
	if e.CompanyID != nil { fields = append(fields, "公司") }
	if e.Department != nil { fields = append(fields, "部门") }
	if e.Position != nil { fields = append(fields, "职位") }
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

func GetDeletedPositionEvents(ctx context.Context, pageNum, pageSize int) ([]model.PositionEvent, int64, error) {
	var list []model.PositionEvent
	var total int64
	tx := dao.DBFromContext(ctx).Unscoped().Model(&model.PositionEvent{}).Where("deleted_at IS NOT NULL")
	tx = OwnFilter(ctx, tx, "person_id")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}

// GetPositionEventBadges 职务事件徽章：无任何事件 gray；最新事件距今超过 2 年 orange（提示该涨薪）；
// 否则 green。未入职（无事件）归 gray，不误报涨薪提醒。
func GetPositionEventBadges(ctx context.Context) ([]PersonBadge, error) {
	twoYearsAgo := utils.DateOnlyFromTime(time.Now().AddDate(-2, 0, 0))
	var rows []struct {
		PersonID uint
		Level    string
	}
	db := dao.DBFromContext(ctx).Table("persons").
		Select(`persons.id AS person_id,
			CASE
				WHEN MAX(e.effective_date) IS NULL THEN 'gray'
				WHEN MAX(e.effective_date) < ? THEN 'orange'
				ELSE 'green'
			END AS level`, twoYearsAgo).
		Joins(`LEFT JOIN position_events e ON e.person_id = persons.id AND e.deleted_at IS NULL`).
		Where("persons.deleted_at IS NULL")
	if pid, ok := dao.OwnPersonID(ctx); ok {
		db = db.Where("persons.id = ?", pid)
	}
	err := db.Group("persons.id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPersonBadges(rows), nil
}
