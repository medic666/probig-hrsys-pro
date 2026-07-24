package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"probig/internal/models"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetPositionEvents(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	events, err := service.GetPositionEvents(uint(personID))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, events)
}

func GetPositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	e, err := service.GetPositionEvent(uint(id))
	if err != nil {
		response.Error(c, "事件不存在")
		return
	}
	response.Success(c, e)
}

func CreatePositionEvent(c *gin.Context) {
	var e models.PositionEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		response.Error(c, "参数错误")
		return
	}
	if e.EffectiveDate == nil {
		now := time.Now()
		e.EffectiveDate = &now
	}
	if err := service.CreatePositionEvent(c, &e); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, e)
}

func UpdatePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e models.PositionEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		response.Error(c, "参数错误")
		return
	}
	e.ID = uint(id)
	if err := service.UpdatePositionEvent(c, &e); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, e)
}

func DeletePositionEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeletePositionEvent(c, uint(id)); err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func GetPositionSnapshots(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	snapshots, err := service.GetPositionSnapshots(uint(personID))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, snapshots)
}

func RebuildSnapshots(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	if err := service.RebuildPositionSnapshots(uint(personID)); err != nil {
		response.Error(c, "重建失败")
		return
	}
	response.Success(c, nil)
}
