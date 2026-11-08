// Package state State management - Todo task list
// Decompiled from qodercli-reverse v0.1.29
package state

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// TodoStatus Task status type
type TodoStatus string

const (
	// TodoStatusInProgress In progress
	TodoStatusInProgress TodoStatus = "in_progress"
	// TodoStatusCompleted Completed
	TodoStatusCompleted TodoStatus = "completed"
	// TodoStatusPending Pending
	TodoStatusPending TodoStatus = "pending"
	// TodoStatusDone Done
	TodoStatusDone TodoStatus = "done"
	// TodoStatusCancelled Cancelled
	TodoStatusCancelled TodoStatus = "cancelled"
)

// IsValidTodoStatus Checks whether the status is valid
func IsValidTodoStatus(status string) bool {
	switch TodoStatus(status) {
	case TodoStatusInProgress, TodoStatusCompleted, TodoStatusPending, TodoStatusDone, TodoStatusCancelled:
		return true
	default:
		return false
	}
}

// Todo Task item
type Todo struct {
	// ID Unique task identifier
	ID string `json:"id"`
	// Content Task content description
	Content string `json:"content"`
	// Status Task status
	Status string `json:"status"`
	// ActiveForm Description of the in-progress state during execution (e.g. "Running tests")
	ActiveForm string `json:"activeForm"`
}

// Validate Validates Todo validity
func (t *Todo) Validate() error {
	if strings.TrimSpace(t.ID) == "" {
		return fmt.Errorf("todo id cannot be empty")
	}
	if strings.TrimSpace(t.Content) == "" {
		return fmt.Errorf("todo content cannot be empty")
	}
	if !IsValidTodoStatus(t.Status) {
		return fmt.Errorf("todo invalid status '%s'", t.Status)
	}
	if strings.TrimSpace(t.ActiveForm) == "" {
		return fmt.Errorf("todo activeForm cannot be empty")
	}
	return nil
}

// TodoState Task state management interface
type TodoState interface {
	// LoadTodos Loads the task list
	LoadTodos() ([]Todo, error)
	// SaveTodos Saves the task list
	SaveTodos(todos []Todo) error
	// GetTodos Retrieves the current task list
	GetTodos() []Todo
	// TodosToText Converts the task list to a text description
	TodosToText() string
}

// defaultTodoState Default implementation
type defaultTodoState struct {
	mu     sync.RWMutex
	todos  []Todo
	memory *SessionMemory // Associated session memory
}

// NewTodoState Creates a new TodoState
func NewTodoState(memory *SessionMemory) TodoState {
	return &defaultTodoState{
		todos:  make([]Todo, 0),
		memory: memory,
	}
}

// LoadTodos Loads the task list from memory
func (s *defaultTodoState) LoadTodos() ([]Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read from session memory
	if s.memory != nil {
		if data, ok := s.memory.Get("todos"); ok {
			var todos []Todo
			if err := json.Unmarshal(data, &todos); err == nil {
				s.todos = todos
			}
		}
	}

	result := make([]Todo, len(s.todos))
	copy(result, s.todos)
	return result, nil
}

// SaveTodos Saves the task list to memory
func (s *defaultTodoState) SaveTodos(todos []Todo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate all tasks
	for i, todo := range todos {
		if err := todo.Validate(); err != nil {
			return fmt.Errorf("todo[%d]: %w", i, err)
		}
	}

	s.todos = make([]Todo, len(todos))
	copy(s.todos, todos)

	// Save to session memory
	if s.memory != nil {
		data, err := json.Marshal(s.todos)
		if err == nil {
			s.memory.Set("todos", data)
		}
	}

	return nil
}

// GetTodos Retrieves the current task list
func (s *defaultTodoState) GetTodos() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Todo, len(s.todos))
	copy(result, s.todos)
	return result
}

// TodosToText Converts the task list to a text description
func (s *defaultTodoState) TodosToText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.todos) == 0 {
		return "Todo list is empty."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Todo list status (%d todos):\n", len(s.todos)))

	for _, todo := range s.todos {
		statusIcon := "⬜"
		switch TodoStatus(todo.Status) {
		case TodoStatusCompleted, TodoStatusDone:
			statusIcon = "✅"
		case TodoStatusInProgress:
			statusIcon = "🔄"
		case TodoStatusCancelled:
			statusIcon = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s [%s] %s\n", statusIcon, todo.Status, todo.Content))
	}

	return sb.String()
}

// AutoFixEmptyTodos Automatically fixes empty arrays
// Rule: empty arrays are not allowed to clear the list; when all tasks are completed, keep the tasks and set their status to completed
func AutoFixEmptyTodos(todos []Todo) []Todo {
	if len(todos) == 0 {
		// Return an empty list, but the caller should handle this case
		return []Todo{}
	}
	return todos
}

// CountByStatus Counts tasks by status
func CountByStatus(todos []Todo, status string) int {
	count := 0
	for _, todo := range todos {
		if todo.Status == status {
			count++
		}
	}
	return count
}

// HasInProgress Checks whether there is an in-progress task
func HasInProgress(todos []Todo) bool {
	for _, todo := range todos {
		if TodoStatus(todo.Status) == TodoStatusInProgress {
			return true
		}
	}
	return false
}
