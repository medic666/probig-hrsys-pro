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

// baselineMigration 全新库一次成型基线：合并全部历史迁移（表结构 + 索引）最终形态，
// 仅对无任何迁移记录的全新数据库执行；存量库不重复执行。
var baselineMigration = Migration{
	ID:   "20260805_01_fresh_baseline",
	Name: "全新库完整结构基线（合并全部历史迁移）",
	Func: migrateV2FreshBaseline,
}

// legacyMigrations 旧增量迁移：仅供存量库升级执行（历史库按序补迁），
// 其内容已并入基线，全新库不再执行（ID 直接标记为已应用）。
var legacyMigrations = []Migration{
	{ID: "20260731_01_init", Name: "初始表结构与索引", Func: migrateV1Init},
	{ID: "20260801_01_person_emergency_contacts_position_fields", Name: "人员紧急联系人表 + 职务公司/部门/职位字段", Func: migrateV1EmergencyContactsAndPositionFields},
	{ID: "20260801_02_unify_punch_time", Name: "打卡时间统一到 daily.punch_time，清除打卡时间戳事件", Func: migrateV1UnifyPunchTime},
	{ID: "20260802_01_config_key_classification", Name: "计薪小时基准配置归入考勤类", Func: migrateV1ConfigKeyClassification},
	{ID: "20260804_01_annual_leave_tier_config", Name: "年假额度配置升级为阶梯交互", Func: migrateV1AnnualLeaveTierConfig},
	{ID: "20260804_03_salary_advance_types", Name: "借款还款拆分工资预支/预支还款", Func: migrateV1SalaryAdvanceTypes},
	{ID: "20260804_04_annual_leave_tier_lower_bound", Name: "年假阶梯配置改下界语义（满X司龄年配发）并迁移默认值", Func: migrateV1AnnualLeaveTierLowerBound},
}

// RunMigrations 数据库结构迁移：
// - 全新库（无任何迁移记录）：只执行基线一次成型，旧增量 ID 标记为已应用（内容已并入基线）；
// - 存量库（有迁移记录）：只执行未应用的旧增量，基线不重复执行（避免对存量数据重复数据迁移）。
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

	record := func(id string) error {
		if err := db.Create(&schemaMigration{MigrationID: id, AppliedAt: time.Now()}).Error; err != nil {
			return fmt.Errorf("记录迁移 %s 失败: %w", id, err)
		}
		return nil
	}

	switch {
	case applied[baselineMigration.ID]:
		// 基线库（全新库或已基线化）：全部旧增量已并入基线，无需执行
		return nil
	case len(applied) == 0:
		// 全新库：只执行基线一次成型（schema_migrations 仅 1 条记录）
		if err := baselineMigration.Func(db); err != nil {
			return fmt.Errorf("迁移 %s(%s) 失败: %w", baselineMigration.ID, baselineMigration.Name, err)
		}
		return record(baselineMigration.ID)
	default:
		// 存量旧库：按序执行未应用的旧增量，基线不重复执行（避免对存量数据重复数据迁移）
		for _, m := range legacyMigrations {
			if applied[m.ID] {
				continue
			}
			if err := m.Func(db); err != nil {
				return fmt.Errorf("迁移 %s(%s) 失败: %w", m.ID, m.Name, err)
			}
			if err := record(m.ID); err != nil {
				return err
			}
		}
		return nil
	}
}

// migrateV2FreshBaseline 全新库完整结构基线：全部表结构（GORM 幂等）+ 全部索引，
// 内容为各历史增量迁移的最终形态合并（含后续补建的紧急联系人表与职务字段索引）。
func migrateV2FreshBaseline(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{}, &model.Role{}, &model.Permission{}, &model.UserRole{}, &model.RolePermission{},
		&model.AuditLog{}, &model.SysConfig{},
		&model.Person{}, &model.PersonPhone{}, &model.PersonEmail{}, &model.PersonBankCard{},
		&model.PersonEmergencyContact{},
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
		// 历史增量补建索引
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

// migrateV1SalaryAdvanceTypes 借款还款拆分：按金额正负拆分历史事件
// （正 = 工资预支，负 = 预支还款，符合拆分后的业务语义）
func migrateV1SalaryAdvanceTypes(db *gorm.DB) error {
	return db.Exec(`UPDATE salary_events
		SET event_type = CASE WHEN amount >= 0 THEN '工资预支' ELSE '预支还款' END
		WHERE event_type = '借款还款'`).Error
}

// migrateV1AnnualLeaveTierConfig 年假额度配置升级为阶梯交互：
// 仅更新 value_type（number → table），保留用户已配置的值（解析层兼容旧单值）
func migrateV1AnnualLeaveTierConfig(db *gorm.DB) error {
	return db.Exec(`UPDATE sys_config SET value_type = 'table' WHERE config_key = 'annual_leave.yearly_hours'`).Error
}

// migrateV1AnnualLeaveTierLowerBound 年假阶梯配置改下界语义（满X司龄年配发）：
// years 由"档位上限"改为"配发门槛"，仅当配置值精确等于旧默认种子时重写为新默认；
// 用户自定义配置（语义已随代码切换为下界）一律不动。
func migrateV1AnnualLeaveTierLowerBound(db *gorm.DB) error {
	const oldSeed = `[{"years":10,"hours":40},{"years":20,"hours":80},{"years":999,"hours":120}]`
	const newSeed = `[{"years":1,"hours":40},{"years":10,"hours":80},{"years":20,"hours":120}]`
	return db.Exec(`UPDATE sys_config SET config_value = ?
		WHERE config_key = 'annual_leave.yearly_hours' AND config_value = ?`, newSeed, oldSeed).Error
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
