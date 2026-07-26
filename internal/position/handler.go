package position

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"probig/internal/pkg/database"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"
)

func InitDB() {
	SetDB(database.DB)
}

type listEventsRequest struct {
	PersonID  uint   `form:"person_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	EventName string `form:"event_name"`
	PageNum   int    `form:"page_num"`
	PageSize  int    `form:"page_size"`
}

func ListEventsHandler(c *gin.Context) {
	var req listEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse(utils.DateFormat, req.StartDate)
		if err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse(utils.DateFormat, req.EndDate)
		if err == nil {
			endDate = &t
		}
	}

	events, total, err := ListEventsWithFilter(req.PersonID, startDate, endDate, req.EventName, req.PageNum, req.PageSize)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: events, Total: total})
}

func ListTrashEventsHandler(c *gin.Context) {
	var req listEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, _ := time.Parse(utils.DateFormat, req.StartDate)
		startDate = &t
	}
	if req.EndDate != "" {
		t, _ := time.Parse(utils.DateFormat, req.EndDate)
		endDate = &t
	}

	events, total, err := ListTrashEvents(req.PersonID, startDate, endDate, req.EventName, req.PageNum, req.PageSize)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: events, Total: total})
}

type createEventRequest struct {
	PersonID              uint     `json:"person_id" binding:"required"`
	EventName             string   `json:"event_name" binding:"required"`
	EffectiveDate         string   `json:"effective_date" binding:"required"`
	AttendanceGroup       *string  `json:"attendance_group"`
	HasAnnualLeave        *bool    `json:"has_annual_leave"`
	HasAttendanceBonus    *bool    `json:"has_attendance_bonus"`
	BaseSalary            *float64 `json:"base_salary"`
	PerformanceSalary     *float64 `json:"performance_salary"`
	SalaryDays            *int     `json:"salary_days"`
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

func CreateEventHandler(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	effDate, err := time.Parse(utils.DateFormat, req.EffectiveDate)
	if err != nil {
		response.Error(c, response.ParamError, "日期格式错误，应为YYYY-MM-DD")
		return
	}

	event := PositionEvent{
		PersonID:              req.PersonID,
		EventName:             req.EventName,
		EffectiveDate:         &effDate,
		AttendanceGroup:       req.AttendanceGroup,
		HasAnnualLeave:        req.HasAnnualLeave,
		HasAttendanceBonus:    req.HasAttendanceBonus,
		BaseSalary:            req.BaseSalary,
		PerformanceSalary:     req.PerformanceSalary,
		SalaryDays:            req.SalaryDays,
		PostAllowance:         req.PostAllowance,
		MealAllowance:         req.MealAllowance,
		HousingAllowance:      req.HousingAllowance,
		TransportAllowance:    req.TransportAllowance,
		HighTempAllowance:     req.HighTempAllowance,
		InsuranceCompensation: req.InsuranceCompensation,
		FundCompensation:      req.FundCompensation,
		SocialSecurityDeduct:  req.SocialSecurityDeduct,
		HousingFundDeduct:     req.HousingFundDeduct,
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := CreateEvent(&event, operatorID, operatorName, ip); err != nil {
		response.Error(c, response.BusinessError, "创建失败: "+err.Error())
		return
	}

	response.Success(c, event)
}

func UpdateEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "ID格式错误")
		return
	}

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	var existing PositionEvent
	if err := db().First(&existing, id).Error; err != nil {
		response.Error(c, response.NotFound, "事件不存在")
		return
	}

	if req.EffectiveDate != "" {
		effDate, err := time.Parse(utils.DateFormat, req.EffectiveDate)
		if err != nil {
			response.Error(c, response.ParamError, "日期格式错误")
			return
		}
		existing.EffectiveDate = &effDate
	}
	if req.EventName != "" {
		existing.EventName = req.EventName
	}

	existing.AttendanceGroup = req.AttendanceGroup
	existing.HasAnnualLeave = req.HasAnnualLeave
	existing.HasAttendanceBonus = req.HasAttendanceBonus
	existing.BaseSalary = req.BaseSalary
	existing.PerformanceSalary = req.PerformanceSalary
	existing.SalaryDays = req.SalaryDays
	existing.PostAllowance = req.PostAllowance
	existing.MealAllowance = req.MealAllowance
	existing.HousingAllowance = req.HousingAllowance
	existing.TransportAllowance = req.TransportAllowance
	existing.HighTempAllowance = req.HighTempAllowance
	existing.InsuranceCompensation = req.InsuranceCompensation
	existing.FundCompensation = req.FundCompensation
	existing.SocialSecurityDeduct = req.SocialSecurityDeduct
	existing.HousingFundDeduct = req.HousingFundDeduct

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := UpdateEvent(&existing, operatorID, operatorName, ip); err != nil {
		response.Error(c, response.BusinessError, "更新失败: "+err.Error())
		return
	}

	response.Success(c, existing)
}

func DeleteEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "ID格式错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := DeleteEventByID(id, operatorID, operatorName, ip); err != nil {
		response.Error(c, response.BusinessError, "删除失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func RestoreEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "ID格式错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := RestoreEventByID(id, operatorID, operatorName, ip); err != nil {
		response.Error(c, response.BusinessError, "恢复失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func GetEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "ID格式错误")
		return
	}

	event, err := GetEventByID(id)
	if err != nil {
		response.Error(c, response.NotFound, "事件不存在")
		return
	}

	response.Success(c, event)
}

func ListSnapshotsHandler(c *gin.Context) {
	personIDStr := c.Query("person_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageNumStr := c.Query("page_num")
	pageSizeStr := c.Query("page_size")

	pageNum, _ := strconv.Atoi(pageNumStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	personID, _ := strconv.ParseUint(personIDStr, 10, 64)

	var startDate, endDate *time.Time
	if startDateStr != "" {
		t, _ := time.Parse(utils.DateFormat, startDateStr)
		startDate = &t
	}
	if endDateStr != "" {
		t, _ := time.Parse(utils.DateFormat, endDateStr)
		endDate = &t
	}

	q := SnapshotQuery{
		PersonID:  uint(personID),
		StartDate: startDate,
		EndDate:   endDate,
		PageNum:   pageNum,
		PageSize:  pageSize,
	}

	snapshots, total, err := ListSnapshotsWithFilter(q)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: snapshots, Total: total})
}

func GetSnapshotHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "ID格式错误")
		return
	}

	snapshot, err := GetSnapshotByID(id)
	if err != nil {
		response.Error(c, response.NotFound, "快照不存在")
		return
	}

	response.Success(c, snapshot)
}

func GetCurrentSnapshotHandler(c *gin.Context) {
	personID, err := utils.ParseID(c, "personId")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID格式错误")
		return
	}

	snapshot, err := GetCurrentSnapshot(uint(personID))
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, snapshot)
}

func GetEmploymentStatusHandler(c *gin.Context) {
	personID, err := utils.ParseID(c, "personId")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID格式错误")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": GetEmploymentStatus(uint(personID)),
	})
}
