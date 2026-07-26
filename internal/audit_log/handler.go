package audit_log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"probig/internal/pkg/audit"
	"probig/internal/pkg/config"
	"probig/internal/pkg/excel"
	"probig/internal/pkg/middleware"
	"probig/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) List(pageNum, pageSize int, operatorName, action, targetType, startDate, endDate string) ([]audit.AuditLog, int64, error) {
	var list []audit.AuditLog
	var total int64
	db := s.DB.Model(&audit.AuditLog{})
	if operatorName != "" {
		db = db.Where("operator_name like ?", "%"+operatorName+"%")
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if targetType != "" {
		db = db.Where("target_type = ?", targetType)
	}
	if startDate != "" {
		db = db.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("created_at <= ?", endDate+" 23:59:59")
	}
	db.Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&list).Error
	return list, total, err
}

func RegisterRoutes(r *gin.RouterGroup) {
	r.Use(middleware.Auth)
	r.GET("", middleware.Permission("audit:read"), List)
	r.GET("/export", middleware.Permission("audit:export"), Export)
}

func List(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.List(pageNum, pageSize,
		c.Query("operator_name"), c.Query("action"), c.Query("target_type"),
		c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func Export(c *gin.Context) {
	svc := NewService()
	list, _, err := svc.List(1, 100000,
		c.Query("operator_name"), c.Query("action"), c.Query("target_type"),
		c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		response.Error(c, "导出失败")
		return
	}
	exp := excel.NewExporter([]string{"操作人", "操作类型", "对象类型", "对象名称", "操作时间", "IP"})
	for _, l := range list {
		exp.AddRow([]string{
			l.OperatorName,
			l.Action,
			l.TargetType,
			l.TargetName,
			l.CreatedAt.Format("2006-01-02 15:04:05"),
			l.IP,
		})
	}
	filePath := filepath.Join(os.TempDir(), fmt.Sprintf("audit_export_%d.csv", time.Now().Unix()))
	if err := exp.WriteCSV(filePath); err != nil {
		response.Error(c, "导出失败")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=audit_export.csv")
	c.Header("Content-Type", "text/csv")
	c.File(filePath)
}
