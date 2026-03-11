// Package agent Agent 配置和提示词集成
//
// 本文件扩展了 Agent 的配置功能，集成系统提示词管理
// 参考官方架构: code.alibaba-inc.com/qoder-core/qodercli/core/agent
package agent

import (
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/prompts"
)

// AgentConfig 扩展的 Agent 配置
type AgentConfig struct {
	// 基础配置（继承自原有 Config）
	SystemPrompt          string
	Model                 string
	MaxTokens             int
	Temperature           float64
	MaxTurns              int
	AllowedTools          []string
	DisallowedTools       []string
	PermissionMode        permission.Mode
	ThinkLevel            provider.ThinkLevel
	SubagentMode          bool
	AgentName             string
	EnableEnhancedCompact bool

	// 提示词配置
	PromptType    prompts.PromptType    // 使用内置提示词类型
	PromptVars    *prompts.TemplateVars // 提示词模板变量
	CustomPrompt  string                // 自定义提示词（覆盖内置）
	PromptOptions PromptOptions         // 提示词选项

	// 子 Agent 配置
	IsSubagent      bool
	SubagentType    string
	ParentAgentName string
}

// PromptOptions 提示词选项
type PromptOptions struct {
	// 提示词组件开关
	IncludeCoreInstructions bool
	IncludeToolRules        bool
	IncludeFileRules        bool
	IncludeEducational      bool
	IncludeSecurityRules    bool
	IncludeOutputFormat     bool

	// 自定义内容
	ExtraSections []ExtraPromptSection

	// 覆盖选项
	OverrideRoleDefinition string
}

// ExtraPromptSection 额外提示词章节
type ExtraPromptSection struct {
	Title    string
	Content  string
	Priority int
}

// DefaultPromptOptions 默认提示词选项
func DefaultPromptOptions() PromptOptions {
	return PromptOptions{
		IncludeCoreInstructions: true,
		IncludeToolRules:        true,
		IncludeFileRules:        true,
		IncludeEducational:      false,
		IncludeSecurityRules:    true,
		IncludeOutputFormat:     true,
	}
}

// ToConfig 转换为基础 Config
func (ac *AgentConfig) ToConfig() *Config {
	return &Config{
		SystemPrompt:          ac.SystemPrompt,
		Model:                 ac.Model,
		MaxTokens:             ac.MaxTokens,
		Temperature:           ac.Temperature,
		MaxTurns:              ac.MaxTurns,
		AllowedTools:          ac.AllowedTools,
		DisallowedTools:       ac.DisallowedTools,
		PermissionMode:        ac.PermissionMode,
		ThinkLevel:            ac.ThinkLevel,
		SubagentMode:          ac.SubagentMode,
		AgentName:             ac.AgentName,
		EnableEnhancedCompact: ac.EnableEnhancedCompact,
	}
}

// BuildSystemPrompt 构建系统提示词
func (ac *AgentConfig) BuildSystemPrompt() (string, error) {
	// 如果指定了自定义提示词，直接使用
	if ac.CustomPrompt != "" {
		return ac.renderCustomPrompt(ac.CustomPrompt)
	}

	// 如果指定了提示词类型，使用对应类型
	if ac.PromptType != "" {
		return ac.buildFromPromptType()
	}

	// 默认使用主 Agent 提示词
	return ac.buildDefaultPrompt()
}

// buildFromPromptType 从提示词类型构建
func (ac *AgentConfig) buildFromPromptType() (string, error) {
	vars := ac.getTemplateVars()

	// 子 Agent 特殊处理
	if ac.IsSubagent && ac.SubagentType != "" {
		return prompts.GetSubagentPrompt(ac.SubagentType, vars)
	}

	// 使用提示词管理器获取
	manager := prompts.NewManager(vars)

	// 如果有选项，使用组合构建
	if ac.hasPromptOptions() {
		return ac.buildWithOptions(manager, vars)
	}

	// 直接获取渲染后的提示词
	return manager.GetRendered(ac.PromptType, vars)
}

// buildDefaultPrompt 构建默认提示词
func (ac *AgentConfig) buildDefaultPrompt() (string, error) {
	vars := ac.getTemplateVars()
	options := ac.PromptOptions

	// 如果没有设置选项，使用默认
	if !ac.hasPromptOptions() {
		options = DefaultPromptOptions()
	}

	// 使用组合器构建
	composer := prompts.NewPromptComposer(prompts.NewRegistry(), vars)

	// 角色定义
	roleDef := options.OverrideRoleDefinition
	if roleDef == "" {
		roleDef = "You are {{.AppName}}, an interactive CLI tool that helps users with software engineering tasks."
	}
	composer.AddRaw(roleDef, 0)

	// 核心指令
	if options.IncludeCoreInstructions {
		composer.AddSection("Core Instructions", prompts.CoreInstructions(), 10)
	}

	// 工具规则
	if options.IncludeToolRules {
		composer.AddSection("Tool Usage Rules", prompts.ToolRules(vars), 20)
	}

	// 文件操作规则
	if options.IncludeFileRules {
		composer.AddSection("File Operation Rules", prompts.FileOperationRules(), 30)
	}

	// 安全规则
	if options.IncludeSecurityRules {
		composer.AddSection("Security Rules", prompts.SecurityRules(), 40)
	}

	// 输出格式规则
	if options.IncludeOutputFormat {
		composer.AddSection("Output Format", prompts.OutputFormatRules(), 50)
	}

	// 额外章节
	for _, section := range options.ExtraSections {
		composer.AddSection(section.Title, section.Content, section.Priority)
	}

	return composer.Compose(), nil
}

// buildWithOptions 使用选项构建提示词
func (ac *AgentConfig) buildWithOptions(manager prompts.Manager, vars *prompts.TemplateVars) (string, error) {
	// 直接使用组合器构建
	composer := prompts.NewSystemPromptBuilder(vars)

	// 角色定义
	roleDef := ac.PromptOptions.OverrideRoleDefinition
	if roleDef == "" {
		roleDef = fmt.Sprintf("You are %s, an interactive CLI tool that helps users with software engineering tasks.", vars.AppName)
	}
	composer.AddRoleDefinition(roleDef)

	// 核心指令
	if ac.PromptOptions.IncludeCoreInstructions {
		composer.AddCoreInstructions(prompts.CoreInstructions())
	}

	// 工具规则
	if ac.PromptOptions.IncludeToolRules {
		composer.AddToolRules(prompts.ToolRules(vars))
	}

	// 文件操作规则
	if ac.PromptOptions.IncludeFileRules {
		composer.AddFileOperationRules(prompts.FileOperationRules())
	}

	// 额外章节
	for _, section := range ac.PromptOptions.ExtraSections {
		composer.AddCustomSection(section.Title, section.Content)
	}

	return composer.Build(), nil
}

// renderCustomPrompt 渲染自定义提示词
func (ac *AgentConfig) renderCustomPrompt(template string) (string, error) {
	vars := ac.getTemplateVars()

	prompt := &prompts.Prompt{
		Template: template,
	}

	return prompt.Render(vars)
}

// getTemplateVars 获取模板变量
func (ac *AgentConfig) getTemplateVars() *prompts.TemplateVars {
	// 使用配置的变量，如果没有则创建默认
	if ac.PromptVars != nil {
		return ac.PromptVars
	}

	vars := prompts.DefaultTemplateVars()

	// 设置基本变量
	if ac.AgentName != "" {
		vars.AgentName = ac.AgentName
	}
	if ac.SubagentType != "" {
		vars.Custom["SubagentType"] = ac.SubagentType
	}
	if ac.ParentAgentName != "" {
		vars.Custom["ParentAgentName"] = ac.ParentAgentName
	}

	return vars
}

// hasPromptOptions 检查是否有提示词选项
func (ac *AgentConfig) hasPromptOptions() bool {
	return ac.PromptOptions.IncludeCoreInstructions ||
		ac.PromptOptions.IncludeToolRules ||
		ac.PromptOptions.IncludeFileRules ||
		ac.PromptOptions.IncludeEducational ||
		ac.PromptOptions.IncludeSecurityRules ||
		ac.PromptOptions.IncludeOutputFormat ||
		len(ac.PromptOptions.ExtraSections) > 0 ||
		ac.PromptOptions.OverrideRoleDefinition != ""
}

// ConfigFromAgentConfig 从扩展配置创建基础配置
func ConfigFromAgentConfig(ac *AgentConfig) (*Config, error) {
	// 构建系统提示词
	systemPrompt, err := ac.BuildSystemPrompt()
	if err != nil {
		return nil, err
	}

	config := ac.ToConfig()
	config.SystemPrompt = systemPrompt

	return config, nil
}

// NewAgentWithConfig 使用扩展配置创建 Agent
func NewAgentWithConfig(ac *AgentConfig) (*Agent, error) {
	config, err := ConfigFromAgentConfig(ac)
	if err != nil {
		return nil, err
	}

	// TODO: 需要传入 provider
	return NewAgent(config, nil)
}
