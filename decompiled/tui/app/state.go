// Package app TUI 应用状态定义
package app

// AppState TUI 应用状态枚举
type AppState int

const (
	// StateReady 等待用户输入
	StateReady AppState = iota
	// StateProcessing Agent 正在处理（包括流式输出和工具执行）
	StateProcessing
	// StateError 错误状态
	StateError
	// StateQuitting 退出中
	StateQuitting
)

// String 返回状态的字符串表示
func (s AppState) String() string {
	switch s {
	case StateReady:
		return "Ready"
	case StateProcessing:
		return "Processing"
	case StateError:
		return "Error"
	case StateQuitting:
		return "Quitting"
	default:
		return "Unknown"
	}
}

// StatusText 返回用于状态栏显示的文本
func (s AppState) StatusText() string {
	switch s {
	case StateReady:
		return "Ready"
	case StateProcessing:
		return "Thinking..."
	case StateError:
		return "Error"
	case StateQuitting:
		return "Quitting..."
	default:
		return "Unknown"
	}
}
