// Package log logging system
// Supports writing debug logs and API errors to files and console output.
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Level log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the log level.
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

// Logger log recorder.
type Logger struct {
	level          Level // minimum log level written to file
	stderrMinLevel Level // minimum log level output to stderr
	file           *os.File
	filePath       string
	prefix         string
	logToStderr    bool
}

// defaultLogger default log recorder.
var defaultLogger *Logger

// Init initializes the logging system.
// logFile: path to the log file; if empty, output only to stderr.
// level: log level (controls the level written to file).
func Init(logFile string, level Level) error {
	var file *os.File
	var filePath string

	if logFile != "" {
		// ensure directory exists
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

	// stderr only outputs ERROR and above; debug messages are written to file only.
	stderrMinLevel := LevelError

	defaultLogger = &Logger{
		level:          level,
		stderrMinLevel: stderrMinLevel,
		file:           file,
		filePath:       filePath,
		logToStderr:    true,
	}

	return nil
}

// SetLevel sets the log level.
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.level = level
	}
}

// GetLevel returns the current log level.
func GetLevel() Level {
	if defaultLogger == nil {
		return LevelInfo
	}
	return defaultLogger.level
}

// Close closes the log file.
func Close() error {
	if defaultLogger != nil && defaultLogger.file != nil {
		return defaultLogger.file.Close()
	}
	return nil
}

// log internal logging function.
func log(level Level, format string, args ...interface{}) {
	if defaultLogger == nil {
		// if not initialized, default output to stderr
		if level >= LevelInfo {
			msg := formatLog(level, "", format, args...)
			fmt.Fprintln(os.Stderr, msg)
		}
		return
	}

	// check log level (file and stderr use their own level control)
	needFile := defaultLogger.file != nil && level >= defaultLogger.level
	needStderr := defaultLogger.logToStderr && level >= defaultLogger.stderrMinLevel

	if !needFile && !needStderr {
		return
	}

	// get caller info
	_, file, line, ok := runtime.Caller(2)
	callerInfo := ""
	if ok {
		// simplify file path
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}
		callerInfo = fmt.Sprintf("%s:%d", file, line)
	}

	msg := formatLog(level, callerInfo, format, args...)

	// output to stderr (judged by stderrMinLevel; by default only ERROR and above, debug mode outputs all)
	if needStderr {
		fmt.Fprintln(os.Stderr, msg)
	}

	// output to file
	if needFile {
		defaultLogger.file.WriteString(msg + "\n")
	}
}

// formatLog formats the log message.
func formatLog(level Level, callerInfo, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)

	if callerInfo != "" {
		return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level.String(), callerInfo, msg)
	}
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), msg)
}

// Debug records a debug log.
func Debug(format string, args ...interface{}) {
	log(LevelDebug, format, args...)
}

// Info records an info log.
func Info(format string, args ...interface{}) {
	log(LevelInfo, format, args...)
}

// Warn records a warning log.
func Warn(format string, args ...interface{}) {
	log(LevelWarn, format, args...)
}

// Error records an error log.
func Error(format string, args ...interface{}) {
	log(LevelError, format, args...)
}

// Fatal records a fatal error and exits.
func Fatal(format string, args ...interface{}) {
	log(LevelFatal, format, args...)
	os.Exit(1)
}

// Debugf records a debug log (with format).
func Debugf(format string, args ...interface{}) {
	Debug(format, args...)
}

// Infof records an info log (with format).
func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

// Warnf records a warning log (with format).
func Warnf(format string, args ...interface{}) {
	Warn(format, args...)
}

// Errorf records an error log (with format).
func Errorf(format string, args ...interface{}) {
	Error(format, args...)
}

// Fatalf records a fatal error and exits (with format).
func Fatalf(format string, args ...interface{}) {
	Fatal(format, args...)
}

// WithPrefix creates a logging function with a prefix.
func WithPrefix(prefix string) func(string, ...interface{}) {
	return func(format string, args ...interface{}) {
		Info("["+prefix+"] "+format, args...)
	}
}

// GetLogFile returns the log file path.
func GetLogFile() string {
	if defaultLogger != nil {
		return defaultLogger.filePath
	}
	return ""
}
