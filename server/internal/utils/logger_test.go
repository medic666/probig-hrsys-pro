package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want Level
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"", InfoLevel},
		{"unknown", InfoLevel},
	}
	for _, c := range cases {
		if got := ParseLevel(c.in); got != c.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestLoggerLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	// 未初始化（logger==nil）：任何级别都不输出
	Debugf("d")
	Infof("i")
	Warnf("w")
	Errorf("e")
	if buf.Len() != 0 {
		t.Fatalf("uninitialized logger should not output, got %q", buf.String())
	}

	// 初始化到 info 级（level 过滤走自持实例，Init 的 writer 供测试捕获）
	dir := t.TempDir()
	w, err := Init(filepath.Join(dir, "test.log"), InfoLevel, 30)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	_ = w
	// 通过包级 logger 输出到临时文件，验证级别过滤
	Debugf("debug-msg")
	Infof("info-msg")
	Warnf("warn-msg")
	Errorf("error-msg")
	data, err := os.ReadFile(filepath.Join(dir, "test.log"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	s := string(data)
	if bytes.Contains([]byte(s), []byte("debug-msg")) {
		t.Errorf("debug level should be filtered at info, got %q", s)
	}
	for _, want := range []string{"info-msg", "warn-msg", "error-msg"} {
		if !bytes.Contains([]byte(s), []byte(want)) {
			t.Errorf("missing %q in log output %q", want, s)
		}
	}
}

func TestDailyRotateWriter(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "probig.log")

	w, err := newDailyRotateWriter(base, 2)
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}
	day1 := time.Date(2026, 8, 1, 10, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 8, 2, 10, 0, 0, 0, time.Local)
	day3 := time.Date(2026, 8, 3, 10, 0, 0, 0, time.Local)
	w.nowFunc = func() time.Time { return day1 }
	w.curDate = day1.Format("20060102")

	if _, err := w.Write([]byte("day1\n")); err != nil {
		t.Fatalf("write day1: %v", err)
	}
	// 跨天 → 归档为 probig.log.20260801，新文件续写 day2
	w.nowFunc = func() time.Time { return day2 }
	if _, err := w.Write([]byte("day2\n")); err != nil {
		t.Fatalf("write day2: %v", err)
	}
	archived, err := os.ReadFile(filepath.Join(dir, "probig.log.20260801"))
	if err != nil {
		t.Fatalf("archived file missing: %v", err)
	}
	if string(archived) != "day1\n" {
		t.Errorf("archived content = %q, want day1", string(archived))
	}
	cur, _ := os.ReadFile(base)
	if string(cur) != "day2\n" {
		t.Errorf("active content = %q, want day2", string(cur))
	}

	// 第 3 天：保留 2 天 → 20260801 应被清理，20260802 保留
	// 先人工造一个超期文件 20260730 验证清理范围
	os.WriteFile(filepath.Join(dir, "probig.log.20260730"), []byte("old"), 0644)
	w.nowFunc = func() time.Time { return day3 }
	if _, err := w.Write([]byte("day3\n")); err != nil {
		t.Fatalf("write day3: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "probig.log.20260801")); !os.IsNotExist(err) {
		t.Errorf("20260801 should be cleaned (older than 2 days), stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "probig.log.20260802")); err != nil {
		t.Errorf("20260802 should be kept, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "probig.log.20260730")); !os.IsNotExist(err) {
		t.Errorf("20260730 (old) should be cleaned, stat err=%v", err)
	}
	w.file.Close()
}
