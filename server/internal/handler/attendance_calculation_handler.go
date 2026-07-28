package handler

import (
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetMonthlyList(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetMonthlyList(c.Query("month"), uint(personID), pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

type calculateReq struct {
	Month     string `json:"month" binding:"required"`
	PersonIDs []uint `json:"person_ids"`
}

func CalculateMonthly(c *gin.Context) {
	var req calculateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if len(req.PersonIDs) == 0 {
		monthStart, err := time.Parse("2006-01", req.Month)
		if err != nil {
			utils.BadRequest(c, "月份格式错误")
			return
		}
		monthEnd := monthStart.AddDate(0, 1, -1)
		dao.DB.Model(&model.PositionSnapshot{}).
			Select("DISTINCT person_id").
			Where("effective_start_date <= ? AND effective_end_date >= ? AND is_active = true",
				utils.DateOnlyFromTime(monthEnd), utils.DateOnlyFromTime(monthStart)).
			Pluck("person_id", &req.PersonIDs)
		if len(req.PersonIDs) == 0 {
			utils.Error(c, "当月无在职人员")
			return
		}
	}

	success, fail, err := service.CalculateMonthlyBatch(req.Month, req.PersonIDs)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "核算完成", gin.H{"success": success, "fail": fail})
}
