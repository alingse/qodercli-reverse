package utils

import (
	"os"
	"path/filepath"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
)

// InitLogger 初始化日志系统
func InitLogger(logFile string, debug bool) error {
	logLevel := log.LevelInfo
	if debug {
		logLevel = log.LevelDebug
	}

	if logFile == "" {
		logFile = GetDefaultLogFile()
	}

	return log.Init(logFile, logLevel)
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
