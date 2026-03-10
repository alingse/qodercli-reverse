// Package prompts 系统提示词管理
// 参考官方架构: code.alibaba-inc.com/qoder-core/qodercli/core/agent
//
// 官方二进制中发现的提示词相关路径:
// - code.alibaba-inc.com/qoder-core/qodercli/core/agent.(*agentContext).GetSystemPrompt
// - code.alibaba-inc.com/qoder-core/qodercli/core/resource/command.(*Command).RenderPrompt
// - code.alibaba-inc.com/qoder-core/qodercli/core/resource/command.loadBuiltinPromptsCommands
// - code.alibaba-inc.com/qoder-core/qodercli/tui/texts.(*Service).GetText
//
// 本包实现了系统提示词的统一管理，支持:
// 1. 内置提示词（常量定义）
// 2. 外部提示词文件加载
// 3. 模板变量替换
// 4. 角色定义组合
package prompts

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TemplateVars 提示词模板变量
type TemplateVars struct {
	// 基础信息
	AppName     string // 应用名称 (如 "qodercli")
	BrandName   string // 品牌名称 (如 "Qoder")
	ProductName string // 产品名称
	Version     string // 版本号

	// Agent 配置
	RoleDefinition string // 角色定义
	AgentName      string // Agent 名称

	// 工具名称（可配置）
	ReadToolName         string
	BashToolName         string
	BashOutputToolName   string
	ImageGenToolName     string
	WebSearchToolName    string
	SearchCodebaseTool   string
	SearchSymbolTool     string
	BrowserUseToolPrefix string

	// 输出样式
	OutputStyleName   string
	OutputStylePrompt string

	// 超时配置
	MaxTimeoutMs      int
	MaxTimeoutMin     int
	DefaultTimeoutMs  int
	DefaultTimeoutMin int

	// 动态内容
	PlanFilePath     string // Plan 文件路径
	ContextDirs      string // 上下文目录
	Language         string // 用户语言
	PreviousSummary  string // 会话摘要

	// 自定义变量
	Custom map[string]string
}

// DefaultTemplateVars 返回默认模板变量
func DefaultTemplateVars() *TemplateVars {
	return &TemplateVars{
		AppName:              "qodercli",
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

// Merge 合并另一个 TemplateVars
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

// Get 获取变量值，优先从 Custom 获取
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

// PromptType 提示词类型
type PromptType string

const (
	// 主 Agent 提示词
	PromptTypeMainAgent        PromptType = "main_agent"
	PromptTypeMainAgentEducational PromptType = "main_agent_educational"
	PromptTypeMainAgentPractice    PromptType = "main_agent_practice"
	PromptTypeGenericTask      PromptType = "generic_task"

	// 子 Agent 提示词
	PromptTypeBrowserSubagent  PromptType = "browser_subagent"
	PromptTypeCodeImplement    PromptType = "code_implement"
	PromptTypeTaskExecutor     PromptType = "task_executor"
	PromptTypeDesignAgent      PromptType = "design_agent"
	PromptTypeSystemDesign     PromptType = "system_design"
	PromptTypeSoftwareArchitect PromptType = "software_architect"
	PromptTypeDesignReview     PromptType = "design_review"
	PromptTypeRequirements     PromptType = "requirements_analysis"
	PromptTypeTestAutomation   PromptType = "test_automation"
	PromptTypeCodeReviewer     PromptType = "code_reviewer"
	PromptTypeDebugger         PromptType = "debugger"
	PromptTypeFileSearch       PromptType = "file_search"
	PromptTypeWorkflowOrchestrator PromptType = "workflow_orchestrator"
	PromptTypeBehaviorAnalyzer PromptType = "behavior_analyzer"
	PromptTypeSkepticalValidator PromptType = "skeptical_validator"
	PromptTypeSecurityAuditor  PromptType = "security_auditor"
	PromptTypeDataScientist    PromptType = "data_scientist"
	PromptTypeGuideAgent       PromptType = "guide_agent"
	PromptTypeQuestHandler     PromptType = "quest_handler"

	// IDE 集成提示词
	PromptTypeQoderWork        PromptType = "qoderwork"
	PromptTypeQoderStudio      PromptType = "qoder_studio"
	PromptTypeQoderDesktop     PromptType = "qoder_desktop"

	// 专项提示词
	PromptTypeConversationSummary PromptType = "conversation_summary"
	PromptTypeAgentArchitect   PromptType = "agent_architect"
	PromptTypeCommandArchitect PromptType = "command_architect"
	PromptTypeUnitTestExpert   PromptType = "unit_test_expert"
	PromptTypeCoordinator      PromptType = "coordinator"
	PromptTypePlanModeReturn   PromptType = "plan_mode_return"
)

// Prompt 提示词定义
type Prompt struct {
	Type        PromptType
	Name        string
	Description string
	Template    string
	IsBuiltIn   bool
	Vars        []string // 所需的模板变量
}

// Render 渲染提示词模板
func (p *Prompt) Render(vars *TemplateVars) (string, error) {
	if vars == nil {
		vars = DefaultTemplateVars()
	}

	tmpl, err := template.New(string(p.Type)).
		Funcs(template.FuncMap{
			"join": strings.Join,
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

// RenderWithVars 使用 map 渲染提示词
func (p *Prompt) RenderWithVars(vars map[string]string) (string, error) {
	tv := DefaultTemplateVars()
	for k, v := range vars {
		tv.Custom[k] = v
	}
	return p.Render(tv)
}

// SystemPromptBuilder 系统提示词构建器
type SystemPromptBuilder struct {
	parts []string
	vars  *TemplateVars
}

// NewSystemPromptBuilder 创建新的构建器
func NewSystemPromptBuilder(vars *TemplateVars) *SystemPromptBuilder {
	if vars == nil {
		vars = DefaultTemplateVars()
	}
	return &SystemPromptBuilder{
		parts: make([]string, 0),
		vars:  vars,
	}
}

// AddRoleDefinition 添加角色定义
func (b *SystemPromptBuilder) AddRoleDefinition(role string) *SystemPromptBuilder {
	if role != "" {
		b.parts = append(b.parts, role)
	}
	return b
}

// AddCoreInstructions 添加核心指令
func (b *SystemPromptBuilder) AddCoreInstructions(instructions string) *SystemPromptBuilder {
	if instructions != "" {
		b.parts = append(b.parts, instructions)
	}
	return b
}

// AddToolRules 添加工具使用规则
func (b *SystemPromptBuilder) AddToolRules(rules string) *SystemPromptBuilder {
	if rules != "" {
		b.parts = append(b.parts, "## Tool Usage Rules", rules)
	}
	return b
}

// AddFileOperationRules 添加文件操作规则
func (b *SystemPromptBuilder) AddFileOperationRules(rules string) *SystemPromptBuilder {
	if rules != "" {
		b.parts = append(b.parts, "## File Operation Rules", rules)
	}
	return b
}

// AddOutputFormatRules 添加输出格式规则
func (b *SystemPromptBuilder) AddOutputFormatRules(rules string) *SystemPromptBuilder {
	if rules != "" {
		b.parts = append(b.parts, "## Output Format Rules", rules)
	}
	return b
}

// AddCustomSection 添加自定义章节
func (b *SystemPromptBuilder) AddCustomSection(title, content string) *SystemPromptBuilder {
	if content != "" {
		b.parts = append(b.parts, fmt.Sprintf("## %s", title), content)
	}
	return b
}

// Build 构建最终提示词
func (b *SystemPromptBuilder) Build() string {
	return strings.Join(b.parts, "\n\n")
}

// BuildWithTemplate 使用模板构建
func (b *SystemPromptBuilder) BuildWithTemplate(tmpl string) (string, error) {
	prompt := &Prompt{
		Template: tmpl,
	}
	return prompt.Render(b.vars)
}

// Manager 提示词管理器接口
type Manager interface {
	// Get 获取指定类型的提示词
	Get(promptType PromptType) (*Prompt, error)

	// GetRendered 获取渲染后的提示词
	GetRendered(promptType PromptType, vars *TemplateVars) (string, error)

	// Register 注册自定义提示词
	Register(prompt *Prompt) error

	// RegisterFromFile 从文件加载并注册提示词
	RegisterFromFile(path string) error

	// List 列出所有可用的提示词类型
	List() []PromptType

	// GetMainAgentPrompt 获取主 Agent 提示词（常用快捷方法）
	GetMainAgentPrompt(vars *TemplateVars) (string, error)

	// GetSubagentPrompt 获取子 Agent 提示词
	GetSubagentPrompt(subagentType string, vars *TemplateVars) (string, error)
}
