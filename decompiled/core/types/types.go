// Package types 核心类型定义
// 反编译自 qodercli v0.1.29
package types

import "time"

// Role 消息角色
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// ModelProvider 模型提供商
type ModelProvider string

const (
	ProviderQoder     ModelProvider = "qoder"
	ProviderAnthropic ModelProvider = "anthropic"
	ProviderOpenAI    ModelProvider = "openai"
	ProviderIdeaLab   ModelProvider = "idealab"
	ProviderDashScope ModelProvider = "dashscope"
)

// ModelId 模型标识
type ModelId string

// FinishReason 完成原因
type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonToolUse       FinishReason = "tool_use"
	FinishReasonLength        FinishReason = "length"
	FinishReasonContentFilter FinishReason = "content_filter"
)

// Message 消息结构
type Message struct {
	Role       Role          `json:"role"`
	Content    []ContentPart `json:"content,omitempty"`
	ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	Name       string        `json:"name,omitempty"`
}

// ContentPart 内容部分
type ContentPart struct {
	Type        string       `json:"type"` // "text", "image", "thinking"
	Text        string       `json:"text,omitempty"`
	ImageSource *ImageSource `json:"image_source,omitempty"`
	Thinking    string       `json:"thinking,omitempty"`
}

// ImageSource 图片源
type ImageSource struct {
	Type      string `json:"type"`       // "base64", "url"
	MediaType string `json:"media_type"` // "image/png", "image/jpeg"
	Data      string `json:"data"`       // base64 数据或 URL
}

// ToolCall 工具调用
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON 字符串
}

// ToolResult 工具结果
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name,omitempty"` // 工具名称，OpenAI API 要求
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// TokenUsage Token 使用量
type TokenUsage struct {
	InputTokens       int `json:"input_tokens"`
	OutputTokens      int `json:"output_tokens"`
	TotalTokens       int `json:"total_tokens"`
	CachedTokens      int `json:"cached_tokens,omitempty"`
	PreContextTokens  int `json:"pre_context_tokens,omitempty"`
	PostContextTokens int `json:"post_context_tokens,omitempty"`
}

// Finish 完成状态
type Finish struct {
	Reason    FinishReason `json:"reason"`
	Timestamp time.Time    `json:"timestamp"`
}

// Delta 增量更新
type Delta struct {
	Type    DeltaType `json:"type"`
	Content string    `json:"content,omitempty"`
}

// DeltaType 增量类型
type DeltaType string

const (
	DeltaTypeText         DeltaType = "text"
	DeltaTypeThinking     DeltaType = "thinking"
	DeltaTypeToolUse      DeltaType = "tool_use"
	DeltaTypeToolCall     DeltaType = "tool_call"
	DeltaTypeContentBlock DeltaType = "content_block"
)

// PermissionRule 权限规则
type PermissionRule struct {
	Pattern string `json:"pattern"`
	Action  string `json:"action"` // "allow", "deny", "ask"
}

// Location 位置信息
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// Attachment 附件
type Attachment struct {
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
	Data      []byte `json:"data,omitempty"`
}

// Memory 记忆项
type Memory struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Status 状态
type Status struct {
	Running    bool      `json:"running"`
	StartTime  time.Time `json:"start_time"`
	LastActive time.Time `json:"last_active"`
}

// ErrorData 错误数据
type ErrorData struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SdkMessage SDK 消息
type SdkMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// FeatureName 功能名称
type FeatureName string

const (
	FeatureCodeReview FeatureName = "code_review"
	FeatureWebSearch  FeatureName = "web_search"
	FeatureImageGen   FeatureName = "image_gen"
	FeatureSubAgent   FeatureName = "sub_agent"
	FeatureGitHub     FeatureName = "github"
	FeatureMCP        FeatureName = "mcp"
)

// BusinessType 业务类型
type BusinessType string

const (
	BusinessTypeCoding   BusinessType = "coding"
	BusinessTypeReview   BusinessType = "review"
	BusinessTypeDebug    BusinessType = "debug"
	BusinessTypeExplain  BusinessType = "explain"
	BusinessTypeRefactor BusinessType = "refactor"
)
