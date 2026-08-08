package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetPersonAnnualLeaveBalance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	snapshot, err := service.GetCurrentAnnualLeaveBalance(uint(id))
	if err != nil {
		utils.Error(c, "暂无年假余额数据")
		return
	}
	utils.Success(c, snapshot)
}

func GetPersonAnnualLeaveHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	history, err := service.GetAnnualLeaveBalanceHistory(uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, history)
}

func GetPersonLILBalance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	snapshot, err := service.GetCurrentLILBalance(uint(id))
	if err != nil {
		utils.Error(c, "暂无调休余额数据")
		return
	}
	utils.Success(c, snapshot)
}

func GetPersonLILHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	history, err := service.GetLILBalanceHistory(uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, history)
}

func GetLILEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	dateStart, dateEnd := utils.BindDateRange(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	// 明细级查询：先过滤（仅已确认组的补班/调休）后分页，避免列表缺失
	list, total, err := service.GetLILEventList(c.Request.Context(), service.AttendanceDailyListQuery{
		PageNum:   pageReq.PageNum,
		PageSize:  pageReq.PageSize,
		PersonID:  uint(personID),
		DateStart: dateStart,
		DateEnd:   dateEnd,
	})
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
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
	userName := c.GetString("username")
	result, err := service.ExecuteCarryover(c.Request.Context(), req.Month, userID, userName)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, result)
}

func CancelCarryover(c *gin.Context) {
	batchID, _ := strconv.ParseUint(c.Param("batchId"), 10, 64)
	if err := service.CancelCarryover(c.Request.Context(), uint(batchID)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "反结账成功", nil)
}

func GetCarryoverBatches(c *gin.Context) {
	batches, err := service.GetCarryoverBatches(c.Request.Context())
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, batches)
}

func GetBatchEvents(c *gin.Context) {
	batchID, _ := strconv.ParseUint(c.Param("batchId"), 10, 64)
	events, err := service.GetBatchEvents(c.Request.Context(), uint(batchID))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, events)
}

func GetAllALBalances(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetAllALBalances(c.Request.Context(), pageReq.PageNum, pageReq.PageSize, uint(personID))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
}

func GetAllLILBalances(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetAllLILBalances(c.Request.Context(), pageReq.PageNum, pageReq.PageSize, uint(personID))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, pageReq.PageNum, pageReq.PageSize)
}

func GetALBalanceDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	detail, err := service.GetAnnualLeaveBalanceDetail(uint(id))
	if err != nil {
		utils.Error(c, "获取余额明细失败")
		return
	}
	utils.Success(c, detail)
}

func GetLILBalanceDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	detail, err := service.GetLILBalanceDetail(uint(id))
	if err != nil {
		utils.Error(c, "获取余额明细失败")
		return
	}
	utils.Success(c, detail)
}
