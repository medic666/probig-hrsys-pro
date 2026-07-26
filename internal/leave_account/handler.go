package leave_account

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
	PersonID   uint   `form:"person_id"`
	LeaveType  string `form:"leave_type"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	SourceType string `form:"source_type"`
	PageNum    int    `form:"page_num"`
	PageSize   int    `form:"page_size"`
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

	events, total, err := ListEventsWithFilter(req.PersonID, req.LeaveType, startDate, endDate, req.SourceType, req.PageNum, req.PageSize)
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

	events, total, err := ListTrashEvents(req.PersonID, req.LeaveType, req.PageNum, req.PageSize)
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
	PersonID      uint    `json:"person_id" binding:"required"`
	LeaveType     string  `json:"leave_type" binding:"required"`
	EventType     string  `json:"event_type" binding:"required"`
	Hours         float64 `json:"hours"`
	EffectiveDate string  `json:"effective_date" binding:"required"`
	Remark        string  `json:"remark"`
}

func CreateEventHandler(c *gin.Context) {
	var req createEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	effDate, err := time.Parse(utils.DateFormat, req.EffectiveDate)
	if err != nil {
		response.Error(c, response.ParamError, "日期格式错误，应为YYYY-MM-DD")
		return
	}

	event := LeaveAccountEvent{
		PersonID:      req.PersonID,
		LeaveType:     req.LeaveType,
		EventType:     req.EventType,
		Hours:         req.Hours,
		EffectiveDate: &effDate,
		Remark:        req.Remark,
		SourceType:    "manual",
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := CreateManualEvent(&event, operatorID, operatorName, ip); err != nil {
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

	var existing LeaveAccountEvent
	if err := db().First(&existing, id).Error; err != nil {
		response.Error(c, response.NotFound, "事件不存在")
		return
	}
	if existing.SourceType != "manual" {
		response.Error(c, response.BusinessError, "系统结转事件不可编辑")
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
	existing.LeaveType = req.LeaveType
	existing.EventType = req.EventType
	existing.Hours = req.Hours
	existing.Remark = req.Remark

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := UpdateManualEvent(&existing, operatorID, operatorName, ip); err != nil {
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

	var event LeaveAccountEvent
	if err := db().First(&event, id).Error; err != nil {
		response.Error(c, response.NotFound, "事件不存在")
		return
	}
	if event.SourceType != "manual" {
		response.Error(c, response.BusinessError, "系统结转事件不可删除")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := DeleteManualEvent(id, operatorID, operatorName, ip); err != nil {
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

func ListBalancesHandler(c *gin.Context) {
	personIDStr := c.Query("person_id")
	leaveType := c.Query("leave_type")

	personID, _ := strconv.ParseUint(personIDStr, 10, 64)

	balances, err := ListBalancesWithFilter(uint(personID), leaveType)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, balances)
}

func GetBalanceDetailHandler(c *gin.Context) {
	personID, err := utils.ParseID(c, "personId")
	if err != nil {
		response.Error(c, response.ParamError, "人员ID格式错误")
		return
	}
	leaveType := c.Param("leaveType")
	if leaveType == "" {
		response.Error(c, response.ParamError, "假期类型不能为空")
		return
	}

	detail, err := GetBalanceDetail(uint(personID), leaveType)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, detail)
}

type carryoverReq struct {
	Month string `json:"month" binding:"required"`
}

func ExecuteCarryoverHandler(c *gin.Context) {
	var req carryoverReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	batchID, successCount, failCount, err := ExecuteCarryover(req.Month, operatorID, operatorName, ip)
	if err != nil {
		response.Error(c, response.BusinessError, "结转失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"batch_id":      batchID,
		"success_count": successCount,
		"fail_count":    failCount,
	})
}

func CancelBatchHandler(c *gin.Context) {
	batchID, err := utils.ParseID(c, "batchId")
	if err != nil {
		response.Error(c, response.ParamError, "批次ID格式错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	ip := c.ClientIP()

	if err := CancelBatchByID(uint(batchID), operatorID, operatorName, ip); err != nil {
		response.Error(c, response.BusinessError, "反结账失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

func ListBatchesHandler(c *gin.Context) {
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

	batches, total, err := ListBatches(pageNum, pageSize)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, response.PageResult{List: batches, Total: total})
}
