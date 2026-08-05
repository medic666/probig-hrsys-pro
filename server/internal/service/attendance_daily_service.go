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

// maxDailySeq 当日最大版本序号（未删除记录）
func maxDailySeq(tx *gorm.DB, personID uint, eventDate utils.DateOnly) int {
	var maxSeq int
	tx.Model(&model.AttendanceDaily{}).
		Where("person_id = ? AND event_date = ? AND deleted_at IS NULL", personID, eventDate).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
	return maxSeq
}

// dayDailies 当日全部未删除考勤组（升序）
func dayDailies(tx *gorm.DB, personID uint, eventDate utils.DateOnly) ([]model.AttendanceDaily, error) {
	var list []model.AttendanceDaily
	if err := tx.Where("person_id = ? AND event_date = ?", personID, eventDate).
		Order("seq ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// dayDetails 收集多组考勤的明细（involved 集合）：保证被降级组的年假/调休消费被正确撤销
func dayDetails(tx *gorm.DB, dailies []model.AttendanceDaily) ([]model.AttendanceEventDetail, error) {
	var all []model.AttendanceEventDetail
	for _, d := range dailies {
		ds, err := GetDetailsByDailyID(tx, d.ID)
		if err != nil {
			return nil, err
		}
		all = append(all, ds...)
	}
	return all, nil
}

// writeDailyGroup 统一写入核心（追加/编辑转正共用）：
// entryID==0 新建组（seq=MAX+1，状态按提交）；entryID>0 就地编辑该组
// （详情整体替换、字段提供即覆盖，seq 不足当日最大时提升为 MAX+1，已最大则保持）。
// 任何写入后，该组必为当日 seq 最大（唯一 confirmed 必为其）；当日其它未删除组全部降级 pending。
// involved = 当日变更前全部明细 ∪ 本次写入明细，保证年假/调休余额按最新确认组重建。
func writeDailyGroup(tx *gorm.DB, entryID uint, in AttendanceDailyUpsert) error {
	oldDailies, err := dayDailies(tx, in.PersonID, in.Date)
	if err != nil {
		return err
	}
	oldDetails, err := dayDetails(tx, oldDailies)
	if err != nil {
		return err
	}

	var target *model.AttendanceDaily
	if entryID > 0 {
		var daily model.AttendanceDaily
		if err := tx.First(&daily, entryID).Error; err != nil {
			return err
		}
		if daily.PersonID != in.PersonID || !daily.EventDate.Equal(in.Date) {
			return errors.New("考勤记录与提交的人员/日期不一致")
		}
		target = &daily
		updates := map[string]interface{}{}
		if target.Seq < maxDailySeq(tx, in.PersonID, in.Date) {
			updates["seq"] = maxDailySeq(tx, in.PersonID, in.Date) + 1
		}
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
			if err := tx.Model(target).Updates(updates).Error; err != nil {
				return err
			}
		}
		if in.Details != nil {
			if err := tx.Where("daily_id = ?", target.ID).Delete(&model.AttendanceEventDetail{}).Error; err != nil {
				return err
			}
			if err := createDetails(tx, target.ID, in.Details); err != nil {
				return err
			}
		}
	} else {
		initStatus := "confirmed"
		if in.Status != nil {
			initStatus = *in.Status
		}
		target = &model.AttendanceDaily{
			PersonID:  in.PersonID,
			EventDate: in.Date,
			Seq:       maxDailySeq(tx, in.PersonID, in.Date) + 1,
			Status:    initStatus,
		}
		if in.PunchTime != nil {
			target.PunchTime = *in.PunchTime
		}
		if in.Remark != nil {
			target.Remark = *in.Remark
		}
		if err := tx.Create(target).Error; err != nil {
			return err
		}
		if in.Details != nil {
			if err := createDetails(tx, target.ID, in.Details); err != nil {
				return err
			}
		}
	}

	// 降级其它组为 pending（逐条 Save 保证审计钩子留存每行前后快照；已是 pending 的跳过零噪音）
	for _, d := range oldDailies {
		if d.ID == target.ID || d.Status == "pending" {
			continue
		}
		d.Status = "pending"
		if err := tx.Save(&d).Error; err != nil {
			return err
		}
	}

	involved := append(oldDetails, in.Details...)
	return RebuildProjectionsAfterAttendanceChange(tx, in.PersonID, in.Date, involved)
}

// createDetails 批量创建明细（强制新行：忽略传入 id，daily_id 归位）
func createDetails(tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail) error {
	for _, d := range details {
		d.ID = 0
		d.DailyID = dailyID
		if err := tx.Create(&d).Error; err != nil {
			return err
		}
	}
	return nil
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
	tx.Order("event_date DESC, seq DESC, person_id ASC").Offset(offset).Limit(q.PageSize).Find(&list)

	ids := make([]uint, len(list))
	for i, d := range list {
		ids[i] = d.PersonID
	}
	nameMap := PersonNameMap(ids)

	result := make([]map[string]interface{}, len(list))
	for i, d := range list {
		item := map[string]interface{}{
			"id": d.ID, "person_id": d.PersonID, "event_date": d.EventDate, "seq": d.Seq,
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

// GetPendingDailyList 待确认考勤（日级有效语义）：仅返回「当日 seq 最大（有效）且 status=pending」的记录。
// 被取代的陈旧记录不进入待确认页（在套卡/列表中处理），避免同一日多版本重复堆积。
func GetPendingDailyList(q AttendanceDailyListQuery) ([]map[string]interface{}, int64, error) {
	tx := dao.DB.Model(&model.AttendanceDaily{}).Preload("Details").
		Where(`(person_id, event_date, seq) IN (
			SELECT person_id, event_date, MAX(seq) FROM attendance_daily
			WHERE deleted_at IS NULL GROUP BY person_id, event_date
		)`).
		Where("status = ?", "pending")
	if q.PersonID > 0 {
		tx = tx.Where("person_id = ?", q.PersonID)
	}
	if q.DateStart != "" {
		tx = tx.Where("event_date >= ?", q.DateStart)
	}
	if q.DateEnd != "" {
		tx = tx.Where("event_date <= ?", q.DateEnd)
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
			"id": d.ID, "person_id": d.PersonID, "event_date": d.EventDate, "seq": d.Seq,
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

// ConfirmDaily 编辑/确认统一入口（就地转正语义）：
// 目标组（任意版本，由 dailyID 定位）就地更新为提交内容，seq 不足当日最大则提升为最新，
// 当日其它组全部降级 pending——与"新录入成为最新版"完全同构，pending 组可随时编辑/转正。
func ConfirmDaily(ctx context.Context, tx *gorm.DB, dailyID uint, details []model.AttendanceEventDetail, status, punchTime, remark string) error {
	var daily model.AttendanceDaily
	if err := tx.First(&daily, dailyID).Error; err != nil {
		return err
	}
	oldDetails, err := GetDetailsByDailyID(tx, dailyID)
	if err != nil {
		return err
	}
	if err := writeDailyGroup(tx, dailyID, AttendanceDailyUpsert{
		PersonID:  daily.PersonID,
		Date:      daily.EventDate,
		Status:    &status,
		PunchTime: &punchTime,
		Remark:    &remark,
		Details:   details,
	}); err != nil {
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

// RestoreAttendanceDaily 恢复 = 复活为新版：seq 重分配为当日最大+1（避免与现存组冲突），
// 当日其它未删除组全部降级 pending（维持"同日至多一条 confirmed 且必为最大 seq"不变式）。
func RestoreAttendanceDaily(ctx context.Context, id uint) error {
	var daily model.AttendanceDaily
	if err := dao.DB.Unscoped().First(&daily, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		daily.Seq = maxDailySeq(tx, daily.PersonID, daily.EventDate) + 1
		if err := tx.Unscoped().Model(&daily).Updates(map[string]interface{}{
			"deleted_at": nil,
			"seq":        daily.Seq,
		}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&model.AttendanceEventDetail{}).Where("daily_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		dailies, err := dayDailies(tx, daily.PersonID, daily.EventDate)
		if err != nil {
			return err
		}
		details, err := dayDetails(tx, dailies)
		if err != nil {
			return err
		}
		// 其它组降级 pending（恢复组即新版）
		for _, d := range dailies {
			if d.ID == daily.ID || d.Status == "pending" {
				continue
			}
			d.Status = "pending"
			if err := tx.Save(&d).Error; err != nil {
				return err
			}
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

// AttendanceDailyUpsert 考勤日记录统一写入入参（追加式/就地转正共用）。
// 规则唯一——提供即覆盖，未提供保持：
//   Status/PunchTime/Remark 为 *string：非 nil 无条件写入（空串=清空），nil 保持原值；
//   Status 同时作为新建记录的初始状态（nil 时新建默认 confirmed）；
//   Details 非 nil 则整体替换目标组明细（编辑场景先软删旧明细再新建），nil 则保持。
// 单条录入/批量录入/钉钉导入/编辑确认共用同一规则，差异仅在于本次数据提供了哪些字段。
type AttendanceDailyUpsert struct {
	PersonID  uint
	Date      utils.DateOnly
	Status    *string
	PunchTime *string
	Remark    *string
	Details   []model.AttendanceEventDetail
}

// AppendAttendanceDaily 新录入统一入口（追加式）：新建当日 seq 最大+1 的考勤组，
// 当日其它组全部降级 pending（最新记录优先，如实记录；陈旧记录标记待处理），最后重建投影与余额。
// 单条录入/批量录入/钉钉导入共用同一规则。
func AppendAttendanceDaily(tx *gorm.DB, in AttendanceDailyUpsert) error {
	return writeDailyGroup(tx, 0, in)
}

// CreateBatchAttendanceDailies 批量录入考勤日记录：
// 对所选人员 × 时间段内每一天应用同一组事件明细（追加式，见 AppendAttendanceDaily）。
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
				return AppendAttendanceDaily(tx, AttendanceDailyUpsert{
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

// GetAttendanceEventBadges 考勤事件徽章（指定月份，日级有效语义）：无考勤记录 gray；
// 存在当日最新记录为待确认的日期 orange；每日最新记录均确认 green。
// 被取代的陈旧记录（同日较低 seq 的 pending）不参与判定。
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
			FROM (
				SELECT dd.*, ROW_NUMBER() OVER (
					PARTITION BY person_id, event_date ORDER BY seq DESC
				) AS rn
				FROM attendance_daily dd
				WHERE deleted_at IS NULL AND event_date >= ? AND event_date <= ?
			) t
			WHERE rn = 1
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
