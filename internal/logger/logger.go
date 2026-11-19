package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"GameTrainerHub/internal/model"
)

var (
	defaultLogger *slog.Logger
	logFile       *os.File
	mu            sync.Mutex
	recentBuffer  = newRingBuffer(500)
	sensitiveReg  = regexp.MustCompile(`(?i)(password|token|key|secret|auth|proxy)=([^&\s]+)`)
)

type ringBuffer struct {
	entries []model.LogEntry
	maxSize int
	mu      sync.RWMutex
}

func newRingBuffer(size int) *ringBuffer {
	if size <= 0 {
		size = 200
	}
	return &ringBuffer{
		entries: make([]model.LogEntry, 0, size),
		maxSize: size,
	}
}

func (rb *ringBuffer) add(entry model.LogEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if len(rb.entries) >= rb.maxSize {
		// 移除最早的条目
		rb.entries = rb.entries[1:]
	}
	rb.entries = append(rb.entries, entry)
}

func (rb *ringBuffer) get(limit int, category string) []model.LogEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var result []model.LogEntry
	catLower := strings.ToLower(category)

	// 从最新到最旧反向选取
	for i := len(rb.entries) - 1; i >= 0; i-- {
		e := rb.entries[i]
		if catLower == "" || catLower == "all" || strings.ToLower(e.Category) == catLower {
			result = append(result, e)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}

// MaskSensitiveData 对包含敏感信息（密码、密钥、代理认证等）的字符串进行脱敏处理
func MaskSensitiveData(input string) string {
	return sensitiveReg.ReplaceAllString(input, "${1}=***")
}

func categorizeMessage(msg string, details string) string {
	combined := strings.ToLower(msg + " " + details)
	if strings.Contains(combined, "download") || strings.Contains(combined, "range") || strings.Contains(combined, "task") {
		return "download"
	}
	if strings.Contains(combined, "launch") || strings.Contains(combined, "process") || strings.Contains(combined, "trainer") {
		return "launch"
	}
	if strings.Contains(combined, "search") || strings.Contains(combined, "query") || strings.Contains(combined, "fts") || strings.Contains(combined, "translate") {
		return "search"
	}
	return "system"
}

// recordEntry 将日志记入内存循环缓冲区
func recordEntry(level string, msg string, args ...any) {
	var details strings.Builder
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			details.WriteString(fmt.Sprintf("%v=%v ", args[i], args[i+1]))
		} else {
			details.WriteString(fmt.Sprintf("%v ", args[i]))
		}
	}
	rawDetails := details.String()
	maskedDetails := MaskSensitiveData(rawDetails)
	category := categorizeMessage(msg, maskedDetails)

	entry := model.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   MaskSensitiveData(msg),
		Details:   strings.TrimSpace(maskedDetails),
	}
	recentBuffer.add(entry)
}

// GetRecentLogs 查询最近结构化分类日志
func GetRecentLogs(limit int, category string) []model.LogEntry {
	return recentBuffer.get(limit, category)
}

// InitLogger 初始化结构化日志
func InitLogger(logDir string, levelStr string) (*slog.Logger, error) {
	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log dir: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(logDir, fmt.Sprintf("app_%s.log", today))

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// 解析日志级别
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	// 组合输出到控制台与文件
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	handler := slog.NewTextHandler(multiWriter, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	defaultLogger.Info("Logger initialized successfully", "log_file", logFilePath, "level", level.String())
	recordEntry("INFO", "Logger initialized successfully", "log_file", logFilePath, "level", level.String())
	return defaultLogger, nil
}

// GetLogger 获取默认日志记录器
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		defaultLogger = slog.Default()
	}
	return defaultLogger
}

// CloseLogger 关闭日志文件流
func CloseLogger() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		_ = logFile.Sync()
		_ = logFile.Close()
		logFile = nil
	}
}

// 便捷记录函数
func Debug(msg string, args ...any) {
	recordEntry("DEBUG", msg, args...)
	GetLogger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	recordEntry("INFO", msg, args...)
	GetLogger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	recordEntry("WARN", msg, args...)
	GetLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	recordEntry("ERROR", msg, args...)
	GetLogger().Error(msg, args...)
}

