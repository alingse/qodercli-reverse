// Package state 状态管理 - Todo 任务列表
// 反编译自 qodercli v0.1.29
package state

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// TodoStatus 任务状态类型
type TodoStatus string

const (
	// TodoStatusInProgress 进行中
	TodoStatusInProgress TodoStatus = "in_progress"
	// TodoStatusCompleted 已完成
	TodoStatusCompleted TodoStatus = "completed"
	// TodoStatusPending 待处理
	TodoStatusPending TodoStatus = "pending"
	// TodoStatusDone 完成
	TodoStatusDone TodoStatus = "done"
	// TodoStatusCancelled 已取消
	TodoStatusCancelled TodoStatus = "cancelled"
)

// IsValidTodoStatus 检查状态是否有效
func IsValidTodoStatus(status string) bool {
	switch TodoStatus(status) {
	case TodoStatusInProgress, TodoStatusCompleted, TodoStatusPending, TodoStatusDone, TodoStatusCancelled:
		return true
	default:
		return false
	}
}

// Todo 任务项
type Todo struct {
	// ID 任务唯一标识
	ID string `json:"id"`
	// Content 任务内容描述
	Content string `json:"content"`
	// Status 任务状态
	Status string `json:"status"`
	// ActiveForm 执行时的进行态描述 (如 "Running tests")
	ActiveForm string `json:"activeForm"`
}

// Validate 验证 Todo 有效性
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

// TodoState 任务状态管理接口
type TodoState interface {
	// LoadTodos 加载任务列表
	LoadTodos() ([]Todo, error)
	// SaveTodos 保存任务列表
	SaveTodos(todos []Todo) error
	// GetTodos 获取当前任务列表
	GetTodos() []Todo
	// TodosToText 将任务列表转换为文本描述
	TodosToText() string
}

// defaultTodoState 默认实现
type defaultTodoState struct {
	mu     sync.RWMutex
	todos  []Todo
	memory *SessionMemory // 关联到会话内存
}

// NewTodoState 创建新的 TodoState
func NewTodoState(memory *SessionMemory) TodoState {
	return &defaultTodoState{
		todos:  make([]Todo, 0),
		memory: memory,
	}
}

// LoadTodos 从内存加载任务列表
func (s *defaultTodoState) LoadTodos() ([]Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 从会话内存中读取
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

// SaveTodos 保存任务列表到内存
func (s *defaultTodoState) SaveTodos(todos []Todo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证所有任务
	for i, todo := range todos {
		if err := todo.Validate(); err != nil {
			return fmt.Errorf("todo[%d]: %w", i, err)
		}
	}

	s.todos = make([]Todo, len(todos))
	copy(s.todos, todos)

	// 保存到会话内存
	if s.memory != nil {
		data, err := json.Marshal(s.todos)
		if err == nil {
			s.memory.Set("todos", data)
		}
	}

	return nil
}

// GetTodos 获取当前任务列表
func (s *defaultTodoState) GetTodos() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Todo, len(s.todos))
	copy(result, s.todos)
	return result
}

// TodosToText 将任务列表转换为文本描述
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

// AutoFixEmptyTodos 自动修复空数组
// 规则: 不允许传递空数组清除列表，当所有任务完成时，保持任务并将状态设为 completed
func AutoFixEmptyTodos(todos []Todo) []Todo {
	if len(todos) == 0 {
		// 返回空列表，但外部应该处理这种情况
		return []Todo{}
	}
	return todos
}

// CountByStatus 按状态统计任务数量
func CountByStatus(todos []Todo, status string) int {
	count := 0
	for _, todo := range todos {
		if todo.Status == status {
			count++
		}
	}
	return count
}

// HasInProgress 检查是否有进行中的任务
func HasInProgress(todos []Todo) bool {
	for _, todo := range todos {
		if TodoStatus(todo.Status) == TodoStatusInProgress {
			return true
		}
	}
	return false
}
