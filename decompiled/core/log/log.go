// Package log 日志系统
// 支持将 debug log 和 API 错误写入文件和控制台输出
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器
type Logger struct {
	level      Level
	file       *os.File
	filePath   string
	prefix     string
	logToStderr bool
}

// defaultLogger 默认日志记录器
var defaultLogger *Logger

// Init 初始化日志系统
// logFile: 日志文件路径，如果为空则只输出到 stderr
// level: 日志级别
func Init(logFile string, level Level) error {
	var file *os.File
	var filePath string

	if logFile != "" {
		// 确保目录存在
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create log directory: %w", err)
		}

		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
		file = f
		filePath = logFile
	}

	defaultLogger = &Logger{
		level:       level,
		file:        file,
		filePath:    filePath,
		logToStderr: true,
	}

	return nil
}

// SetLevel 设置日志级别
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.level = level
	}
}

// GetLevel 获取当前日志级别
func GetLevel() Level {
	if defaultLogger == nil {
		return LevelInfo
	}
	return defaultLogger.level
}

// Close 关闭日志文件
func Close() error {
	if defaultLogger != nil && defaultLogger.file != nil {
		return defaultLogger.file.Close()
	}
	return nil
}

// log 内部日志记录函数
func log(level Level, format string, args ...interface{}) {
	if defaultLogger == nil {
		// 如果未初始化，默认输出到 stderr
		if level >= LevelInfo {
			msg := formatLog(level, "", format, args...)
			fmt.Fprintln(os.Stderr, msg)
		}
		return
	}

	// 检查日志级别
	if level < defaultLogger.level {
		return
	}

	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	callerInfo := ""
	if ok {
		// 简化文件路径
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}
		callerInfo = fmt.Sprintf("%s:%d", file, line)
	}

	msg := formatLog(level, callerInfo, format, args...)

	// 输出到 stderr
	if defaultLogger.logToStderr && level >= LevelInfo {
		fmt.Fprintln(os.Stderr, msg)
	}

	// 输出到文件
	if defaultLogger.file != nil {
		defaultLogger.file.WriteString(msg + "\n")
	}
}

// formatLog 格式化日志消息
func formatLog(level Level, callerInfo, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)

	if callerInfo != "" {
		return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level.String(), callerInfo, msg)
	}
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), msg)
}

// Debug 记录调试日志
func Debug(format string, args ...interface{}) {
	log(LevelDebug, format, args...)
}

// Info 记录信息日志
func Info(format string, args ...interface{}) {
	log(LevelInfo, format, args...)
}

// Warn 记录警告日志
func Warn(format string, args ...interface{}) {
	log(LevelWarn, format, args...)
}

// Error 记录错误日志
func Error(format string, args ...interface{}) {
	log(LevelError, format, args...)
}

// Fatal 记录致命错误并退出
func Fatal(format string, args ...interface{}) {
	log(LevelFatal, format, args...)
	os.Exit(1)
}

// Debugf 记录调试日志（带格式）
func Debugf(format string, args ...interface{}) {
	Debug(format, args...)
}

// Infof 记录信息日志（带格式）
func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

// Warnf 记录警告日志（带格式）
func Warnf(format string, args ...interface{}) {
	Warn(format, args...)
}

// Errorf 记录错误日志（带格式）
func Errorf(format string, args ...interface{}) {
	Error(format, args...)
}

// Fatalf 记录致命错误并退出（带格式）
func Fatalf(format string, args ...interface{}) {
	Fatal(format, args...)
}

// WithPrefix 创建带前缀的日志函数
func WithPrefix(prefix string) func(string, ...interface{}) {
	return func(format string, args ...interface{}) {
		Info("["+prefix+"] "+format, args...)
	}
}

// GetLogFile 获取日志文件路径
func GetLogFile() string {
	if defaultLogger != nil {
		return defaultLogger.filePath
	}
	return ""
}
