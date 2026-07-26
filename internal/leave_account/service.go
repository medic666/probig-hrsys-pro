package leave_account

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"
	"probig/internal/pkg/audit"
	"probig/internal/pkg/batch"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"
)

func getAnnualLeaveYearlyHours() float64 {
	v := config.Get("annual_leave.yearly_hours")
	if v == "" {
		return 40.0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 40.0
	}
	return f
}

func RebuildBalance(personID uint, leaveType string) error {
	leaveEvents, err := GetEventsByPersonAndType(personID, leaveType)
	if err != nil {
		return fmt.Errorf("failed to get leave account events: %w", err)
	}

	var subType string
	switch leaveType {
	case "annual_leave":
		subType = "年假"
	case "time_off":
		subType = "调休"
	default:
		return fmt.Errorf("unknown leave type: %s", leaveType)
	}

	attEvents, err := GetAttendanceEventsByPersonAndSubType(personID, subType)
	if err != nil {
		return fmt.Errorf("failed to get attendance events: %w", err)
	}

	var totalAccrued float64
	for _, e := range leaveEvents {
		totalAccrued += e.Hours
	}

	var totalTaken float64
	for _, e := range attEvents {
		totalTaken += e.Hours
	}

	balance := totalAccrued - totalTaken
	balance = math.Round(balance*10) / 10

	now := time.Now()
	lb := LeaveAccountBalance{
		PersonID:     personID,
		LeaveType:    leaveType,
		BalanceHours: balance,
		LastCalcAt:   &now,
	}

	return UpsertBalance(&lb)
}

func RebuildBalanceForPerson(personID uint, leaveType string) error {
	return RebuildBalance(personID, leaveType)
}

func CreateManualEvent(event *LeaveAccountEvent, operatorID uint, operatorName string, ip string) error {
	if event.PersonID == 0 {
		return fmt.Errorf("person_id is required")
	}
	if event.LeaveType == "" {
		return fmt.Errorf("leave_type is required")
	}
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.EffectiveDate == nil {
		return fmt.Errorf("effective_date is required")
	}

	event.SourceType = "manual"

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "leave_account_event", event.ID, personName, "新增", nil, event, ip); err != nil {
			return err
		}

		return RebuildBalance(event.PersonID, event.LeaveType)
	})
}

func UpdateManualEvent(event *LeaveAccountEvent, operatorID uint, operatorName string, ip string) error {
	if event.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if event.SourceType != "manual" {
		return fmt.Errorf("only manual events can be edited")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var oldEvent LeaveAccountEvent
		if err := tx.First(&oldEvent, event.ID).Error; err != nil {
			return err
		}
		if oldEvent.SourceType != "manual" {
			return fmt.Errorf("only manual events can be edited")
		}

		if err := tx.Save(event).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "leave_account_event", event.ID, personName, "修改", oldEvent, event, ip); err != nil {
			return err
		}

		return RebuildBalance(event.PersonID, event.LeaveType)
	})
}

func DeleteManualEvent(id uint, operatorID uint, operatorName string, ip string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event LeaveAccountEvent
		if err := tx.First(&event, id).Error; err != nil {
			return err
		}
		if event.SourceType != "manual" {
			return fmt.Errorf("only manual events can be deleted")
		}

		if err := tx.Delete(&event).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "leave_account_event", event.ID, personName, "删除", event, nil, ip); err != nil {
			return err
		}

		return RebuildBalance(event.PersonID, event.LeaveType)
	})
}

func RestoreEventByID(id uint, operatorID uint, operatorName string, ip string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var event LeaveAccountEvent
		if err := tx.Unscoped().First(&event, id).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Model(&event).Update("deleted_at", nil).Error; err != nil {
			return err
		}

		personName := GetPersonName(event.PersonID)
		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "leave_account_event", event.ID, personName, "恢复", nil, event, ip); err != nil {
			return err
		}

		return RebuildBalance(event.PersonID, event.LeaveType)
	})
}

func GetBalanceDetail(personID uint, leaveType string) (*BalanceDetail, error) {
	var subType string
	switch leaveType {
	case "annual_leave":
		subType = "年假"
	case "time_off":
		subType = "调休"
	default:
		return nil, fmt.Errorf("unknown leave type: %s", leaveType)
	}

	leaveEvents, err := GetEventsByPersonAndType(personID, leaveType)
	if err != nil {
		return nil, err
	}

	attEvents, err := GetAttendanceEventsByPersonAndSubType(personID, subType)
	if err != nil {
		return nil, err
	}

	var totalAccrued float64
	var totalAdjusted float64
	var totalCarryover float64
	for _, e := range leaveEvents {
		totalAccrued += e.Hours
		switch e.EventType {
		case "adjust":
			totalAdjusted += e.Hours
		case "carryover_deduct":
			totalCarryover += e.Hours
		}
	}

	var totalTaken float64
	for _, e := range attEvents {
		totalTaken += e.Hours
	}

	balance := math.Round((totalAccrued-totalTaken)*10) / 10

	personName := GetPersonName(personID)

	return &BalanceDetail{
		PersonID:       personID,
		PersonName:     personName,
		LeaveType:      leaveType,
		BalanceHours:   balance,
		TotalAccrued:   math.Round(totalAccrued*10) / 10,
		TotalTaken:     math.Round(totalTaken*10) / 10,
		TotalAdjusted:  math.Round(totalAdjusted*10) / 10,
		TotalCarryover: math.Round(totalCarryover*10) / 10,
	}, nil
}

func ExecuteCarryover(month string, operatorID uint, operatorName string, ip string) (uint, int, int, error) {
	monthStart, err := time.Parse("2006-01-02", month+"-01")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid month format: %s", month)
	}

	targetMonth := int(monthStart.Month())

	var eligiblePersons []struct {
		PersonID uint
		Name     string
	}

	farFutureStr := FarFutureDate.Format("2006-01-02")
	err = db().Table("position_snapshots ps").
		Joins("JOIN persons p ON p.id = ps.person_id").
		Where("ps.effective_end_date = ? AND ps.entry_date IS NOT NULL AND ps.has_annual_leave = ? AND ps.leave_date IS NULL",
			farFutureStr, true).
		Select("ps.person_id, p.name").
		Find(&eligiblePersons).Error
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to find eligible persons: %w", err)
	}

	var personsToProcess []struct {
		PersonID uint
		Name     string
	}

	for _, ep := range eligiblePersons {
		var snapshot PositionSnapshot
		err := db().Where("person_id = ? AND entry_date IS NOT NULL", ep.PersonID).
			Order("entry_date ASC").First(&snapshot).Error
		if err != nil {
			continue
		}
		if snapshot.EntryDate != nil && int(snapshot.EntryDate.Month()) == targetMonth {
			personsToProcess = append(personsToProcess, ep)
		}
	}

	if len(personsToProcess) == 0 {
		return 0, 0, 0, nil
	}

	var batchID uint
	var successCount int
	var failCount int

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		b, err := batch.CreateBatch(tx, "annual_leave_carryover", month, operatorID, operatorName, len(personsToProcess), fmt.Sprintf("年假周年结转 - %s", month))
		if err != nil {
			return fmt.Errorf("failed to create batch: %w", err)
		}
		batchID = b.ID

		yearlyHours := getAnnualLeaveYearlyHours()
		now := time.Now()

		for _, ep := range personsToProcess {
			if err := RebuildBalance(ep.PersonID, "annual_leave"); err != nil {
				failCount++
				continue
			}

			balance, err := GetBalance(ep.PersonID, "annual_leave")
			if err != nil {
				failCount++
				continue
			}

			if balance.BalanceHours > 0 {
				carryoverEvent := LeaveAccountEvent{
					PersonID:      ep.PersonID,
					LeaveType:     "annual_leave",
					EventType:     "carryover_deduct",
					SourceType:    "system_period",
					BatchID:       &batchID,
					Hours:         -balance.BalanceHours,
					EffectiveDate: &now,
					Remark:        fmt.Sprintf("周年结转扣减 - %s", month),
				}
				if err := tx.Create(&carryoverEvent).Error; err != nil {
					failCount++
					continue
				}
			}

			grantEvent := LeaveAccountEvent{
				PersonID:      ep.PersonID,
				LeaveType:     "annual_leave",
				EventType:     "grant",
				SourceType:    "system_period",
				BatchID:       &batchID,
				Hours:         yearlyHours,
				EffectiveDate: &now,
				Remark:        fmt.Sprintf("年度配发 - %s", month),
			}
			if err := tx.Create(&grantEvent).Error; err != nil {
				failCount++
				continue
			}

			if err := RebuildBalance(ep.PersonID, "annual_leave"); err != nil {
				failCount++
				continue
			}

			successCount++
		}

		b.TotalCount = successCount + failCount

		if failCount > 0 {
			b.Status = 4
		} else {
			b.Status = 2
		}
		execTime := time.Now()
		b.ExecutedAt = &execTime
		if err := tx.Save(b).Error; err != nil {
			return err
		}

		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "leave_account_carryover", batchID, month, "结转", nil, b, ip); err != nil {
			return err
		}

		return nil
	})

	return batchID, successCount, failCount, err
}

func CancelBatchByID(batchID uint, operatorID uint, operatorName string, ip string) error {
	b, err := GetBatchByID(batchID)
	if err != nil {
		return fmt.Errorf("batch not found: %w", err)
	}
	if b.BusinessType != "annual_leave_carryover" {
		return fmt.Errorf("batch is not annual leave carryover")
	}
	if b.Status == batch.StatusCanceled {
		return fmt.Errorf("batch already canceled")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		events, err := GetEventsByBatchID(batchID)
		if err != nil {
			return err
		}

		if err := SoftDeleteEventsByBatchID(batchID); err != nil {
			return err
		}

		for _, e := range events {
			if err := RebuildBalance(e.PersonID, e.LeaveType); err != nil {
				return err
			}
		}

		if err := batch.CancelBatch(batchID); err != nil {
			return err
		}

		if err := audit.CreateAuditLog(tx, operatorID, operatorName, "leave_account_carryover", batchID, b.BusinessPeriod, "反结账", b, nil, ip); err != nil {
			return err
		}

		return nil
	})
}

func CreateTimeOffAccrual(personID uint, hours float64, eventDate time.Time, operatorID uint, operatorName string, ip string) error {
	event := LeaveAccountEvent{
		PersonID:      personID,
		LeaveType:     "time_off",
		EventType:     "time_off_accrue",
		SourceType:    "system_period",
		Hours:         hours,
		EffectiveDate: &eventDate,
		Remark:        "补班累计调休",
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		if err := RebuildBalance(personID, "time_off"); err != nil {
			return err
		}

		return nil
	})
}

func ListEventsWithFilter(personID uint, leaveType string, startDate, endDate *time.Time, sourceType string, pageNum, pageSize int) ([]LeaveAccountEventWithName, int64, error) {
	return ListEvents(personID, leaveType, startDate, endDate, sourceType, pageNum, pageSize)
}

func ListBalancesWithFilter(personID uint, leaveType string) ([]LeaveAccountBalanceWithName, error) {
	return ListBalances(personID, leaveType)
}
