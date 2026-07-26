package position

import (
	"sort"
	"time"

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

func (s *Service) ListEvents(pageNum, pageSize int, personID uint, startDate, endDate, eventName string) ([]PositionEvent, int64, error) {
	var list []PositionEvent
	var total int64
	db := s.DB.Model(&PositionEvent{})
	if personID > 0 {
		db = db.Where("person_id = ?", personID)
	}
	if startDate != "" {
		db = db.Where("effective_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("effective_date <= ?", endDate)
	}
	if eventName != "" {
		db = db.Where("event_name = ?", eventName)
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("effective_date desc, created_at desc").Find(&list).Error
	return list, total, err
}

func (s *Service) CreateEvent(req map[string]interface{}) (uint, error) {
	event := PositionEvent{
		PersonID:      uint(getF(req, "person_id")),
		EventName:     getS(req, "event_name"),
		EffectiveDate: getS(req, "effective_date"),
	}
	if v, ok := req["attendance_group"]; ok {
		s := ""
		if v != nil {
			s, _ = v.(string)
		}
		event.AttendanceGroup = &s
	}
	if v, ok := req["has_annual_leave"]; ok {
		b := false
		if v != nil {
			b, _ = v.(bool)
		}
		event.HasAnnualLeave = &b
	}
	if v, ok := req["has_attendance_bonus"]; ok {
		b := false
		if v != nil {
			b, _ = v.(bool)
		}
		event.HasAttendanceBonus = &b
	}
	setFloatPtr(req, "base_salary", &event.BaseSalary)
	setFloatPtr(req, "performance_salary", &event.PerformanceSalary)
	setIntPtr(req, "salary_days", &event.SalaryDays)
	setFloatPtr(req, "post_allowance", &event.PostAllowance)
	setFloatPtr(req, "meal_allowance", &event.MealAllowance)
	setFloatPtr(req, "housing_allowance", &event.HousingAllowance)
	setFloatPtr(req, "transport_allowance", &event.TransportAllowance)
	setFloatPtr(req, "high_temp_allowance", &event.HighTempAllowance)
	setFloatPtr(req, "insurance_compensation", &event.InsuranceComp)
	setFloatPtr(req, "fund_compensation", &event.FundComp)
	setFloatPtr(req, "social_security_deduct", &event.SocialSecurityDeduct)
	setFloatPtr(req, "housing_fund_deduct", &event.HousingFundDeduct)

	if err := s.DB.Create(&event).Error; err != nil {
		return 0, err
	}
	go s.RebuildSnapshots(event.PersonID)
	return event.ID, nil
}

func (s *Service) UpdateEvent(id uint, req map[string]interface{}) error {
	var event PositionEvent
	if err := s.DB.First(&event, id).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if v, ok := req["event_name"]; ok {
		updates["event_name"] = v
	}
	if v, ok := req["effective_date"]; ok && v.(string) != "" {
		updates["effective_date"] = v
	}
	for _, f := range []string{"attendance_group", "base_salary", "performance_salary", "salary_days",
		"post_allowance", "meal_allowance", "housing_allowance", "transport_allowance",
		"high_temp_allowance", "insurance_compensation", "fund_compensation",
		"social_security_deduct", "housing_fund_deduct", "has_annual_leave", "has_attendance_bonus"} {
		if v, ok := req[f]; ok {
			updates[f] = v
		}
	}
	if err := s.DB.Model(&event).Updates(updates).Error; err != nil {
		return err
	}
	go s.RebuildSnapshots(event.PersonID)
	return nil
}

func (s *Service) DeleteEvent(id uint) error {
	var event PositionEvent
	if err := s.DB.First(&event, id).Error; err != nil {
		return err
	}
	if err := s.DB.Delete(&event).Error; err != nil {
		return err
	}
	go s.RebuildSnapshots(event.PersonID)
	return nil
}

func (s *Service) RebuildSnapshots(personID uint) error {
	var events []PositionEvent
	s.DB.Where("person_id = ?", personID).Order("effective_date asc, created_at asc").Find(&events)

	if len(events) == 0 {
		return s.DB.Where("person_id = ?", personID).Delete(&PositionSnapshot{}).Error
	}

	type evtInfo struct {
		evt       PositionEvent
		createdAt time.Time
	}
	var sortedEvents []evtInfo
	for _, e := range events {
		sortedEvents = append(sortedEvents, evtInfo{evt: e, createdAt: e.CreatedAt})
	}
	sort.Slice(sortedEvents, func(i, j int) bool {
		if sortedEvents[i].evt.EffectiveDate != sortedEvents[j].evt.EffectiveDate {
			return sortedEvents[i].evt.EffectiveDate < sortedEvents[j].evt.EffectiveDate
		}
		return !sortedEvents[i].createdAt.Before(sortedEvents[j].createdAt)
	})

	dates := make(map[string]bool)
	for _, se := range sortedEvents {
		dates[se.evt.EffectiveDate] = true
	}
	var dateList []string
	for d := range dates {
		dateList = append(dateList, d)
	}
	sort.Strings(dateList)

	state := initState()
	entryDate := ""
	var leaveDate *string

	var snapshots []PositionSnapshot
	hasLeft := false

	for i, d := range dateList {
		for _, se := range sortedEvents {
			if se.evt.EffectiveDate == d {
				applyEvent(state, se.evt)
				if se.evt.EventName == EventTypeEntry {
					if entryDate != "" && hasLeft {
						entryDate = se.evt.EffectiveDate
						leaveDate = nil
						hasLeft = false
						state = initState()
						applyEvent(state, se.evt)
					} else if entryDate == "" {
						entryDate = se.evt.EffectiveDate
					}
				}
				if se.evt.EventName == EventTypeLeave {
					leaveDate = &se.evt.EffectiveDate
					hasLeft = true
				}
			}
		}

		endDate := utils.InfiniteDate
		if i < len(dateList)-1 {
			nextDate := utils.ParseDate(dateList[i+1])
			endDate = utils.DateStr(nextDate.Add(-24 * time.Hour))
		}

		snapshot := PositionSnapshot{
			PersonID:           personID,
			EffectiveStartDate: d,
			EffectiveEndDate:   endDate,
			EntryDate:          entryDate,
			LastCalcAt:         time.Now(),
		}
		if leaveDate != nil {
			snapshot.LeaveDate = leaveDate
		}
		applyToSnapshot(&snapshot, state)
		snapshots = append(snapshots, snapshot)
	}

	if len(snapshots) > 0 {
		snapshots[len(snapshots)-1].EffectiveEndDate = utils.InfiniteDate
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("person_id = ?", personID).Delete(&PositionSnapshot{}).Error; err != nil {
			return err
		}
		for _, snap := range snapshots {
			if err := tx.Create(&snap).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) GetCurrentSnapshot(personID uint) (*PositionSnapshot, error) {
	var snap PositionSnapshot
	date := time.Now().Format("2006-01-02")
	err := s.DB.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?", personID, date, date).First(&snap).Error
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

func (s *Service) GetSnapshots(personID uint) ([]PositionSnapshot, error) {
	var list []PositionSnapshot
	err := s.DB.Where("person_id = ?", personID).Order("effective_start_date asc").Find(&list).Error
	return list, err
}

func (s *Service) GetSnapshotsInRange(personID uint, monthStart, monthEnd string) ([]PositionSnapshot, error) {
	var list []PositionSnapshot
	err := s.DB.Where("person_id = ? AND effective_start_date <= ? AND effective_end_date >= ?", personID, monthEnd, monthStart).Order("effective_start_date asc").Find(&list).Error
	return list, err
}

func (s *Service) GetMaxEventUpdatedAt(personID uint, dateStart, dateEnd string) *time.Time {
	var maxT time.Time
	s.DB.Model(&PositionEvent{}).Unscoped().
		Where("person_id = ? AND effective_date >= ? AND effective_date <= ?", personID, dateStart, dateEnd).
		Select("MAX(updated_at)").
		Scan(&maxT)
	var maxD time.Time
	s.DB.Model(&PositionEvent{}).Unscoped().
		Where("person_id = ? AND effective_date >= ? AND effective_date <= ? AND deleted_at IS NOT NULL", personID, dateStart, dateEnd).
		Select("MAX(deleted_at)").
		Scan(&maxD)
	if maxT.IsZero() && maxD.IsZero() {
		return nil
	}
	if maxT.After(maxD) {
		return &maxT
	}
	return &maxD
}

func initState() map[string]interface{} {
	return map[string]interface{}{
		"attendance_group":       "",
		"has_annual_leave":       true,
		"has_attendance_bonus":   true,
		"base_salary":            0.0,
		"performance_salary":     0.0,
		"salary_days":            0,
		"post_allowance":         0.0,
		"meal_allowance":         0.0,
		"housing_allowance":      0.0,
		"transport_allowance":    0.0,
		"high_temp_allowance":    0.0,
		"insurance_comp":         0.0,
		"fund_comp":              0.0,
		"social_security_deduct": 0.0,
		"housing_fund_deduct":    0.0,
	}
}

func applyEvent(state map[string]interface{}, e PositionEvent) {
	if e.AttendanceGroup != nil {
		state["attendance_group"] = *e.AttendanceGroup
	}
	if e.HasAnnualLeave != nil {
		state["has_annual_leave"] = *e.HasAnnualLeave
	}
	if e.HasAttendanceBonus != nil {
		state["has_attendance_bonus"] = *e.HasAttendanceBonus
	}
	if e.BaseSalary != nil {
		state["base_salary"] = *e.BaseSalary
	}
	if e.PerformanceSalary != nil {
		state["performance_salary"] = *e.PerformanceSalary
	}
	if e.SalaryDays != nil {
		state["salary_days"] = *e.SalaryDays
	}
	if e.PostAllowance != nil {
		state["post_allowance"] = *e.PostAllowance
	}
	if e.MealAllowance != nil {
		state["meal_allowance"] = *e.MealAllowance
	}
	if e.HousingAllowance != nil {
		state["housing_allowance"] = *e.HousingAllowance
	}
	if e.TransportAllowance != nil {
		state["transport_allowance"] = *e.TransportAllowance
	}
	if e.HighTempAllowance != nil {
		state["high_temp_allowance"] = *e.HighTempAllowance
	}
	if e.InsuranceComp != nil {
		state["insurance_comp"] = *e.InsuranceComp
	}
	if e.FundComp != nil {
		state["fund_comp"] = *e.FundComp
	}
	if e.SocialSecurityDeduct != nil {
		state["social_security_deduct"] = *e.SocialSecurityDeduct
	}
	if e.HousingFundDeduct != nil {
		state["housing_fund_deduct"] = *e.HousingFundDeduct
	}
}

func applyToSnapshot(sn *PositionSnapshot, state map[string]interface{}) {
	sn.AttendanceGroup = state["attendance_group"].(string)
	sn.HasAnnualLeave = state["has_annual_leave"].(bool)
	sn.HasAttendanceBonus = state["has_attendance_bonus"].(bool)
	sn.BaseSalary = state["base_salary"].(float64)
	sn.PerformanceSalary = state["performance_salary"].(float64)
	sn.SalaryDays = state["salary_days"].(int)
	sn.PostAllowance = state["post_allowance"].(float64)
	sn.MealAllowance = state["meal_allowance"].(float64)
	sn.HousingAllowance = state["housing_allowance"].(float64)
	sn.TransportAllowance = state["transport_allowance"].(float64)
	sn.HighTempAllowance = state["high_temp_allowance"].(float64)
	sn.InsuranceComp = state["insurance_comp"].(float64)
	sn.FundComp = state["fund_comp"].(float64)
	sn.SocialSecurityDeduct = state["social_security_deduct"].(float64)
	sn.HousingFundDeduct = state["housing_fund_deduct"].(float64)
}

func setFloatPtr(req map[string]interface{}, key string, target **float64) {
	if v, ok := req[key]; ok {
		if v == nil {
			*target = nil
		} else {
			f := toFloat(v)
			*target = &f
		}
	}
}

func setIntPtr(req map[string]interface{}, key string, target **int) {
	if v, ok := req[key]; ok {
		if v == nil {
			*target = nil
		} else {
			i := int(toFloat(v))
			*target = &i
		}
	}
}

func getF(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		return toFloat(v)
	}
	return 0
}

func getS(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		s, _ := v.(string)
		return s
	}
	return ""
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		return utils.ParseFloat(val)
	}
	return 0
}
