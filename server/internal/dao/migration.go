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

// migrations 全部版本化迁移（按序执行，仅追加、禁止修改已发布条目）。
// 20260805_01_fresh_baseline 为上线聚合基线：上线前测试阶段的全部增量迁移已合并于此
// （生产库启动时已应用）；其后为上线后增量。新库从第 1 条逐条执行到最新，
// 存量库跳过已应用条目、从最后已应用的下一条继续——两条路径执行完全相同的迁移代码。
// 后续新增结构变更：在列表末尾追加一条幂等迁移即可（无需修改基线）。
var migrations = []Migration{
	{ID: "20260805_01_fresh_baseline", Name: "上线聚合基线（合并测试阶段全部迁移）", Func: migrateV2FreshBaseline},
	{ID: "20260805_02_attendance_daily_seq", Name: "考勤日记录支持同日多版本(seq)", Func: migrateAttendanceDailySeq},
}

// RunMigrations 数据库结构迁移：加载已应用集合，按序跳过已应用、执行未应用的迁移并记录。
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

// migrateV2FreshBaseline 上线聚合基线：全部表结构（GORM 幂等）+ 全部索引，
// 内容为上线前各增量迁移的最终形态合并（含紧急联系人表与职务字段索引）。
// 注意：attendance_daily 唯一索引保持上线时的 (person_id, event_date) 形态，
// seq 多版本改造由 20260805_02 迁移完成，保证新库与生产升级走同一条迁移路径。
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

// migrateAttendanceDailySeq 考勤日记录支持同日多版本（seq 追加式写入）：
// ① AutoMigrate 新增 seq 列（存量行默认 1）；② 旧唯一索引 (person_id, event_date)
// 会阻止同日多记录，必须先删除；③ 以 (person_id, event_date, seq) 部分唯一索引替代。
func migrateAttendanceDailySeq(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.AttendanceDaily{}); err != nil {
		return err
	}
	if err := db.Exec(`UPDATE attendance_daily SET seq = 1 WHERE seq = 0`).Error; err != nil {
		return err
	}
	if err := db.Exec(`DROP INDEX IF EXISTS idx_attendance_daily_person_date`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_daily_person_date_seq
		ON attendance_daily(person_id, event_date, seq) WHERE deleted_at IS NULL`).Error; err != nil {
		return err
	}
	return nil
}
