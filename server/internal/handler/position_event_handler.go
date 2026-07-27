package handler

import (
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetPositionEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personIDStr := c.Query("person_id")
	personID, _ := strconv.ParseUint(personIDStr, 10, 64)
	list, total, err := service.GetPositionEventList(pageReq.PageNum, pageReq.PageSize,
		uint(personID), c.Query("start_date"), c.Query("end_date"), c.Query("event_type"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func GetPositionEventByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	event, err := service.GetPositionEvent(uint(id))
	if err != nil {
		utils.Error(c, "职务事件不存在")
		return
	}
	utils.Success(c, event)
}

type positionEventReq struct {
	PersonID            uint     `json:"person_id"`
	EventType           string   `json:"event_type"`
	Remark              string   `json:"remark"`
	EffectiveDate       string   `json:"effective_date"`
	EntryDate           *string  `json:"entry_date"`
	LeaveDate           *string  `json:"leave_date"`
	AttendanceGroup     *string  `json:"attendance_group"`
	HasAnnualLeave      *bool    `json:"has_annual_leave"`
	HasAttendanceBonus  *bool    `json:"has_attendance_bonus"`
	BaseSalary          *float64 `json:"base_salary"`
	PerformanceSalary   *float64 `json:"performance_salary"`
	SalaryDays          *int     `json:"salary_days"`
	PostAllowance       *float64 `json:"post_allowance"`
	MealAllowance       *float64 `json:"meal_allowance"`
	HousingAllowance    *float64 `json:"housing_allowance"`
	TransportAllowance  *float64 `json:"transport_allowance"`
	HighTempAllowance   *float64 `json:"high_temp_allowance"`
	InsuranceCompensation *float64 `json:"insurance_compensation"`
	FundCompensation    *float64 `json:"fund_compensation"`
	SocialSecurityDeduct *float64 `json:"social_security_deduct"`
	HousingFundDeduct   *float64 `json:"housing_fund_deduct"`
}

func reqToModel(req positionEventReq) model.PositionEvent {
	e := model.PositionEvent{
		PersonID:      req.PersonID,
		EventType:     req.EventType,
		Remark:        req.Remark,
		AttendanceGroup:    req.AttendanceGroup,
		HasAnnualLeave:     req.HasAnnualLeave,
		HasAttendanceBonus: req.HasAttendanceBonus,
		BaseSalary:         req.BaseSalary,
		PerformanceSalary:  req.PerformanceSalary,
		SalaryDays:         req.SalaryDays,
		PostAllowance:      req.PostAllowance,
		MealAllowance:      req.MealAllowance,
		HousingAllowance:   req.HousingAllowance,
		TransportAllowance: req.TransportAllowance,
		HighTempAllowance:  req.HighTempAllowance,
		InsuranceCompensation: req.InsuranceCompensation,
		FundCompensation:    req.FundCompensation,
		SocialSecurityDeduct: req.SocialSecurityDeduct,
		HousingFundDeduct:  req.HousingFundDeduct,
	}
	if req.EffectiveDate != "" {
		var d utils.DateOnly
		d.UnmarshalJSON([]byte(`"` + req.EffectiveDate + `"`))
		e.EffectiveDate = d
	}
	if req.EntryDate != nil && *req.EntryDate != "" {
		var d utils.DateOnly
		d.UnmarshalJSON([]byte(`"` + *req.EntryDate + `"`))
		e.EntryDate = &d
	}
	if req.LeaveDate != nil && *req.LeaveDate != "" {
		var d utils.DateOnly
		d.UnmarshalJSON([]byte(`"` + *req.LeaveDate + `"`))
		e.LeaveDate = &d
	}
	return e
}

func CreatePositionEvent(c *gin.Context) {
	var req positionEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if req.PersonID == 0 || req.EventType == "" || req.EffectiveDate == "" {
		utils.BadRequest(c, "人员、事件类型、生效日期为必填项")
		return
	}

	e := reqToModel(req)
	if err := service.CreatePositionEvent(&e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": e.ID})
}

func UpdatePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req positionEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	e := reqToModel(req)
	if err := service.UpdatePositionEvent(uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePositionEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestorePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestorePositionEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedPositionEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedPositionEvents(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}
