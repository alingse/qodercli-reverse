// Package config configuration management
package config

import "os"

// MarkdownConfig Markdown rendering configuration
type MarkdownConfig struct {
	Enabled  bool   // whether markdown rendering is enabled
	Style    string // theme name (e.g. dark, light, github-dark, etc.)
	WordWrap int    // auto line-wrap width (0 means auto based on terminal width)
}

// DefaultMarkdownConfig returns the default markdown configuration
func DefaultMarkdownConfig() MarkdownConfig {
	return MarkdownConfig{
		Enabled:  true,
		Style:    "", // empty string means auto-detect
		WordWrap: 0,
	}
}

// Config configuration structure
type Config struct {
	// Model configuration
	Model string

	// API configuration
	Provider string
	APIKey   string
	BaseURL  string

	// Request configuration
	MaxTokens   int
	Temperature float64
	MaxTurns    int

	// Permission configuration
	PermissionMode string

	// Output configuration
	OutputFormat string

	// Tool configuration
	AllowedTools    []string
	DisallowedTools []string

	// Session configuration
	ContinueSession bool
	ResumeSessionID string

	// Workspace configuration
	Workspace string

	// Markdown rendering configuration
	Markdown MarkdownConfig
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	cfg := &Config{
		Markdown: DefaultMarkdownConfig(),
	}

	// Model configuration (priority: QODER_MODEL > OPENAI_MODEL > default)
	cfg.Model = getEnvOrDefault("QODER_MODEL",
		getEnvOrDefault("OPENAI_MODEL", "auto"))

	// API configuration
	cfg.Provider = getEnvOrDefault("QODER_PROVIDER", "openai")
	cfg.APIKey = getEnvOrDefault("OPENAI_API_KEY",
		getEnvOrDefault("QODER_PERSONAL_ACCESS_TOKEN", ""))
	cfg.BaseURL = getEnvOrDefault("OPENAI_BASE_URL",
		getEnvOrDefault("QODER_OPENAI_BASE_URL", "https://api.openai.com/v1"))

	// Markdown configuration
	if os.Getenv("QODER_MARKDOWN_DISABLED") == "true" {
		cfg.Markdown.Enabled = false
	}
	cfg.Markdown.Style = getEnvOrDefault("QODER_MARKDOWN_STYLE", "")

	return cfg
}

// getEnvOrDefault gets an environment variable or its default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
