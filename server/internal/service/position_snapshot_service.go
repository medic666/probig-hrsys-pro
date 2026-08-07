package service

import (
	"reflect"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

var realFarFuture = func() utils.DateOnly {
	t, _ := utils.ParseDate("9999-12-31")
	return utils.DateOnlyFromTime(t)
}()

// positionState 快照段内容状态（与 PositionSnapshot 业务字段一一对应）。
// 局部重建时由「截断保留段」反解为状态机起点（positionStateFromSnapshot），
// 经 applyPositionEvent 演进、makePositionSnapshot 固化——与全量重建共用同一状态机。
type positionState struct {
	IsActive           bool
	EntryDate          *utils.DateOnly
	LeaveDate          *utils.DateOnly
	AttendanceGroup    string
	HasAnnualLeave     bool
	HasAttendanceBonus bool
	CompanyID          uint
	Department         string
	Position           string
	BaseSalary         float64
	PerformanceSalary  float64
	SalaryDays         float64
	PostAllowance      float64
	MealAllowance      float64
	HousingAllowance   float64
	TransportAllowance float64
	HighTempAllowance  float64
	InsuranceCompensation float64
	FundCompensation      float64
	SocialSecurityDeduct  float64
	HousingFundDeduct     float64
}

func positionStateFromSnapshot(s model.PositionSnapshot) positionState {
	return positionState{
		IsActive:             s.IsActive,
		EntryDate:            s.EntryDate,
		LeaveDate:            s.LeaveDate,
		AttendanceGroup:      s.AttendanceGroup,
		HasAnnualLeave:       s.HasAnnualLeave,
		HasAttendanceBonus:   s.HasAttendanceBonus,
		CompanyID:            s.CompanyID,
		Department:           s.Department,
		Position:             s.Position,
		BaseSalary:           s.BaseSalary,
		PerformanceSalary:    s.PerformanceSalary,
		SalaryDays:           s.SalaryDays,
		PostAllowance:        s.PostAllowance,
		MealAllowance:        s.MealAllowance,
		HousingAllowance:     s.HousingAllowance,
		TransportAllowance:   s.TransportAllowance,
		HighTempAllowance:    s.HighTempAllowance,
		InsuranceCompensation: s.InsuranceCompensation,
		FundCompensation:      s.FundCompensation,
		SocialSecurityDeduct:  s.SocialSecurityDeduct,
		HousingFundDeduct:     s.HousingFundDeduct,
	}
}

// applyPositionEvent 将单个事件叠加到状态机上，返回状态是否发生变更
func applyPositionEvent(s *positionState, e model.PositionEvent) bool {
	changed := false
	if e.EntryDate != nil {
		s.EntryDate = e.EntryDate
		s.IsActive = true
		changed = true
	}
	if e.LeaveDate != nil {
		s.LeaveDate = e.LeaveDate
		s.IsActive = false
		changed = true
	}
	if e.AttendanceGroup != nil {
		if s.AttendanceGroup != *e.AttendanceGroup { changed = true }
		s.AttendanceGroup = *e.AttendanceGroup
	}
	if e.CompanyID != nil {
		if s.CompanyID != *e.CompanyID { changed = true }
		s.CompanyID = *e.CompanyID
	}
	if e.Department != nil {
		if s.Department != *e.Department { changed = true }
		s.Department = *e.Department
	}
	if e.Position != nil {
		if s.Position != *e.Position { changed = true }
		s.Position = *e.Position
	}
	if e.HasAnnualLeave != nil {
		if s.HasAnnualLeave != *e.HasAnnualLeave { changed = true }
		s.HasAnnualLeave = *e.HasAnnualLeave
	}
	if e.HasAttendanceBonus != nil {
		if s.HasAttendanceBonus != *e.HasAttendanceBonus { changed = true }
		s.HasAttendanceBonus = *e.HasAttendanceBonus
	}
	if e.BaseSalary != nil {
		if s.BaseSalary != *e.BaseSalary { changed = true }
		s.BaseSalary = *e.BaseSalary
	}
	if e.PerformanceSalary != nil {
		if s.PerformanceSalary != *e.PerformanceSalary { changed = true }
		s.PerformanceSalary = *e.PerformanceSalary
	}
	if e.SalaryDays != nil {
		if s.SalaryDays != *e.SalaryDays { changed = true }
		s.SalaryDays = *e.SalaryDays
	}
	if e.PostAllowance != nil { applyFloat(&s.PostAllowance, e.PostAllowance, &changed) }
	if e.MealAllowance != nil { applyFloat(&s.MealAllowance, e.MealAllowance, &changed) }
	if e.HousingAllowance != nil { applyFloat(&s.HousingAllowance, e.HousingAllowance, &changed) }
	if e.TransportAllowance != nil { applyFloat(&s.TransportAllowance, e.TransportAllowance, &changed) }
	if e.HighTempAllowance != nil { applyFloat(&s.HighTempAllowance, e.HighTempAllowance, &changed) }
	if e.InsuranceCompensation != nil { applyFloat(&s.InsuranceCompensation, e.InsuranceCompensation, &changed) }
	if e.FundCompensation != nil { applyFloat(&s.FundCompensation, e.FundCompensation, &changed) }
	if e.SocialSecurityDeduct != nil { applyFloat(&s.SocialSecurityDeduct, e.SocialSecurityDeduct, &changed) }
	if e.HousingFundDeduct != nil { applyFloat(&s.HousingFundDeduct, e.HousingFundDeduct, &changed) }
	return changed
}

func applyFloat(target *float64, src *float64, changed *bool) {
	if *target != *src {
		*changed = true
		*target = *src
	}
}

func makePositionSnapshot(personID uint, s positionState, start, end utils.DateOnly) model.PositionSnapshot {
	return model.PositionSnapshot{
		PersonID:             personID,
		EffectiveStartDate:   start,
		EffectiveEndDate:     end,
		IsActive:             s.IsActive,
		EntryDate:            s.EntryDate,
		LeaveDate:            s.LeaveDate,
		AttendanceGroup:      s.AttendanceGroup,
		HasAnnualLeave:       s.HasAnnualLeave,
		HasAttendanceBonus:   s.HasAttendanceBonus,
		CompanyID:            s.CompanyID,
		Department:           s.Department,
		Position:             s.Position,
		BaseSalary:           s.BaseSalary,
		PerformanceSalary:    s.PerformanceSalary,
		SalaryDays:           s.SalaryDays,
		PostAllowance:        s.PostAllowance,
		MealAllowance:        s.MealAllowance,
		HousingAllowance:     s.HousingAllowance,
		TransportAllowance:   s.TransportAllowance,
		HighTempAllowance:    s.HighTempAllowance,
		InsuranceCompensation: s.InsuranceCompensation,
		FundCompensation:      s.FundCompensation,
		SocialSecurityDeduct:  s.SocialSecurityDeduct,
		HousingFundDeduct:     s.HousingFundDeduct,
		LastCalcAt:            time.Now(),
	}
}

// RebuildPositionSnapshots 全量重建（从最早事件生效日切点）：无事件时清空快照。
// 事件变动入口按「以时间为轴」的局部重建（RebuildPositionSnapshotsFrom）工作；
// 全量重建仅作为初始化/兼容入口（等价于从最早事件日局部重建）。
func RebuildPositionSnapshots(tx *gorm.DB, personID uint) error {
	var events []model.PositionEvent
	tx.Where("person_id = ?", personID).Order("effective_date ASC, seq ASC").Find(&events)

	if len(events) == 0 {
		tx.Where("person_id = ?", personID).Delete(&model.PositionSnapshot{})
		return nil
	}
	return RebuildPositionSnapshotsFrom(tx, personID, events[0].EffectiveDate)
}

// RebuildPositionSnapshotsFrom 以时间为轴的局部重建：
// fromDate 之前的快照段原样保留（LastCalcAt 不变，仅将与 fromDate 相交的段截断至 fromDate-1），
// fromDate（含）之后的段全部物理删除，以「截断保留段的内容」为起始状态，
// 重放 fromDate 及之后的事件生成新段。
// 由此，过去月份的 last_calc_at 不被无关的后续事件变动刷新——考勤核算/工资汇总
// 及其徽章的 stale 判定仅对「内容真实变化之后」的月份标橙。
// 鲁棒性：
//   - G1 兜底：快照表缺失/损坏（无保留段）时回放 cut 前事件从零计算起始状态，结果自足；
//   - 混合方案：重建段与被删旧段内容相同则沿用旧 LastCalcAt（内容未变月份不标橙）。
func RebuildPositionSnapshotsFrom(tx *gorm.DB, personID uint, fromDate utils.DateOnly) error {
	var oldSnaps []model.PositionSnapshot
	tx.Where("person_id = ?", personID).Order("effective_start_date ASC").Find(&oldSnaps)

	// 分区：cut 前的段原样保留（含截断段，LastCalcAt 不动）；cut 起的段删除并收集，
	// 作为混合方案「重建段内容比对」的时间戳沿用参照
	var deletedOld []model.PositionSnapshot
	var startState positionState
	preserved := false
	for _, s := range oldSnaps {
		switch {
		case s.EffectiveEndDate.Before(fromDate):
			// 过去段：原样保留（LastCalcAt 不动），内容即 fromDate 前的最近状态
			startState = positionStateFromSnapshot(s)
			preserved = true
		case s.EffectiveStartDate.Before(fromDate) || s.EffectiveStartDate.Equal(fromDate):
			// 包含 fromDate 的段：截断至 fromDate-1（LastCalcAt 不动）；
			// 截断后为空（fromDate 恰为段起点）则整段删除，起始状态继承上一保留段
			end := fromDate.AddDate(0, 0, -1)
			if end.Before(s.EffectiveStartDate) {
				if err := tx.Delete(&s).Error; err != nil {
					return err
				}
				deletedOld = append(deletedOld, s)
			} else {
				if err := tx.Model(&s).Update("effective_end_date", end).Error; err != nil {
					return err
				}
				startState = positionStateFromSnapshot(s)
				preserved = true
			}
		default:
			// fromDate 起的段：删除，随后重建
			if err := tx.Delete(&s).Error; err != nil {
				return err
			}
			deletedOld = append(deletedOld, s)
		}
	}

	var events []model.PositionEvent
	tx.Where("person_id = ?", personID).Order("effective_date ASC, seq ASC").Find(&events)

	// 无任何事件：清空快照（含已保留段），与全量重建语义一致
	if len(events) == 0 {
		tx.Where("person_id = ?", personID).Delete(&model.PositionSnapshot{})
		return nil
	}

	// G1 兜底：无保留段（快照表缺失/损坏）时回放 cut 前事件从零计算起始状态，
	// 保证重建不静默丢失 cut 前事件效果
	if !preserved {
		for _, e := range events {
			if e.EffectiveDate.Before(fromDate) {
				applyPositionEvent(&startState, e)
			}
		}
	}

	current := startState
	snapshotStart := fromDate
	var snapshots []model.PositionSnapshot

	for _, e := range events {
		if e.EffectiveDate.Before(fromDate) {
			continue
		}
		prev := current
		if applyPositionEvent(&current, e) && snapshotStart.Before(e.EffectiveDate) {
			snapshots = append(snapshots, makePositionSnapshot(personID, prev, snapshotStart, e.EffectiveDate.AddDate(0, 0, -1)))
			snapshotStart = e.EffectiveDate
		}
	}

	snapshots = append(snapshots, makePositionSnapshot(personID, current, snapshotStart, realFarFuture))

	// 混合方案：重建段与被删旧段（包含其起点）内容相同 → 沿用旧 LastCalcAt——
	// 该段覆盖的月份核算结果未变，不应标橙；内容不同或无参照 → 当前时间
	for i := range snapshots {
		snapshots[i].LastCalcAt = rebuiltSegmentCalcAt(snapshots[i], deletedOld)
	}

	for _, s := range snapshots {
		if err := tx.Create(&s).Error; err != nil {
			return err
		}
	}

	return nil
}

// rebuiltSegmentCalcAt 混合方案的时间戳判定：
// 在被删旧段中找「包含新段起点」的段（旧段互不重叠，至多一个），
// 内容（positionState）相同则沿用其 LastCalcAt，否则取当前时间。
// 已知残留保守性：纯记录型新增事件（无任何内容字段）从生效日起无被删旧段参照，
// 仍会标橙——编辑型 no-op（内容未变）已被精确消除。
func rebuiltSegmentCalcAt(s model.PositionSnapshot, deletedOld []model.PositionSnapshot) time.Time {
	newState := positionStateFromSnapshot(s)
	for _, o := range deletedOld {
		if o.EffectiveEndDate.Before(s.EffectiveStartDate) {
			continue
		}
		if !o.EffectiveStartDate.Before(s.EffectiveStartDate) && !o.EffectiveStartDate.Equal(s.EffectiveStartDate) {
			continue
		}
		if reflect.DeepEqual(positionStateFromSnapshot(o), newState) {
			return o.LastCalcAt
		}
	}
	return time.Now()
}

func GetCurrentPositionSnapshot(personID uint) (*model.PositionSnapshot, error) {
	var snapshot model.PositionSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date DESC").First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func GetPositionSnapshotHistory(personID uint) ([]model.PositionSnapshot, error) {
	var snapshots []model.PositionSnapshot
	err := dao.DB.Where("person_id = ?", personID).Order("effective_start_date ASC").Find(&snapshots).Error
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}
