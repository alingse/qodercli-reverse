// Package prompts System Prompt Builder full implementation
//
// Reference official architecture: code.alibaba-inc.com/qoder-core/qodercli-reverse/acp.(*QoderAcpAgent).buildPrompt
// Features include:
// - Basic role definition
// - Tool usage guide
// - Permission rules description
// - Environment info (OS, Git status)
// - Project-specific instructions (AGENTS.md, .claude/)
// - Coding standards and best practices
package prompts

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// SystemPromptBuilderV2 Enhanced system prompt builder
type SystemPromptBuilderV2 struct {
	// Basic config
	vars *TemplateVars

	// Component toggles
	enableRoleDefinition  bool
	enableToolGuide       bool
	enablePermissionRules bool
	enableEnvironmentInfo bool
	enableProjectContext  bool
	enableCodingStandards bool
	enableSystemReminder  bool
	enableSessionContext  bool

	// Collectors
	envCollector   *EnvironmentCollector
	projectLoader  *ProjectContextLoader
	sessionContext *SessionContext

	// Custom content
	customSections []Section
	prebuiltParts  map[string]string
}

// Section Prompt section
type Section struct {
	Title     string
	Content   string
	Priority  int
	Condition func() bool
}

// EnvironmentInfo Environment info
type EnvironmentInfo struct {
	// System info
	OS           string
	Architecture string
	Shell        string
	HomeDir      string
	WorkingDir   string
	TempDir      string

	// Git info
	GitRepo   bool
	GitBranch string
	GitCommit string
	GitRemote string
	GitStatus string

	// Dev environment
	GoVersion     string
	NodeVersion   string
	PythonVersion string
	JavaVersion   string
	RustVersion   string

	// Editor/IDE
	Editor        string
	EditorVersion string

	// Time
	Timezone    string
	CurrentTime string
}

// ProjectContext Project context
type ProjectContext struct {
	// Basic project info
	Name     string
	RootPath string
	Type     string // go, node, python, rust, etc.
	Language string

	// Config files
	HasGoMod       bool
	HasPackageJSON bool
	HasPyProject   bool
	HasCargoToml   bool
	HasPomXML      bool

	// Project-specific instructions
	AgentsMDContent  string
	ClaudeDirContent string
	ReadmeContent    string

	// Coding standards
	CodingStandards string
	StyleGuide      string

	// Ignore rules
	GitignorePatterns []string
}

// SessionContext Session context
type SessionContext struct {
	SessionID       string
	StartTime       time.Time
	PreviousSummary string
	CurrentTask     string
	TodoList        []string
	ContextDirs     []string
}

// NewSystemPromptBuilderV2 Creates a new builder
func NewSystemPromptBuilderV2(vars *TemplateVars) *SystemPromptBuilderV2 {
	if vars == nil {
		vars = DefaultTemplateVars()
	}

	return &SystemPromptBuilderV2{
		vars:                  vars,
		prebuiltParts:         make(map[string]string),
		customSections:        make([]Section, 0),
		enableRoleDefinition:  true,
		enableToolGuide:       true,
		enablePermissionRules: true,
		enableEnvironmentInfo: true,
		enableProjectContext:  true,
		enableCodingStandards: true,
		enableSystemReminder:  true,
		enableSessionContext:  true,
	}
}

// ==================== Config methods ====================

// BuilderOption Builder option function
type BuilderOption func(*SystemPromptBuilderV2)

// WithOptions Applies builder options in batch
func (b *SystemPromptBuilderV2) WithOptions(opts ...BuilderOption) *SystemPromptBuilderV2 {
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Feature toggle constants
const (
	FeatureRoleDefinition  = 1 << iota // Role definition
	FeatureToolGuide                   // Tool guide
	FeaturePermissionRules             // Permission rules
	FeatureEnvironmentInfo             // Environment info
	FeatureProjectContext              // Project context
	FeatureCodingStandards             // Coding standards
	FeatureSessionContext              // Session context
)

// WithFeatures Enables/disables features via bitmask
//
//	Enable features: WithFeatures(FeatureRoleDefinition | FeatureToolGuide)
//	Disable features: WithFeatures(0) - disable all
func WithFeatures(mask int) BuilderOption {
	return func(b *SystemPromptBuilderV2) {
		b.enableRoleDefinition = (mask & FeatureRoleDefinition) != 0
		b.enableToolGuide = (mask & FeatureToolGuide) != 0
		b.enablePermissionRules = (mask & FeaturePermissionRules) != 0
		b.enableEnvironmentInfo = (mask & FeatureEnvironmentInfo) != 0
		b.enableProjectContext = (mask & FeatureProjectContext) != 0
		b.enableCodingStandards = (mask & FeatureCodingStandards) != 0
		b.enableSessionContext = (mask & FeatureSessionContext) != 0
	}
}

// WithAllFeatures Enables all features
func WithAllFeatures() BuilderOption {
	return WithFeatures(0xFF)
}

// WithMinimalFeatures Enables only core features (role, tools, permissions)
func WithMinimalFeatures() BuilderOption {
	return WithFeatures(FeatureRoleDefinition | FeatureToolGuide | FeaturePermissionRules)
}

// WithRoleDefinition Enables/disables role definition
func (b *SystemPromptBuilderV2) WithRoleDefinition(enable bool) *SystemPromptBuilderV2 {
	b.enableRoleDefinition = enable
	return b
}

// WithToolGuide Enables/disables tool guide
func (b *SystemPromptBuilderV2) WithToolGuide(enable bool) *SystemPromptBuilderV2 {
	b.enableToolGuide = enable
	return b
}

// WithPermissionRules Enables/disables permission rules
func (b *SystemPromptBuilderV2) WithPermissionRules(enable bool) *SystemPromptBuilderV2 {
	b.enablePermissionRules = enable
	return b
}

// WithEnvironmentInfo Enables/disables environment info
func (b *SystemPromptBuilderV2) WithEnvironmentInfo(enable bool) *SystemPromptBuilderV2 {
	b.enableEnvironmentInfo = enable
	return b
}

// WithProjectContext Enables/disables project context
func (b *SystemPromptBuilderV2) WithProjectContext(enable bool) *SystemPromptBuilderV2 {
	b.enableProjectContext = enable
	return b
}

// WithCodingStandards Enables/disables coding standards
func (b *SystemPromptBuilderV2) WithCodingStandards(enable bool) *SystemPromptBuilderV2 {
	b.enableCodingStandards = enable
	return b
}

// WithSessionContext Enables/disables session context
func (b *SystemPromptBuilderV2) WithSessionContext(enable bool) *SystemPromptBuilderV2 {
	b.enableSessionContext = enable
	return b
}

// WithCustomRole Sets a custom role definition
func (b *SystemPromptBuilderV2) WithCustomRole(role string) *SystemPromptBuilderV2 {
	b.prebuiltParts["role"] = role
	return b
}

// WithCustomToolGuide Sets a custom tool guide
func (b *SystemPromptBuilderV2) WithCustomToolGuide(guide string) *SystemPromptBuilderV2 {
	b.prebuiltParts["tool_guide"] = guide
	return b
}

// AddCustomSection Adds a custom section
func (b *SystemPromptBuilderV2) AddCustomSection(title, content string, priority int) *SystemPromptBuilderV2 {
	b.customSections = append(b.customSections, Section{
		Title:    title,
		Content:  content,
		Priority: priority,
	})
	return b
}

// AddConditionalSection Adds a conditional section
func (b *SystemPromptBuilderV2) AddConditionalSection(title, content string, priority int, condition func() bool) *SystemPromptBuilderV2 {
	b.customSections = append(b.customSections, Section{
		Title:     title,
		Content:   content,
		Priority:  priority,
		Condition: condition,
	})
	return b
}

// ==================== Data collection ====================

// CollectEnvironment Collects environment info
func (b *SystemPromptBuilderV2) CollectEnvironment() (*EnvironmentInfo, error) {
	if b.envCollector == nil {
		b.envCollector = NewEnvironmentCollector()
	}

	info, err := b.envCollector.Collect()
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CollectProjectContext Collects project context
func (b *SystemPromptBuilderV2) CollectProjectContext(workDir string) (*ProjectContext, error) {
	if b.projectLoader == nil {
		b.projectLoader = NewProjectContextLoader(workDir)
	}

	ctx, err := b.projectLoader.Load()
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

// SetSessionContext Sets the session context
func (b *SystemPromptBuilderV2) SetSessionContext(ctx *SessionContext) *SystemPromptBuilderV2 {
	b.sessionContext = ctx
	return b
}

// ==================== Build methods ====================

// Build Builds the complete system prompt
func (b *SystemPromptBuilderV2) Build() (string, error) {
	var sections []Section

	// 1. Role definition (P0)
	if b.enableRoleDefinition {
		section := b.buildRoleSection()
		if section != nil {
			sections = append(sections, *section)
		}
	}

	// 2. Core instructions (P0)
	section := b.buildCoreInstructionsSection()
	if section != nil {
		sections = append(sections, *section)
	}

	// 3. Tool usage guide (P0)
	if b.enableToolGuide {
		section := b.buildToolGuideSection()
		if section != nil {
			sections = append(sections, *section)
		}
	}

	// 4. Permission rules (P0)
	if b.enablePermissionRules {
		section := b.buildPermissionSection()
		if section != nil {
			sections = append(sections, *section)
		}
	}

	// 4.5 File operation rules (P0)
	section = b.buildFileOperationSection()
	if section != nil {
		sections = append(sections, *section)
	}

	// 5. Environment info (P1)
	if b.enableEnvironmentInfo {
		section := b.buildEnvironmentSection()
		if section != nil {
			sections = append(sections, *section)
		}
	}

	// 6. Project context (P1)
	if b.enableProjectContext {
		section := b.buildProjectSection()
		if section != nil {
			sections = append(sections, *section)
		}
	}

	// 7. Coding standards (P1)
	if b.enableCodingStandards {
		section := b.buildCodingStandardsSection()
		if section != nil {
			sections = append(sections, *section)
		}
	}

	// 8. Session context (P2)
	if b.enableSessionContext && b.sessionContext != nil {
		section := b.buildSessionSection()
		if section != nil {
			sections = append(sections, *section)
		}
	}

	// 9. Custom sections
	for _, section := range b.customSections {
		if section.Condition == nil || section.Condition() {
			sections = append(sections, section)
		}
	}

	// Sort by priority
	sections = sortSectionsByPriority(sections)

	// Render all sections
	return b.renderSections(sections)
}

// BuildMinimal Builds a minimal prompt (core components only)
func (b *SystemPromptBuilderV2) BuildMinimal() (string, error) {
	return b.WithOptions(WithMinimalFeatures()).Build()
}

// BuildForSubagent Builds a prompt for a sub-agent
func (b *SystemPromptBuilderV2) BuildForSubagent(subagentType string, parentContext string) (string, error) {
	// Sub-agents use a simplified version
	b.enableEnvironmentInfo = false
	b.enableSessionContext = false

	// Add parent Agent context
	if parentContext != "" {
		b.AddCustomSection("Parent Context", parentContext, 50)
	}

	// Add sub-agent specific instructions
	subagentPrompt, err := GetSubagentPrompt(subagentType, b.vars)
	if err == nil {
		b.WithCustomRole(subagentPrompt)
	}

	return b.Build()
}

// ==================== Section build methods ====================

func (b *SystemPromptBuilderV2) buildRoleSection() *Section {
	role := b.prebuiltParts["role"]
	if role == "" {
		role = b.vars.RoleDefinition
	}
	if role == "" {
		role = fmt.Sprintf("You are %s, an interactive CLI tool that helps users with software engineering tasks.", b.vars.AppName)
	}

	return &Section{
		Title:    "Role",
		Content:  role,
		Priority: 0,
	}
}

func (b *SystemPromptBuilderV2) buildCoreInstructionsSection() *Section {
	return &Section{
		Title:    "Core Instructions",
		Content:  coreInstructions,
		Priority: 10,
	}
}

func (b *SystemPromptBuilderV2) buildToolGuideSection() *Section {
	guide := b.prebuiltParts["tool_guide"]
	if guide == "" {
		guide = toolRules
	}

	return &Section{
		Title:    "Tool Usage Guide",
		Content:  guide,
		Priority: 20,
	}
}

func (b *SystemPromptBuilderV2) buildPermissionSection() *Section {
	return &Section{
		Title: "Permission Rules",
		Content: `Permission Rules:
- ALWAYS ask for permission before:
  * Modifying files outside the current working directory
  * Executing commands that modify system state
  * Accessing sensitive files (credentials, keys, etc.)
  * Installing new software or packages
  * Making network requests to external services

- You may proceed WITHOUT asking when:
  * Reading files within the working directory
  * Running read-only commands (ls, grep, cat, etc.)
  * Writing to files the user explicitly asked you to modify
  * Using tools the user explicitly invoked`,
		Priority: 30,
	}
}

func (b *SystemPromptBuilderV2) buildFileOperationSection() *Section {
	return &Section{
		Title:    "File Operation Rules",
		Content:  fileOperationRules,
		Priority: 35,
	}
}

func (b *SystemPromptBuilderV2) buildEnvironmentSection() *Section {
	if b.envCollector == nil {
		return nil
	}

	info, err := b.CollectEnvironment()
	if err != nil {
		return nil
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Operating System: %s (%s)", info.OS, info.Architecture))
	parts = append(parts, fmt.Sprintf("Shell: %s", info.Shell))
	parts = append(parts, fmt.Sprintf("Working Directory: %s", info.WorkingDir))

	if info.GoVersion != "" {
		parts = append(parts, fmt.Sprintf("Go Version: %s", info.GoVersion))
	}
	if info.NodeVersion != "" {
		parts = append(parts, fmt.Sprintf("Node.js Version: %s", info.NodeVersion))
	}
	if info.PythonVersion != "" {
		parts = append(parts, fmt.Sprintf("Python Version: %s", info.PythonVersion))
	}

	if info.GitRepo {
		parts = append(parts, "")
		parts = append(parts, "Git Information:")
		parts = append(parts, fmt.Sprintf("  Branch: %s", info.GitBranch))
		parts = append(parts, fmt.Sprintf("  Commit: %s", info.GitCommit))
		if info.GitRemote != "" {
			parts = append(parts, fmt.Sprintf("  Remote: %s", info.GitRemote))
		}
		if info.GitStatus != "" {
			parts = append(parts, fmt.Sprintf("  Status: %s", info.GitStatus))
		}
	}

	return &Section{
		Title:    "Environment Information",
		Content:  strings.Join(parts, "\n"),
		Priority: 40,
	}
}

func (b *SystemPromptBuilderV2) buildProjectSection() *Section {
	if b.projectLoader == nil {
		return nil
	}

	ctx, err := b.CollectProjectContext("")
	if err != nil {
		return nil
	}

	var parts []string

	// Project type
	if ctx.Type != "" {
		parts = append(parts, fmt.Sprintf("Project Type: %s", ctx.Type))
	}

	// AGENTS.md content
	if ctx.AgentsMDContent != "" {
		parts = append(parts, "")
		parts = append(parts, "Project Guidelines (from AGENTS.md):")
		parts = append(parts, ctx.AgentsMDContent)
	}

	// .claude/ content
	if ctx.ClaudeDirContent != "" {
		parts = append(parts, "")
		parts = append(parts, "Project Configuration (from .claude/):")
		parts = append(parts, ctx.ClaudeDirContent)
	}

	if len(parts) == 0 {
		return nil
	}

	return &Section{
		Title:    "Project Context",
		Content:  strings.Join(parts, "\n"),
		Priority: 50,
	}
}

func (b *SystemPromptBuilderV2) buildCodingStandardsSection() *Section {
	if b.projectLoader == nil {
		return nil
	}

	ctx, err := b.CollectProjectContext("")
	if err != nil {
		return nil
	}

	var parts []string

	// Project-specific coding standards
	if ctx.CodingStandards != "" {
		parts = append(parts, ctx.CodingStandards)
	} else {
		// Default coding standards
		parts = append(parts, `General Coding Standards:
- Follow the existing code style in the project
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Write tests for new functionality
- Ensure code is properly formatted`)
	}

	if ctx.StyleGuide != "" {
		parts = append(parts, "")
		parts = append(parts, "Style Guide:")
		parts = append(parts, ctx.StyleGuide)
	}

	return &Section{
		Title:    "Coding Standards",
		Content:  strings.Join(parts, "\n"),
		Priority: 60,
	}
}

func (b *SystemPromptBuilderV2) buildSessionSection() *Section {
	if b.sessionContext == nil {
		return nil
	}

	var parts []string

	if b.sessionContext.PreviousSummary != "" {
		parts = append(parts, "Previous Session Summary:")
		parts = append(parts, b.sessionContext.PreviousSummary)
		parts = append(parts, "")
	}

	if b.sessionContext.CurrentTask != "" {
		parts = append(parts, fmt.Sprintf("Current Task: %s", b.sessionContext.CurrentTask))
	}

	if len(b.sessionContext.TodoList) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Todo List:")
		for i, todo := range b.sessionContext.TodoList {
			parts = append(parts, fmt.Sprintf("  %d. %s", i+1, todo))
		}
	}

	if len(b.sessionContext.ContextDirs) > 0 {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("Context Directories: %s", strings.Join(b.sessionContext.ContextDirs, ", ")))
	}

	if len(parts) == 0 {
		return nil
	}

	return &Section{
		Title:    "Session Context",
		Content:  strings.Join(parts, "\n"),
		Priority: 70,
	}
}

// ==================== Helper methods ====================

func (b *SystemPromptBuilderV2) renderSections(sections []Section) (string, error) {
	var parts []string

	for _, section := range sections {
		content := strings.TrimSpace(section.Content)
		if content == "" {
			continue
		}

		// Render template variables
		rendered, err := b.renderTemplate(content)
		if err != nil {
			return "", fmt.Errorf("render section %s: %w", section.Title, err)
		}

		if section.Title != "" {
			parts = append(parts, fmt.Sprintf("## %s\n%s", section.Title, rendered))
		} else {
			parts = append(parts, rendered)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

func (b *SystemPromptBuilderV2) renderTemplate(content string) (string, error) {
	tmpl, err := template.New("section").
		Funcs(template.FuncMap{
			"join":  strings.Join,
			"upper": strings.ToUpper,
			"lower": strings.ToLower,
		}).
		Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, b.vars); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func sortSectionsByPriority(sections []Section) []Section {
	// Simple bubble sort
	result := make([]Section, len(sections))
	copy(result, sections)

	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Priority > result[j].Priority {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
