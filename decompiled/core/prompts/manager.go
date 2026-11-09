// Package prompts Prompt manager implementation
package prompts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DefaultManager Default prompt manager instance
var DefaultManager Manager = NewDefaultManager()

// defaultManager Default prompt manager implementation
type defaultManager struct {
	mu      sync.RWMutex
	prompts map[PromptType]*Prompt
	vars    *TemplateVars
	dataDir string // External prompt files directory
}

// NewDefaultManager Creates the default prompt manager
func NewDefaultManager() Manager {
	m := &defaultManager{
		prompts: make(map[PromptType]*Prompt),
		vars:    DefaultTemplateVars(),
	}

	// Load built-in prompts
	for t, p := range builtinPrompts {
		m.prompts[t] = p
	}

	return m
}

// NewManager Creates a custom prompt manager
func NewManager(vars *TemplateVars) Manager {
	if vars == nil {
		vars = DefaultTemplateVars()
	}
	m := &defaultManager{
		prompts: make(map[PromptType]*Prompt),
		vars:    vars,
	}

	// Load built-in prompts
	for t, p := range builtinPrompts {
		m.prompts[t] = p
	}

	return m
}

// SetDataDir Sets the external prompt files directory
func (m *defaultManager) SetDataDir(dir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("invalid data directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("data path is not a directory: %s", dir)
	}

	m.dataDir = dir

	// Auto-load prompt files in the directory
	return m.loadFromDataDir()
}

// loadFromDataDir Loads prompts from the data directory
func (m *defaultManager) loadFromDataDir() error {
	if m.dataDir == "" {
		return nil
	}

	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		return fmt.Errorf("read data directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)

		// Supports .json, .yaml, .yml, .md formats
		if ext != ".json" && ext != ".yaml" && ext != ".yml" && ext != ".md" {
			continue
		}

		path := filepath.Join(m.dataDir, name)
		if err := m.loadPromptFile(path, ext); err != nil {
			// Log the error but continue loading other files
			fmt.Fprintf(os.Stderr, "Warning: failed to load prompt file %s: %v\n", name, err)
		}
	}

	return nil
}

// loadPromptFile Loads a single prompt file
func (m *defaultManager) loadPromptFile(path string, ext string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	switch ext {
	case ".json":
		return m.loadJSONPrompt(data)
	case ".yaml", ".yml":
		return m.loadYAMLPrompt(data)
	case ".md":
		return m.loadMarkdownPrompt(path, data)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}

// loadJSONPrompt Loads a JSON format prompt
func (m *defaultManager) loadJSONPrompt(data []byte) error {
	var prompt Prompt
	if err := json.Unmarshal(data, &prompt); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}

	if prompt.Type == "" {
		return fmt.Errorf("prompt type is required")
	}

	prompt.IsBuiltIn = false
	m.prompts[prompt.Type] = &prompt
	return nil
}

// loadYAMLPrompt Loads a YAML format prompt
func (m *defaultManager) loadYAMLPrompt(data []byte) error {
	// Simple implementation, using JSON as an intermediate format
	// In a real project the yaml package could be imported
	var prompt Prompt
	if err := json.Unmarshal(data, &prompt); err != nil {
		return fmt.Errorf("unmarshal YAML (as JSON): %w", err)
	}

	if prompt.Type == "" {
		return fmt.Errorf("prompt type is required")
	}

	prompt.IsBuiltIn = false
	m.prompts[prompt.Type] = &prompt
	return nil
}

// loadMarkdownPrompt Loads a Markdown format prompt
func (m *defaultManager) loadMarkdownPrompt(path string, data []byte) error {
	// Extract prompt type from file name
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	promptType := PromptType(name)

	// Try to parse frontmatter from file content
	content := string(data)
	var title, description string

	// Simple YAML frontmatter parsing
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// Parse frontmatter
			frontmatter := parts[1]
			content = strings.TrimSpace(parts[2])

			// Simple key: value line parsing
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "title:") {
					title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				} else if strings.HasPrefix(line, "description:") {
					description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
		}
	}

	if title == "" {
		title = name
	}

	prompt := &Prompt{
		Type:        promptType,
		Name:        title,
		Description: description,
		Template:    content,
		IsBuiltIn:   false,
	}

	m.prompts[promptType] = prompt
	return nil
}

// Get Retrieves a prompt
func (m *defaultManager) Get(promptType PromptType) (*Prompt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prompt, ok := m.prompts[promptType]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", promptType)
	}

	// Return a copy to avoid external modification
	p := *prompt
	return &p, nil
}

// GetRendered Retrieves a rendered prompt
func (m *defaultManager) GetRendered(promptType PromptType, vars *TemplateVars) (string, error) {
	prompt, err := m.Get(promptType)
	if err != nil {
		return "", err
	}

	// Merge variables
	if vars == nil {
		vars = m.vars
	} else {
		merged := *m.vars
		merged.Merge(vars)
		vars = &merged
	}

	return prompt.Render(vars)
}

// Register Registers a custom prompt
func (m *defaultManager) Register(prompt *Prompt) error {
	if prompt == nil {
		return fmt.Errorf("prompt cannot be nil")
	}
	if prompt.Type == "" {
		return fmt.Errorf("prompt type is required")
	}
	if prompt.Template == "" {
		return fmt.Errorf("prompt template is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.prompts[prompt.Type] = prompt
	return nil
}

// RegisterFromFile Registers a prompt from a file
func (m *defaultManager) RegisterFromFile(path string) error {
	ext := filepath.Ext(path)
	if ext == "" {
		return fmt.Errorf("file has no extension")
	}

	return m.loadPromptFile(path, ext)
}

// List Lists all prompt types
func (m *defaultManager) List() []PromptType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]PromptType, 0, len(m.prompts))
	for t := range m.prompts {
		types = append(types, t)
	}
	return types
}

// GetMainAgentPrompt Retrieves the main Agent prompt
func (m *defaultManager) GetMainAgentPrompt(vars *TemplateVars) (string, error) {
	// Try to get a custom main Agent prompt
	if prompt, err := m.GetRendered("custom_main_agent", vars); err == nil {
		return prompt, nil
	}
	// Use the default main Agent prompt
	return m.GetRendered(PromptTypeMainAgent, vars)
}

// GetSubagentPrompt Retrieves a sub-agent prompt
func (m *defaultManager) GetSubagentPrompt(subagentType string, vars *TemplateVars) (string, error) {
	// Build prompt type
	promptType := PromptType(subagentType)

	// Try direct retrieval
	if prompt, err := m.GetRendered(promptType, vars); err == nil {
		return prompt, nil
	}

	// Try adding the "subagent_" prefix
	promptType = PromptType("subagent_" + subagentType)
	if prompt, err := m.GetRendered(promptType, vars); err == nil {
		return prompt, nil
	}

	return "", fmt.Errorf("subagent prompt not found: %s", subagentType)
}

// ========== Convenience functions ==========

// Get Retrieves a prompt using the default manager
func Get(promptType PromptType) (*Prompt, error) {
	return DefaultManager.Get(promptType)
}

// GetRendered Retrieves a rendered prompt using the default manager
func GetRendered(promptType PromptType, vars *TemplateVars) (string, error) {
	return DefaultManager.GetRendered(promptType, vars)
}

// Register Registers a prompt using the default manager
func Register(prompt *Prompt) error {
	return DefaultManager.Register(prompt)
}

// RegisterFromFile Registers a prompt from a file using the default manager
func RegisterFromFile(path string) error {
	return DefaultManager.RegisterFromFile(path)
}

// List Lists prompts using the default manager
func List() []PromptType {
	return DefaultManager.List()
}

// GetMainAgentPrompt Retrieves the main Agent prompt using the default manager
func GetMainAgentPrompt(vars *TemplateVars) (string, error) {
	return DefaultManager.GetMainAgentPrompt(vars)
}

// GetSubagentPrompt Retrieves a sub-agent prompt using the default manager
func GetSubagentPrompt(subagentType string, vars *TemplateVars) (string, error) {
	return DefaultManager.GetSubagentPrompt(subagentType, vars)
}
