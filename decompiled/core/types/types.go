// Package types core type definitions
// Decompiled from qodercli-reverse v0.1.29
package types

import "time"

// Role message role
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// ModelProvider model provider
type ModelProvider string

const (
	ProviderQoder     ModelProvider = "qoder"
	ProviderAnthropic ModelProvider = "anthropic"
	ProviderOpenAI    ModelProvider = "openai"
	ProviderIdeaLab   ModelProvider = "idealab"
	ProviderDashScope ModelProvider = "dashscope"
)

// ModelId model identifier
type ModelId string

// FinishReason finish reason
type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonToolUse       FinishReason = "tool_use"
	FinishReasonLength        FinishReason = "length"
	FinishReasonContentFilter FinishReason = "content_filter"
)

// Message message structure
type Message struct {
	Role       Role          `json:"role"`
	Content    []ContentPart `json:"content,omitempty"`
	ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	Name       string        `json:"name,omitempty"`
}

// ContentPart content part
type ContentPart struct {
	Type        string       `json:"type"` // "text", "image", "thinking"
	Text        string       `json:"text,omitempty"`
	ImageSource *ImageSource `json:"image_source,omitempty"`
	Thinking    string       `json:"thinking,omitempty"`
}

// ImageSource image source
type ImageSource struct {
	Type      string `json:"type"`       // "base64", "url"
	MediaType string `json:"media_type"` // "image/png", "image/jpeg"
	Data      string `json:"data"`       // base64 data or URL
}

// ToolCall tool call
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// ToolResult tool result
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name,omitempty"` // tool name, required by OpenAI API
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// ToolInfo tool information
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// TokenUsage token usage
type TokenUsage struct {
	InputTokens       int `json:"input_tokens"`
	OutputTokens      int `json:"output_tokens"`
	TotalTokens       int `json:"total_tokens"`
	CachedTokens      int `json:"cached_tokens,omitempty"`
	PreContextTokens  int `json:"pre_context_tokens,omitempty"`
	PostContextTokens int `json:"post_context_tokens,omitempty"`
}

// Finish finish status
type Finish struct {
	Reason    FinishReason `json:"reason"`
	Timestamp time.Time    `json:"timestamp"`
}

// Delta incremental update
type Delta struct {
	Type    DeltaType `json:"type"`
	Content string    `json:"content,omitempty"`
}

// DeltaType delta type
type DeltaType string

const (
	DeltaTypeText         DeltaType = "text"
	DeltaTypeThinking     DeltaType = "thinking"
	DeltaTypeToolUse      DeltaType = "tool_use"
	DeltaTypeToolCall     DeltaType = "tool_call"
	DeltaTypeContentBlock DeltaType = "content_block"
)

// PermissionRule permission rule
type PermissionRule struct {
	Pattern string `json:"pattern"`
	Action  string `json:"action"` // "allow", "deny", "ask"
}

// Location location info
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// Attachment attachment
type Attachment struct {
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
	Data      []byte `json:"data,omitempty"`
}

// Memory memory item
type Memory struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Status status
type Status struct {
	Running    bool      `json:"running"`
	StartTime  time.Time `json:"start_time"`
	LastActive time.Time `json:"last_active"`
}

// ErrorData error data
type ErrorData struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SdkMessage SDK message
type SdkMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ContentBlock content block
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// FeatureName feature name
type FeatureName string

const (
	FeatureCodeReview FeatureName = "code_review"
	FeatureWebSearch  FeatureName = "web_search"
	FeatureImageGen   FeatureName = "image_gen"
	FeatureSubAgent   FeatureName = "sub_agent"
	FeatureGitHub     FeatureName = "github"
	FeatureMCP        FeatureName = "mcp"
)

// BusinessType business type
type BusinessType string

const (
	BusinessTypeCoding   BusinessType = "coding"
	BusinessTypeReview   BusinessType = "review"
	BusinessTypeDebug    BusinessType = "debug"
	BusinessTypeExplain  BusinessType = "explain"
	BusinessTypeRefactor BusinessType = "refactor"
)
