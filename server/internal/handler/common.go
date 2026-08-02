package handler

import (
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// successPage 列表分页统一响应（list 接口与导出/查询共用同一构造）
func successPage(c *gin.Context, list interface{}, total int64, pageNum, pageSize int) {
	utils.Success(c, utils.NewPageResult(list, total, utils.PageRequest{PageNum: pageNum, PageSize: pageSize}))
}
