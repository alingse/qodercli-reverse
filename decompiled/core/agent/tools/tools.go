// Package tools 工具系统接口和基础实现
// 反编译自 qodercli v0.1.29
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// Tool 工具接口
type Tool interface {
	// Name 工具名称
	Name() string

	// Description 工具描述
	Description() string

	// InputSchema 输入 Schema (JSON Schema)
	InputSchema() map[string]interface{}

	// Execute 执行工具
	Execute(ctx context.Context, input string) (string, error)
}

// BaseTool 基础工具
type BaseTool struct {
	name        string
	description string
	inputSchema map[string]interface{}
}

// Name 返回工具名称
func (t *BaseTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *BaseTool) Description() string {
	return t.description
}

// InputSchema 返回输入 Schema
func (t *BaseTool) InputSchema() map[string]interface{} {
	return t.inputSchema
}

// Registry 工具注册表
type Registry struct {
	tools map[string]Tool
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *Registry) Register(tool Tool) error {
	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}
	r.tools[tool.Name()] = tool
	return nil
}

// Get 获取工具
func (r *Registry) Get(name string) (Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

// List 列出所有工具
func (r *Registry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ListNames 列出所有工具名称
func (r *Registry) ListNames() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ToToolInfo 转换为 ToolInfo 列表
func (r *Registry) ToToolInfo() []types.ToolInfo {
	tools := r.List()
	infos := make([]types.ToolInfo, len(tools))
	for i, tool := range tools {
		infos[i] = types.ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		}
	}
	return infos
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ExecuteRequest 执行请求
type ExecuteRequest struct {
	ToolName string          `json:"tool_name"`
	Input    json.RawMessage `json:"input"`
}

// ExecuteResponse 执行响应
type ExecuteResponse struct {
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
	IsError bool   `json:"is_error"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult 工具结果
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// Executor 工具执行器
type Executor struct {
	registry *Registry
}

// NewExecutor 创建新的执行器
func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

// Execute 执行工具调用
func (e *Executor) Execute(ctx context.Context, call *ToolCall) (*ToolResult, error) {
	tool, err := e.registry.Get(call.Name)
	if err != nil {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Tool %s not found: %v", call.Name, err),
			IsError:    true,
		}, nil
	}

	content, err := tool.Execute(ctx, string(call.Arguments))
	if err != nil {
		return &ToolResult{
			ToolCallID: call.ID,
			Content:    err.Error(),
			IsError:    true,
		}, nil
	}

	return &ToolResult{
		ToolCallID: call.ID,
		Content:    content,
		IsError:    false,
	}, nil
}

// ExecuteBatch 批量执行工具调用
func (e *Executor) ExecuteBatch(ctx context.Context, calls []*ToolCall) ([]*ToolResult, error) {
	results := make([]*ToolResult, len(calls))
	for i, call := range calls {
		result, err := e.Execute(ctx, call)
		if err != nil {
			results[i] = &ToolResult{
				ToolCallID: call.ID,
				Content:    err.Error(),
				IsError:    true,
			}
			continue
		}
		results[i] = result
	}
	return results, nil
}

// BuildStringSchema 构建字符串参数 Schema
func BuildStringSchema(properties map[string]struct {
	Description string
	Required    bool
}) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	props := schema["properties"].(map[string]interface{})
	var required []string

	for name, prop := range properties {
		props[name] = map[string]interface{}{
			"type":        "string",
			"description": prop.Description,
		}
		if prop.Required {
			required = append(required, name)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// BuildFileSchema 构建文件操作 Schema
func BuildFileSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the file",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "Line number to start reading from",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Number of lines to read",
			},
		},
		"required": []string{"file_path"},
	}
}

// BuildEditSchema 构建编辑操作 Schema
func BuildEditSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The absolute path to the file",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "The text to replace",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "The replacement text",
			},
			"replace_all": map[string]interface{}{
				"type":        "boolean",
				"description": "Replace all occurrences",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

// BuildBashSchema 构建 Bash 命令 Schema
func BuildBashSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to execute",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of what the command does",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in milliseconds",
			},
			"run_in_background": map[string]interface{}{
				"type":        "boolean",
				"description": "Run the command in background",
			},
		},
		"required": []string{"command"},
	}
}

// BuildGrepSchema 构建 Grep 搜索 Schema
func BuildGrepSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The regex pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to search in",
			},
			"output_mode": map[string]interface{}{
				"type":        "string",
				"description": "Output mode: content, files_with_matches, count",
				"enum":        []string{"content", "files_with_matches", "count"},
			},
			"glob": map[string]interface{}{
				"type":        "string",
				"description": "File glob pattern",
			},
			"head_limit": map[string]interface{}{
				"type":        "integer",
				"description": "Limit number of results",
			},
			"multiline": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable multiline matching",
			},
		},
		"required": []string{"pattern"},
	}
}

// BuildGlobSchema 构建 Glob 模式 Schema
func BuildGlobSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "The glob pattern to match",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The directory to search in",
			},
		},
		"required": []string{"pattern"},
	}
}

// BuildTaskSchema 构建 Task 工具 Schema
func BuildTaskSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"subagent_type": map[string]interface{}{
				"type":        "string",
				"description": "The type of subagent to use",
				"enum": []string{
					"code-reviewer",
					"design-agent",
					"spec-review-agent",
					"task-executor",
					"general-purpose",
				},
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of the task",
			},
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "The prompt for the subagent",
			},
		},
		"required": []string{"subagent_type", "description", "prompt"},
	}
}
