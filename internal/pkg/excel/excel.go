package excel

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
)

type Exporter struct {
	Headers []string
	Rows    [][]string
}

func NewExporter(headers []string) *Exporter {
	return &Exporter{Headers: headers}
}

func (e *Exporter) AddRow(row []string) {
	e.Rows = append(e.Rows, row)
}

func (e *Exporter) WriteCSV(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(e.Headers); err != nil {
		return err
	}
	for _, row := range e.Rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func StructToHeaders(s interface{}) ([]string, []string) {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var tags []string
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("excel")
		if tag != "" && tag != "-" {
			tags = append(tags, t.Field(i).Name)
			headers = append(headers, tag)
		}
	}
	return tags, headers
}

func FormatDecimal(v float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, v)
}
