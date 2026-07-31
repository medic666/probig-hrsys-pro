package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetAuditLogs(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, total, err := service.GetAuditLogList(pageReq.PageNum, pageReq.PageSize,
		c.Query("operator_name"), c.Query("action"), c.Query("target_type"), c.Query("date_start"), c.Query("date_end"))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, pageReq))
}

func GetAuditLogDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	log, err := service.GetAuditLogByID(uint(id))
	if err != nil {
		utils.Error(c, "审计记录不存在")
		return
	}
	utils.Success(c, log)
}
