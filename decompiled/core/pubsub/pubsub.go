// Package pubsub 提供发布订阅事件系统
// 用于组件间通信和解耦
package pubsub

import (
	"context"
	"sync"
)

// EventType 定义事件类型
type EventType string

const (
	// Agent 相关事件
	EventTypeAgentStart    EventType = "agent:start"
	EventTypeAgentStop     EventType = "agent:stop"
	EventTypeAgentThinking EventType = "agent:thinking"
	EventTypeAgentResponse EventType = "agent:response"
	EventTypeAgentError    EventType = "agent:error"

	// Tool 相关事件
	EventTypeToolStart  EventType = "tool:start"
	EventTypeToolEnd    EventType = "tool:end"
	EventTypeToolError  EventType = "tool:error"
	EventTypeToolOutput EventType = "tool:output"

	// 消息相关事件
	EventTypeMessageAdd    EventType = "message:add"
	EventTypeMessageUpdate EventType = "message:update"
	EventTypeMessageDelete EventType = "message:delete"

	// 会话相关事件
	EventTypeSessionStart   EventType = "session:start"
	EventTypeSessionResume  EventType = "session:resume"
	EventTypeSessionExport  EventType = "session:export"
	EventTypeSessionCompact EventType = "session:compact"

	// 用户交互事件
	EventTypeUserInput    EventType = "user:input"
	EventTypeUserInterrupt EventType = "user:interrupt"
	EventTypeCommandExec  EventType = "command:exec"

	// 状态事件
	EventTypeStatusUpdate EventType = "status:update"
	EventTypeTokenUsage   EventType = "token:usage"
)

// Event 定义事件结构
type Event struct {
	Type    EventType
	Payload interface{}
}

// Handler 事件处理器
type Handler func(ctx context.Context, event Event)

// PubSub 发布订阅系统
type PubSub struct {
	subscribers map[EventType][]Handler
	mu          sync.RWMutex
}

// New 创建新的 PubSub 实例
func New() *PubSub {
	return &PubSub{
		subscribers: make(map[EventType][]Handler),
	}
}

// Subscribe 订阅事件
func (ps *PubSub) Subscribe(eventType EventType, handler Handler) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.subscribers[eventType] = append(ps.subscribers[eventType], handler)
}

// Unsubscribe 取消订阅（简化实现：清空所有处理器）
func (ps *PubSub) Unsubscribe(eventType EventType) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.subscribers, eventType)
}

// Publish 发布事件
func (ps *PubSub) Publish(ctx context.Context, event Event) {
	ps.mu.RLock()
	handlers := ps.subscribers[event.Type]
	ps.mu.RUnlock()

	for _, handler := range handlers {
		go handler(ctx, event)
	}
}

// PublishSync 同步发布事件
func (ps *PubSub) PublishSync(ctx context.Context, event Event) {
	ps.mu.RLock()
	handlers := ps.subscribers[event.Type]
	ps.mu.RUnlock()

	for _, handler := range handlers {
		handler(ctx, event)
	}
}
