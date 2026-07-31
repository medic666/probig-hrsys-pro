package dao

import (
	"context"
	"encoding/json"
	"reflect"

	"probig/server/internal/model"

	"gorm.io/gorm"
)

// AuditInfo 请求级审计上下文，由中间件注入，随 context 贯穿整个请求
type AuditInfo struct {
	OperatorID   uint
	OperatorName string
	IP           string
}

type auditCtxKey struct{}

// WithAuditContext 将审计信息注入 context
func WithAuditContext(ctx context.Context, info AuditInfo) context.Context {
	return context.WithValue(ctx, auditCtxKey{}, info)
}

// AuditFromContext 从 context 读取审计信息，无则返回零值
func AuditFromContext(ctx context.Context) AuditInfo {
	if ctx == nil {
		return AuditInfo{}
	}
	info, ok := ctx.Value(auditCtxKey{}).(AuditInfo)
	if !ok {
		return AuditInfo{}
	}
	return info
}

// DBFromContext 返回携带审计 context 的数据库实例，写操作统一经此发起
func DBFromContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return DB
	}
	return DB.WithContext(ctx)
}

// WriteBusinessAudit 显式业务动作审计（结转/反结账/核算/确认等），
// 独立于业务事务写入；审计失败不影响业务
func WriteBusinessAudit(ctx context.Context, action, targetType string, targetID uint, targetName, before, after string) {
	info := AuditFromContext(ctx)
	if err := DBFromContext(ctx).Create(&model.AuditLog{
		OperatorID:     info.OperatorID,
		OperatorName:   info.OperatorName,
		TargetType:     targetType,
		TargetID:       targetID,
		TargetName:     targetName,
		Action:         action,
		BeforeSnapshot: before,
		AfterSnapshot:  after,
		IP:             info.IP,
	}).Error; err != nil {
		return
	}
}

func shouldSkipAudit(table string) bool {
	switch table {
	case "audit_logs", "sys_batches",
		"position_snapshots", "attendance_daily_projections",
		"attendance_calculation_monthly", "annual_leave_balance_snapshots",
		"leave_in_lieu_balance_snapshots", "salary_summaries", "salary_summary_versions":
		return true
	}
	return false
}

func RegisterAuditHooks(db *gorm.DB) {
	db.Callback().Create().After("gorm:after_create").Register("audit:create", auditCreate)
	db.Callback().Update().Before("gorm:update").Register("audit:update-before", auditUpdateBefore)
	db.Callback().Update().After("gorm:after_update").Register("audit:update", auditUpdate)
	db.Callback().Delete().Before("gorm:delete").Register("audit:delete-before", auditDeleteBefore)
	db.Callback().Delete().After("gorm:after_delete").Register("audit:delete", auditDelete)
}

func auditCreate(db *gorm.DB) {
	if db.Statement == nil || shouldSkipAudit(db.Statement.Table) {
		return
	}
	rv := db.Statement.ReflectValue
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		for i := 0; i < rv.Len(); i++ {
			writeAudit(db, "新增", "", marshalValue(rv.Index(i)), rv.Index(i))
		}
		return
	}
	writeAudit(db, "新增", "", marshalValue(rv), rv)
}

func auditUpdateBefore(db *gorm.DB) {
	if db.Statement == nil || shouldSkipAudit(db.Statement.Table) {
		return
	}
	db.Statement.Settings.Store("audit:before", querySnapshot(db))
	db.Statement.Settings.Store("audit:restore", isRestoreOp(db))
}

func auditUpdate(db *gorm.DB) {
	if db.Statement == nil || shouldSkipAudit(db.Statement.Table) {
		return
	}
	before, _ := db.Statement.Settings.Load("audit:before")
	beforeStr, _ := before.(string)
	restore, _ := db.Statement.Settings.Load("audit:restore")
	isRestore, _ := restore.(bool)

	action := "修改"
	if isRestore {
		action = "恢复"
	} else if db.Statement.Table == "sys_config" {
		action = "配置修改"
	}
	writeAudit(db, action, beforeStr, querySnapshot(db), db.Statement.ReflectValue)
}

func auditDeleteBefore(db *gorm.DB) {
	if db.Statement == nil || shouldSkipAudit(db.Statement.Table) {
		return
	}
	db.Statement.Settings.Store("audit:before", querySnapshot(db))
}

func auditDelete(db *gorm.DB) {
	if db.Statement == nil || shouldSkipAudit(db.Statement.Table) {
		return
	}
	before, _ := db.Statement.Settings.Load("audit:before")
	beforeStr, _ := before.(string)
	writeAudit(db, "删除", beforeStr, "", db.Statement.ReflectValue)
}

// isRestoreOp 识别恢复操作：Dest 仅含 deleted_at 且值为 nil
func isRestoreOp(db *gorm.DB) bool {
	m, ok := db.Statement.Dest.(map[string]interface{})
	if !ok || len(m) != 1 {
		return false
	}
	v, exists := m["deleted_at"]
	if !exists {
		return false
	}
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

// querySnapshot 按主键查询当前完整数据作为快照
func querySnapshot(db *gorm.DB) string {
	id := extractID(db)
	if id == 0 {
		return ""
	}
	var row map[string]interface{}
	err := db.Session(&gorm.Session{NewDB: true}).Unscoped().
		Table(db.Statement.Table).Where("id = ?", id).Take(&row).Error
	if err != nil {
		return ""
	}
	b, _ := json.Marshal(row)
	return string(b)
}

// extractID 提取 Statement.Model / Dest 中的主键
func extractID(db *gorm.DB) uint {
	for _, src := range []interface{}{db.Statement.Model, db.Statement.Dest} {
		if src == nil {
			continue
		}
		rv := reflect.ValueOf(src)
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				break
			}
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			continue
		}
		if f := rv.FieldByName("ID"); f.IsValid() {
			return uint(f.Uint())
		}
	}
	return 0
}

func marshalValue(rv reflect.Value) string {
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	b, err := json.Marshal(rv.Interface())
	if err != nil {
		return ""
	}
	return string(b)
}

// writeAudit 从 Statement.Context 读取操作人，按实体维度写入审计（批量拆单条）
// Session(NewDB) 继承当前连接的 ConnPool：事务内随业务事务提交/回滚，事务外独立写入，
// 避免审计写入与业务事务的 SQLite 写锁竞争
func writeAudit(db *gorm.DB, action, before, after string, rv reflect.Value) {
	info := AuditFromContext(db.Statement.Context)
	auditDB := db.Session(&gorm.Session{NewDB: true})
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		for i := 0; i < rv.Len(); i++ {
			writeAuditRow(auditDB, info, db.Statement.Table, action, before, after, rv.Index(i))
		}
		return
	}
	writeAuditRow(auditDB, info, db.Statement.Table, action, before, after, rv)
}

func writeAuditRow(auditDB *gorm.DB, info AuditInfo, table, action, before, after string, rv reflect.Value) {
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	var targetID uint
	if f := rv.FieldByName("ID"); f.IsValid() {
		targetID = uint(f.Uint())
	}
	var targetName string
	if f := rv.FieldByName("Name"); f.IsValid() {
		targetName = f.String()
	}
	if err := auditDB.Create(&model.AuditLog{
		OperatorID:     info.OperatorID,
		OperatorName:   info.OperatorName,
		TargetType:     table,
		TargetID:       targetID,
		TargetName:     targetName,
		Action:         action,
		BeforeSnapshot: before,
		AfterSnapshot:  after,
		IP:             info.IP,
	}).Error; err != nil {
		return
	}
}
