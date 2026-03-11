package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
)

// InitLogger 初始化日志系统
func InitLogger(logFile string, debug bool) error {
	logLevel := log.LevelInfo
	if debug {
		logLevel = log.LevelDebug
	}

	if logFile == "" {
		if debug {
			logFile = GetDebugLogFile()
		} else {
			logFile = GetDefaultLogFile()
		}
	}

	if err := log.Init(logFile, logLevel); err != nil {
		return err
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Debug log: %s\n", logFile)
	}

	return nil
}

// GetDefaultLogFile 获取默认日志文件路径
func GetDefaultLogFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "qodercli.log"
	}

	logDir := filepath.Join(homeDir, ".qoder")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "qodercli.log"
	}

	return filepath.Join(logDir, "qodercli.log")
}

// GetDebugLogFile 获取带时间戳的 debug 日志文件路径
func GetDebugLogFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "debug.log"
	}

	logDir := filepath.Join(homeDir, ".qoder")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "debug.log"
	}

	ts := time.Now().Format("20060102-150405")
	return filepath.Join(logDir, fmt.Sprintf("debug.%s.log", ts))
}
