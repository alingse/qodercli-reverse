// Package pubsub provides a publish/subscribe event system
// used for component communication and decoupling.
package pubsub

import (
	"context"
	"sync"
)

// EventType defines an event type.
type EventType string

const (
	// Agent-related events
	EventTypeAgentStart    EventType = "agent:start"
	EventTypeAgentStop     EventType = "agent:stop"
	EventTypeAgentThinking EventType = "agent:thinking"
	EventTypeAgentResponse EventType = "agent:response"
	EventTypeAgentError    EventType = "agent:error"

	// Tool-related events
	EventTypeToolStart  EventType = "tool:start"
	EventTypeToolEnd    EventType = "tool:end"
	EventTypeToolError  EventType = "tool:error"
	EventTypeToolOutput EventType = "tool:output"

	// Message-related events
	EventTypeMessageAdd    EventType = "message:add"
	EventTypeMessageUpdate EventType = "message:update"
	EventTypeMessageDelete EventType = "message:delete"

	// Session-related events
	EventTypeSessionStart   EventType = "session:start"
	EventTypeSessionResume  EventType = "session:resume"
	EventTypeSessionExport  EventType = "session:export"
	EventTypeSessionCompact EventType = "session:compact"

	// User interaction events
	EventTypeUserInput     EventType = "user:input"
	EventTypeUserInterrupt EventType = "user:interrupt"
	EventTypeCommandExec   EventType = "command:exec"

	// Status events
	EventTypeStatusUpdate EventType = "status:update"
	EventTypeTokenUsage   EventType = "token:usage"
)

// Event defines the event structure.
type Event struct {
	Type    EventType
	Payload interface{}
}

// Handler is an event handler.
type Handler func(ctx context.Context, event Event)

// PubSub is a publish/subscribe system.
type PubSub struct {
	subscribers map[EventType][]Handler
	mu          sync.RWMutex
}

// New creates a new PubSub instance.
func New() *PubSub {
	return &PubSub{
		subscribers: make(map[EventType][]Handler),
	}
}

// Subscribe subscribes to an event.
func (ps *PubSub) Subscribe(eventType EventType, handler Handler) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.subscribers[eventType] = append(ps.subscribers[eventType], handler)
}

// Unsubscribe unsubscribes from an event (simplified implementation: clears all handlers).
func (ps *PubSub) Unsubscribe(eventType EventType) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.subscribers, eventType)
}

// Publish publishes an event.
func (ps *PubSub) Publish(ctx context.Context, event Event) {
	ps.mu.RLock()
	handlers := ps.subscribers[event.Type]
	ps.mu.RUnlock()

	for _, handler := range handlers {
		go handler(ctx, event)
	}
}

// PublishSync publishes an event synchronously.
func (ps *PubSub) PublishSync(ctx context.Context, event Event) {
	ps.mu.RLock()
	handlers := ps.subscribers[event.Type]
	ps.mu.RUnlock()

	for _, handler := range handlers {
		handler(ctx, event)
	}
}
