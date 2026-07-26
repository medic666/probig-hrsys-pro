package attendance

import (
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

type listEventsQuery struct {
	PersonID  uint   `form:"person_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	EventType string `form:"event_type"`
	SubType   string `form:"sub_type"`
	PageNum   int    `form:"page_num"`
	PageSize  int    `form:"page_size"`
}

func ListEventsHandler(c *gin.Context) {
	var req listEventsQuery
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

	events, total, err := ListEvents(req.PersonID, startDate, endDate, req.EventType, req.SubType, req.PageNum, req.PageSize)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, response.PageResult{List: events, Total: total})
}

func ListTrashHandler(c *gin.Context) {
	var req listEventsQuery
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

	events, total, err := ListTrashEvents(req.PersonID, startDate, endDate, req.EventType, req.SubType, req.PageNum, req.PageSize)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, response.PageResult{List: events, Total: total})
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

type createEventReq struct {
	PersonID          uint    `json:"person_id" binding:"required"`
	EventDate         string  `json:"event_date" binding:"required"`
	PunchTime         string  `json:"punch_time"`
	EventType         string  `json:"event_type" binding:"required"`
	SubType           string  `json:"sub_type" binding:"required"`
	Hours             float64 `json:"hours"`
	LateMinutes       int     `json:"late_minutes"`
	IsSpecialApproval bool    `json:"is_special_approval"`
	Remark            string  `json:"remark"`
}

func CreateEventHandler(c *gin.Context) {
	var req createEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	eventDate, err := time.Parse(utils.DateFormat, req.EventDate)
	if err != nil {
		response.Error(c, response.ParamError, "日期格式错误，应为YYYY-MM-DD")
		return
	}
	ed := time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 0, 0, 0, 0, time.UTC)

	event := AttendanceEvent{
		PersonID:          req.PersonID,
		EventDate:         &ed,
		PunchTime:         req.PunchTime,
		EventType:         req.EventType,
		SubType:           req.SubType,
		Hours:             req.Hours,
		LateMinutes:       req.LateMinutes,
		IsSpecialApproval: req.IsSpecialApproval,
		Remark:            req.Remark,
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := CreateAttendanceEvent(&event, operatorID, operatorName, ip); err != nil {
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

	var req createEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	var existing AttendanceEvent
	if err := db().First(&existing, id).Error; err != nil {
		response.Error(c, response.NotFound, "事件不存在")
		return
	}

	if req.EventDate != "" {
		eventDate, err := time.Parse(utils.DateFormat, req.EventDate)
		if err != nil {
			response.Error(c, response.ParamError, "日期格式错误")
			return
		}
		ed := time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 0, 0, 0, 0, time.UTC)
		existing.EventDate = &ed
	}
	existing.EventType = req.EventType
	existing.SubType = req.SubType
	existing.Hours = req.Hours
	existing.PunchTime = req.PunchTime
	existing.LateMinutes = req.LateMinutes
	existing.IsSpecialApproval = req.IsSpecialApproval
	existing.Remark = req.Remark

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := UpdateAttendanceEvent(&existing, operatorID, operatorName, ip); err != nil {
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

	if err := DeleteAttendanceEvent(id, operatorID, operatorName, ip); err != nil {
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

	if err := RestoreAttendanceEvent(id, operatorID, operatorName, ip); err != nil {
		response.Error(c, response.BusinessError, "恢复失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

type batchCreateReq struct {
	PersonIDs         []uint  `json:"person_ids" binding:"required"`
	EventDate         string  `json:"event_date" binding:"required"`
	EventType         string  `json:"event_type" binding:"required"`
	SubType           string  `json:"sub_type" binding:"required"`
	Hours             float64 `json:"hours"`
	PunchTime         string  `json:"punch_time"`
	IsSpecialApproval bool    `json:"is_special_approval"`
	Remark            string  `json:"remark"`
}

func BatchCreateHandler(c *gin.Context) {
	var req batchCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	eventDate, err := time.Parse(utils.DateFormat, req.EventDate)
	if err != nil {
		response.Error(c, response.ParamError, "日期格式错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	events, errors, err := CreateBatchEvents(req.PersonIDs, eventDate, req.EventType, req.SubType, req.Hours, req.PunchTime, req.Remark, req.IsSpecialApproval, operatorID, operatorName, ip)
	if err != nil {
		response.Error(c, response.BusinessError, "批量创建失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"success_count": len(events),
		"fail_count":    len(errors),
		"events":        events,
		"errors":        errors,
	})
}

type crossDayReq struct {
	PersonID          uint    `json:"person_id" binding:"required"`
	StartDate         string  `json:"start_date" binding:"required"`
	EndDate           string  `json:"end_date" binding:"required"`
	EventType         string  `json:"event_type" binding:"required"`
	SubType           string  `json:"sub_type" binding:"required"`
	DailyHours        float64 `json:"daily_hours"`
	PunchTime         string  `json:"punch_time"`
	IsSpecialApproval bool    `json:"is_special_approval"`
	Remark            string  `json:"remark"`
}

func CrossDayCreateHandler(c *gin.Context) {
	var req crossDayReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	startDate, err := time.Parse(utils.DateFormat, req.StartDate)
	if err != nil {
		response.Error(c, response.ParamError, "起始日期格式错误")
		return
	}
	endDate, err := time.Parse(utils.DateFormat, req.EndDate)
	if err != nil {
		response.Error(c, response.ParamError, "结束日期格式错误")
		return
	}
	if endDate.Before(startDate) {
		response.Error(c, response.ParamError, "结束日期不能早于起始日期")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	events, errors, err := CreateCrossDayEvents(req.PersonID, startDate, endDate, req.EventType, req.SubType, req.DailyHours, req.PunchTime, req.Remark, req.IsSpecialApproval, operatorID, operatorName, ip)
	if err != nil {
		response.Error(c, response.BusinessError, "跨天创建失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"success_count": len(events),
		"fail_count":    len(errors),
		"events":        events,
		"errors":        errors,
	})
}

func ListDailyHandler(c *gin.Context) {
	personIDStr := c.Query("person_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageNumStr := c.Query("page_num")
	pageSizeStr := c.Query("page_size")

	personID, _ := strconv.ParseUint(personIDStr, 10, 64)
	pageNum, _ := strconv.Atoi(pageNumStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		t, _ := time.Parse(utils.DateFormat, startDateStr)
		startDate = &t
	}
	if endDateStr != "" {
		t, _ := time.Parse(utils.DateFormat, endDateStr)
		endDate = &t
	}

	projections, total, err := ListDailyProjections(uint(personID), startDate, endDate, pageNum, pageSize)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, response.PageResult{List: projections, Total: total})
}

func ListMonthlyHandler(c *gin.Context) {
	personIDStr := c.Query("person_id")
	belongMonth := c.Query("belong_month")
	startMonth := c.Query("start_month")
	endMonth := c.Query("end_month")
	pageNumStr := c.Query("page_num")
	pageSizeStr := c.Query("page_size")

	personID, _ := strconv.ParseUint(personIDStr, 10, 64)
	pageNum, _ := strconv.Atoi(pageNumStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	q := MonthlySalaryQuery{
		PersonID:    uint(personID),
		BelongMonth: belongMonth,
		StartMonth:  startMonth,
		EndMonth:    endMonth,
		PageNum:     pageNum,
		PageSize:    pageSize,
	}

	results, total, err := ListMonthlySalary(q)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, response.PageResult{List: results, Total: total})
}

type calcMonthlyReq struct {
	BelongMonth string `json:"belong_month" binding:"required"`
	PersonIDs   []uint `json:"person_ids"`
}

func CalcMonthlyHandler(c *gin.Context) {
	var req calcMonthlyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	results, errors, err := CalcMonthlySalary(req.BelongMonth, req.PersonIDs, operatorID, operatorName, ip)
	if err != nil {
		response.Error(c, response.BusinessError, "核算失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"success_count": len(results),
		"fail_count":    len(errors),
		"results":       results,
		"errors":        errors,
	})
}

func ExportMonthlyHandler(c *gin.Context) {
	belongMonth := c.Query("belong_month")
	if belongMonth == "" {
		response.Error(c, response.ParamError, "请指定月份")
		return
	}

	results, err := GetMonthlyByMonth(belongMonth)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, results)
}
