package auditlog

import (
	"probig/internal/pkg/excel"
)

func List(filter AuditLogFilter) ([]AuditLogVO, int64, error) {
	logs, total, err := ListLogs(filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]AuditLogVO, 0, len(logs))
	for _, l := range logs {
		result = append(result, AuditLogVO{
			ID:           l.ID,
			OperatorID:   l.OperatorID,
			OperatorName: l.OperatorName,
			TargetType:   l.TargetType,
			TargetID:     l.TargetID,
			TargetName:   l.TargetName,
			Action:       l.Action,
			IP:           l.IP,
			CreatedAt:    l.CreatedAt,
		})
	}

	return result, total, nil
}

func GetDetail(id uint) (*AuditLog, error) {
	return GetLogByID(id)
}

func Export(filter AuditLogFilter) (*excel.SheetData, error) {
	logs, err := GetAllLogs(filter)
	if err != nil {
		return nil, err
	}

	headers := []string{"操作时间", "操作人", "操作类型", "操作对象类型", "操作对象名称", "IP"}

	var rows [][]interface{}
	for _, l := range logs {
		rows = append(rows, []interface{}{
			l.CreatedAt.Format("2006-01-02 15:04:05"),
			l.OperatorName,
			l.Action,
			l.TargetType,
			l.TargetName,
			l.IP,
		})
	}

	return &excel.SheetData{
		SheetName: "审计日志",
		Headers:   headers,
		Rows:      rows,
	}, nil
}
