// Package config 配置管理
package config

import "os"

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
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() *Config {
	cfg := &Config{}

	// 模型配置（优先级：QODER_MODEL > OPENAI_MODEL > 默认值）
	cfg.Model = getEnvOrDefault("QODER_MODEL",
		getEnvOrDefault("OPENAI_MODEL", "auto"))

	// API 配置
	cfg.Provider = getEnvOrDefault("QODER_PROVIDER", "openai")
	cfg.APIKey = getEnvOrDefault("OPENAI_API_KEY",
		getEnvOrDefault("QODER_PERSONAL_ACCESS_TOKEN", ""))
	cfg.BaseURL = getEnvOrDefault("OPENAI_BASE_URL",
		getEnvOrDefault("QODER_OPENAI_BASE_URL", "https://api.openai.com/v1"))

	return cfg
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
