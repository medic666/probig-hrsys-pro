package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"probig/server/internal/config"
	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/router"
	"probig/server/internal/service"

	"gorm.io/gorm"
)

func main() {
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Printf("警告: 加载配置文件失败，使用默认配置: %v", err)
	}

	dbPath := config.AppConfig.Database.Path
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("创建数据库目录失败: %v", err)
	}

	db, err := dao.InitDB(dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := initSystem(db); err != nil {
		log.Fatalf("初始化系统失败: %v", err)
	}

	r := router.SetupRouter()

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("服务启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func initSystem(db *gorm.DB) error {
	if err := autoMigrate(db); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	if err := service.InitSysConfig(db); err != nil {
		return fmt.Errorf("初始化系统配置失败: %w", err)
	}

	if err := service.SeedDefaultAdmin(db); err != nil {
		return fmt.Errorf("初始化默认管理员失败: %w", err)
	}

	log.Println("系统初始化完成")
	return nil
}

func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.AuditLog{},
		&model.SysConfig{},
		&model.Person{},
		&model.PersonPhone{},
		&model.PersonEmail{},
		&model.PersonBankCard{},
		&model.Company{},
		&model.File{},
		&model.FileRelation{},
		&model.PositionEvent{},
		&model.PositionSnapshot{},
		&model.AttendanceEvent{},
		&model.AttendanceDailyProjection{},
		&model.AttendanceCalculationMonthly{},
		&model.AnnualLeaveAccountEvent{},
		&model.AnnualLeaveBalanceSnapshot{},
		&model.LeaveInLieuBalanceSnapshot{},
		&model.SysBatch{},
	); err != nil {
		return err
	}

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_persons_id_card ON persons(id_card) WHERE deleted_at IS NULL AND id_card != ''")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_credit_code ON companies(credit_code) WHERE deleted_at IS NULL AND credit_code != ''")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_position_events_person_seq ON position_events(person_id, seq) WHERE deleted_at IS NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_events_person_seq ON attendance_events(person_id, seq) WHERE deleted_at IS NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_annual_leave_events_person_seq ON annual_leave_account_events(person_id, seq) WHERE deleted_at IS NULL")

	return nil
}
