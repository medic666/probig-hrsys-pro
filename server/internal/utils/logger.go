package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level 日志级别（数字越小越详细）
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// ParseLevel 解析级别字符串（debug/info/warn/error），非法或空值回退 Info。
func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return DebugLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// dailyRotateWriter 按天轮转 writer：
// 活跃文件恒为 basePath（最新一天），写入时检测跨天 → 已写内容改名为 basePath.YYYYMMDD
// （离开那天的日期，服务停多天后重启也能正确归档）→ 重开新文件续写；
// 仅保留最近 retainDays 份历史文件（含当日），超期自动删除。
type dailyRotateWriter struct {
	mu         sync.Mutex
	file       *os.File
	basePath   string
	curDate    string
	retainDays int
	nowFunc    func() time.Time // 可注入时钟（测试）
}

func newDailyRotateWriter(basePath string, retainDays int) (*dailyRotateWriter, error) {
	if retainDays < 1 {
		retainDays = 1
	}
	w := &dailyRotateWriter{
		basePath:   basePath,
		curDate:    time.Now().Format("20060102"),
		retainDays: retainDays,
		nowFunc:    time.Now,
	}
	f, err := os.OpenFile(basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	w.file = f
	return w, nil
}

func (w *dailyRotateWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	today := w.nowFunc().Format("20060102")
	if today != w.curDate {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

func (w *dailyRotateWriter) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	archived := fmt.Sprintf("%s.%s", w.basePath, w.curDate)
	if err := os.Rename(w.basePath, archived); err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(w.basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = f
	w.curDate = w.nowFunc().Format("20060102")
	w.cleanup()
	return nil
}

// cleanup 扫描 basePath.YYYYMMDD 历史文件，保留最近 retainDays 天（活跃文件为当日，
// 另保留 retainDays-1 份历史），更早的旧文件删除。
func (w *dailyRotateWriter) cleanup() {
	dir := filepath.Dir(w.basePath)
	base := filepath.Base(w.basePath)
	matches, err := filepath.Glob(filepath.Join(dir, base+".*"))
	if err != nil {
		return
	}
	cutoff := w.nowFunc().AddDate(0, 0, -(w.retainDays - 1)).Format("20060102")
	var dates []string
	for _, m := range matches {
		name := filepath.Base(m)
		date := strings.TrimPrefix(name, base+".")
		if len(date) == 8 {
			dates = append(dates, date)
		}
	}
	sort.Strings(dates)
	for _, date := range dates {
		if date >= cutoff {
			continue
		}
		os.Remove(filepath.Join(dir, base+"."+date))
	}
}

// logger 业务日志实例：级别过滤 + 轮转落盘/控制台双写（输出由 Init 统一接入）
var (
	logMu    sync.RWMutex
	logger   *log.Logger
	minLevel = InfoLevel
	writer   io.Writer
)

// Init 初始化业务日志：打开/创建按天轮转文件，日志双写（控制台 + 文件，文件轮转保留 retainDays 天）。
// 返回统一 writer（供 gin.DefaultWriter 对接 access log 同源轮转）。
func Init(filePath string, level Level, retainDays int) (io.Writer, error) {
	w, err := newDailyRotateWriter(filePath, retainDays)
	if err != nil {
		return nil, err
	}
	out := io.MultiWriter(os.Stdout, w)
	logMu.Lock()
	writer = out
	logger = log.New(out, "probig ", log.LstdFlags)
	minLevel = level
	logMu.Unlock()
	return out, nil
}

// Writer 返回当前日志 writer（access log 对接；未初始化时回退控制台）
func Writer() io.Writer {
	logMu.RLock()
	defer logMu.RUnlock()
	if writer != nil {
		return writer
	}
	return os.Stdout
}

func logf(level Level, format string, args ...any) {
	logMu.RLock()
	l := logger
	min := minLevel
	logMu.RUnlock()
	if l == nil || level < min {
		return
	}
	l.Printf(format, args...)
}

func Debugf(format string, args ...any) { logf(DebugLevel, format, args...) }
func Infof(format string, args ...any)  { logf(InfoLevel, format, args...) }
func Warnf(format string, args ...any)  { logf(WarnLevel, format, args...) }
func Errorf(format string, args ...any) { logf(ErrorLevel, format, args...) }
