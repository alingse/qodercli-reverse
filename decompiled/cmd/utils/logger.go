package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
)

// InitLogger initializes the logging system
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

// GetDefaultLogFile returns the default log file path
func GetDefaultLogFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "qodercli-reverse.log"
	}

	logDir := filepath.Join(homeDir, ".qodercli-reverse")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "qodercli-reverse.log"
	}

	return filepath.Join(logDir, "qodercli-reverse.log")
}

// GetDebugLogFile returns the timestamped debug log file path
func GetDebugLogFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "debug.log"
	}

	logDir := filepath.Join(homeDir, ".qodercli-reverse")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "debug.log"
	}

	ts := time.Now().Format("20060102-150405")
	return filepath.Join(logDir, fmt.Sprintf("debug.%s.log", ts))
}
