package leave_account

import (
	"probig/internal/pkg/audit"
	"probig/internal/pkg/middleware"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.Use(middleware.Auth)
	events := r.Group("/events")
	{
		events.GET("", middleware.Permission("leave:read"), ListEvents)
		events.POST("", middleware.Permission("leave:write"), CreateEvent)
		events.DELETE("/:id", middleware.Permission("leave:delete"), DeleteEvent)
	}

	balance := r.Group("/balances")
	{
		balance.GET("", middleware.Permission("leave:read"), ListBalances)
		balance.GET("/detail/:personID/:leaveType", middleware.Permission("leave:read"), GetBalanceDetail)
	}

	carryover := r.Group("/carryover")
	{
		carryover.GET("/batches", middleware.Permission("leave:read"), ListBatches)
		carryover.GET("/batches/:id/events", middleware.Permission("leave:read"), GetBatchEvents)
		carryover.POST("/execute", middleware.Permission("leave:carryover"), ExecuteCarryover)
		carryover.POST("/cancel/:id", middleware.Permission("leave:carryover"), CancelCarryover)
	}
}

func ListEvents(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.ListEvents(pageNum, pageSize,
		utils.ParseUint(c.Query("person_id")),
		c.Query("leave_type"), c.Query("start_date"), c.Query("end_date"), c.Query("source_type"))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func CreateEvent(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	id, err := svc.CreateManualEvent(req)
	if err != nil {
		response.Error(c, "创建失败")
		return
	}
	response.Success(c, gin.H{"id": id})
}

func DeleteEvent(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeleteEvent(id); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func ListBalances(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.GetBalanceList(pageNum, pageSize,
		utils.ParseUint(c.Query("person_id")),
		c.Query("leave_type"))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func GetBalanceDetail(c *gin.Context) {
	personID := utils.ParseUint(c.Param("personID"))
	leaveType := c.Param("leaveType")
	svc := NewService()
	detail, err := svc.GetBalanceDetail(personID, leaveType)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, detail)
}

func ListBatches(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.GetBatchList(pageNum, pageSize)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func GetBatchEvents(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	events, err := svc.GetBatchEvents(id)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, events)
}

func ExecuteCarryover(c *gin.Context) {
	var req struct {
		TargetMonth string `json:"target_month" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	svc := NewService()
	count, err := svc.CarryoverAnnualLeave(req.TargetMonth, userID.(uint), uname.(string))
	if err != nil {
		response.Error(c, "结转失败: "+err.Error())
		return
	}
	audit.Write(c, userID.(uint), uname.(string), "年假结转", 0, req.TargetMonth, "结转", nil, nil)
	response.Success(c, gin.H{"processed_count": count})
}

func CancelCarryover(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.CancelCarryover(id); err != nil {
		response.Error(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "年假结转", id, "", "反结账", nil, nil)
	response.SuccessMsg(c, "反结账成功")
}
