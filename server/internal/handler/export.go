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

// exportField 导出字段定义：key 取值、label 表头、fmt 取值后格式化钩子（nil 直出）。
// 字段表即导出列的唯一事实源，与列表/详情/追溯展示口径保持一致。
type exportField struct {
	key, label string
	fmt        func(v interface{}) interface{}
}

// buildExportRows 按字段表生成导出行与表头（字段顺序即导出列顺序）
func buildExportRows(list []map[string]interface{}, fields []exportField) ([][]interface{}, []string) {
	rows := make([][]interface{}, 0, len(list))
	headers := make([]string, len(fields))
	for i, f := range fields {
		headers[i] = f.label
	}
	for _, s := range list {
		row := make([]interface{}, 0, len(fields))
		for _, f := range fields {
			v := s[f.key]
			if f.fmt != nil {
				v = f.fmt(v)
			}
			row = append(row, v)
		}
		rows = append(rows, row)
	}
	return rows, headers
}

// exportTimeFmt time.Time → "2006-01-02 15:04:05"（非 time.Time 原样返回）
func exportTimeFmt(v interface{}) interface{} {
	t, ok := v.(time.Time)
	if !ok {
		return v
	}
	return t.Format("2006-01-02 15:04:05")
}

// exportTimeStrFmt RFC3339 字符串 → "2006-01-02 15:04:05"（解析失败原样返回）
func exportTimeStrFmt(v interface{}) interface{} {
	s, ok := v.(string)
	if !ok || s == "" {
		return s
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return s
}

// exportStatusFmt 核算/汇总状态映射（calculated/data_changed → 已核算/数据已变动）
func exportStatusFmt(v interface{}) interface{} {
	return map[string]string{"calculated": "已核算", "data_changed": "数据已变动"}[fmt.Sprint(v)]
}

// exportBoolFmt exportBool 的字段表钩子包装
func exportBoolFmt(v interface{}) interface{} {
	return exportBool(v.(bool))
}
