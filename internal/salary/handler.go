package salary

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
		events.GET("", middleware.Permission("salary:read"), ListEvents)
		events.POST("", middleware.Permission("salary:write"), CreateEvent)
		events.PUT("/:id", middleware.Permission("salary:write"), UpdateEvent)
		events.DELETE("/:id", middleware.Permission("salary:delete"), DeleteEvent)
	}

	summary := r.Group("/summaries")
	{
		summary.GET("", middleware.Permission("salary:read"), ListSummaries)
		summary.POST("/cal", middleware.Permission("salary:calc"), CalculateSummaries)
	}
}

func ListEvents(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.ListEvents(pageNum, pageSize,
		utils.ParseUint(c.Query("person_id")),
		c.Query("belong_month"), c.Query("event_type"))
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

func ListSummaries(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.GetSummaryList(
		c.Query("belong_month"),
		utils.ParseUint(c.Query("person_id")),
		pageNum, pageSize)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func CalculateSummaries(c *gin.Context) {
	var req struct {
		PersonIDs   []uint `json:"person_ids"`
		BelongMonth string `json:"belong_month" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()

	var skipped []uint
	successCount := 0

	for _, pid := range req.PersonIDs {
		if !svc.HasAttendaceSalary(pid, req.BelongMonth) {
			skipped = append(skipped, pid)
			continue
		}
		if err := svc.CalculateSummary(pid, req.BelongMonth); err != nil {
			skipped = append(skipped, pid)
		} else {
			successCount++
		}
	}

	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "工资汇总", 0, req.BelongMonth, "核算", nil, nil)
	response.Success(c, gin.H{
		"success_count": successCount,
		"skipped":       skipped,
	})
}
