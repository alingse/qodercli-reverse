// Package utils system prompt building tools
// Strictly aligned with official qodercli-reverse architecture
package utils

import (
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/prompts"
)

// BuildSystemPromptAuto auto-builds system prompt (official behavior)
// Auto-builds internally when user doesn't provide --system-prompt
func BuildSystemPromptAuto(workDir string, withClaudeConfig bool) (string, error) {
	vars := prompts.DefaultTemplateVars()

	// Create builder
	builder := prompts.NewSystemPromptBuilderV2(vars)

	// Enable standard components (official default behavior)
	builder.WithRoleDefinition(true).
		WithToolGuide(true).
		WithPermissionRules(true).
		WithEnvironmentInfo(true).
		WithProjectContext(true).
		WithCodingStandards(true).
		WithSessionContext(false) // Session context managed server-side

	// Collect environment info
	if _, err := builder.CollectEnvironment(); err != nil {
		log.Debug("Failed to collect environment: %v", err)
	}

	// Collect project context
	if workDir != "" {
		if _, err := builder.CollectProjectContext(workDir); err != nil {
			log.Debug("Failed to collect project context: %v", err)
		}
	}

	// If --with-claude-config enabled, additionally load .claude/ config
	if withClaudeConfig {
		// Project context loader already loads .claude/ directory
		// Additional Claude-specific config can be added here
		log.Debug("Loading .claude configuration")
	}

	return builder.Build()
}

// GetDefaultSystemPrompt returns default system prompt (backward compatible)
func GetDefaultSystemPrompt() string {
	// Use minimal build
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
		// Final fallback
		return "You are qodercli-reverse, an interactive CLI tool that helps users with software engineering tasks."
	}
	return prompt
}

// LoadProjectContext loads project context (for external use)
func LoadProjectContext(workDir string) (*prompts.ProjectContext, error) {
	loader := prompts.NewProjectContextLoader(workDir)
	return loader.Load()
}

// CollectEnvironmentInfo collects environment info (for external use)
func CollectEnvironmentInfo() (*prompts.EnvironmentInfo, error) {
	collector := prompts.NewEnvironmentCollector()
	return collector.Collect()
}
