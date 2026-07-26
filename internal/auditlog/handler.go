package auditlog

import (
	"probig/internal/pkg/excel"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

func ListHandler(c *gin.Context) {
	var filter AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	logs, total, err := List(filter)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: logs, Total: total})
}

func GetDetailHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	log, err := GetDetail(id)
	if err != nil {
		response.Error(c, response.NotFound, "审计日志不存在")
		return
	}

	response.Success(c, log)
}

func ExportHandler(c *gin.Context) {
	var filter AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	sheet, err := Export(filter)
	if err != nil {
		response.Error(c, response.InternalError, "导出失败: "+err.Error())
		return
	}

	f, err := excel.ExportExcel([]excel.SheetData{*sheet})
	if err != nil {
		response.Error(c, response.InternalError, "导出失败: "+err.Error())
		return
	}

	excel.WriteToResponse(c, f, "audit_logs.xlsx")
}
