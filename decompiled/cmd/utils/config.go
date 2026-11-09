package utils

import (
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
)

// Flags contains all CLI flags - strictly aligned with official
type Flags struct {
	Model           string
	MaxTokens       int
	Temperature     float64
	MaxTurns        int
	PermissionMode  string
	OutputFormat    string
	AllowedTools    []string
	DisallowedTools []string
	Workspace       string

	// Official flags
	WithClaudeConfig bool // Load .claude configuration
	Debug            bool // Enable debug mode
}

// LoadConfig loads config from flags and env vars
func LoadConfig(flags *Flags) *config.Config {
	cfg := config.LoadFromEnv()

	// CLI flags override env vars
	if flags.Model != "" && flags.Model != "auto" {
		cfg.Model = flags.Model
	}
	if flags.MaxTokens > 0 {
		cfg.MaxTokens = flags.MaxTokens
	}
	if flags.Temperature > 0 {
		cfg.Temperature = flags.Temperature
	}
	if flags.MaxTurns > 0 {
		cfg.MaxTurns = flags.MaxTurns
	}
	if flags.PermissionMode != "" {
		cfg.PermissionMode = flags.PermissionMode
	}
	if flags.OutputFormat != "" {
		cfg.OutputFormat = flags.OutputFormat
	}
	if len(flags.AllowedTools) > 0 {
		cfg.AllowedTools = flags.AllowedTools
	}
	if len(flags.DisallowedTools) > 0 {
		cfg.DisallowedTools = flags.DisallowedTools
	}
	if flags.Workspace != "" {
		cfg.Workspace = flags.Workspace
	}

	return cfg
}
