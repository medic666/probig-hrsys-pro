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
	CompanyID           *uint    `json:"company_id"`
	Department          *string  `json:"department"`
	Position            *string  `json:"position"`
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
		CompanyID:          req.CompanyID,
		Department:         req.Department,
		Position:           req.Position,
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
	if err := service.CreatePositionEvent(c.Request.Context(), &e); err != nil {
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
	if req.EventType == "" || req.EffectiveDate == "" {
		utils.BadRequest(c, "事件类型、生效日期为必填项")
		return
	}

	updates := map[string]interface{}{
		"event_type":     req.EventType,
		"remark":         req.Remark,
		"effective_date": req.EffectiveDate,
	}
	if req.EntryDate != nil && *req.EntryDate != "" {
		updates["entry_date"] = *req.EntryDate
	}
	if req.LeaveDate != nil && *req.LeaveDate != "" {
		updates["leave_date"] = *req.LeaveDate
	}
	if req.AttendanceGroup != nil {
		updates["attendance_group"] = *req.AttendanceGroup
	}
	if req.CompanyID != nil {
		updates["company_id"] = *req.CompanyID
	}
	if req.Department != nil {
		updates["department"] = *req.Department
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.HasAnnualLeave != nil {
		updates["has_annual_leave"] = *req.HasAnnualLeave
	}
	if req.HasAttendanceBonus != nil {
		updates["has_attendance_bonus"] = *req.HasAttendanceBonus
	}
	if req.BaseSalary != nil {
		updates["base_salary"] = *req.BaseSalary
	}
	if req.PerformanceSalary != nil {
		updates["performance_salary"] = *req.PerformanceSalary
	}
	if req.SalaryDays != nil {
		updates["salary_days"] = *req.SalaryDays
	}
	if req.PostAllowance != nil {
		updates["post_allowance"] = *req.PostAllowance
	}
	if req.MealAllowance != nil {
		updates["meal_allowance"] = *req.MealAllowance
	}
	if req.HousingAllowance != nil {
		updates["housing_allowance"] = *req.HousingAllowance
	}
	if req.TransportAllowance != nil {
		updates["transport_allowance"] = *req.TransportAllowance
	}
	if req.HighTempAllowance != nil {
		updates["high_temp_allowance"] = *req.HighTempAllowance
	}
	if req.InsuranceCompensation != nil {
		updates["insurance_compensation"] = *req.InsuranceCompensation
	}
	if req.FundCompensation != nil {
		updates["fund_compensation"] = *req.FundCompensation
	}
	if req.SocialSecurityDeduct != nil {
		updates["social_security_deduct"] = *req.SocialSecurityDeduct
	}
	if req.HousingFundDeduct != nil {
		updates["housing_fund_deduct"] = *req.HousingFundDeduct
	}

	if err := service.UpdatePositionEvent(c.Request.Context(), uint(id), updates); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeletePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePositionEvent(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestorePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestorePositionEvent(c.Request.Context(), uint(id)); err != nil {
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

func ExportPositionEvents(c *gin.Context) {
	list, _, _ := service.GetPositionEventList(1, 10000, 0, c.Query("start_date"), c.Query("end_date"), c.Query("event_type"))

	var rows [][]interface{}
	for _, e := range list {
		rows = append(rows, []interface{}{
			e["person_name"], e["event_type"], e["changed_fields"], e["effective_date"], e["remark"],
		})
	}
	writeExcel(c, "职务事件", "position_events",
		[]string{"人员", "事件类型", "变更字段", "生效日期", "备注"}, rows)
}
