package handlers

import (
	"strconv"
	"time"

	"probig/database"
	"probig/middleware"
	"probig/models"
	"probig/services"
	"probig/utils"

	"github.com/gin-gonic/gin"
)

type attEventInput struct {
	PersonID          uint     `json:"person_id"`
	EventDate         string   `json:"event_date"`
	EventType         string   `json:"event_type"`
	SubType           string   `json:"sub_type"`
	Hours             *float64 `json:"hours"`
	LateMinutes       *int     `json:"late_minutes"`
	LeaveAdjustAmount *float64 `json:"leave_adjust_amount"`
	IsSpecialApproval *bool    `json:"is_special_approval"`
	Remark            string   `json:"remark"`
}

func toAttEvent(input attEventInput) models.AttendanceEvent {
	date, _ := time.Parse("2006-01-02", input.EventDate)
	return models.AttendanceEvent{
		PersonID:          input.PersonID,
		EventDate:         date,
		EventType:         input.EventType,
		SubType:           input.SubType,
		Hours:             input.Hours,
		LateMinutes:       input.LateMinutes,
		LeaveAdjustAmount: input.LeaveAdjustAmount,
		IsSpecialApproval: input.IsSpecialApproval,
		Remark:            input.Remark,
	}
}

func ListAttendanceEvents(c *gin.Context) {
	personIDStr := c.Query("person_id")
	var personID uint
	if personIDStr != "" {
		pid, _ := strconv.ParseUint(personIDStr, 10, 64)
		personID = uint(pid)
	}
	eventType := c.Query("event_type")
	subType := c.Query("sub_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize

	events, total, err := services.ListAttendanceEvents(personID, eventType, subType, startDate, endDate, offset, pageSize)
	if err != nil {
		utils.ErrInternal(c, err.Error())
		return
	}
	utils.SuccessPage(c, events, total, page, pageSize)
}

func CreateAttendanceEvent(c *gin.Context) {
	var input attEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	event := toAttEvent(input)
	if err := services.CreateAttendanceEvent(&event); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "attendance_event", event.ID, "新增", "{}", event)
	utils.Success(c, event)
}

func BatchCreateAttendanceEvents(c *gin.Context) {
	var input struct {
		PersonIDs          []uint   `json:"person_ids"`
		StartDate          string   `json:"start_date"`
		EndDate            string   `json:"end_date"`
		EventType          string   `json:"event_type"`
		SubType            string   `json:"sub_type"`
		Hours              *float64 `json:"hours"`
		LateMinutes        *int     `json:"late_minutes"`
		LeaveAdjustAmount  *float64 `json:"leave_adjust_amount"`
		IsSpecialApproval  *bool    `json:"is_special_approval"`
		Remark             string   `json:"remark"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		utils.ErrBadRequest(c, "开始日期格式错误")
		return
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		utils.ErrBadRequest(c, "结束日期格式错误")
		return
	}

	batchID, err := services.BatchCreateAttendanceEvents(c, input.PersonIDs, startDate, endDate,
		input.EventType, input.SubType, input.Hours, input.LateMinutes, input.LeaveAdjustAmount,
		input.IsSpecialApproval, input.Remark)
	if err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}

	middleware.WriteAuditLog(c, "attendance_event", 0, "批量新增", "{}", "batch created", batchID)
	utils.Success(c, gin.H{"batch_id": batchID})
}

func UpdateAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	c.ShouldBindJSON(&updates)
	var event models.AttendanceEvent
	database.DB.First(&event, uint(id))
	before := event
	if err := services.UpdateAttendanceEvent(uint(id), updates); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "attendance_event", uint(id), "修改", before, event)
	utils.Success(c, nil)
}

func DeleteAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var event models.AttendanceEvent
	database.DB.First(&event, uint(id))
	before := event
	if err := services.DeleteAttendanceEvent(uint(id)); err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.AuditAction(c, "attendance_event", uint(id), "删除", before, "{}")
	utils.Success(c, nil)
}

func CalcAttendanceSummary(c *gin.Context) {
	var input struct {
		BelongMonth string `json:"belong_month"`
		PersonIDs   []uint `json:"person_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrBadRequest(c, "参数错误")
		return
	}
	summaries, err := services.CalcAttendanceSummary(c, input.BelongMonth, input.PersonIDs)
	if err != nil {
		utils.ErrBadRequest(c, err.Error())
		return
	}
	middleware.WriteAuditLog(c, "attendance_summary", 0, "核算", "{}", input, "")
	utils.Success(c, summaries)
}

func ListAttendanceSummaries(c *gin.Context) {
	month := c.Query("belong_month")
	page, pageSize := utils.GetPageParams(c)
	offset := (page - 1) * pageSize
	summaries, total, _ := services.ListAttendanceSummaries(month, offset, pageSize)
	utils.SuccessPage(c, summaries, total, page, pageSize)
}

func LockAttendanceSummary(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		IsLocked bool `json:"is_locked"`
	}
	c.ShouldBindJSON(&input)
	services.LockAttendanceSummary(uint(id), input.IsLocked)
	action := "解锁"
	if input.IsLocked {
		action = "锁定"
	}
	middleware.AuditAction(c, "attendance_summary", uint(id), action, "{}", input)
	utils.Success(c, nil)
}

func GetAnnualLeaveBalance(c *gin.Context) {
	personIDStr := c.Param("person_id")
	personID, _ := strconv.ParseUint(personIDStr, 10, 64)
	balance, _ := services.CalcAnnualLeaveBalance(uint(personID))
	utils.Success(c, gin.H{"balance": balance})
}

func AnnualLeaveAnniversary(c *gin.Context) {
	personIDs, _ := services.AnnualLeaveAnniversary()
	utils.Success(c, gin.H{"eligible_person_ids": personIDs})
}
