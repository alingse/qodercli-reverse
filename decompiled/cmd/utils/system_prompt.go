// Package utils 系统提示词构建工具
// 严格对齐官方 qodercli 架构
package utils

import (
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/prompts"
)

// BuildSystemPromptAuto 自动构建系统提示词（官方行为）
// 当用户没有提供 --system-prompt 时，内部自动构建
func BuildSystemPromptAuto(workDir string, withClaudeConfig bool) (string, error) {
	vars := prompts.DefaultTemplateVars()

	// 创建构建器
	builder := prompts.NewSystemPromptBuilderV2(vars)

	// 启用标准组件（官方默认行为）
	builder.WithRoleDefinition(true).
		WithToolGuide(true).
		WithPermissionRules(true).
		WithEnvironmentInfo(true).
		WithProjectContext(true).
		WithCodingStandards(true).
		WithSessionContext(false) // 会话上下文由服务器端管理

	// 收集环境信息
	if _, err := builder.CollectEnvironment(); err != nil {
		log.Debug("Failed to collect environment: %v", err)
	}

	// 收集项目上下文
	if workDir != "" {
		if _, err := builder.CollectProjectContext(workDir); err != nil {
			log.Debug("Failed to collect project context: %v", err)
		}
	}

	// 如果启用了 --with-claude-config，额外加载 .claude/ 配置
	if withClaudeConfig {
		// 项目上下文加载器已经会加载 .claude/ 目录
		// 这里可以添加额外的 Claude 特定配置
		log.Debug("Loading .claude configuration")
	}

	return builder.Build()
}

// GetDefaultSystemPrompt 获取默认系统提示词（向后兼容）
func GetDefaultSystemPrompt() string {
	// 使用最小化构建
	vars := prompts.DefaultTemplateVars()
	builder := prompts.NewSystemPromptBuilderV2(vars)
	builder.WithRoleDefinition(true).
		WithToolGuide(true).
		WithPermissionRules(true).
		WithEnvironmentInfo(false).
		WithProjectContext(false).
		WithCodingStandards(false).
		WithSessionContext(false)

	prompt, err := builder.BuildMinimal()
	if err != nil {
		// 最后的回退
		return "You are qodercli, an interactive CLI tool that helps users with software engineering tasks."
	}
	return prompt
}

// LoadProjectContext 加载项目上下文（供外部使用）
func LoadProjectContext(workDir string) (*prompts.ProjectContext, error) {
	loader := prompts.NewProjectContextLoader(workDir)
	return loader.Load()
}

// CollectEnvironmentInfo 收集环境信息（供外部使用）
func CollectEnvironmentInfo() (*prompts.EnvironmentInfo, error) {
	collector := prompts.NewEnvironmentCollector()
	return collector.Collect()
}
