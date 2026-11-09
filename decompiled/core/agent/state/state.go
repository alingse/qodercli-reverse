// Package state Agent state management
package state

import (
	"sync"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// Stats Session statistics
type Stats struct {
	TotalInputTokens  int
	TotalOutputTokens int
	TotalTokens       int
	ToolCallCount     int
	AssistantReplies  int
}

// State Agent state
type State struct {
	messages    []types.Message
	toolResults map[string]*types.ToolResult
	stats       Stats
	mu          sync.RWMutex
}

// New Creates a new state instance
func New() *State {
	return &State{
		messages:    make([]types.Message, 0),
		toolResults: make(map[string]*types.ToolResult),
		stats:       Stats{},
	}
}

// AddMessage Adds a message
func (s *State) AddMessage(msg *types.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, *msg)
}

// AddToolResult Adds a tool result
func (s *State) AddToolResult(result *types.ToolResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolResults[result.ToolCallID] = result

	// Also add the tool result as a message with the tool role to the message list
	// so the API can correctly identify the tool call response
	// Note: the OpenAI API requires tool response messages to include the name field
	toolMsg := types.Message{
		Role:       types.RoleTool,
		Name:       result.Name,
		Content:    []types.ContentPart{{Type: "text", Text: result.Content}},
		ToolCallID: result.ToolCallID,
	}
	s.messages = append(s.messages, toolMsg)
}

// GetMessages Returns all messages
func (s *State) GetMessages() []types.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.messages
}

// SetMessages Sets messages
func (s *State) SetMessages(messages []types.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = messages
}

// ClearMessages Clears messages
func (s *State) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make([]types.Message, 0)
	s.toolResults = make(map[string]*types.ToolResult)
}

// GetToolResult Retrieves a tool result
func (s *State) GetToolResult(toolCallID string) *types.ToolResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.toolResults[toolCallID]
}

// UpdateTokenUsage Updates token usage statistics
func (s *State) UpdateTokenUsage(inputTokens, outputTokens int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.TotalInputTokens += inputTokens
	s.stats.TotalOutputTokens += outputTokens
	s.stats.TotalTokens += inputTokens + outputTokens
}

// IncrementToolCallCount Increments the tool call count
func (s *State) IncrementToolCallCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.ToolCallCount++
}

// IncrementAssistantReplies Increments the assistant reply count
func (s *State) IncrementAssistantReplies() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.AssistantReplies++
}

// GetStats Retrieves statistics
func (s *State) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}
