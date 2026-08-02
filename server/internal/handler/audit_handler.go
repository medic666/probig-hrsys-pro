package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// auditLogListQuery 审计日志列表筛选解析（列表与导出共用）
func auditLogListQuery(c *gin.Context) service.AuditLogListQuery {
	pageReq := utils.BindPage(c)
	return service.AuditLogListQuery{
		PageNum:      pageReq.PageNum,
		PageSize:     pageReq.PageSize,
		OperatorName: c.Query("operator_name"),
		Action:       c.Query("action"),
		TargetType:   c.Query("target_type"),
		DateStart:    c.Query("date_start"),
		DateEnd:      c.Query("date_end"),
	}
}

func GetAuditLogs(c *gin.Context) {
	q := auditLogListQuery(c)
	list, total, err := service.GetAuditLogList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, utils.NewPageResult(list, total, utils.PageRequest{PageNum: q.PageNum, PageSize: q.PageSize}))
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

// auditExportFilters 审计日志导出文件名筛选摘要
func auditExportFilters(q service.AuditLogListQuery) []string {
	var parts []string
	if q.OperatorName != "" {
		parts = append(parts, "操作人="+q.OperatorName)
	}
	if q.Action != "" {
		parts = append(parts, "操作类型="+q.Action)
	}
	if q.TargetType != "" {
		parts = append(parts, "对象="+q.TargetType)
	}
	if p := dateRangePiece("时间", q.DateStart, q.DateEnd); p != "" {
		parts = append(parts, p)
	}
	return parts
}

func ExportAuditLogs(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := auditLogListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, err := service.GetAuditLogList(q)
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
	writeExcel(c, "审计日志",
		[]string{"操作人", "操作类型", "对象类型", "对象名称", "操作时间", "操作前快照", "操作后快照"}, rows,
		auditExportFilters(q)...)
}
