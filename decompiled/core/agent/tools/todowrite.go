// Package tools TodoWrite tool implementation
// Decompiled from qodercli-reverse v0.1.29
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
)

// TodoWriteParams TodoWrite tool parameters
type TodoWriteParams struct {
	// Todos task list
	Todos []state.Todo `json:"todos"`
}

// NotPresent indicates this tool requires no permission parameters
type NotPresent struct{}

// TodoWriteResponseMetadata response metadata
type TodoWriteResponseMetadata struct {
	// OldTodos task list before update
	OldTodos []state.Todo `json:"oldTodos"`
	// NewTodos task list after update
	NewTodos []state.Todo `json:"newTodos"`
}

// todoWriteTool TodoWrite tool implementation
type todoWriteTool struct {
	todoState state.TodoState
}

// NewTodoWriteTool creates the TodoWrite tool
func NewTodoWriteTool(todoState state.TodoState) Tool {
	return &todoWriteTool{
		todoState: todoState,
	}
}

// Name returns the tool name
func (t *todoWriteTool) Name() string {
	return "TodoWrite"
}

// Description returns the tool description
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

// InputSchema returns the input schema
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

// Execute executes the tool
func (t *todoWriteTool) Execute(ctx context.Context, input string) (string, error) {
	var params TodoWriteParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse TodoWrite params: %w", err)
	}

	// load old todos
	oldTodos, err := t.todoState.LoadTodos()
	if err != nil {
		return "", fmt.Errorf("failed to load todos: %w", err)
	}

	// handle empty array case
	params.Todos = autoFixEmptyTodos(params.Todos)

	// validate new todos
	if err := t.validateTodos(params.Todos); err != nil {
		return "", err
	}

	// save new todos
	if err := t.todoState.SaveTodos(params.Todos); err != nil {
		return "", fmt.Errorf("failed to save todos: %w", err)
	}

	// build response metadata
	metadata := TodoWriteResponseMetadata{
		OldTodos: oldTodos,
		NewTodos: params.Todos,
	}

	// build response text
	result := t.buildResponse(metadata)
	return result, nil
}

// validateTodos validates the task list
func (t *todoWriteTool) validateTodos(todos []state.Todo) error {
	for i, todo := range todos {
		if err := todo.Validate(); err != nil {
			return fmt.Errorf("todo[%d]: %w", i, err)
		}
	}
	return nil
}

// autoFixEmptyTodos auto-fixes empty array
// rule: empty array not allowed to clear list
// when all tasks complete, keep tasks with status completed
func autoFixEmptyTodos(todos []state.Todo) []state.Todo {
	if len(todos) == 0 {
		// return empty list, but needs special handling
		// this tool should not actually be called with empty array
		return []state.Todo{}
	}
	return todos
}

// buildResponse builds response text
func (t *todoWriteTool) buildResponse(metadata TodoWriteResponseMetadata) string {
	// calculate statistics
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

	// build response
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

// TodoWriteToolName tool name constant
const TodoWriteToolName = "TodoWrite"
