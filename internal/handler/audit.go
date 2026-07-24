package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"probig/internal/service"
	"probig/pkg/response"
)

func GetAuditLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	operatorID, _ := strconv.ParseUint(c.Query("operator_id"), 10, 64)
	targetType := c.Query("target_type")
	action := c.Query("action")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	batchID := c.Query("batch_id")
	list, total, err := service.GetAuditLogList(page, pageSize, uint(operatorID), targetType, action, startDate, endDate, batchID)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageSuccess(c, response.PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}
