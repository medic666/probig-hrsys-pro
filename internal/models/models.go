package models

var AllModels = []interface{}{
	&Person{},
	&PersonPhone{},
	&PersonEmail{},
	&PersonBankCard{},
	&Company{},
	&File{},
	&FileRelation{},
	&PositionEvent{},
	&AttendanceEvent{},
	&SalaryEvent{},
	&PositionSnapshot{},
	&AttendanceSummary{},
	&SalarySummary{},
	&AuditLog{},
	&SysConfig{},
	&User{},
	&Role{},
	&Permission{},
	&UserRole{},
	&RolePermission{},
}
