// Package prompts System prompt management
// Reference official architecture: code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent
//
// Prompt-related paths discovered in the official binary:
// - code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent.(*agentContext).GetSystemPrompt
// - code.alibaba-inc.com/qoder-core/qodercli-reverse/core/resource/command.(*Command).RenderPrompt
// - code.alibaba-inc.com/qoder-core/qodercli-reverse/core/resource/command.loadBuiltinPromptsCommands
// - code.alibaba-inc.com/qoder-core/qodercli-reverse/tui/texts.(*Service).GetText
//
// This package implements unified management of system prompts, supporting:
// 1. Built-in prompts (constant definitions)
// 2. Loading external prompt files
// 3. Template variable substitution
// 4. Role definition composition
package prompts

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TemplateVars Prompt template variables
type TemplateVars struct {
	// Basic info
	AppName     string // Application name (e.g. "qodercli-reverse")
	BrandName   string // Brand name (e.g. "Qoder")
	ProductName string // Product name
	Version     string // Version number

	// Agent config
	RoleDefinition string // Role definition
	AgentName      string // Agent name

	// Tool names (configurable)
	ReadToolName         string
	BashToolName         string
	BashOutputToolName   string
	ImageGenToolName     string
	WebSearchToolName    string
	SearchCodebaseTool   string
	SearchSymbolTool     string
	BrowserUseToolPrefix string

	// Output style
	OutputStyleName   string
	OutputStylePrompt string

	// Timeout config
	MaxTimeoutMs      int
	MaxTimeoutMin     int
	DefaultTimeoutMs  int
	DefaultTimeoutMin int

	// Dynamic content
	PlanFilePath    string // Plan file path
	ContextDirs     string // Context directories
	Language        string // User language
	PreviousSummary string // Session summary

	// Custom variables
	Custom map[string]string
}

// DefaultTemplateVars Returns default template variables
func DefaultTemplateVars() *TemplateVars {
	return &TemplateVars{
		AppName:              "qodercli-reverse",
		BrandName:            "Qoder",
		ProductName:          "Qoder CLI",
		Version:              "0.1.29",
		ReadToolName:         "Read",
		BashToolName:         "Bash",
		BashOutputToolName:   "BashOutput",
		ImageGenToolName:     "ImageGen",
		WebSearchToolName:    "WebSearch",
		SearchCodebaseTool:   "SearchCodebase",
		SearchSymbolTool:     "SearchSymbol",
		BrowserUseToolPrefix: "browser",
		MaxTimeoutMs:         600000,
		MaxTimeoutMin:        10,
		DefaultTimeoutMs:     300000,
		DefaultTimeoutMin:    5,
		Custom:               make(map[string]string),
	}
}

// Merge Merges another TemplateVars
func (v *TemplateVars) Merge(other *TemplateVars) *TemplateVars {
	if other.AppName != "" {
		v.AppName = other.AppName
	}
	if other.BrandName != "" {
		v.BrandName = other.BrandName
	}
	if other.ProductName != "" {
		v.ProductName = other.ProductName
	}
	if other.Version != "" {
		v.Version = other.Version
	}
	if other.RoleDefinition != "" {
		v.RoleDefinition = other.RoleDefinition
	}
	if other.AgentName != "" {
		v.AgentName = other.AgentName
	}
	if other.Language != "" {
		v.Language = other.Language
	}
	for k, val := range other.Custom {
		v.Custom[k] = val
	}
	return v
}

// Get Retrieves a variable value, preferring Custom
func (v *TemplateVars) Get(key string) string {
	if val, ok := v.Custom[key]; ok {
		return val
	}
	switch key {
	case "AppName":
		return v.AppName
	case "BrandName":
		return v.BrandName
	case "ProductName":
		return v.ProductName
	case "Version":
		return v.Version
	case "RoleDefinition":
		return v.RoleDefinition
	case "AgentName":
		return v.AgentName
	case "ReadToolName":
		return v.ReadToolName
	case "BashToolName":
		return v.BashToolName
	case "BashOutputToolName":
		return v.BashOutputToolName
	case "ImageGenToolName":
		return v.ImageGenToolName
	case "WebSearchToolName":
		return v.WebSearchToolName
	case "SearchCodebaseTool":
		return v.SearchCodebaseTool
	case "SearchSymbolTool":
		return v.SearchSymbolTool
	case "BrowserUseToolPrefix":
		return v.BrowserUseToolPrefix
	case "OutputStyleName":
		return v.OutputStyleName
	case "OutputStylePrompt":
		return v.OutputStylePrompt
	case "PlanFilePath":
		return v.PlanFilePath
	case "ContextDirs":
		return v.ContextDirs
	case "Language":
		return v.Language
	case "PreviousSummary":
		return v.PreviousSummary
	default:
		return ""
	}
}

// PromptType Prompt type
type PromptType string

const (
	// Main Agent prompts
	PromptTypeMainAgent            PromptType = "main_agent"
	PromptTypeMainAgentEducational PromptType = "main_agent_educational"
	PromptTypeMainAgentPractice    PromptType = "main_agent_practice"
	PromptTypeGenericTask          PromptType = "generic_task"

	// Sub-agent prompts
	PromptTypeBrowserSubagent      PromptType = "browser_subagent"
	PromptTypeCodeImplement        PromptType = "code_implement"
	PromptTypeTaskExecutor         PromptType = "task_executor"
	PromptTypeDesignAgent          PromptType = "design_agent"
	PromptTypeSystemDesign         PromptType = "system_design"
	PromptTypeSoftwareArchitect    PromptType = "software_architect"
	PromptTypeDesignReview         PromptType = "design_review"
	PromptTypeRequirements         PromptType = "requirements_analysis"
	PromptTypeTestAutomation       PromptType = "test_automation"
	PromptTypeCodeReviewer         PromptType = "code_reviewer"
	PromptTypeDebugger             PromptType = "debugger"
	PromptTypeFileSearch           PromptType = "file_search"
	PromptTypeWorkflowOrchestrator PromptType = "workflow_orchestrator"
	PromptTypeBehaviorAnalyzer     PromptType = "behavior_analyzer"
	PromptTypeSkepticalValidator   PromptType = "skeptical_validator"
	PromptTypeSecurityAuditor      PromptType = "security_auditor"
	PromptTypeDataScientist        PromptType = "data_scientist"
	PromptTypeGuideAgent           PromptType = "guide_agent"
	PromptTypeQuestHandler         PromptType = "quest_handler"
	PromptTypeSpecHLDDesigner      PromptType = "spec_hld_designer"
	PromptTypeSpecLLDDesigner      PromptType = "spec_lld_designer"
	PromptTypeSpecImplementer      PromptType = "spec_implementer"
	PromptTypeSpecLeader           PromptType = "spec_leader"
	PromptTypeExploreAgent         PromptType = "explore_agent"

	// IDE integration prompts
	PromptTypeQoderWork    PromptType = "qoderwork"
	PromptTypeQoderStudio  PromptType = "qoder_studio"
	PromptTypeQoderDesktop PromptType = "qoder_desktop"

	// Specialized prompts
	PromptTypeConversationSummary PromptType = "conversation_summary"
	PromptTypeAgentArchitect      PromptType = "agent_architect"
	PromptTypeCommandArchitect    PromptType = "command_architect"
	PromptTypeUnitTestExpert      PromptType = "unit_test_expert"
	PromptTypeCoordinator         PromptType = "coordinator"
	PromptTypePlanModeReturn      PromptType = "plan_mode_return"
)

// Prompt Prompt definition
type Prompt struct {
	Type        PromptType
	Name        string
	Description string
	Template    string
	IsBuiltIn   bool
	Vars        []string // Required template variables
}

// Render Renders the prompt template
func (p *Prompt) Render(vars *TemplateVars) (string, error) {
	if vars == nil {
		vars = DefaultTemplateVars()
	}

	tmpl, err := template.New(string(p.Type)).
		Funcs(template.FuncMap{
			"join":  strings.Join,
			"upper": strings.ToUpper,
			"lower": strings.ToLower,
			"title": strings.Title,
		}).
		Parse(p.Template)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderWithVars Renders the prompt using a map
func (p *Prompt) RenderWithVars(vars map[string]string) (string, error) {
	tv := DefaultTemplateVars()
	for k, v := range vars {
		tv.Custom[k] = v
	}
	return p.Render(tv)
}

// SystemPromptBuilder System prompt builder
type SystemPromptBuilder struct {
	parts []string
	vars  *TemplateVars
}

// NewSystemPromptBuilder Creates a new builder
func NewSystemPromptBuilder(vars *TemplateVars) *SystemPromptBuilder {
	if vars == nil {
		vars = DefaultTemplateVars()
	}
	return &SystemPromptBuilder{
		parts: make([]string, 0),
		vars:  vars,
	}
}

// AddRoleDefinition Adds a role definition
func (b *SystemPromptBuilder) AddRoleDefinition(role string) *SystemPromptBuilder {
	if role != "" {
		b.parts = append(b.parts, role)
	}
	return b
}

// AddCoreInstructions Adds core instructions
func (b *SystemPromptBuilder) AddCoreInstructions(instructions string) *SystemPromptBuilder {
	if instructions != "" {
		b.parts = append(b.parts, instructions)
	}
	return b
}

// AddToolRules Adds tool usage rules
func (b *SystemPromptBuilder) AddToolRules(rules string) *SystemPromptBuilder {
	if rules != "" {
		b.parts = append(b.parts, "## Tool Usage Rules", rules)
	}
	return b
}

// AddFileOperationRules Adds file operation rules
func (b *SystemPromptBuilder) AddFileOperationRules(rules string) *SystemPromptBuilder {
	if rules != "" {
		b.parts = append(b.parts, "## File Operation Rules", rules)
	}
	return b
}

// AddOutputFormatRules Adds output format rules
func (b *SystemPromptBuilder) AddOutputFormatRules(rules string) *SystemPromptBuilder {
	if rules != "" {
		b.parts = append(b.parts, "## Output Format Rules", rules)
	}
	return b
}

// AddCustomSection Adds a custom section
func (b *SystemPromptBuilder) AddCustomSection(title, content string) *SystemPromptBuilder {
	if content != "" {
		b.parts = append(b.parts, fmt.Sprintf("## %s", title), content)
	}
	return b
}

// Build Builds the final prompt
func (b *SystemPromptBuilder) Build() string {
	return strings.Join(b.parts, "\n\n")
}

// BuildWithTemplate Builds using a template
func (b *SystemPromptBuilder) BuildWithTemplate(tmpl string) (string, error) {
	prompt := &Prompt{
		Template: tmpl,
	}
	return prompt.Render(b.vars)
}

// Manager Prompt manager interface
type Manager interface {
	// Get Retrieves the prompt of the specified type
	Get(promptType PromptType) (*Prompt, error)

	// GetRendered Retrieves the rendered prompt
	GetRendered(promptType PromptType, vars *TemplateVars) (string, error)

	// Register Registers a custom prompt
	Register(prompt *Prompt) error

	// RegisterFromFile Loads and registers a prompt from a file
	RegisterFromFile(path string) error

	// List Lists all available prompt types
	List() []PromptType

	// GetMainAgentPrompt Retrieves the main Agent prompt (common shortcut method)
	GetMainAgentPrompt(vars *TemplateVars) (string, error)

	// GetSubagentPrompt Retrieves a sub-agent prompt
	GetSubagentPrompt(subagentType string, vars *TemplateVars) (string, error)
}
