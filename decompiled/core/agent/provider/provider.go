// Package provider LLM Provider interfaces and implementation
// Decompiled from qodercli-reverse v0.1.29
package provider

import (
	"context"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// EventType event type
type EventType int

const (
	EventTypeMessageStart EventType = iota
	EventTypeMessageDelta
	EventTypeMessageStop
	EventTypeContentBlockStart
	EventTypeContentBlockDelta
	EventTypeContentBlockStop
	EventTypeToolUseStart
	EventTypeToolUseDelta
	EventTypeToolUseStop
	EventTypeThinkingDelta
	EventTypeError
)

// Event streaming event
type Event struct {
	Type         EventType          `json:"type"`
	Content      string             `json:"content,omitempty"`
	ToolCall     *ToolCall          `json:"tool_call,omitempty"`
	ToolUse      *ToolUse           `json:"tool_use,omitempty"`
	Thinking     string             `json:"thinking,omitempty"`
	Error        *ErrorData         `json:"error,omitempty"`
	TokenUsage   *TokenUsage        `json:"token_usage,omitempty"`
	FinishReason types.FinishReason `json:"finish_reason,omitempty"`
}

// ToolCall tool call event
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolUse tool use event
type ToolUse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ErrorData error data
type ErrorData struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// TokenUsage token usage
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ThinkLevel thinking level
type ThinkLevel int

const (
	ThinkLevelNone ThinkLevel = iota
	ThinkLevelLow
	ThinkLevelMedium
	ThinkLevelHigh
)

// ReasoningEffort reasoning effort
type ReasoningEffort string

const (
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortHigh   ReasoningEffort = "high"
)

// ModelRequest model request
type ModelRequest struct {
	Model           string            `json:"model"`
	Messages        []types.Message   `json:"messages"`
	Tools           []types.ToolInfo  `json:"tools,omitempty"`
	MaxTokens       int               `json:"max_tokens,omitempty"`
	Temperature     float64           `json:"temperature,omitempty"`
	TopP            float64           `json:"top_p,omitempty"`
	ThinkLevel      ThinkLevel        `json:"think_level,omitempty"`
	ReasoningEffort ReasoningEffort   `json:"reasoning_effort,omitempty"`
	Stream          bool              `json:"stream"`
	SystemPrompt    string            `json:"system_prompt,omitempty"`
	StopSequences   []string          `json:"stop_sequences,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// Response model response
type Response struct {
	Content      string             `json:"content"`
	ToolCalls    []types.ToolCall   `json:"tool_calls,omitempty"`
	FinishReason types.FinishReason `json:"finish_reason"`
	TokenUsage   *TokenUsage        `json:"token_usage"`
	Thinking     string             `json:"thinking,omitempty"`
	ID           string             `json:"id"`
	Model        string             `json:"model"`
}

// Client Provider client interface
type Client interface {
	// Stream streaming request
	Stream(ctx context.Context, req *ModelRequest) <-chan Event

	// Send sync request
	Send(ctx context.Context, req *ModelRequest) (*Response, error)

	// Close closes the client
	Close() error
}

// ClientBuilder client builder interface
type ClientBuilder interface {
	WithAPIKey(apiKey string) ClientBuilder
	WithBaseURL(baseURL string) ClientBuilder
	WithDebug(debug bool) ClientBuilder
	WithAgentName(name string) ClientBuilder
	WithSubagent(isSubagent bool) ClientBuilder
	Build() (Client, error)
}

// baseClient base client
type baseClient struct {
	apiKey     string
	baseURL    string
	debug      bool
	agentName  string
	isSubagent bool
	httpClient HTTPClient
}

// HTTPClient HTTP client interface
type HTTPClient interface {
	Do(req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPRequest HTTP request
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Timeout time.Duration
}

// HTTPResponse HTTP response
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}
