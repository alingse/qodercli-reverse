// Package prompts 提示词注册表
// 提供动态提示词管理和组合功能
package prompts

import (
	"fmt"
	"strings"
	"sync"
)

// Registry 提示词注册表
type Registry struct {
	mu       sync.RWMutex
	prompts  map[PromptType]*Prompt
	builders map[PromptType]PromptBuilderFunc
}

// PromptBuilderFunc 提示词构建函数
type PromptBuilderFunc func(vars *TemplateVars) (string, error)

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		prompts:  make(map[PromptType]*Prompt),
		builders: make(map[PromptType]PromptBuilderFunc),
	}
}

// Register 注册提示词
func (r *Registry) Register(prompt *Prompt) error {
	if prompt == nil {
		return fmt.Errorf("prompt cannot be nil")
	}
	if prompt.Type == "" {
		return fmt.Errorf("prompt type is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.prompts[prompt.Type] = prompt
	return nil
}

// RegisterBuilder 注册构建函数
func (r *Registry) RegisterBuilder(promptType PromptType, builder PromptBuilderFunc) error {
	if builder == nil {
		return fmt.Errorf("builder cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.builders[promptType] = builder
	return nil
}

// Get 获取提示词
func (r *Registry) Get(promptType PromptType) (*Prompt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if prompt, ok := r.prompts[promptType]; ok {
		p := *prompt
		return &p, nil
	}

	return nil, fmt.Errorf("prompt not found: %s", promptType)
}

// Build 构建提示词
func (r *Registry) Build(promptType PromptType, vars *TemplateVars) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 优先使用构建函数
	if builder, ok := r.builders[promptType]; ok {
		return builder(vars)
	}

	// 然后使用静态提示词
	if prompt, ok := r.prompts[promptType]; ok {
		return prompt.Render(vars)
	}

	return "", fmt.Errorf("prompt not found: %s", promptType)
}

// List 列出所有提示词类型
func (r *Registry) List() []PromptType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]PromptType, 0, len(r.prompts)+len(r.builders))
	for t := range r.prompts {
		types = append(types, t)
	}
	for t := range r.builders {
		// 去重
		found := false
		for _, existing := range types {
			if existing == t {
				found = true
				break
			}
		}
		if !found {
			types = append(types, t)
		}
	}
	return types
}

// Unregister 注销提示词
func (r *Registry) Unregister(promptType PromptType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.prompts, promptType)
	delete(r.builders, promptType)
}

// CompositePrompt 组合提示词
// 用于将多个提示词片段组合成完整提示词
type CompositePrompt struct {
	parts []CompositePart
}

// CompositePart 组合部件
type CompositePart struct {
	PromptType PromptType
	Condition  func(vars *TemplateVars) bool
	Required   bool
}

// NewCompositePrompt 创建组合提示词
func NewCompositePrompt() *CompositePrompt {
	return &CompositePrompt{
		parts: make([]CompositePart, 0),
	}
}

// Add 添加部件
func (c *CompositePrompt) Add(promptType PromptType, required bool) *CompositePrompt {
	c.parts = append(c.parts, CompositePart{
		PromptType: promptType,
		Required:   required,
	})
	return c
}

// AddConditional 添加条件部件
func (c *CompositePrompt) AddConditional(
	promptType PromptType,
	condition func(vars *TemplateVars) bool,
) *CompositePrompt {
	c.parts = append(c.parts, CompositePart{
		PromptType: promptType,
		Condition:  condition,
	})
	return c
}

// Build 构建组合提示词
func (c *CompositePrompt) Build(registry *Registry, vars *TemplateVars) (string, error) {
	var parts []string

	for _, part := range c.parts {
		// 检查条件
		if part.Condition != nil && !part.Condition(vars) {
			continue
		}

		// 构建提示词
		content, err := registry.Build(part.PromptType, vars)
		if err != nil {
			if part.Required {
				return "", fmt.Errorf("required prompt %s failed: %w", part.PromptType, err)
			}
			continue
		}

		if content != "" {
			parts = append(parts, content)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// PromptComposer 提示词组合器
// 提供更灵活的组合方式
type PromptComposer struct {
	registry *Registry
	vars     *TemplateVars
	sections []composedSection
}

// composedSection 组合章节
type composedSection struct {
	Title    string
	Content  string
	Priority int
}

// NewPromptComposer 创建提示词组合器
func NewPromptComposer(registry *Registry, vars *TemplateVars) *PromptComposer {
	if vars == nil {
		vars = DefaultTemplateVars()
	}
	return &PromptComposer{
		registry: registry,
		vars:     vars,
		sections: make([]composedSection, 0),
	}
}

// AddFromRegistry 从注册表添加
func (pc *PromptComposer) AddFromRegistry(
	promptType PromptType,
	title string,
	priority int,
) (*PromptComposer, error) {
	content, err := pc.registry.Build(promptType, pc.vars)
	if err != nil {
		return pc, err
	}

	pc.sections = append(pc.sections, composedSection{
		Title:    title,
		Content:  content,
		Priority: priority,
	})

	return pc, nil
}

// AddSection 添加自定义章节
func (pc *PromptComposer) AddSection(title, content string, priority int) *PromptComposer {
	pc.sections = append(pc.sections, composedSection{
		Title:    title,
		Content:  content,
		Priority: priority,
	})
	return pc
}

// AddRaw 添加原始内容
func (pc *PromptComposer) AddRaw(content string, priority int) *PromptComposer {
	pc.sections = append(pc.sections, composedSection{
		Content:  content,
		Priority: priority,
	})
	return pc
}

// Compose 组合最终提示词
func (pc *PromptComposer) Compose() string {
	// 按优先级排序
	for i := 0; i < len(pc.sections); i++ {
		for j := i + 1; j < len(pc.sections); j++ {
			if pc.sections[i].Priority > pc.sections[j].Priority {
				pc.sections[i], pc.sections[j] = pc.sections[j], pc.sections[i]
			}
		}
	}

	var parts []string
	for _, section := range pc.sections {
		if section.Title != "" {
			parts = append(parts, fmt.Sprintf("## %s\n%s", section.Title, section.Content))
		} else {
			parts = append(parts, section.Content)
		}
	}

	return strings.Join(parts, "\n\n")
}

// Clear 清空所有章节
func (pc *PromptComposer) Clear() {
	pc.sections = pc.sections[:0]
}

// SetVars 设置变量
func (pc *PromptComposer) SetVars(vars *TemplateVars) {
	pc.vars = vars
}

// GetVars 获取变量
func (pc *PromptComposer) GetVars() *TemplateVars {
	return pc.vars
}

// ========== 预定义组合配置 ==========

// ComposeMainAgent 组合主 Agent 提示词
func ComposeMainAgent(manager *defaultManager, vars *TemplateVars, options MainAgentOptions) (string, error) {
	registry := NewRegistry()
	// 复制管理器中的提示词到注册表
	for _, t := range manager.List() {
		if p, err := manager.Get(t); err == nil {
			registry.Register(p)
		}
	}
	composer := NewPromptComposer(registry, vars)

	// 角色定义（最高优先级）
	if options.RoleDefinition != "" {
		composer.AddRaw(options.RoleDefinition, 0)
	} else {
		composer.AddFromRegistry(PromptTypeMainAgent, "", 0)
	}

	// 核心指令
	if options.IncludeCoreInstructions {
		composer.AddFromRegistry(PromptType("core_instructions"), "Core Instructions", 10)
	}

	// 工具规则
	if options.IncludeToolRules {
		composer.AddFromRegistry(PromptType("tool_rules"), "Tool Rules", 20)
	}

	// 文件操作规则
	if options.IncludeFileRules {
		composer.AddFromRegistry(PromptType("file_rules"), "File Operation Rules", 30)
	}

	// 教育内容（可选）
	if options.IncludeEducational {
		composer.AddFromRegistry(PromptType("educational_rules"), "Educational Guidelines", 40)
	}

	// 自定义章节
	for _, section := range options.CustomSections {
		composer.AddSection(section.Title, section.Content, section.Priority)
	}

	return composer.Compose(), nil
}

// MainAgentOptions 主 Agent 选项
type MainAgentOptions struct {
	RoleDefinition          string
	IncludeCoreInstructions bool
	IncludeToolRules        bool
	IncludeFileRules        bool
	IncludeEducational      bool
	CustomSections          []CustomSection
}

// CustomSection 自定义章节
type CustomSection struct {
	Title    string
	Content  string
	Priority int
}

// ComposeSubagent 组合子 Agent 提示词
func ComposeSubagent(
	registry *Registry,
	vars *TemplateVars,
	baseType PromptType,
	customRules []string,
) (string, error) {
	composer := NewPromptComposer(registry, vars)

	// 基础提示词
	if _, err := composer.AddFromRegistry(baseType, "", 0); err != nil {
		return "", err
	}

	// 自定义规则
	for i, rule := range customRules {
		composer.AddSection("", rule, 100+i)
	}

	return composer.Compose(), nil
}
