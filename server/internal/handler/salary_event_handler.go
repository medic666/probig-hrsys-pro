package handler

import (
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetSalaryEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetSalaryEventList(pageReq.PageNum, pageReq.PageSize,
		uint(personID), c.Query("belong_month"), c.Query("event_type"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func CreateSalaryEvent(c *gin.Context) {
	var e model.SalaryEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if e.PersonID == 0 || e.BelongMonth == "" || e.EventType == "" {
		utils.BadRequest(c, "人员、月份、事件类型为必填项")
		return
	}
	if err := service.CreateSalaryEvent(c.Request.Context(), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": e.ID})
}

func UpdateSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e model.SalaryEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdateSalaryEvent(c.Request.Context(), uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteSalaryEvent(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreSalaryEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreSalaryEvent(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedSalaryEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetDeletedSalaryEvents(pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func ExportSalaryEvents(c *gin.Context) {
	list, _, _ := service.GetSalaryEventList(1, 10000, 0, c.Query("belong_month"), c.Query("event_type"))

	var rows [][]interface{}
	for _, e := range list {
		rows = append(rows, []interface{}{
			e["person_name"], e["belong_month"], e["event_type"], e["amount"], e["remark"],
		})
	}
	writeExcel(c, "工资事件", "salary_events",
		[]string{"人员", "归属月份", "类型", "值", "备注"}, rows)
}
