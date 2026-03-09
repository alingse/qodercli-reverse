// Package app TUI 应用启动器
package app

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
)

// Options 应用启动选项
type Options struct {
	Config   *config.Config
	Agent    *agent.Agent
	PubSub   *pubsub.PubSub
	ResumeID string
}

// Run 启动 TUI 应用
func Run(opts Options) error {
	// 创建应用模型
	model := New(opts.Config, opts.Agent, opts.PubSub)

	// 如果有恢复会话 ID
	if opts.ResumeID != "" {
		ctx := context.Background()
		opts.PubSub.Publish(ctx, pubsub.Event{
			Type:    pubsub.EventTypeSessionResume,
			Payload: opts.ResumeID,
		})
	}

	// 创建 Bubble Tea 程序
	// 启用备用屏幕，以便支持滚动查看历史消息
	p := tea.NewProgram(model, tea.WithAltScreen())

	// 运行程序
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// RunWithInput 以非交互模式运行，处理单次输入
func RunWithInput(opts Options, input string) error {
	ctx := context.Background()

	// 发送用户输入
	opts.PubSub.Publish(ctx, pubsub.Event{
		Type:    pubsub.EventTypeUserInput,
		Payload: input,
	})

	// 调用 Agent 处理
	if err := opts.Agent.ProcessUserInput(ctx, input); err != nil {
		return fmt.Errorf("error processing input: %w", err)
	}

	return nil
}

// RunSDKMode 以 SDK 模式运行（JSON 协议）
func RunSDKMode(opts Options) error {
	// SDK 模式通过 stdin/stdout 进行 JSON 通信
	decoder := NewJSONDecoder(os.Stdin)
	encoder := NewJSONEncoder(os.Stdout)

	for {
		// 读取请求
		var req SDKRequest
		if err := decoder.Decode(&req); err != nil {
			return fmt.Errorf("error decoding request: %w", err)
		}

		// 处理请求
		resp := handleSDKRequest(opts, &req)

		// 发送响应
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("error encoding response: %w", err)
		}
	}
}

// SDKRequest SDK 请求
type SDKRequest struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// SDKResponse SDK 响应
type SDKResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *SDKError   `json:"error,omitempty"`
}

// SDKError SDK 错误
type SDKError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSONDecoder JSON 解码器
type JSONDecoder struct {
	r *os.File
}

// NewJSONDecoder 创建新的 JSON 解码器
func NewJSONDecoder(r *os.File) *JSONDecoder {
	return &JSONDecoder{r: r}
}

// Decode 解码
func (d *JSONDecoder) Decode(v interface{}) error {
	// 简化实现，实际使用 encoding/json
	return nil
}

// JSONEncoder JSON 编码器
type JSONEncoder struct {
	w *os.File
}

// NewJSONEncoder 创建新的 JSON 编码器
func NewJSONEncoder(w *os.File) *JSONEncoder {
	return &JSONEncoder{w: w}
}

// Encode 编码
func (e *JSONEncoder) Encode(v interface{}) error {
	// 简化实现，实际使用 encoding/json
	return nil
}

// handleSDKRequest 处理 SDK 请求
func handleSDKRequest(opts Options, req *SDKRequest) *SDKResponse {
	switch req.Method {
	case "sendMessage":
		content, _ := req.Params["content"].(string)
		ctx := context.Background()

		opts.PubSub.Publish(ctx, pubsub.Event{
			Type:    pubsub.EventTypeUserInput,
			Payload: content,
		})

		err := opts.Agent.ProcessUserInput(ctx, content)
		if err != nil {
			return &SDKResponse{
				ID: req.ID,
				Error: &SDKError{
					Code:    -1,
					Message: err.Error(),
				},
			}
		}

		return &SDKResponse{
			ID:     req.ID,
			Result: map[string]string{"status": "ok"},
		}

	case "getStatus":
		return &SDKResponse{
			ID: req.ID,
			Result: map[string]interface{}{
				"status": "ready",
				"model":  opts.Config.Model,
			},
		}

	default:
		return &SDKResponse{
			ID: req.ID,
			Error: &SDKError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}
