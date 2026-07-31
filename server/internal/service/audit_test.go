package service

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"probig/server/internal/dao"
	"probig/server/internal/model"

	"gorm.io/gorm"
)

func migrateAuditTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&model.AuditLog{}, &model.Person{}, &model.File{}, &model.SysConfig{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	dao.RegisterAuditHooks(db)
}

func auditCtx(id uint, name string) context.Context {
	return dao.WithAuditContext(context.Background(), dao.AuditInfo{OperatorID: id, OperatorName: name, IP: "127.0.0.1"})
}

func lastAudit(db *gorm.DB, targetType string) model.AuditLog {
	var log model.AuditLog
	db.Where("target_type = ?", targetType).Order("id DESC").First(&log)
	return log
}

func TestAuditCreateAndUpdateAndDeleteAndRestore(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateAuditTables(t, db)
		ctx := auditCtx(1, "admin")

		// 新增
		p := model.Person{Name: "张三", IDCard: "110101199001011234"}
		if err := dao.DBFromContext(ctx).Create(&p).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		log := lastAudit(db, "persons")
		if log.Action != "新增" || log.OperatorName != "admin" || log.IP != "127.0.0.1" {
			t.Fatalf("create audit wrong: %+v", log)
		}
		if log.BeforeSnapshot != "" {
			t.Errorf("create before should be empty, got %s", log.BeforeSnapshot)
		}
		var after map[string]interface{}
		if err := json.Unmarshal([]byte(log.AfterSnapshot), &after); err != nil || after["name"] != "张三" {
			t.Errorf("create after snapshot wrong: %s", log.AfterSnapshot)
		}

		// 修改
		if err := dao.DBFromContext(ctx).Model(&p).Update("name", "李四").Error; err != nil {
			t.Fatalf("update: %v", err)
		}
		log = lastAudit(db, "persons")
		if log.Action != "修改" {
			t.Fatalf("update action wrong: %s", log.Action)
		}
		var before map[string]interface{}
		json.Unmarshal([]byte(log.BeforeSnapshot), &before)
		if before["name"] != "张三" {
			t.Errorf("update before should contain old name 张三, got %s", log.BeforeSnapshot)
		}
		var after2 map[string]interface{}
		json.Unmarshal([]byte(log.AfterSnapshot), &after2)
		if after2["name"] != "李四" {
			t.Errorf("update after should contain new name 李四, got %s", log.AfterSnapshot)
		}

		// 删除
		if err := dao.DBFromContext(ctx).Delete(&p).Error; err != nil {
			t.Fatalf("delete: %v", err)
		}
		log = lastAudit(db, "persons")
		if log.Action != "删除" || log.AfterSnapshot != "" {
			t.Fatalf("delete audit wrong: %+v", log)
		}
		var delBefore map[string]interface{}
		json.Unmarshal([]byte(log.BeforeSnapshot), &delBefore)
		if delBefore["name"] != "李四" {
			t.Errorf("delete before should contain full record, got %s", log.BeforeSnapshot)
		}

		// 恢复
		if err := dao.DBFromContext(ctx).Unscoped().Model(&p).Update("deleted_at", nil).Error; err != nil {
			t.Fatalf("restore: %v", err)
		}
		log = lastAudit(db, "persons")
		if log.Action != "恢复" {
			t.Fatalf("restore action should be 恢复, got %s", log.Action)
		}
	})
}

func TestAuditConcurrentOperators(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateAuditTables(t, db)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				var ctx context.Context
				var name string
				if n < 5 {
					ctx = auditCtx(uint(100+n), "user-a")
					name = "A"
				} else {
					ctx = auditCtx(uint(200+n-5), "user-b")
					name = "B"
				}
				dao.DBFromContext(ctx).Create(&model.Person{Name: name})
			}(i)
		}
		wg.Wait()

		var logs []model.AuditLog
		db.Where("target_type = ?", "persons").Order("id ASC").Find(&logs)
		if len(logs) != 10 {
			t.Fatalf("expected 10 audit logs, got %d", len(logs))
		}
		for _, l := range logs {
			if l.OperatorName == "" {
				t.Errorf("audit log %d has empty operator", l.ID)
			}
			want := "user-a"
			if l.TargetName == "B" {
				want = "user-b"
			}
			if l.OperatorName != want {
				t.Errorf("audit %d target=%s operator=%s, want %s (operator leaked across requests)", l.ID, l.TargetName, l.OperatorName, want)
			}
		}
	})
}

func TestAuditBatchCreateSplitsPerRow(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateAuditTables(t, db)
		ctx := auditCtx(1, "admin")

		if err := dao.DBFromContext(ctx).Create(&[]model.Person{
			{Name: "王五"}, {Name: "赵六"},
		}).Error; err != nil {
			t.Fatalf("batch create: %v", err)
		}
		var count int64
		db.Model(&model.AuditLog{}).Where("target_type = ? AND action = ?", "persons", "新增").Count(&count)
		if count != 2 {
			t.Fatalf("batch create should produce 2 audit rows, got %d", count)
		}
	})
}

func TestAuditConfigChangeAction(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateAuditTables(t, db)
		ctx := auditCtx(1, "admin")

		cfg := model.SysConfig{ConfigKey: "test.key", ConfigValue: "1", ConfigName: "测试", ValueType: "number"}
		if err := dao.DBFromContext(ctx).Create(&cfg).Error; err != nil {
			t.Fatalf("create cfg: %v", err)
		}
		if err := dao.DBFromContext(ctx).Model(&cfg).Update("config_value", "2").Error; err != nil {
			t.Fatalf("update cfg: %v", err)
		}
		log := lastAudit(db, "sys_config")
		if log.Action != "配置修改" {
			t.Fatalf("sys_config update action should be 配置修改, got %s", log.Action)
		}
	})
}

func TestAuditFileTable(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateAuditTables(t, db)
		ctx := auditCtx(1, "admin")

		f := model.File{Name: "a.txt", OriginalName: "a.txt", Path: "/tmp/a.txt", Size: 10}
		if err := dao.DBFromContext(ctx).Create(&f).Error; err != nil {
			t.Fatalf("create file: %v", err)
		}
		log := lastAudit(db, "files")
		if log.Action != "新增" {
			t.Fatalf("file create should be audited, got %s", log.Action)
		}
		if err := dao.DBFromContext(ctx).Delete(&f).Error; err != nil {
			t.Fatalf("delete file: %v", err)
		}
		log = lastAudit(db, "files")
		if log.Action != "删除" {
			t.Fatalf("file delete should be audited, got %s", log.Action)
		}
	})
}

func TestWriteBusinessAudit(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		migrateAuditTables(t, db)
		ctx := auditCtx(7, "ops")

		dao.WriteBusinessAudit(ctx, "结转", "annual_leave_carryover", 9, "ALC-1", "", `{"success":3}`)
		var log model.AuditLog
		db.Where("action = ?", "结转").First(&log)
		if log.OperatorName != "ops" || log.TargetID != 9 || log.TargetName != "ALC-1" {
			t.Fatalf("business audit wrong: %+v", log)
		}
		if log.AfterSnapshot != `{"success":3}` {
			t.Errorf("after snapshot wrong: %s", log.AfterSnapshot)
		}
	})
}
