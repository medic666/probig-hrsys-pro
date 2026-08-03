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

// dailyDetailKey 明细业务键：事件类型 + 子类型（一天内同类型明细的业务对齐依据）
func dailyDetailKey(d model.AttendanceEventDetail) string {
	return d.EventType + "\x00" + d.SubType
}

// dailyDetailsEqual 明细内容比较（不含 id/时间戳等技术字段）
func dailyDetailsEqual(a, b model.AttendanceEventDetail) bool {
	return a.EventType == b.EventType && a.SubType == b.SubType &&
		a.Hours == b.Hours && a.Minutes == b.Minutes && a.Remark == b.Remark
}

// SyncDailyDetails 考勤明细颗粒化同步（UPSERT）：
// 匹配优先级 P1 主键 → P2 业务键（事件类型+子类型）；内容相同零操作零审计；
// 变化更新（保留原行，一条"修改"审计）；新增创建；缺失软删除。事务由调用方包裹。
// 与 SyncChildRecords 语义一致，并扩展无主键输入（录入/导入场景）的内容对齐能力。
func SyncDailyDetails(tx *gorm.DB, dailyID uint, incoming []model.AttendanceEventDetail) error {
	var old []model.AttendanceEventDetail
	if err := tx.Where("daily_id = ?", dailyID).Find(&old).Error; err != nil {
		return err
	}
	oldByID := make(map[uint]int, len(old))
	keyOfOld := make(map[string][]int, len(old))
	for i, o := range old {
		oldByID[o.ID] = i
		key := dailyDetailKey(o)
		keyOfOld[key] = append(keyOfOld[key], i)
	}
	used := make([]bool, len(old))

	for _, in := range incoming {
		// P1 主键对齐（编辑回显场景）
		if in.ID != 0 {
			idx, ok := oldByID[in.ID]
			if !ok || used[idx] {
				continue // id 不属于当日或已被匹配，忽略
			}
			if !dailyDetailsEqual(old[idx], in) {
				updated := in
				updated.DailyID = dailyID // in.ID 即原行主键，保留
				if err := tx.Save(&updated).Error; err != nil {
					return err
				}
			}
			used[idx] = true
			continue
		}
		// P2 业务键对齐（录入/导入场景）：内容相同优先，否则顺序配对
		idx := -1
		for _, c := range keyOfOld[dailyDetailKey(in)] {
			if !used[c] && dailyDetailsEqual(old[c], in) {
				idx = c
				break
			}
		}
		if idx == -1 {
			for _, c := range keyOfOld[dailyDetailKey(in)] {
				if !used[c] {
					idx = c
					break
				}
			}
		}
		if idx == -1 {
			created := in
			created.DailyID = dailyID
			if err := tx.Create(&created).Error; err != nil {
				return err
			}
			continue
		}
		used[idx] = true
		if !dailyDetailsEqual(old[idx], in) {
			updated := in
			updated.ID = old[idx].ID
			updated.DailyID = dailyID
			if err := tx.Save(&updated).Error; err != nil {
				return err
			}
		}
	}
	// 剩余旧行：软删除
	for i, o := range old {
		if !used[i] {
			if err := tx.Delete(&o).Error; err != nil {
				return err
			}
		}
	}
	return nil
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
	// 颗粒化同步：未变化明细零操作零审计；变化更新、新增创建、缺失软删除
	if err := SyncDailyDetails(tx, dailyID, details); err != nil {
		return err
	}
	return tx.Model(&model.AttendanceDaily{}).Where("id = ?", dailyID).Update("status", status).Error
}

// AttendanceDailyListQuery 考勤日记录列表查询（列表与导出共用）
type AttendanceDailyListQuery struct {
	PageNum   int
	PageSize  int
	PersonID  uint
	DateStart string
	DateEnd   string
	Status    string
}

func GetAttendanceDailyList(q AttendanceDailyListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.AttendanceDaily{}).Preload("Details")
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}
	if q.DateStart != "" {
		tx = tx.Where("event_date >= ?", q.DateStart)
	}
	if q.DateEnd != "" {
		tx = tx.Where("event_date <= ?", q.DateEnd)
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	tx.Count(&total)

	var list []model.AttendanceDaily
	offset := (q.PageNum - 1) * q.PageSize
	tx.Order("event_date DESC, person_id ASC").Offset(offset).Limit(q.PageSize).Find(&list)

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

func GetPendingDailyList(q AttendanceDailyListQuery) ([]map[string]interface{}, int64, error) {
	q.DateStart, q.DateEnd = "", ""
	q.Status = "pending"
	return GetAttendanceDailyList(q)
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

func ConfirmDaily(ctx context.Context, tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail, status string) error {
	var daily model.AttendanceDaily
	if err := tx.First(&daily, dailyID).Error; err != nil {
		return err
	}
	oldDetails, err := GetDetailsByDailyID(tx, dailyID)
	if err != nil {
		return err
	}
	if err := UpdateDailyDetails(tx, dailyID, details, status); err != nil {
		return err
	}
	if err := RebuildProjectionsAfterAttendanceChange(tx, daily.PersonID, daily.EventDate, append(oldDetails, details...)); err != nil {
		return err
	}
	writeConfirmAudit(ctx, tx, daily, oldDetails, details)
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
	PersonIDs []uint                     `json:"person_ids"`
	StartDate string                     `json:"start_date"`
	EndDate   string                     `json:"end_date"`
	Status    string                     `json:"status"`
	PunchTime string                     `json:"punch_time"`
	Remark    string                     `json:"remark"`
	Details   []model.AttendanceEventDetail `json:"details"`
}

// AttendanceDailyUpsert 考勤日记录统一写入入参（颗粒化 upsert）。
// 规则唯一——提供即覆盖，未提供保持：
//   Status/PunchTime/Remark 为 *string：非 nil 无条件写入（空串=清空），nil 保持原值；
//   Status 同时作为新建记录的初始状态（nil 时新建默认 confirmed）；
//   Details 非 nil 则整体替换当日明细（先清空再写入），nil 则保持。
// 单条录入/批量录入/钉钉导入共用同一规则，差异仅在于本次数据提供了哪些字段。
type AttendanceDailyUpsert struct {
	PersonID  uint
	Date      utils.DateOnly
	Status    *string
	PunchTime *string
	Remark    *string
	Details   []model.AttendanceEventDetail
}

// UpsertAttendanceDaily 颗粒化 upsert 统一写入入口：明细经 SyncDailyDetails 细粒度同步
// （未变化零操作零审计、变化更新、新增创建、缺失软删除，重复提交不产生重复记录），
// 再按提供字段覆盖 daily 附加属性，最后重建投影。
func UpsertAttendanceDaily(tx *gorm.DB, in AttendanceDailyUpsert) error {
	initStatus := "confirmed"
	if in.Status != nil {
		initStatus = *in.Status
	}
	daily, err := GetOrCreateDaily(tx, in.PersonID, in.Date, initStatus)
	if err != nil {
		return err
	}
	if in.Details != nil {
		if err := SyncDailyDetails(tx, daily.ID, in.Details); err != nil {
			return err
		}
	}
	updates := map[string]interface{}{}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if in.PunchTime != nil {
		updates["punch_time"] = *in.PunchTime
	}
	if in.Remark != nil {
		updates["remark"] = *in.Remark
	}
	if len(updates) > 0 {
		if err := tx.Model(daily).Updates(updates).Error; err != nil {
			return err
		}
	}
	return RebuildProjectionsAfterAttendanceChange(tx, in.PersonID, in.Date, in.Details)
}

// CreateBatchAttendanceDailies 批量录入考勤日记录：
// 对所选人员 × 时间段内每一天应用同一组事件明细（颗粒化 upsert，见 UpsertAttendanceDaily）。
func CreateBatchAttendanceDailies(ctx context.Context, req BatchAttendanceReq) (int, int, error) {
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	if end.Before(start) {
		return 0, 0, errors.New("结束日期不能早于开始日期")
	}
	success, fail := 0, 0
	for _, pid := range req.PersonIDs {
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateOnly := utils.DateOnlyFromTime(d)
			err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
				return UpsertAttendanceDaily(tx, AttendanceDailyUpsert{
					PersonID:  pid,
					Date:      dateOnly,
					Status:    &req.Status,
					PunchTime: &req.PunchTime,
					Remark:    &req.Remark,
					Details:   req.Details,
				})
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

// GetAttendanceEventBadges 考勤事件徽章（指定月份）：无考勤记录 gray；存在待确认记录 orange；全确认 green。
// 未入职（无记录）归 gray。
func GetAttendanceEventBadges(month string) ([]PersonBadge, error) {
	start, err := utils.MonthStart(month)
	if err != nil {
		return nil, err
	}
	startD := utils.DateOnlyFromTime(start)
	endD := utils.DateOnlyFromTime(start.AddDate(0, 1, -1))

	var rows []struct {
		PersonID uint
		Level    string
	}
	err = dao.DB.Table("persons").
		Select(`persons.id AS person_id,
			CASE
				WHEN d.cnt IS NULL THEN 'gray'
				WHEN d.pending > 0 THEN 'orange'
				ELSE 'green'
			END AS level`).
		Joins(`LEFT JOIN (
			SELECT person_id, COUNT(*) AS cnt,
				SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending
			FROM attendance_daily
			WHERE deleted_at IS NULL AND event_date >= ? AND event_date <= ?
			GROUP BY person_id
		) d ON d.person_id = persons.id`, startD, endD).
		Where("persons.deleted_at IS NULL").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPersonBadges(rows), nil
}

// GetDailyProjectionBadges 日记工时徽章（指定月份）：无投影 gray；
// 同月既有事假天又有加班天 orange（少见情况，一般按补班/调休记录）；否则 green。
func GetDailyProjectionBadges(month string) ([]PersonBadge, error) {
	start, err := utils.MonthStart(month)
	if err != nil {
		return nil, err
	}
	startD := utils.DateOnlyFromTime(start)
	endD := utils.DateOnlyFromTime(start.AddDate(0, 1, -1))

	var rows []struct {
		PersonID uint
		Level    string
	}
	err = dao.DB.Table("persons").
		Select(`persons.id AS person_id,
			CASE
				WHEN d.cnt IS NULL THEN 'gray'
				WHEN d.leave_days > 0 AND d.ot_days > 0 THEN 'orange'
				ELSE 'green'
			END AS level`).
		Joins(`LEFT JOIN (
			SELECT person_id, COUNT(*) AS cnt,
				SUM(CASE WHEN has_personal_leave = 1 THEN 1 ELSE 0 END) AS leave_days,
				SUM(CASE WHEN overtime_workday_hours > 0 OR overtime_holiday_hours > 0 THEN 1 ELSE 0 END) AS ot_days
			FROM attendance_daily_projections
			WHERE work_date >= ? AND work_date <= ?
			GROUP BY person_id
		) d ON d.person_id = persons.id`, startD, endD).
		Where("persons.deleted_at IS NULL").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return toPersonBadges(rows), nil
}
