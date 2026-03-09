// Package tools TodoWrite 工具实现
// 反编译自 qodercli v0.1.29
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
)

// TodoWriteParams TodoWrite 工具参数
type TodoWriteParams struct {
	// Todos 任务列表
	Todos []state.Todo `json:"todos"`
}

// NotPresent 表示该工具不需要权限参数
type NotPresent struct{}

// TodoWriteResponseMetadata 响应元数据
type TodoWriteResponseMetadata struct {
	// OldTodos 更新前的任务列表
	OldTodos []state.Todo `json:"oldTodos"`
	// NewTodos 更新后的任务列表
	NewTodos []state.Todo `json:"newTodos"`
}

// todoWriteTool TodoWrite 工具实现
type todoWriteTool struct {
	todoState state.TodoState
}

// NewTodoWriteTool 创建 TodoWrite 工具
func NewTodoWriteTool(todoState state.TodoState) Tool {
	return &todoWriteTool{
		todoState: todoState,
	}
}

// Name 返回工具名称
func (t *todoWriteTool) Name() string {
	return "TodoWrite"
}

// Description 返回工具描述
func (t *todoWriteTool) Description() string {
	return `Use this tool to create and update a todo list to track your progress on the user's task.

The todo list helps you organize complex tasks into manageable steps.

Guidelines:
- **High-level stages only**: Track major phases like requirements, design, review, coding, testing, delivery
- **Language alignment**: Use the same language as the user's conversation
- **ActiveForm**: Use present continuous form (e.g., "Analyzing code", "Implementing feature")
- **NEVER pass empty array**: When all tasks complete, keep them with status "completed"

Task States:
- in_progress: Currently working on this
- completed/done: Finished successfully
- pending: Waiting to start
- cancelled: No longer needed`
}

// InputSchema 返回输入 Schema
func (t *todoWriteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"todos": map[string]interface{}{
				"type":        "array",
				"description": "The updated todo list. NEVER pass an empty array to clear the list. When all tasks are done, keep them with status 'completed'.",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "The ID of the TODO item",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "The content of the TODO item",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "The status of the TODO item",
							"enum":        []string{"in_progress", "completed", "pending", "done", "cancelled"},
						},
						"activeForm": map[string]interface{}{
							"type":        "string",
							"description": "The activeForm of the TODO item - present continuous form shown during execution (e.g., 'Running tests', 'Building the project')",
						},
					},
					"required": []string{"id", "content", "status", "activeForm"},
				},
			},
		},
		"required": []string{"todos"},
	}
}

// Execute 执行工具
func (t *todoWriteTool) Execute(ctx context.Context, input string) (string, error) {
	var params TodoWriteParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse TodoWrite params: %w", err)
	}

	// 加载旧的 todos
	oldTodos, err := t.todoState.LoadTodos()
	if err != nil {
		return "", fmt.Errorf("failed to load todos: %w", err)
	}

	// 处理空数组情况
	params.Todos = autoFixEmptyTodos(params.Todos)

	// 验证新的 todos
	if err := t.validateTodos(params.Todos); err != nil {
		return "", err
	}

	// 保存新的 todos
	if err := t.todoState.SaveTodos(params.Todos); err != nil {
		return "", fmt.Errorf("failed to save todos: %w", err)
	}

	// 构建响应元数据
	metadata := TodoWriteResponseMetadata{
		OldTodos: oldTodos,
		NewTodos: params.Todos,
	}

	// 构建响应文本
	result := t.buildResponse(metadata)
	return result, nil
}

// validateTodos 验证任务列表
func (t *todoWriteTool) validateTodos(todos []state.Todo) error {
	for i, todo := range todos {
		if err := todo.Validate(); err != nil {
			return fmt.Errorf("todo[%d]: %w", i, err)
		}
	}
	return nil
}

// autoFixEmptyTodos 自动修复空数组
// 规则: 不允许传递空数组清除列表
// 当所有任务完成时，保持任务并将状态设为 completed
func autoFixEmptyTodos(todos []state.Todo) []state.Todo {
	if len(todos) == 0 {
		// 返回空列表，但需要特殊处理
		// 实际上不应该用空数组调用此工具
		return []state.Todo{}
	}
	return todos
}

// buildResponse 构建响应文本
func (t *todoWriteTool) buildResponse(metadata TodoWriteResponseMetadata) string {
	// 计算统计信息
	oldCount := len(metadata.OldTodos)
	newCount := len(metadata.NewTodos)
	
	completedCount := 0
	inProgressCount := 0
	pendingCount := 0
	
	for _, todo := range metadata.NewTodos {
		switch state.TodoStatus(todo.Status) {
		case state.TodoStatusCompleted, state.TodoStatusDone:
			completedCount++
		case state.TodoStatusInProgress:
			inProgressCount++
		case state.TodoStatusPending:
			pendingCount++
		}
	}

	// 构建响应
	var result string
	if oldCount == 0 {
		result = fmt.Sprintf("Created todo list with %d task(s).\n", newCount)
	} else {
		result = fmt.Sprintf("Updated todo list: %d -> %d task(s).\n", oldCount, newCount)
	}

	result += fmt.Sprintf("Progress: %d completed, %d in_progress, %d pending\n", 
		completedCount, inProgressCount, pendingCount)
	
	if newCount > 0 {
		result += "\nCurrent tasks:\n"
		for _, todo := range metadata.NewTodos {
			statusIcon := "⬜"
			switch state.TodoStatus(todo.Status) {
			case state.TodoStatusCompleted, state.TodoStatusDone:
				statusIcon = "✅"
			case state.TodoStatusInProgress:
				statusIcon = "🔄"
			case state.TodoStatusCancelled:
				statusIcon = "❌"
			}
			result += fmt.Sprintf("  %s [%s] %s\n", statusIcon, todo.Status, todo.Content)
		}
	}

	return result
}

// TodoWriteToolName 工具名称常量
const TodoWriteToolName = "TodoWrite"
