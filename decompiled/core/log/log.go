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
	level         Level      // 写入文件的最低日志级别
	stderrMinLevel Level     // 输出到 stderr 的最低日志级别
	file          *os.File
	filePath      string
	prefix        string
	logToStderr   bool
}

// defaultLogger 默认日志记录器
var defaultLogger *Logger

// Init 初始化日志系统
// logFile: 日志文件路径，如果为空则只输出到 stderr
// level: 日志级别（控制写入文件的级别，以及 debug 模式下 stderr 的级别）
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

	// 根据 level 决定 stderr 的最小输出级别
	// 如果是 debug 模式 (LevelDebug)，则 stderr 输出所有级别
	// 否则，stderr 只输出 ERROR 及以上级别
	stderrMinLevel := LevelError
	if level == LevelDebug {
		stderrMinLevel = LevelDebug
	}

	defaultLogger = &Logger{
		level:          level,
		stderrMinLevel: stderrMinLevel,
		file:           file,
		filePath:       filePath,
		logToStderr:    true,
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

	// 检查日志级别（文件和 stderr 使用各自的级别控制）
	needFile := defaultLogger.file != nil && level >= defaultLogger.level
	needStderr := defaultLogger.logToStderr && level >= defaultLogger.stderrMinLevel

	if !needFile && !needStderr {
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

	// 输出到 stderr（根据 stderrMinLevel 判断，默认只输出 ERROR 及以上，debug 模式输出所有）
	if needStderr {
		fmt.Fprintln(os.Stderr, msg)
	}

	// 输出到文件
	if needFile {
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
