package leave_account

import (
	"fmt"
	"strconv"
	"time"

	"probig/internal/pkg/batch"
	"probig/internal/pkg/config"
	"probig/internal/pkg/utils"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) ListEvents(pageNum, pageSize int, personID uint, leaveType, startDate, endDate, sourceType string) ([]LeaveAccountEvent, int64, error) {
	var list []LeaveAccountEvent
	var total int64
	db := s.DB.Model(&LeaveAccountEvent{})
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if leaveType != "" {
		db = db.Where("leave_type = ?", leaveType)
	}
	if startDate != "" {
		db = db.Where("effective_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("effective_date <= ?", endDate)
	}
	if sourceType != "" {
		db = db.Where("source_type = ?", sourceType)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("effective_date desc, created_at desc").Find(&list).Error
	return list, total, err
}

func (s *Service) CreateManualEvent(req map[string]interface{}) (uint, error) {
	event := LeaveAccountEvent{
		PersonID:      uint(getF(req, "person_id")),
		LeaveType:     getS(req, "leave_type"),
		EventType:     EventTypeAdjust,
		SourceType:    SourceManual,
		Hours:         getF(req, "hours"),
		EffectiveDate: getS(req, "effective_date"),
		Remark:        getS(req, "remark"),
	}
	if err := s.DB.Create(&event).Error; err != nil {
		return 0, err
	}
	go s.RebuildBalance(event.PersonID, event.LeaveType)
	return event.ID, nil
}

func (s *Service) DeleteEvent(id uint) error {
	var event LeaveAccountEvent
	if err := s.DB.First(&event, id).Error; err != nil {
		return err
	}
	if event.SourceType == SourceSystemPeriod {
		return fmt.Errorf("系统结转事件不可单条删除")
	}
	if err := s.DB.Delete(&event).Error; err != nil {
		return err
	}
	go s.RebuildBalance(event.PersonID, event.LeaveType)
	return nil
}

func (s *Service) RebuildBalance(personID uint, leaveType string) error {
	s.DB.Where("person_id = ? AND leave_type = ?", personID, leaveType).Delete(&LeaveAccountBalance{})

	var events []LeaveAccountEvent
	s.DB.Where("person_id = ? AND leave_type = ?", personID, leaveType).Find(&events)

	var balanceHours float64
	for _, e := range events {
		if e.EventType == EventTypeCarryoverDeduct {
			balanceHours -= e.Hours
		} else {
			balanceHours += e.Hours
		}
	}

	var attEvents []struct {
		Hours float64
	}
	dbType := "年假"
	if leaveType == LeaveTypeTimeOff {
		dbType = "调休"
	}
	s.DB.Table("attendance_event").
		Select("hours").
		Where("person_id = ? AND sub_type = ?", personID, dbType).
		Scan(&attEvents)
	for _, e := range attEvents {
		balanceHours -= e.Hours
	}

	bal := LeaveAccountBalance{
		PersonID:     personID,
		LeaveType:    leaveType,
		BalanceHours: balanceHours,
		LastCalcAt:   time.Now(),
	}
	return s.DB.Create(&bal).Error
}

func (s *Service) GetBalance(personID uint, leaveType string) (*LeaveAccountBalance, error) {
	var bal LeaveAccountBalance
	err := s.DB.Where("person_id = ? AND leave_type = ?", personID, leaveType).First(&bal).Error
	return &bal, err
}

func (s *Service) GetBalanceList(pageNum, pageSize int, personID uint, leaveType string) ([]LeaveAccountBalance, int64, error) {
	var list []LeaveAccountBalance
	var total int64
	db := s.DB.Model(&LeaveAccountBalance{})
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if leaveType != "" {
		db = db.Where("leave_type = ?", leaveType)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *Service) GetBalanceDetail(personID uint, leaveType string) (map[string]float64, error) {
	result := map[string]float64{
		"grant":      0,
		"adjust":     0,
		"carryover":  0,
		"time_off":   0,
		"used":       0,
	}

	var events []LeaveAccountEvent
	s.DB.Where("person_id = ? AND leave_type = ?", personID, leaveType).Find(&events)
	for _, e := range events {
		switch e.EventType {
		case EventTypeGrant:
			result["grant"] += e.Hours
		case EventTypeAdjust:
			result["adjust"] += e.Hours
		case EventTypeCarryoverDeduct:
			result["carryover"] += e.Hours
		case EventTypeTimeOffAccrue:
			result["time_off"] += e.Hours
		}
	}

	dbType := "年假"
	if leaveType == LeaveTypeTimeOff {
		dbType = "调休"
	}
	var attEvents []struct{ Hours float64 }
	s.DB.Table("attendance_event").
		Select("hours").
		Where("person_id = ? AND sub_type = ?", personID, dbType).
		Scan(&attEvents)
	for _, e := range attEvents {
		result["used"] += e.Hours
	}

	return result, nil
}

func (s *Service) CarryoverAnnualLeave(targetMonth string, operatorID uint, operatorName string) (int, error) {
	year, _ := strconv.Atoi(targetMonth[:4])
	month, _ := strconv.Atoi(targetMonth[5:7])

	annualQuota := 40.0
	if v := config.GetConfig("leave.annual_quota"); v != "" {
		annualQuota = utils.ParseFloat(v)
	}

	processed := 0
	var persons []uint
	s.DB.Table("person").Select("id").Pluck("id", &persons)

	type carryoverPerson struct {
		personID   uint
		negBalance float64
		balance    float64
	}
	var toProcess []carryoverPerson

	for _, personID := range persons {
		var snap struct {
			EntryDate      string
			HasAnnualLeave bool
		}
		err := s.DB.Table("position_snapshot").
			Select("entry_date, has_annual_leave").
			Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?",
				personID, targetMonth+"-01", targetMonth+"-01").
			First(&snap).Error
		if err != nil || !snap.HasAnnualLeave {
			continue
		}

		entryMonthStr := snap.EntryDate[:7]
		entryYear, _ := strconv.Atoi(entryMonthStr[:4])
		entryMonth, _ := strconv.Atoi(entryMonthStr[5:7])

		isAnniversaryMonth := false
		if year > entryYear {
			if entryMonth == month-1 || (month == 1 && entryMonth == 12) {
				isAnniversaryMonth = true
			}
		} else if entryMonth == month-1 {
			isAnniversaryMonth = true
		}
		if !isAnniversaryMonth {
			continue
		}

		var negBalance float64
		bal, _ := s.GetBalance(personID, LeaveTypeAnnual)
		if bal != nil {
			if bal.BalanceHours < 0 {
				negBalance = bal.BalanceHours
			}
		}
		toProcess = append(toProcess, carryoverPerson{
			personID:   personID,
			negBalance: negBalance,
			balance:    bal.BalanceHours,
		})
	}

	if len(toProcess) == 0 {
		return 0, nil
	}

	bo, err := batch.CreateBatch(s.DB, "annual_leave_carryover", targetMonth, operatorID, operatorName)
	if err != nil {
		return 0, err
	}

	for _, cp := range toProcess {
		if cp.balance > 0 {
			s.DB.Create(&LeaveAccountEvent{
				PersonID:      cp.personID,
				LeaveType:     LeaveTypeAnnual,
				EventType:     EventTypeCarryoverDeduct,
				SourceType:    SourceSystemPeriod,
				BatchID:       &bo.ID,
				Hours:         cp.balance,
				EffectiveDate: targetMonth + "-01",
				Remark:        "年假周年结转扣减",
			})
		}

		hoursToGrant := annualQuota
		if cp.negBalance < 0 {
			hoursToGrant = annualQuota + cp.negBalance
		}

		s.DB.Create(&LeaveAccountEvent{
			PersonID:      cp.personID,
			LeaveType:     LeaveTypeAnnual,
			EventType:     EventTypeGrant,
			SourceType:    SourceSystemPeriod,
			BatchID:       &bo.ID,
			Hours:         hoursToGrant,
			EffectiveDate: targetMonth + "-01",
			Remark:        "年假年度配发",
		})

		go s.RebuildBalance(cp.personID, LeaveTypeAnnual)
		processed++
	}

	batch.ExecuteBatch(s.DB, bo.ID, processed)
	return processed, nil
}

func (s *Service) CancelCarryover(batchID uint) error {
	var b batch.SysBatch
	if err := s.DB.First(&b, batchID).Error; err != nil {
		return err
	}
	if b.Status != batch.BatchStatusExecuted {
		return fmt.Errorf("只能冲销已生效的批次")
	}

	var events []LeaveAccountEvent
	s.DB.Where("batch_id = ?", batchID).Find(&events)
	for _, e := range events {
		s.DB.Delete(&e)
		go s.RebuildBalance(e.PersonID, e.LeaveType)
	}
	return batch.CancelBatch(s.DB, batchID)
}

func (s *Service) GetBatchEvents(batchID uint) ([]LeaveAccountEvent, error) {
	var events []LeaveAccountEvent
	err := s.DB.Where("batch_id = ?", batchID).Find(&events).Error
	return events, err
}

func (s *Service) GetBatchList(pageNum, pageSize int) ([]batch.SysBatch, int64, error) {
	var list []batch.SysBatch
	var total int64
	db := s.DB.Model(&batch.SysBatch{}).Where("business_type = ?", "annual_leave_carryover")
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *Service) GetMaxEventUpdatedAt(personID uint, leaveType string) *time.Time {
	var maxT time.Time
	s.DB.Model(&LeaveAccountEvent{}).Unscoped().
		Where("person_id = ? AND leave_type = ?", personID, leaveType).
		Select("MAX(updated_at)").Scan(&maxT)
	var maxD time.Time
	s.DB.Model(&LeaveAccountEvent{}).Unscoped().
		Where("person_id = ? AND leave_type = ? AND deleted_at IS NOT NULL", personID, leaveType).
		Select("MAX(deleted_at)").Scan(&maxD)
	if maxT.IsZero() && maxD.IsZero() {
		return nil
	}
	if maxT.After(maxD) {
		return &maxT
	}
	return &maxD
}

func getS(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getF(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case string:
			return utils.ParseFloat(val)
		}
	}
	return 0
}
