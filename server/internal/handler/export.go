package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// writeExcel 统一 Excel 导出：生成单 sheet 工作簿并写入响应流
func writeExcel(c *gin.Context, sheetName, filename string, headers []string, rows [][]interface{}) {
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
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename+"_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
}

func exportBool(v bool) string {
	if v {
		return "是"
	}
	return "否"
}
