package excel

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type SheetData struct {
	SheetName string
	Headers   []string
	Rows      [][]interface{}
}

func ExportExcel(sheets []SheetData) (*excelize.File, error) {
	f := excelize.NewFile()
	defaultSheet := "Sheet1"
	f.SetSheetName(defaultSheet, sheets[0].SheetName)

	for i, sheet := range sheets {
		sheetName := sheet.SheetName
		if i > 0 {
			_, err := f.NewSheet(sheetName)
			if err != nil {
				return nil, err
			}
		}

		for colIdx, header := range sheet.Headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
		})
		endCol, _ := excelize.CoordinatesToCellName(len(sheet.Headers), 1)
		f.SetCellStyle(sheetName, "A1", endCol, headerStyle)

		for rowIdx, row := range sheet.Rows {
			for colIdx, val := range row {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
				f.SetCellValue(sheetName, cell, val)
			}
		}
	}

	return f, nil
}

func WriteToResponse(c *gin.Context, f *excelize.File, filename string) {
	if filename == "" {
		filename = fmt.Sprintf("export_%s.xlsx", time.Now().Format("20060102_150405"))
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")
	if err := f.Write(c.Writer); err != nil {
		c.Error(err)
	}
}
