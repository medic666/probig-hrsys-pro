package handler

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// writeExcel 统一 Excel 导出：生成单 sheet 工作簿并写入响应流。
// 文件名 = {业务名}[_{筛选摘要}]_{YYYYMMDD_HHmm}.xlsx（RFC 5987 filename* 编码，前端可正确解析）
func writeExcel(c *gin.Context, sheetName string, headers []string, rows [][]interface{}, pieces ...string) {
	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", sheetName)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}
	for i, row := range rows {
		for j, v := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheetName, cell, v)
		}
	}

	// 筛选摘要：剔除空片段，超长截断
	var summary []string
	for _, p := range pieces {
		if p != "" {
			summary = append(summary, p)
		}
	}
	name := sheetName
	if len(summary) > 0 {
		joined := strings.Join(summary, "_")
		if len(joined) > 60 {
			joined = joined[:60]
		}
		name = name + "_" + joined
	}
	name = fmt.Sprintf("%s_%s.xlsx", name, time.Now().Format("20060102_1504"))
	encoded := url.PathEscape(name)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=download.xlsx; filename*=UTF-8''%s", encoded))
	f.Write(c.Writer)
}

// dateRangePiece 日期区间筛选摘要片段：起至止 / 自X起 / 至X止
func dateRangePiece(prefix, start, end string) string {
	switch {
	case start != "" && end != "":
		return prefix + "=" + start + "至" + end
	case start != "":
		return prefix + "=自" + start + "起"
	case end != "":
		return prefix + "=至" + end + "止"
	}
	return ""
}

func exportBool(v bool) string {
	if v {
		return "是"
	}
	return "否"
}
