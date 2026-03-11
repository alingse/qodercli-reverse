// Package app TUI 应用启动器
package app

import (
	"context"
	"fmt"

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
	Debug    bool // 启用调试模式和流式输出追踪
}

// Run 启动 TUI 应用
func Run(opts Options) error {
	// 创建应用模型
	model := New(opts.Config, opts.Agent, opts.PubSub, opts.Debug)

	// 如果有恢复会话 ID
	if opts.ResumeID != "" {
		ctx := context.Background()
		opts.PubSub.Publish(ctx, pubsub.Event{
			Type:    pubsub.EventTypeSessionResume,
			Payload: opts.ResumeID,
		})
	}

	// 创建 Bubble Tea 程序
	// 不使用 AltScreen，让输出可以随终端滚动
	p := tea.NewProgram(model)

	// 运行程序
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
