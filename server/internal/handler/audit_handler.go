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

func ExportAuditLogs(c *gin.Context) {
	list, _, err := service.GetAuditLogList(1, 10000,
		c.Query("operator_name"), c.Query("action"), c.Query("target_type"), c.Query("date_start"), c.Query("date_end"))
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	var rows [][]interface{}
	for _, l := range list {
		rows = append(rows, []interface{}{
			l.OperatorName, l.Action, l.TargetType, l.TargetName,
			l.CreatedAt.Format("2006-01-02 15:04:05"), l.BeforeSnapshot, l.AfterSnapshot,
		})
	}
	writeExcel(c, "审计日志", "audit_logs",
		[]string{"操作人", "操作类型", "对象类型", "对象名称", "操作时间", "操作前快照", "操作后快照"}, rows)
}
