// Package state Agent 状态管理
package state

import (
	"sync"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// State Agent 状态
type State struct {
	messages    []types.Message
	toolResults map[string]*types.ToolResult
	mu          sync.RWMutex
}

// New 创建新的状态实例
func New() *State {
	return &State{
		messages:    make([]types.Message, 0),
		toolResults: make(map[string]*types.ToolResult),
	}
}

// AddMessage 添加消息
func (s *State) AddMessage(msg *types.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, *msg)
}

// AddToolResult 添加工具结果
func (s *State) AddToolResult(result *types.ToolResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolResults[result.ToolCallID] = result

	// 同时将工具结果作为 tool 角色的消息添加到消息列表
	// 这样 API 才能正确识别工具调用的响应
	toolMsg := types.Message{
		Role:       types.RoleTool,
		Content:    []types.ContentPart{{Type: "text", Text: result.Content}},
		ToolCallID: result.ToolCallID,
	}
	s.messages = append(s.messages, toolMsg)
}

// GetMessages 获取所有消息
func (s *State) GetMessages() []types.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.messages
}

// SetMessages 设置消息
func (s *State) SetMessages(messages []types.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = messages
}

// ClearMessages 清空消息
func (s *State) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make([]types.Message, 0)
	s.toolResults = make(map[string]*types.ToolResult)
}

// GetToolResult 获取工具结果
func (s *State) GetToolResult(toolCallID string) *types.ToolResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.toolResults[toolCallID]
}
