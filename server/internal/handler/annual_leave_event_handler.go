package handler

import (
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetAnnualLeaveEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetAnnualLeaveEventList(pageReq.PageNum, pageReq.PageSize, uint(personID),
		c.Query("date_start"), c.Query("date_end"), c.Query("event_type"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func CreateAnnualLeaveEvent(c *gin.Context) {
	var e model.AnnualLeaveAccountEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if e.PersonID == 0 || e.EventType == "" {
		utils.BadRequest(c, "人员和事件类型为必填项")
		return
	}
	e.SourceType = "manual"
	if err := service.CreateAnnualLeaveEvent(&e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": e.ID})
}

func UpdateAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e model.AnnualLeaveAccountEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdateAnnualLeaveEvent(uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAnnualLeaveEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreAnnualLeaveEvent(uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedAnnualLeaveEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, _, err := service.GetDeletedAnnualLeaveEvents(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	var filtered []model.AnnualLeaveAccountEvent
	for _, e := range list {
		if e.SourceType != "system_period" {
			filtered = append(filtered, e)
		}
	}
	utils.Success(c, utils.NewPageResult(filtered, int64(len(filtered)), pageReq))
}
