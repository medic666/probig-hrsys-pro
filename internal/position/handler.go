package position

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
	r.GET("/events", middleware.Permission("position:read"), ListEvents)
	r.POST("/events", middleware.Permission("position:write"), CreateEvent)
	r.PUT("/events/:id", middleware.Permission("position:write"), UpdateEvent)
	r.DELETE("/events/:id", middleware.Permission("position:delete"), DeleteEvent)
	r.GET("/snapshots/:personID", middleware.Permission("position:read"), GetSnapshots)
	r.GET("/current-snapshot/:personID", middleware.Permission("position:read"), GetCurrentSnapshot)
}

func ListEvents(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	personID := utils.ParseUint(c.Query("person_id"))
	svc := NewService()
	list, total, err := svc.ListEvents(pageNum, pageSize, personID, c.Query("start_date"), c.Query("end_date"), c.Query("event_name"))
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
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "职务事件", id, "", "新增", nil, req)
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
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "职务事件", id, "", "修改", nil, req)
	response.SuccessMsg(c, "更新成功")
}

func DeleteEvent(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeleteEvent(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "职务事件", id, "", "删除", nil, nil)
	response.SuccessMsg(c, "删除成功")
}

func GetSnapshots(c *gin.Context) {
	personID := utils.ParseUint(c.Param("personID"))
	svc := NewService()
	list, err := svc.GetSnapshots(personID)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func GetCurrentSnapshot(c *gin.Context) {
	personID := utils.ParseUint(c.Param("personID"))
	svc := NewService()
	snap, err := svc.GetCurrentSnapshot(personID)
	if err != nil {
		response.Success(c, nil)
		return
	}
	response.Success(c, snap)
}
