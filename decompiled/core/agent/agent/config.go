// Package agent Agent config and prompt integration
//
// This file extends Agent config with system prompt management
// Reference official architecture: code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent
package agent

import (
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/prompts"
)

// AgentConfig extended Agent config
type AgentConfig struct {
	// base config (inherited from original Config)
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

	// prompt config
	PromptType    prompts.PromptType    // use built-in prompt type
	PromptVars    *prompts.TemplateVars // prompt template variables
	CustomPrompt  string                // custom prompt (overrides built-in)
	PromptOptions PromptOptions         // prompt options

	// subagent config
	IsSubagent      bool
	SubagentType    string
	ParentAgentName string
}

// PromptOptions prompt options
type PromptOptions struct {
	// prompt component toggles
	IncludeCoreInstructions bool
	IncludeToolRules        bool
	IncludeFileRules        bool
	IncludeEducational      bool
	IncludeSecurityRules    bool
	IncludeOutputFormat     bool

	// custom content
	ExtraSections []ExtraPromptSection

	// override options
	OverrideRoleDefinition string
}

// ExtraPromptSection extra prompt section
type ExtraPromptSection struct {
	Title    string
	Content  string
	Priority int
}

// DefaultPromptOptions default prompt options
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

// ToConfig converts to base Config
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

// BuildSystemPrompt builds system prompt
func (ac *AgentConfig) BuildSystemPrompt() (string, error) {
	// if custom prompt specified, use directly
	if ac.CustomPrompt != "" {
		return ac.renderCustomPrompt(ac.CustomPrompt)
	}

	// if prompt type specified, use corresponding type
	if ac.PromptType != "" {
		return ac.buildFromPromptType()
	}

	// default to main Agent prompt
	return ac.buildDefaultPrompt()
}

// buildFromPromptType builds from prompt type
func (ac *AgentConfig) buildFromPromptType() (string, error) {
	vars := ac.getTemplateVars()

	// subagent special handling
	if ac.IsSubagent && ac.SubagentType != "" {
		return prompts.GetSubagentPrompt(ac.SubagentType, vars)
	}

	// get via prompt manager
	manager := prompts.NewManager(vars)

	// if options present, build with composition
	if ac.hasPromptOptions() {
		return ac.buildWithOptions(manager, vars)
	}

	// directly get rendered prompt
	return manager.GetRendered(ac.PromptType, vars)
}

// buildDefaultPrompt builds default prompt
func (ac *AgentConfig) buildDefaultPrompt() (string, error) {
	vars := ac.getTemplateVars()
	options := ac.PromptOptions

	// if no options set, use default
	if !ac.hasPromptOptions() {
		options = DefaultPromptOptions()
	}

	// build with composer
	composer := prompts.NewPromptComposer(prompts.NewRegistry(), vars)

	// role definition
	roleDef := options.OverrideRoleDefinition
	if roleDef == "" {
		roleDef = "You are {{.AppName}}, an interactive CLI tool that helps users with software engineering tasks."
	}
	composer.AddRaw(roleDef, 0)

	// core instructions
	if options.IncludeCoreInstructions {
		composer.AddSection("Core Instructions", prompts.CoreInstructions(), 10)
	}

	// tool rules
	if options.IncludeToolRules {
		composer.AddSection("Tool Usage Rules", prompts.ToolRules(vars), 20)
	}

	// file operation rules
	if options.IncludeFileRules {
		composer.AddSection("File Operation Rules", prompts.FileOperationRules(), 30)
	}

	// security rules
	if options.IncludeSecurityRules {
		composer.AddSection("Security Rules", prompts.SecurityRules(), 40)
	}

	// output format rules
	if options.IncludeOutputFormat {
		composer.AddSection("Output Format", prompts.OutputFormatRules(), 50)
	}

	// extra sections
	for _, section := range options.ExtraSections {
		composer.AddSection(section.Title, section.Content, section.Priority)
	}

	return composer.Compose(), nil
}

// buildWithOptions builds prompt with options
func (ac *AgentConfig) buildWithOptions(manager prompts.Manager, vars *prompts.TemplateVars) (string, error) {
	// build directly with composer
	composer := prompts.NewSystemPromptBuilder(vars)

	// role definition
	roleDef := ac.PromptOptions.OverrideRoleDefinition
	if roleDef == "" {
		roleDef = fmt.Sprintf("You are %s, an interactive CLI tool that helps users with software engineering tasks.", vars.AppName)
	}
	composer.AddRoleDefinition(roleDef)

	// core instructions
	if ac.PromptOptions.IncludeCoreInstructions {
		composer.AddCoreInstructions(prompts.CoreInstructions())
	}

	// tool rules
	if ac.PromptOptions.IncludeToolRules {
		composer.AddToolRules(prompts.ToolRules(vars))
	}

	// file operation rules
	if ac.PromptOptions.IncludeFileRules {
		composer.AddFileOperationRules(prompts.FileOperationRules())
	}

	// extra sections
	for _, section := range ac.PromptOptions.ExtraSections {
		composer.AddCustomSection(section.Title, section.Content)
	}

	return composer.Build(), nil
}

// renderCustomPrompt renders custom prompt
func (ac *AgentConfig) renderCustomPrompt(template string) (string, error) {
	vars := ac.getTemplateVars()

	prompt := &prompts.Prompt{
		Template: template,
	}

	return prompt.Render(vars)
}

// getTemplateVars gets template variables
func (ac *AgentConfig) getTemplateVars() *prompts.TemplateVars {
	// use configured vars, or create default if none
	if ac.PromptVars != nil {
		return ac.PromptVars
	}

	vars := prompts.DefaultTemplateVars()

	// set basic variables
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

// hasPromptOptions checks if prompt options present
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

// ConfigFromAgentConfig creates base config from extended config
func ConfigFromAgentConfig(ac *AgentConfig) (*Config, error) {
	// build system prompt
	systemPrompt, err := ac.BuildSystemPrompt()
	if err != nil {
		return nil, err
	}

	config := ac.ToConfig()
	config.SystemPrompt = systemPrompt

	return config, nil
}

// NewAgentWithConfig creates Agent with extended config
func NewAgentWithConfig(ac *AgentConfig) (*Agent, error) {
	config, err := ConfigFromAgentConfig(ac)
	if err != nil {
		return nil, err
	}

	// TODO: need to pass provider
	return NewAgent(config, nil)
}
