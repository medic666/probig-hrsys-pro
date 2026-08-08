package dao

import (
	"fmt"
	"path/filepath"
	"testing"

	"probig/server/internal/model"

	"gorm.io/gorm"
)

func newMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// 临时文件库：与生产同路径打开（_txlock=immediate）
	db, err := gorm.Open(GetSQLiteDialector(fmt.Sprintf("file:%s?_busy_timeout=10000", filepath.Join(t.TempDir(), "test.db"))), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.Permission{}, &model.RolePermission{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedLegacyPerms 预置旧动作体系数据（delete 存在、无 calculate、user 带无端点 export）
func seedLegacyPerms(t *testing.T, db *gorm.DB) {
	t.Helper()
	role := model.Role{Name: "超级管理员", IsDefault: true}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	legacy := []model.Permission{
		{Module: "person", Action: "read", Name: "人员管理查看"},
		{Module: "person", Action: "write", Name: "人员管理编辑"},
		{Module: "person", Action: "delete", Name: "人员管理删除"},
		{Module: "person", Action: "export", Name: "人员管理导出"},
		{Module: "user", Action: "export", Name: "用户管理导出"},
	}
	for i := range legacy {
		if err := db.Create(&legacy[i]).Error; err != nil {
			t.Fatalf("seed perm: %v", err)
		}
		if err := db.Create(&model.RolePermission{RoleID: role.ID, PermissionID: legacy[i].ID}).Error; err != nil {
			t.Fatalf("seed rp: %v", err)
		}
	}
}

func TestMigratePermissionActions(t *testing.T) {
	db := newMigrationTestDB(t)
	seedLegacyPerms(t, db)

	if err := migratePermissionActions(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// ① delete 权限已清（含 role_permissions 关联）
	var deleteCount int64
	db.Model(&model.Permission{}).Where("action = ?", "delete").Count(&deleteCount)
	if deleteCount != 0 {
		t.Errorf("delete perms remain: %d", deleteCount)
	}
	var orphanRP int64
	db.Model(&model.RolePermission{}).
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("permissions.action = ?", "delete").Count(&orphanRP)
	if orphanRP != 0 {
		t.Errorf("delete role_permissions remain: %d", orphanRP)
	}

	// ② 无端点的 user/role 导出权限已清
	var ueCount int64
	db.Model(&model.Permission{}).Where("action = ? AND module IN ?", "export", []string{"user", "role"}).Count(&ueCount)
	if ueCount != 0 {
		t.Errorf("user/role export remain: %d", ueCount)
	}

	// ③ calculate 已补全（attendance_monthly/annual_leave_carryover/salary_summary）
	for _, mod := range []string{"attendance_monthly", "annual_leave_carryover", "salary_summary"} {
		var c int64
		db.Model(&model.Permission{}).Where("module = ? AND action = ?", mod, "calculate").Count(&c)
		if c != 1 {
			t.Errorf("module %s calculate missing", mod)
		}
	}

	// ④ 权限总数与 ModuleActions 定义一致
	want := 0
	for _, m := range model.ModuleActions {
		want += len(m.Actions)
	}
	var total int64
	db.Model(&model.Permission{}).Count(&total)
	if int(total) != want {
		t.Errorf("perm total = %d, want %d", total, want)
	}

	// ⑤ 幂等：重复执行不报错、权限行数不变
	if err := migratePermissionActions(db); err != nil {
		t.Fatalf("re-migrate: %v", err)
	}
	var total2 int64
	db.Model(&model.Permission{}).Count(&total2)
	if total2 != total {
		t.Errorf("re-migrate changed count: %d -> %d", total, total2)
	}

	// ⑥ (module, action) 唯一索引存在
	var idxCount int64
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_permissions_module_action'").Scan(&idxCount)
	if idxCount != 1 {
		t.Errorf("unique index missing")
	}
}

// TestMigratePermissionModules 模块叶子化：旧模块权限语义映射到新叶子模块，home 清理，幂等
func TestMigratePermissionModules(t *testing.T) {
	db := newMigrationTestDB(t)
	role := model.Role{Name: "考勤专员", IsDefault: false}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	legacy := []model.Permission{
		{Module: "attendance", Action: "read", Name: "考勤管理查看"},
		{Module: "attendance", Action: "write", Name: "考勤管理编辑"},
		{Module: "attendance", Action: "export", Name: "考勤管理导出"},
		{Module: "home", Action: "read", Name: "首页查看"},
		{Module: "person", Action: "read", Name: "人员管理查看"},
	}
	for i := range legacy {
		if err := db.Create(&legacy[i]).Error; err != nil {
			t.Fatalf("seed perm: %v", err)
		}
		if err := db.Create(&model.RolePermission{RoleID: role.ID, PermissionID: legacy[i].ID}).Error; err != nil {
			t.Fatalf("seed rp: %v", err)
		}
	}

	if err := migratePermissionModules(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// ① 旧模块权限已清理（含 home）
	var oldCount int64
	db.Model(&model.Permission{}).Where("module IN ?", []string{"attendance", "annual_leave", "salary", "home"}).Count(&oldCount)
	if oldCount != 0 {
		t.Errorf("old module perms remain: %d", oldCount)
	}

	// ② 语义保留：考勤专员获得 attendance.read → 三个新模块 read
	expect := map[string]bool{
		"attendance_event.read":     true,
		"attendance_daily.read":     true,
		"attendance_monthly.read":   true,
		"attendance_event.write":    true,
		"attendance_event.export":   true,
		"attendance_daily.export":   true,
		"attendance_monthly.export": true,
	}
	for key := range expect {
		var cnt int64
		db.Table("role_permissions rp").
			Joins("JOIN permissions p ON p.id = rp.permission_id").
			Where("rp.role_id = ? AND p.module || '.' || p.action = ?", role.ID, key).
			Count(&cnt)
		if cnt != 1 {
			t.Errorf("role missing mapped perm %s", key)
		}
	}
	// person.read 原样保留
	var personCnt int64
	db.Table("role_permissions rp").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("rp.role_id = ? AND p.module = ? AND p.action = ?", role.ID, "person", "read").
		Count(&personCnt)
	if personCnt != 1 {
		t.Errorf("person.read not preserved")
	}

	// ③ 幂等：重复执行不报错、权限行数不变
	var total int64
	db.Model(&model.Permission{}).Count(&total)
	if err := migratePermissionModules(db); err != nil {
		t.Fatalf("re-migrate: %v", err)
	}
	var total2 int64
	db.Model(&model.Permission{}).Count(&total2)
	if total2 != total {
		t.Errorf("re-migrate changed count: %d -> %d", total, total2)
	}
}
