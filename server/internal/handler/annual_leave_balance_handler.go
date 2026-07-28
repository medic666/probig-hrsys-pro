package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetPersonAnnualLeaveBalance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	snapshot, err := service.GetCurrentAnnualLeaveBalance(uint(id))
	if err != nil {
		utils.Error(c, "暂无年假余额数据")
		return
	}
	utils.Success(c, snapshot)
}

func GetPersonAnnualLeaveHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	history, err := service.GetAnnualLeaveBalanceHistory(uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, history)
}

func GetPersonLILBalance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	snapshot, err := service.GetCurrentLILBalance(uint(id))
	if err != nil {
		utils.Error(c, "暂无调休余额数据")
		return
	}
	utils.Success(c, snapshot)
}

func GetPersonLILHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	history, err := service.GetLILBalanceHistory(uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, history)
}

func GetLILEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, _, err := service.GetAttendanceEventList(pageReq.PageNum, pageReq.PageSize,
		uint(personID), c.Query("date_start"), c.Query("date_end"), "", "")
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	var filtered []map[string]interface{}
	for _, e := range list {
		sub, ok := e["sub_type"].(string)
		if ok && (sub == "补班出勤" || sub == "调休") {
			filtered = append(filtered, e)
		}
	}
	utils.Success(c, utils.NewPageResult(filtered, int64(len(filtered)), pageReq))
}

type carryoverReq struct {
	Month string `json:"month" binding:"required"`
}

func ExecuteAnnualLeaveCarryover(c *gin.Context) {
	var req carryoverReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	userID := c.GetUint("userID")
	result, err := service.ExecuteCarryover(req.Month, userID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, result)
}

func CancelCarryover(c *gin.Context) {
	batchID, _ := strconv.ParseUint(c.Param("batchId"), 10, 64)
	if err := service.CancelCarryover(uint(batchID)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "反结账成功", nil)
}

func GetCarryoverBatches(c *gin.Context) {
	batches, err := service.GetCarryoverBatches()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, batches)
}

func GetBatchEvents(c *gin.Context) {
	batchID, _ := strconv.ParseUint(c.Param("batchId"), 10, 64)
	events, err := service.GetBatchEvents(uint(batchID))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, events)
}
