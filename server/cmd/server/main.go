package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"probig/server/internal/config"
	"probig/server/internal/dao"
	"probig/server/internal/router"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	loaded := config.LoadConfig()

	if !loaded {
		if err := config.WriteDefaultConfig(); err != nil {
			log.Fatalf("生成默认配置文件失败: %v", err)
		}
	}

	dbPath := config.ResolvePath(config.AppConfig.Database.Path)
	uploadDir := config.ResolvePath(config.AppConfig.FileStorage.Path)
	logPath := config.ResolvePath(config.AppConfig.Log.File)

	for _, p := range []string{filepath.Dir(dbPath), uploadDir, filepath.Dir(logPath)} {
		if err := os.MkdirAll(p, 0755); err != nil {
			log.Fatalf("创建目录失败 %s: %v", p, err)
		}
	}

	// 日志初始化：业务日志级别过滤 + 按天轮转落盘（保留 retain_days 天），
	// 返回统一 writer 供 gin access log 同源输出；打开失败降级仅控制台
	logWriter, err := utils.Init(logPath,
		utils.ParseLevel(config.AppConfig.Log.Level),
		config.AppConfig.Log.RetainDays)
	if err != nil {
		log.Printf("日志文件打开失败，仅输出控制台: %v", err)
		logWriter = os.Stdout
	}
	gin.DefaultWriter = logWriter
	gin.DefaultErrorWriter = logWriter

	// GIN 模式按配置生效（须在 engine 创建前设置）；非法值回退 release（生产安全）
	gin.SetMode(resolveGinMode(config.AppConfig.Server.Mode))

	utils.Infof("数据库路径: %s", dbPath)
	utils.Infof("文件存储路径: %s", uploadDir)

	db, err := dao.InitDB(dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := initSystem(db); err != nil {
		log.Fatalf("初始化系统失败: %v", err)
	}

	r := router.SetupRouter()

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Status(http.StatusNotFound)
			return
		}
		serveEmbeddedFile(c)
	})

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		utils.Infof("服务启动在 http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	<-ctx.Done()
	utils.Infof("收到退出信号，正在优雅关闭...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		utils.Warnf("优雅关闭失败: %v", err)
	}
	utils.Infof("服务已退出")
}

// resolveGinMode 校验配置的 GIN 模式，非法值回退 release
func resolveGinMode(mode string) string {
	switch mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		return mode
	default:
		return gin.ReleaseMode
	}
}

func initSystem(db *gorm.DB) error {
	if err := dao.RunMigrations(db); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	if err := service.InitSysConfig(db); err != nil {
		return fmt.Errorf("初始化系统配置失败: %w", err)
	}

	dao.RegisterAuditHooks(db)

	if err := service.SyncPermissionRows(db); err != nil {
		return fmt.Errorf("同步权限定义失败: %w", err)
	}

	if err := service.SeedDefaultAdmin(db); err != nil {
		return fmt.Errorf("初始化默认管理员失败: %w", err)
	}

	utils.Infof("系统初始化完成")
	return nil
}

var contentTypeMap = map[string]string{
	".html": "text/html; charset=utf-8",
	".js":   "application/javascript",
	".css":  "text/css",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".woff": "font/woff",
	".woff2": "font/woff2",
}

func serveEmbeddedFile(c *gin.Context) {
	urlPath := c.Request.URL.Path
	filePath := "static" + urlPath
	if urlPath == "/" {
		filePath = "static/index.html"
	}

	data, err := staticFiles.ReadFile(filePath)
	if err != nil {
		filePath = "static/index.html"
		data, err = staticFiles.ReadFile(filePath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
	}

	// 缓存策略：带内容哈希的构建资源可长期缓存（immutable），入口文档每次重新验证
	if strings.HasPrefix(urlPath, "/assets/") {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "no-cache")
	}

	contentType := "text/html; charset=utf-8"
	for ext, ct := range contentTypeMap {
		if strings.HasSuffix(filePath, ext) {
			contentType = ct
			break
		}
	}
	c.Data(http.StatusOK, contentType, data)
}
