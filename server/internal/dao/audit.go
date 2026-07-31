package dao

import (
	"encoding/json"
	"sync"

	"probig/server/internal/model"

	"gorm.io/gorm"
)

var auditCtx struct {
	mu           sync.Mutex
	operatorID   uint
	operatorName string
}

func SetAuditOperator(id uint, name string) {
	auditCtx.mu.Lock()
	defer auditCtx.mu.Unlock()
	auditCtx.operatorID = id
	auditCtx.operatorName = name
}

func RegisterAuditHooks(db *gorm.DB) {
	db.Callback().Create().After("gorm:after_create").Register("audit:create", func(db *gorm.DB) {
		writeAudit(db, "新增", "", "*")
	})
	db.Callback().Update().After("gorm:after_update").Register("audit:update", func(db *gorm.DB) {
		writeAudit(db, "修改", "*", "*")
	})
	db.Callback().Delete().After("gorm:after_delete").Register("audit:delete", func(db *gorm.DB) {
		writeAudit(db, "删除", "*", "")
	})
}

func writeAudit(db *gorm.DB, action, beforeMask, afterMask string) {
	if db.Statement == nil || db.Statement.Table == "" {
		return
	}
	table := db.Statement.Table
	if table == "audit_logs" || table == "sys_batches" {
		return
	}

	excludedProjections := map[string]bool{
		"position_snapshots":                true,
		"attendance_daily_projections":      true,
		"attendance_calculation_monthly":    true,
		"annual_leave_balance_snapshots":    true,
		"leave_in_lieu_balance_snapshots":   true,
		"salary_summaries":                  true,
		"salary_summary_versions":          true,
		"files":                             true,
		"file_relations":                    true,
	}
	if excludedProjections[table] {
		return
	}

	var beforeJSON, afterJSON string
	if beforeMask != "" {
		b, _ := json.Marshal(db.Statement.Dest)
		beforeJSON = string(b)
	}
	if afterMask != "" {
		if db.Statement.Dest != nil {
			a, _ := json.Marshal(db.Statement.Dest)
			afterJSON = string(a)
		}
	}

	var targetID uint
	if db.Statement.ReflectValue.Kind().String() == "struct" {
		idField := db.Statement.ReflectValue.FieldByName("ID")
		if idField.IsValid() {
			targetID = uint(idField.Uint())
		}
	}

	var targetName string
	nameField := db.Statement.ReflectValue.FieldByName("Name")
	if nameField.IsValid() {
		targetName = nameField.String()
	}

	db.Session(&gorm.Session{NewDB: true}).Create(&model.AuditLog{
		OperatorID:     auditCtx.operatorID,
		OperatorName:   auditCtx.operatorName,
		TargetType:     table,
		TargetID:       targetID,
		TargetName:     targetName,
		Action:         action,
		BeforeSnapshot: beforeJSON,
		AfterSnapshot:  afterJSON,
	})
}
