// Package tools tool system interfaces and base implementation
// Decompiled from qodercli-reverse v0.1.29
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// Tool tool interface
type Tool interface {
	// Name tool name
	Name() string

	// Description tool description
	Description() string

	// InputSchema input schema (JSON Schema)
	InputSchema() map[string]interface{}

	// Execute executes the tool
	Execute(ctx context.Context, input string) (string, error)
}

// BaseTool base tool
type BaseTool struct {
	name        string
	description string
	inputSchema map[string]interface{}
}

// Name returns the tool name
func (t *BaseTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t *BaseTool) Description() string {
	return t.description
}

// InputSchema returns the input schema
func (t *BaseTool) InputSchema() map[string]interface{} {
	return t.inputSchema
}

// Registry tool registry
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool
func (r *Registry) Register(tool Tool) error {
	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}
	r.tools[tool.Name()] = tool
	return nil
}

// Get gets a tool
func (r *Registry) Get(name string) (Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

// List lists all tools
func (r *Registry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ListNames lists all tool names
func (r *Registry) ListNames() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ToToolInfo converts to ToolInfo list
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

// ToolInfo tool info
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ExecuteRequest execute request
type ExecuteRequest struct {
	ToolName string          `json:"tool_name"`
	Input    json.RawMessage `json:"input"`
}

// ExecuteResponse execute response
type ExecuteResponse struct {
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
	IsError bool   `json:"is_error"`
}

// ToolCall tool call
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult tool result
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name,omitempty"` // tool name, used for OpenAI API
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// Executor tool executor
type Executor struct {
	registry *Registry
}

// NewExecutor creates a new executor
func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

// Execute executes the tool call
func (e *Executor) Execute(ctx context.Context, call *ToolCall) (*ToolResult, error) {
	tool, err := e.registry.Get(call.Name)
	if err != nil {
		return &ToolResult{
			ToolCallID: call.ID,
			Name:       call.Name,
			Content:    fmt.Sprintf("Tool %s not found: %v", call.Name, err),
			IsError:    true,
		}, nil
	}

	content, err := tool.Execute(ctx, string(call.Arguments))
	if err != nil {
		return &ToolResult{
			ToolCallID: call.ID,
			Name:       call.Name,
			Content:    err.Error(),
			IsError:    true,
		}, nil
	}

	return &ToolResult{
		ToolCallID: call.ID,
		Name:       call.Name,
		Content:    content,
		IsError:    false,
	}, nil
}

// ExecuteBatch batch executes tool calls
func (e *Executor) ExecuteBatch(ctx context.Context, calls []*ToolCall) ([]*ToolResult, error) {
	results := make([]*ToolResult, len(calls))
	for i, call := range calls {
		result, err := e.Execute(ctx, call)
		if err != nil {
			results[i] = &ToolResult{
				ToolCallID: call.ID,
				Name:       call.Name,
				Content:    err.Error(),
				IsError:    true,
			}
			continue
		}
		results[i] = result
	}
	return results, nil
}

// BuildStringSchema builds string parameter schema
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

// BuildFileSchema builds file operation schema
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

// BuildEditSchema builds edit operation schema
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

// BuildBashSchema builds Bash command schema
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

// BuildGrepSchema builds Grep search schema
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

// BuildGlobSchema builds Glob pattern schema
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

// BuildTaskSchema builds Task tool schema
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
