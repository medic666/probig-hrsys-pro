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
	dateStart := c.Query("date_start")
	dateEnd := c.Query("date_end")
	eventType := c.Query("event_type")

	list, _, err := service.GetAnnualLeaveEventList(pageReq.PageNum, pageReq.PageSize, uint(personID),
		dateStart, dateEnd, eventType)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if len(list) > 0 {
		for i := range list {
			list[i]["source"] = "account"
		}
	}

	if eventType == "" || eventType == "休假" {
		attDailyList, _, _ := service.GetAttendanceDailyList(uint(personID), dateStart, dateEnd, "", pageReq.PageNum, pageReq.PageSize)
		for _, daily := range attDailyList {
			if details, ok := daily["details"].([]map[string]interface{}); ok {
				for _, d := range details {
					if sub, _ := d["sub_type"].(string); sub == "年假" {
						h, _ := d["hours"].(float64)
						d["event_type"] = "休假(年假)"
						d["source_type"] = "attendance"
						d["source"] = "attendance"
						d["person_id"] = daily["person_id"]
						d["person_name"] = daily["person_name"]
						d["event_date"] = daily["event_date"]
						d["hours"] = -h
						list = append(list, d)
					}
				}
			}
		}
	}

	utils.Success(c, utils.NewPageResult(list, int64(len(list)), pageReq))
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
	if err := service.CreateAnnualLeaveEvent(c.Request.Context(), &e); err != nil {
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
	if err := service.UpdateAnnualLeaveEvent(c.Request.Context(), uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAnnualLeaveEvent(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreAnnualLeaveEvent(c.Request.Context(), uint(id)); err != nil {
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

func ExportAnnualLeaveEvents(c *gin.Context) {
	list, _, _ := service.GetAnnualLeaveEventList(1, 10000, 0, c.Query("date_start"), c.Query("date_end"), c.Query("event_type"))

	var rows [][]interface{}
	for _, e := range list {
		rows = append(rows, []interface{}{
			e["person_name"], e["event_type"], e["source_type"], e["hours"], e["effective_date"], e["remark"],
		})
	}
	writeExcel(c, "年假事件", "annual_leave_events",
		[]string{"人员", "类型", "来源", "变动时长(小时)", "生效日期", "备注"}, rows)
}
