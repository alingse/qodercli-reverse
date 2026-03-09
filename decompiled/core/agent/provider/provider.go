// Package provider LLM Provider 接口和实现
// 反编译自 qodercli v0.1.29
package provider

import (
	"context"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// EventType 事件类型
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

// Event 流式事件
type Event struct {
	Type      EventType   `json:"type"`
	Content   string      `json:"content,omitempty"`
	ToolCall  *ToolCall   `json:"tool_call,omitempty"`
	ToolUse   *ToolUse    `json:"tool_use,omitempty"`
	Thinking  string      `json:"thinking,omitempty"`
	Error     *ErrorData  `json:"error,omitempty"`
	TokenUsage *TokenUsage `json:"token_usage,omitempty"`
	FinishReason types.FinishReason `json:"finish_reason,omitempty"`
}

// ToolCall 工具调用事件
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolUse 工具使用事件
type ToolUse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ErrorData 错误数据
type ErrorData struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// TokenUsage Token 使用量
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ThinkLevel 思考级别
type ThinkLevel int

const (
	ThinkLevelNone ThinkLevel = iota
	ThinkLevelLow
	ThinkLevelMedium
	ThinkLevelHigh
)

// ReasoningEffort 推理努力程度
type ReasoningEffort string

const (
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortHigh   ReasoningEffort = "high"
)

// ModelRequest 模型请求
type ModelRequest struct {
	Model            string              `json:"model"`
	Messages         []types.Message     `json:"messages"`
	Tools            []types.ToolInfo    `json:"tools,omitempty"`
	MaxTokens        int                 `json:"max_tokens,omitempty"`
	Temperature      float64             `json:"temperature,omitempty"`
	TopP             float64             `json:"top_p,omitempty"`
	ThinkLevel       ThinkLevel          `json:"think_level,omitempty"`
	ReasoningEffort  ReasoningEffort     `json:"reasoning_effort,omitempty"`
	Stream           bool                `json:"stream"`
	SystemPrompt     string              `json:"system_prompt,omitempty"`
	StopSequences    []string            `json:"stop_sequences,omitempty"`
	Metadata         map[string]string   `json:"metadata,omitempty"`
}

// Response 模型响应
type Response struct {
	Content       string              `json:"content"`
	ToolCalls     []types.ToolCall    `json:"tool_calls,omitempty"`
	FinishReason  types.FinishReason  `json:"finish_reason"`
	TokenUsage    *TokenUsage         `json:"token_usage"`
	Thinking      string              `json:"thinking,omitempty"`
	ID            string              `json:"id"`
	Model         string              `json:"model"`
}

// Client Provider 客户端接口
type Client interface {
	// Stream 流式请求
	Stream(ctx context.Context, req *ModelRequest) <-chan Event
	
	// Send 同步请求
	Send(ctx context.Context, req *ModelRequest) (*Response, error)
	
	// Close 关闭客户端
	Close() error
}

// ClientBuilder 客户端构建器接口
type ClientBuilder interface {
	WithAPIKey(apiKey string) ClientBuilder
	WithBaseURL(baseURL string) ClientBuilder
	WithDebug(debug bool) ClientBuilder
	WithAgentName(name string) ClientBuilder
	WithSubagent(isSubagent bool) ClientBuilder
	Build() (Client, error)
}

// baseClient 基础客户端
type baseClient struct {
	apiKey     string
	baseURL    string
	debug      bool
	agentName  string
	isSubagent bool
	httpClient HTTPClient
}

// HTTPClient HTTP 客户端接口
type HTTPClient interface {
	Do(req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPRequest HTTP 请求
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Timeout time.Duration
}

// HTTPResponse HTTP 响应
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}
