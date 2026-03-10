// Package prompts 提示词管理器实现
package prompts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DefaultManager 默认提示词管理器实例
var DefaultManager Manager = NewDefaultManager()

// defaultManager 提示词管理器默认实现
type defaultManager struct {
	mu      sync.RWMutex
	prompts map[PromptType]*Prompt
	vars    *TemplateVars
	dataDir string // 外部提示词文件目录
}

// NewDefaultManager 创建默认提示词管理器
func NewDefaultManager() Manager {
	m := &defaultManager{
		prompts: make(map[PromptType]*Prompt),
		vars:    DefaultTemplateVars(),
	}

	// 加载内置提示词
	for t, p := range builtinPrompts {
		m.prompts[t] = p
	}

	return m
}

// NewManager 创建自定义提示词管理器
func NewManager(vars *TemplateVars) Manager {
	if vars == nil {
		vars = DefaultTemplateVars()
	}
	m := &defaultManager{
		prompts: make(map[PromptType]*Prompt),
		vars:    vars,
	}

	// 加载内置提示词
	for t, p := range builtinPrompts {
		m.prompts[t] = p
	}

	return m
}

// SetDataDir 设置外部提示词文件目录
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

	// 自动加载目录中的提示词文件
	return m.loadFromDataDir()
}

// loadFromDataDir 从数据目录加载提示词
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

		// 支持 .json, .yaml, .yml, .md 格式
		if ext != ".json" && ext != ".yaml" && ext != ".yml" && ext != ".md" {
			continue
		}

		path := filepath.Join(m.dataDir, name)
		if err := m.loadPromptFile(path, ext); err != nil {
			// 记录错误但继续加载其他文件
			fmt.Fprintf(os.Stderr, "Warning: failed to load prompt file %s: %v\n", name, err)
		}
	}

	return nil
}

// loadPromptFile 加载单个提示词文件
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

// loadJSONPrompt 加载 JSON 格式提示词
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

// loadYAMLPrompt 加载 YAML 格式提示词
func (m *defaultManager) loadYAMLPrompt(data []byte) error {
	// 简单实现，使用 JSON 作为中间格式
	// 实际项目中可以导入 yaml 包
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

// loadMarkdownPrompt 加载 Markdown 格式提示词
func (m *defaultManager) loadMarkdownPrompt(path string, data []byte) error {
	// 从文件名提取提示词类型
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	promptType := PromptType(name)

	// 尝试从文件内容解析 frontmatter
	content := string(data)
	var title, description string

	// 简单解析 YAML frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// 解析 frontmatter
			frontmatter := parts[1]
			content = strings.TrimSpace(parts[2])

			// 简单解析 key: value 行
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

// Get 获取提示词
func (m *defaultManager) Get(promptType PromptType) (*Prompt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prompt, ok := m.prompts[promptType]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", promptType)
	}

	// 返回副本以避免外部修改
	p := *prompt
	return &p, nil
}

// GetRendered 获取渲染后的提示词
func (m *defaultManager) GetRendered(promptType PromptType, vars *TemplateVars) (string, error) {
	prompt, err := m.Get(promptType)
	if err != nil {
		return "", err
	}

	// 合并变量
	if vars == nil {
		vars = m.vars
	} else {
		merged := *m.vars
		merged.Merge(vars)
		vars = &merged
	}

	return prompt.Render(vars)
}

// Register 注册自定义提示词
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

// RegisterFromFile 从文件注册提示词
func (m *defaultManager) RegisterFromFile(path string) error {
	ext := filepath.Ext(path)
	if ext == "" {
		return fmt.Errorf("file has no extension")
	}

	return m.loadPromptFile(path, ext)
}

// List 列出所有提示词类型
func (m *defaultManager) List() []PromptType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]PromptType, 0, len(m.prompts))
	for t := range m.prompts {
		types = append(types, t)
	}
	return types
}

// GetMainAgentPrompt 获取主 Agent 提示词
func (m *defaultManager) GetMainAgentPrompt(vars *TemplateVars) (string, error) {
	// 尝试获取自定义主 Agent 提示词
	if prompt, err := m.GetRendered("custom_main_agent", vars); err == nil {
		return prompt, nil
	}
	// 使用默认主 Agent 提示词
	return m.GetRendered(PromptTypeMainAgent, vars)
}

// GetSubagentPrompt 获取子 Agent 提示词
func (m *defaultManager) GetSubagentPrompt(subagentType string, vars *TemplateVars) (string, error) {
	// 构建提示词类型
	promptType := PromptType(subagentType)

	// 尝试直接获取
	if prompt, err := m.GetRendered(promptType, vars); err == nil {
		return prompt, nil
	}

	// 尝试添加 "subagent_" 前缀
	promptType = PromptType("subagent_" + subagentType)
	if prompt, err := m.GetRendered(promptType, vars); err == nil {
		return prompt, nil
	}

	return "", fmt.Errorf("subagent prompt not found: %s", subagentType)
}

// ========== 便捷函数 ==========

// Get 使用默认管理器获取提示词
func Get(promptType PromptType) (*Prompt, error) {
	return DefaultManager.Get(promptType)
}

// GetRendered 使用默认管理器获取渲染后的提示词
func GetRendered(promptType PromptType, vars *TemplateVars) (string, error) {
	return DefaultManager.GetRendered(promptType, vars)
}

// Register 使用默认管理器注册提示词
func Register(prompt *Prompt) error {
	return DefaultManager.Register(prompt)
}

// RegisterFromFile 使用默认管理器从文件注册
func RegisterFromFile(path string) error {
	return DefaultManager.RegisterFromFile(path)
}

// List 使用默认管理器列出提示词
func List() []PromptType {
	return DefaultManager.List()
}

// GetMainAgentPrompt 使用默认管理器获取主 Agent 提示词
func GetMainAgentPrompt(vars *TemplateVars) (string, error) {
	return DefaultManager.GetMainAgentPrompt(vars)
}

// GetSubagentPrompt 使用默认管理器获取子 Agent 提示词
func GetSubagentPrompt(subagentType string, vars *TemplateVars) (string, error) {
	return DefaultManager.GetSubagentPrompt(subagentType, vars)
}
