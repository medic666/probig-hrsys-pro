package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func CreatePositionEvent(ctx context.Context, event *model.PositionEvent) error {
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
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

func UpdatePositionEvent(ctx context.Context, id uint, updates map[string]interface{}) error {
	var existing model.PositionEvent
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("职务事件不存在")
	}

	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		return RebuildPositionSnapshots(tx, existing.PersonID)
	})
}

func DeletePositionEvent(ctx context.Context, id uint) error {
	var event model.PositionEvent
	if err := dao.DB.First(&event, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Delete(&event).Error; err != nil {
			return err
		}
		return RebuildPositionSnapshots(tx, event.PersonID)
	})
}

func RestorePositionEvent(ctx context.Context, id uint) error {
	var event model.PositionEvent
	if err := dao.DB.Unscoped().First(&event, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
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

// PositionEventListQuery 职务事件列表查询（列表与导出共用）
type PositionEventListQuery struct {
	PageNum    int
	PageSize   int
	PersonID   uint
	StartDate  string
	EndDate    string
	EventType  string
}

func GetPositionEventList(q PositionEventListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.PositionEvent{})
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

func GetDeletedPositionEvents(pageNum, pageSize int) ([]model.PositionEvent, int64, error) {
	var list []model.PositionEvent
	var total int64
	tx := dao.DB.Unscoped().Model(&model.PositionEvent{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}

// GetPositionEventBadges 职务事件徽章：无任何事件 gray；最新事件距今超过 2 年 orange（提示该涨薪）；
// 否则 green。未入职（无事件）归 gray，不误报涨薪提醒。
func GetPositionEventBadges() ([]PersonBadge, error) {
	twoYearsAgo := utils.DateOnlyFromTime(time.Now().AddDate(-2, 0, 0))
	var rows []struct {
		PersonID uint
		Level    string
	}
	err := dao.DB.Table("persons").
		Select(fmt.Sprintf(`persons.id AS person_id,
			CASE
				WHEN MAX(e.effective_date) IS NULL THEN 'gray'
				WHEN MAX(e.effective_date) < '%s' THEN 'orange'
				ELSE 'green'
			END AS level`, twoYearsAgo.String())).
		Joins(`LEFT JOIN position_events e ON e.person_id = persons.id AND e.deleted_at IS NULL`).
		Where("persons.deleted_at IS NULL").
		Group("persons.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPersonBadges(rows), nil
}
