package common

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func ExportToExcel(headers []string, rows [][]string) (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "Sheet1"

	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, 1), h)
	}

	for i, row := range rows {
		rowNum := i + 2
		for j, val := range row {
			col, _ := excelize.ColumnNumberToName(j + 1)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, rowNum), val)
		}
	}

	return f, nil
}
