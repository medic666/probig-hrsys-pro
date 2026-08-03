package handler

import (
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// 人员徽章聚合接口：各模块返回 [{person_id, level}]（gray/green/orange/red），
// 前端直接渲染卡片右上角颜色点，不参与运算。month 缺省为当前时间的上一个月。

func GetPositionEventBadges(c *gin.Context) {
	badges, err := service.GetPositionEventBadges()
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, badges)
}

func GetAttendanceEventBadges(c *gin.Context) {
	month := c.DefaultQuery("month", service.DefaultBadgeMonth())
	badges, err := service.GetAttendanceEventBadges(month)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, badges)
}

func GetDailyProjectionBadges(c *gin.Context) {
	month := c.DefaultQuery("month", service.DefaultBadgeMonth())
	badges, err := service.GetDailyProjectionBadges(month)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, badges)
}

func GetAttendanceMonthlyBadges(c *gin.Context) {
	month := c.DefaultQuery("month", service.DefaultBadgeMonth())
	badges, err := service.GetAttendanceMonthlyBadges(month)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, badges)
}
