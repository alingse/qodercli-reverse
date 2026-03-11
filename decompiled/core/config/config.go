// Package config 配置管理
package config

import "os"

// MarkdownConfig Markdown 渲染配置
type MarkdownConfig struct {
	Enabled  bool   // 是否启用 markdown 渲染
	Style    string // 主题名称（如 dark, light, github-dark 等）
	WordWrap int    // 自动换行宽度（0 表示自动根据终端宽度）
}

// DefaultMarkdownConfig 返回默认 markdown 配置
func DefaultMarkdownConfig() MarkdownConfig {
	return MarkdownConfig{
		Enabled:  true,
		Style:    "", // 空字符串表示自动检测
		WordWrap: 0,
	}
}

// Config 配置结构
type Config struct {
	// 模型配置
	Model string

	// API 配置
	Provider string
	APIKey   string
	BaseURL  string

	// 请求配置
	MaxTokens   int
	Temperature float64
	MaxTurns    int

	// 权限配置
	PermissionMode string

	// 输出配置
	OutputFormat string

	// 工具配置
	AllowedTools    []string
	DisallowedTools []string

	// 会话配置
	ContinueSession bool
	ResumeSessionID string

	// 工作区配置
	Workspace string

	// Markdown 渲染配置
	Markdown MarkdownConfig
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() *Config {
	cfg := &Config{
		Markdown: DefaultMarkdownConfig(),
	}

	// 模型配置（优先级：QODER_MODEL > OPENAI_MODEL > 默认值）
	cfg.Model = getEnvOrDefault("QODER_MODEL",
		getEnvOrDefault("OPENAI_MODEL", "auto"))

	// API 配置
	cfg.Provider = getEnvOrDefault("QODER_PROVIDER", "openai")
	cfg.APIKey = getEnvOrDefault("OPENAI_API_KEY",
		getEnvOrDefault("QODER_PERSONAL_ACCESS_TOKEN", ""))
	cfg.BaseURL = getEnvOrDefault("OPENAI_BASE_URL",
		getEnvOrDefault("QODER_OPENAI_BASE_URL", "https://api.openai.com/v1"))

	// Markdown 配置
	if os.Getenv("QODER_MARKDOWN_DISABLED") == "true" {
		cfg.Markdown.Enabled = false
	}
	cfg.Markdown.Style = getEnvOrDefault("QODER_MARKDOWN_STYLE", "")

	return cfg
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
