package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetAttendanceEventList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	eventType := c.Query("event_type")
	subType := c.Query("sub_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	list, total, err := service.GetAttendanceEventList(page, pageSize, uint(personID), eventType, subType, startDate, endDate)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

func GetAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	e, err := service.GetAttendanceEvent(uint(id))
	if err != nil {
		response.Error(c, "事件不存在")
		return
	}
	response.Success(c, e)
}

func CreateAttendanceEvent(c *gin.Context) {
	var req struct {
		PersonID          uint     `json:"person_id"`
		EventDate         string   `json:"event_date"`
		EndDate           string   `json:"end_date"`
		PersonIDs         []uint   `json:"person_ids"`
		AttendanceGroup   string   `json:"attendance_group"`
		IsCrossDay        bool     `json:"is_cross_day"`
		IsBatch           bool     `json:"is_batch"`
		EventType         string   `json:"event_type"`
		SubType           string   `json:"sub_type"`
		Hours             *float64 `json:"hours"`
		LateMinutes       *int     `json:"late_minutes"`
		LeaveAdjustAmount *float64 `json:"leave_adjust_amount"`
		IsSpecialApproval bool     `json:"is_special_approval"`
		Remark            string   `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}

	if req.IsCrossDay && req.EndDate != "" {
		if err := service.CreateAttendanceEventRange(c, req.PersonID, req.EventDate, req.EndDate,
			req.EventType, req.SubType, req.Hours, req.Remark); err != nil {
			response.Error(c, err.Error())
			return
		}
		response.Success(c, nil)
		return
	}

	if req.IsBatch && len(req.PersonIDs) > 0 {
		if err := service.BatchCreateAttendanceEvents(c, req.PersonIDs, req.EventDate,
			req.EventType, req.SubType, req.Hours, req.Remark); err != nil {
			response.Error(c, err.Error())
			return
		}
		response.Success(c, nil)
		return
	}

	date, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		response.Error(c, "日期格式错误")
		return
	}
	e := &models.AttendanceEvent{
		PersonID:          req.PersonID,
		EventDate:         &date,
		EventType:         req.EventType,
		SubType:           req.SubType,
		Hours:             req.Hours,
		LateMinutes:       req.LateMinutes,
		LeaveAdjustAmount: req.LeaveAdjustAmount,
		IsSpecialApproval: req.IsSpecialApproval,
		Remark:            req.Remark,
	}
	if err := service.CreateAttendanceEvent(c, e); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, e)
}

func UpdateAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e models.AttendanceEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		response.Error(c, "参数错误")
		return
	}
	e.ID = uint(id)
	if err := service.UpdateAttendanceEvent(c, &e); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, e)
}

func DeleteAttendanceEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAttendanceEvent(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}
