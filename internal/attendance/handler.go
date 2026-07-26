package attendance

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
		events.GET("", middleware.Permission("attendance:read"), ListEvents)
		events.POST("", middleware.Permission("attendance:write"), CreateEvent)
		events.POST("/cross-day", middleware.Permission("attendance:write"), CreateCrossDayEvent)
		events.POST("/batch", middleware.Permission("attendance:write"), CreateBatchEvents)
		events.PUT("/:id", middleware.Permission("attendance:write"), UpdateEvent)
		events.DELETE("/:id", middleware.Permission("attendance:delete"), DeleteEvent)
	}

	daily := r.Group("/daily")
	{
		daily.GET("", middleware.Permission("attendance:read"), ListDaily)
		daily.GET("/events/:personID/:date", middleware.Permission("attendance:read"), GetDailyEvents)
	}

	salary := r.Group("/salary")
	{
		salary.GET("", middleware.Permission("attendance:read"), ListMonthlySalary)
		salary.POST("/cal", middleware.Permission("attendance:calc"), CalculateMonthlySalary)
	}
}

func ListEvents(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.ListEvents(pageNum, pageSize,
		utils.ParseUint(c.Query("person_id")),
		c.Query("start_date"), c.Query("end_date"),
		c.Query("event_type"), c.Query("sub_type"))
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
	id, err := svc.CreateEvent(req)
	if err != nil {
		response.Error(c, "创建失败")
		return
	}
	response.Success(c, gin.H{"id": id})
}

func CreateCrossDayEvent(c *gin.Context) {
	var req struct {
		PersonID    uint    `json:"person_id" binding:"required"`
		StartDate   string  `json:"start_date" binding:"required"`
		EndDate     string  `json:"end_date" binding:"required"`
		EventType   string  `json:"event_type" binding:"required"`
		SubType     string  `json:"sub_type" binding:"required"`
		HoursPerDay float64 `json:"hours_per_day" binding:"required"`
		Remark      string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.CreateCrossDayEvent(req.PersonID, req.StartDate, req.EndDate, req.EventType, req.SubType, req.HoursPerDay, req.Remark); err != nil {
		response.Error(c, "创建失败")
		return
	}
	response.SuccessMsg(c, "创建成功")
}

func CreateBatchEvents(c *gin.Context) {
	var req struct {
		PersonIDs []uint                 `json:"person_ids" binding:"required"`
		Event     map[string]interface{} `json:"event" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	successCount, errors := svc.CreateEventsBatch(req.PersonIDs, req.Event)
	response.Success(c, gin.H{
		"success_count": successCount,
		"errors":        errors,
	})
}

func UpdateEvent(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.UpdateEvent(id, req); err != nil {
		response.Error(c, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func DeleteEvent(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeleteEvent(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func ListDaily(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.GetDailyList(pageNum, pageSize,
		utils.ParseUint(c.Query("person_id")),
		c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func GetDailyEvents(c *gin.Context) {
	personID := utils.ParseUint(c.Param("personID"))
	date := c.Param("date")
	svc := NewService()
	events, err := svc.GetDailyEvents(personID, date)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, events)
}

func ListMonthlySalary(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.GetMonthlySalaryList(
		c.Query("belong_month"),
		utils.ParseUint(c.Query("person_id")),
		pageNum, pageSize)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func CalculateMonthlySalary(c *gin.Context) {
	var req struct {
		PersonIDs   []string `json:"person_ids"`
		BelongMonth string   `json:"belong_month" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()

	if len(req.PersonIDs) == 0 {
		var persons []uint
		svc.DB.Table("person").Select("id").Pluck("id", &persons)
		for _, pid := range persons {
			req.PersonIDs = append(req.PersonIDs, strconv.FormatUint(uint64(pid), 10))
		}
	}

	successCount := 0
	var skipped []string
	for _, pidStr := range req.PersonIDs {
		pid := utils.ParseUint(pidStr)
		_, dailiesCount, _ := svc.GetDailyList(1, 1, pid, req.BelongMonth+"-01", "")
		if dailiesCount > 0 {
			if err := svc.CalculateMonthlySalary(pid, req.BelongMonth); err != nil {
				skipped = append(skipped, pidStr)
			} else {
				successCount++
			}
		} else {
			skipped = append(skipped, pidStr)
		}
	}

	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "假勤工资", 0, req.BelongMonth, "核算", nil, nil)
	response.Success(c, gin.H{
		"success_count": successCount,
		"skipped":       skipped,
	})
}
