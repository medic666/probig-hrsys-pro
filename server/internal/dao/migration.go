package dao

import (
	"fmt"
	"time"

	"probig/server/internal/model"

	"gorm.io/gorm"
)

// schemaMigration 迁移版本记录表
type schemaMigration struct {
	ID          uint      `gorm:"primarykey"`
	MigrationID string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	AppliedAt   time.Time
}

func (schemaMigration) TableName() string { return "schema_migrations" }

// Migration 一次版本化迁移：仅追加，禁止修改已发布条目
type Migration struct {
	ID   string
	Name string
	Func func(db *gorm.DB) error
}

var migrations = []Migration{
	{ID: "20260731_01_init", Name: "初始表结构与索引", Func: migrateV1Init},
	{ID: "20260801_01_person_emergency_contacts_position_fields", Name: "人员紧急联系人表 + 职务公司/部门/职位字段", Func: migrateV1EmergencyContactsAndPositionFields},
	{ID: "20260801_02_unify_punch_time", Name: "打卡时间统一到 daily.punch_time，清除打卡时间戳事件", Func: migrateV1UnifyPunchTime},
	{ID: "20260802_01_config_key_classification", Name: "计薪小时基准配置归入考勤类", Func: migrateV1ConfigKeyClassification},
}

// RunMigrations 顺序执行未应用的迁移；新库或未迁移库（无版本记录）从头执行全部
func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return fmt.Errorf("初始化迁移版本表失败: %w", err)
	}
	applied := make(map[string]bool)
	var rows []schemaMigration
	if err := db.Find(&rows).Error; err != nil {
		return err
	}
	for _, r := range rows {
		applied[r.MigrationID] = true
	}
	for _, m := range migrations {
		if applied[m.ID] {
			continue
		}
		if err := m.Func(db); err != nil {
			return fmt.Errorf("迁移 %s(%s) 失败: %w", m.ID, m.Name, err)
		}
		if err := db.Create(&schemaMigration{MigrationID: m.ID, AppliedAt: time.Now()}).Error; err != nil {
			return fmt.Errorf("记录迁移 %s 失败: %w", m.ID, err)
		}
	}
	return nil
}

// migrateV1ConfigKeyClassification 计薪小时基准配置键归入考勤类（system.work_hours_per_day → attendance.work_hours_per_day）
func migrateV1ConfigKeyClassification(db *gorm.DB) error {
	return db.Exec(`UPDATE sys_config SET config_key = 'attendance.work_hours_per_day'
		WHERE config_key = 'system.work_hours_per_day'`).Error
}

// migrateV1UnifyPunchTime 打卡时间统一到 daily.punch_time（唯一载体），清除"打卡时间戳"事件：
// ① punch_time 为空的 daily 从打卡时间戳事件 remark 回填；② 物理删除打卡时间戳事件
func migrateV1UnifyPunchTime(db *gorm.DB) error {
	sqls := []string{
		`UPDATE attendance_daily
		 SET punch_time = (SELECT e.remark FROM attendance_event_details e
		                   WHERE e.daily_id = attendance_daily.id
		                     AND e.event_type = '打卡时间戳' AND e.deleted_at IS NULL
		                   LIMIT 1)
		 WHERE punch_time = '' AND EXISTS (SELECT 1 FROM attendance_event_details e
		                                   WHERE e.daily_id = attendance_daily.id
		                                     AND e.event_type = '打卡时间戳' AND e.deleted_at IS NULL)`,
		`DELETE FROM attendance_event_details WHERE event_type = '打卡时间戳'`,
	}
	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

// migrateV1EmergencyContactsAndPositionFields 人员紧急联系人关联表 + 职务事件/快照补充公司、部门、职位
func migrateV1EmergencyContactsAndPositionFields(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.PersonEmergencyContact{},
		&model.PositionEvent{}, &model.PositionSnapshot{},
	); err != nil {
		return err
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_person_emergency_contacts_person ON person_emergency_contacts(person_id)",
		"CREATE INDEX IF NOT EXISTS idx_position_events_company ON position_events(company_id)",
	}
	for _, sql := range indexes {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

// migrateV1Init 基线：全部表结构（GORM 幂等）+ 唯一/查询索引
func migrateV1Init(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{}, &model.Role{}, &model.Permission{}, &model.UserRole{}, &model.RolePermission{},
		&model.AuditLog{}, &model.SysConfig{},
		&model.Person{}, &model.PersonPhone{}, &model.PersonEmail{}, &model.PersonBankCard{},
		&model.Company{}, &model.File{}, &model.FileRelation{},
		&model.PositionEvent{}, &model.PositionSnapshot{},
		&model.AttendanceDaily{}, &model.AttendanceEventDetail{}, &model.AttendanceDailyProjection{},
		&model.AttendanceCalculationMonthly{},
		&model.AnnualLeaveAccountEvent{}, &model.AnnualLeaveBalanceSnapshot{}, &model.LeaveInLieuBalanceSnapshot{},
		&model.SysBatch{}, &model.SalaryEvent{}, &model.SalarySummary{}, &model.SalarySummaryVersion{},
	); err != nil {
		return err
	}

	indexes := []string{
		// 部分唯一索引（WHERE 条件，模型 tag 无法表达）
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_persons_id_card ON persons(id_card) WHERE deleted_at IS NULL AND id_card != ''",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_credit_code ON companies(credit_code) WHERE deleted_at IS NULL AND credit_code != ''",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_position_events_person_seq ON position_events(person_id, seq) WHERE deleted_at IS NULL",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_daily_person_date ON attendance_daily(person_id, event_date) WHERE deleted_at IS NULL",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_annual_leave_events_person_seq ON annual_leave_account_events(person_id, seq) WHERE deleted_at IS NULL",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_salary_events_person_seq ON salary_events(person_id, seq) WHERE deleted_at IS NULL",
		// 查询与并发兜底索引
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_projection_person_date ON attendance_daily_projections(person_id, work_date)",
		"CREATE INDEX IF NOT EXISTS idx_position_snapshots_person_start ON position_snapshots(person_id, effective_start_date)",
		"CREATE INDEX IF NOT EXISTS idx_al_balance_person_start ON annual_leave_balance_snapshots(person_id, effective_start_date)",
		"CREATE INDEX IF NOT EXISTS idx_lil_balance_person_start ON leave_in_lieu_balance_snapshots(person_id, effective_start_date)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_salary_version_person_month_ver ON salary_summary_versions(person_id, belong_month, version)",
		"CREATE INDEX IF NOT EXISTS idx_salary_events_person_month ON salary_events(person_id, belong_month)",
	}
	for _, sql := range indexes {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
