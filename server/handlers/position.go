package handlers

import (
	"strconv"
	"time"

	"probig/middleware"
	"probig/models"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

func ListPositionEvents(c *gin.Context) {
	personIDStr := c.Query("person_id")
	var personID uint
	if personIDStr != "" {
		pid, _ := strconv.ParseUint(personIDStr, 10, 64)
		personID = uint(pid)
	}
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	events, total, err := services.ListPositionEvents(personID, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, events, total, page, pageSize)
}

func GetPositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	event, err := services.GetPositionEvent(uint(id))
	if err != nil {
		utils.ErrBadRequest(c, "事件不存在")
		return
	}
	utils.Success(c, event)
}

type positionEventInput struct {
	PersonID              uint    `json:"person_id"`
	EventName             string  `json:"event_name"`
	EffectiveDate         string  `json:"effective_date"`
	AttendanceGroup       *string `json:"attendance_group"`
	EntryDate             *string `json:"entry_date"`
	LeaveDate             *string `json:"leave_date"`
	HasAnnualLeave        *bool   `json:"has_annual_leave"`
	HasAttendanceBonus    *bool   `json:"has_attendance_bonus"`
	BaseSalary            *float64 `json:"base_salary"`
	PerformanceSalary     *float64 `json:"performance_salary"`
	SalaryDays            *int    `json:"salary_days"`
	PostAllowance         *float64 `json:"post_allowance"`
	MealAllowance         *float64 `json:"meal_allowance"`
	HousingAllowance      *float64 `json:"housing_allowance"`
	TransportAllowance    *float64 `json:"transport_allowance"`
	HighTempAllowance     *float64 `json:"high_temp_allowance"`
	InsuranceCompensation *float64 `json:"insurance_compensation"`
	FundCompensation      *float64 `json:"fund_compensation"`
	SocialSecurityDeduct  *float64 `json:"social_security_deduct"`
	HousingFundDeduct     *float64 `json:"housing_fund_deduct"`
}

func parseDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func toEvent(input positionEventInput) models.PositionEvent {
	effDate, _ := time.Parse("2006-01-02", input.EffectiveDate)
	return models.PositionEvent{
		PersonID:              input.PersonID,
		EventName:             input.EventName,
		EffectiveDate:         effDate,
		AttendanceGroup:       input.AttendanceGroup,
		EntryDate:             parseDate(input.EntryDate),
		LeaveDate:             parseDate(input.LeaveDate),
		HasAnnualLeave:        input.HasAnnualLeave,
		HasAttendanceBonus:    input.HasAttendanceBonus,
		BaseSalary:            input.BaseSalary,
		PerformanceSalary:     input.PerformanceSalary,
		SalaryDays:            input.SalaryDays,
		PostAllowance:         input.PostAllowance,
		MealAllowance:         input.MealAllowance,
		HousingAllowance:      input.HousingAllowance,
		TransportAllowance:    input.TransportAllowance,
		HighTempAllowance:     input.HighTempAllowance,
		InsuranceCompensation: input.InsuranceCompensation,
		FundCompensation:      input.FundCompensation,
		SocialSecurityDeduct:  input.SocialSecurityDeduct,
		HousingFundDeduct:     input.HousingFundDeduct,
	}
}

func CreatePositionEvent(c *gin.Context) {
	var input positionEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	event := toEvent(input)
	if err := services.CreatePositionEvent(&event); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "position_event", event.ID, "新增", "{}", event)
	utils.Success(c, event)
}

func UpdatePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	before, _ := services.GetPositionEvent(uint(id))
	if err := services.UpdatePositionEvent(uint(id), updates); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	after, _ := services.GetPositionEvent(uint(id))
	middleware.AuditAction(c, "position_event", uint(id), "修改", before, after)
	utils.Success(c, nil)
}

func DeletePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	before, _ := services.GetPositionEvent(uint(id))
	if err := services.DeletePositionEvent(uint(id)); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "position_event", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func ListPositionSnapshots(c *gin.Context) {
	personIDStr := c.Param("person_id")
	personID, _ := strconv.ParseUint(personIDStr, 10, 64)
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	snaps, total, _ := services.ListPositionSnapshots(uint(personID), offset, pageSize)
	utils.SuccessPage(c, snaps, total, page, pageSize)
}

func GetLatestSnapshot(c *gin.Context) {
	personIDStr := c.Param("person_id")
	personID, _ := strconv.ParseUint(personIDStr, 10, 64)
	snap, err := services.GetPersonLatestSnapshot(uint(personID))
	if err != nil {
		utils.ErrBadRequest(c, "无快照数据")
		return
	}
	utils.Success(c, snap)
}

func RebuildSnapshots(c *gin.Context) {
	personIDStr := c.Param("person_id")
	personID, _ := strconv.ParseUint(personIDStr, 10, 64)
	services.RebuildSnapshots(uint(personID))
	utils.Success(c, nil)
}
