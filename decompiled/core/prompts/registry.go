// Package prompts Prompt registry
// Provides dynamic prompt management and composition
package prompts

import (
	"fmt"
	"strings"
	"sync"
)

// Registry Prompt registry
type Registry struct {
	mu       sync.RWMutex
	prompts  map[PromptType]*Prompt
	builders map[PromptType]PromptBuilderFunc
}

// PromptBuilderFunc Prompt builder function
type PromptBuilderFunc func(vars *TemplateVars) (string, error)

// NewRegistry Creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		prompts:  make(map[PromptType]*Prompt),
		builders: make(map[PromptType]PromptBuilderFunc),
	}
}

// Register Registers a prompt
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

// RegisterBuilder Registers a builder function
func (r *Registry) RegisterBuilder(promptType PromptType, builder PromptBuilderFunc) error {
	if builder == nil {
		return fmt.Errorf("builder cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.builders[promptType] = builder
	return nil
}

// Get Retrieves a prompt
func (r *Registry) Get(promptType PromptType) (*Prompt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if prompt, ok := r.prompts[promptType]; ok {
		p := *prompt
		return &p, nil
	}

	return nil, fmt.Errorf("prompt not found: %s", promptType)
}

// Build Builds a prompt
func (r *Registry) Build(promptType PromptType, vars *TemplateVars) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Prefer builder functions
	if builder, ok := r.builders[promptType]; ok {
		return builder(vars)
	}

	// Then static prompts
	if prompt, ok := r.prompts[promptType]; ok {
		return prompt.Render(vars)
	}

	return "", fmt.Errorf("prompt not found: %s", promptType)
}

// List Lists all prompt types
func (r *Registry) List() []PromptType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]PromptType, 0, len(r.prompts)+len(r.builders))
	for t := range r.prompts {
		types = append(types, t)
	}
	for t := range r.builders {
		// Deduplicate
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

// Unregister Unregisters a prompt
func (r *Registry) Unregister(promptType PromptType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.prompts, promptType)
	delete(r.builders, promptType)
}

// CompositePrompt Composite prompt
// Used to combine multiple prompt fragments into a complete prompt
type CompositePrompt struct {
	parts []CompositePart
}

// CompositePart Composite part
type CompositePart struct {
	PromptType PromptType
	Condition  func(vars *TemplateVars) bool
	Required   bool
}

// NewCompositePrompt Creates a composite prompt
func NewCompositePrompt() *CompositePrompt {
	return &CompositePrompt{
		parts: make([]CompositePart, 0),
	}
}

// Add Adds a part
func (c *CompositePrompt) Add(promptType PromptType, required bool) *CompositePrompt {
	c.parts = append(c.parts, CompositePart{
		PromptType: promptType,
		Required:   required,
	})
	return c
}

// AddConditional Adds a conditional part
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

// Build Builds the composite prompt
func (c *CompositePrompt) Build(registry *Registry, vars *TemplateVars) (string, error) {
	var parts []string

	for _, part := range c.parts {
		// Check condition
		if part.Condition != nil && !part.Condition(vars) {
			continue
		}

		// Build prompt
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

// PromptComposer Prompt composer
// Provides a more flexible composition approach
type PromptComposer struct {
	registry *Registry
	vars     *TemplateVars
	sections []composedSection
}

// composedSection Composed section
type composedSection struct {
	Title    string
	Content  string
	Priority int
}

// NewPromptComposer Creates a prompt composer
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

// AddFromRegistry Adds from the registry
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

// AddSection Adds a custom section
func (pc *PromptComposer) AddSection(title, content string, priority int) *PromptComposer {
	pc.sections = append(pc.sections, composedSection{
		Title:    title,
		Content:  content,
		Priority: priority,
	})
	return pc
}

// AddRaw Adds raw content
func (pc *PromptComposer) AddRaw(content string, priority int) *PromptComposer {
	pc.sections = append(pc.sections, composedSection{
		Content:  content,
		Priority: priority,
	})
	return pc
}

// Compose Composes the final prompt
func (pc *PromptComposer) Compose() string {
	// Sort by priority
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

// Clear Clears all sections
func (pc *PromptComposer) Clear() {
	pc.sections = pc.sections[:0]
}

// SetVars Sets variables
func (pc *PromptComposer) SetVars(vars *TemplateVars) {
	pc.vars = vars
}

// GetVars Gets variables
func (pc *PromptComposer) GetVars() *TemplateVars {
	return pc.vars
}

// ========== Predefined composition configs ==========

// ComposeMainAgent Composes the main Agent prompt
func ComposeMainAgent(manager *defaultManager, vars *TemplateVars, options MainAgentOptions) (string, error) {
	registry := NewRegistry()
	// Copy prompts from the manager into the registry
	for _, t := range manager.List() {
		if p, err := manager.Get(t); err == nil {
			registry.Register(p)
		}
	}
	composer := NewPromptComposer(registry, vars)

	// Role definition (highest priority)
	if options.RoleDefinition != "" {
		composer.AddRaw(options.RoleDefinition, 0)
	} else {
		composer.AddFromRegistry(PromptTypeMainAgent, "", 0)
	}

	// Core instructions
	if options.IncludeCoreInstructions {
		composer.AddFromRegistry(PromptType("core_instructions"), "Core Instructions", 10)
	}

	// Tool rules
	if options.IncludeToolRules {
		composer.AddFromRegistry(PromptType("tool_rules"), "Tool Rules", 20)
	}

	// File operation rules
	if options.IncludeFileRules {
		composer.AddFromRegistry(PromptType("file_rules"), "File Operation Rules", 30)
	}

	// Educational content (optional)
	if options.IncludeEducational {
		composer.AddFromRegistry(PromptType("educational_rules"), "Educational Guidelines", 40)
	}

	// Custom sections
	for _, section := range options.CustomSections {
		composer.AddSection(section.Title, section.Content, section.Priority)
	}

	return composer.Compose(), nil
}

// MainAgentOptions Main Agent options
type MainAgentOptions struct {
	RoleDefinition          string
	IncludeCoreInstructions bool
	IncludeToolRules        bool
	IncludeFileRules        bool
	IncludeEducational      bool
	CustomSections          []CustomSection
}

// CustomSection Custom section
type CustomSection struct {
	Title    string
	Content  string
	Priority int
}

// ComposeSubagent Composes a sub-agent prompt
func ComposeSubagent(
	registry *Registry,
	vars *TemplateVars,
	baseType PromptType,
	customRules []string,
) (string, error) {
	composer := NewPromptComposer(registry, vars)

	// Base prompt
	if _, err := composer.AddFromRegistry(baseType, "", 0); err != nil {
		return "", err
	}

	// Custom rules
	for i, rule := range customRules {
		composer.AddSection("", rule, 100+i)
	}

	return composer.Compose(), nil
}
