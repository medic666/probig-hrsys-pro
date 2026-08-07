package handler

import (
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// positionEventListQuery 职务事件列表筛选解析（列表与导出共用）
func positionEventListQuery(c *gin.Context) service.PositionEventListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	return service.PositionEventListQuery{
		PageNum:   pageReq.PageNum,
		PageSize:  pageReq.PageSize,
		PersonID:  uint(personID),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		EventType: c.Query("event_type"),
	}
}

func GetPositionEvents(c *gin.Context) {
	q := positionEventListQuery(c)
	list, total, err := service.GetPositionEventList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, q.PageNum, q.PageSize)
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
	SalaryDays          *float64 `json:"salary_days"`
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
	// 入职事件必须配置计薪天数（工资核算前置），且必须大于 0
	if req.EventType == "入职" && (req.SalaryDays == nil || *req.SalaryDays <= 0) {
		utils.BadRequest(c, "入职事件必须填写计薪天数，且必须大于 0")
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
	// 编辑时若提供计薪天数，必须大于 0（入职与调薪调岗共用）
	if req.SalaryDays != nil && *req.SalaryDays <= 0 {
		utils.BadRequest(c, "计薪天数必须大于 0")
		return
	}

	// 与创建同构：reqToModel 统一字段解析，颗粒化更新语义由服务层 buildPositionEventUpdates 承担
	e := reqToModel(req)
	if err := service.UpdatePositionEvent(c.Request.Context(), uint(id), &e); err != nil {
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
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
}

// positionEventExportFilters 职务事件导出文件名筛选摘要
func positionEventExportFilters(q service.PositionEventListQuery) []string {
	var parts []string
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	if q.EventType != "" {
		parts = append(parts, "类型="+q.EventType)
	}
	if p := dateRangePiece("日期", q.StartDate, q.EndDate); p != "" {
		parts = append(parts, p)
	}
	return parts
}

func ExportPositionEvents(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := positionEventListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, _ := service.GetPositionEventList(q)

	var rows [][]interface{}
	for _, e := range list {
		rows = append(rows, []interface{}{
			e["person_name"], e["event_type"], e["changed_fields"], e["effective_date"], e["remark"],
		})
	}
	writeExcel(c, "职务事件",
		[]string{"人员", "事件类型", "变更字段", "生效日期", "备注"}, rows,
		positionEventExportFilters(q)...)
}
