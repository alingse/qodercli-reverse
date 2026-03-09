// Package app TUI 应用主入口
// 1:1 还原原版 qodercli TUI 架构
package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/interaction/editor"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/messages"
)

// appModel TUI 应用模型 - 对应原版的 appModel
type appModel struct {
	// 核心组件
	editor  *editor.EditorComponent
	msgView *messages.MessageView

	// 配置
	config *config.Config

	// Agent
	agent *agent.Agent

	// 事件系统
	pubsub *pubsub.PubSub

	// 尺寸
	width  int
	height int

	// 状态
	quitting      bool
	processing    bool
	showHelp      bool
	errorMsg      string
	sessionActive bool

	// 状态栏
	status      string
	model       string
	tokenUsage  *types.TokenUsage
}

// New 创建新的 TUI 应用模型
func New(cfg *config.Config, ag *agent.Agent, ps *pubsub.PubSub) *appModel {
	m := &appModel{
		config:        cfg,
		agent:         ag,
		pubsub:        ps,
		sessionActive: true,
		status:        "Ready",
		model:         cfg.Model,
	}

	// 初始化子组件
	m.editor = editor.NewEditorComponent(ps)
	m.msgView = messages.NewMessageView()

	// 设置事件订阅
	m.setupSubscriptions()

	return m
}

// Init 初始化 Bubble Tea 程序
func (m appModel) Init() tea.Cmd {
	log.Debug("Initializing TUI components")
	return tea.Batch(
		m.editor.Init(),
		m.msgView.Init(),
		func() tea.Msg {
			return InitMsg{}
		},
	)
}

// Update 处理消息更新
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 全局快捷键处理
		if cmd := m.handleGlobalKeys(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()

	case InitMsg:
		log.Info("TUI session started")
		m.pubsub.Publish(context.Background(), pubsub.Event{
			Type:    pubsub.EventTypeSessionStart,
			Payload: m.config,
		})

	case editor.SendMsg:
		// 处理编辑器发送的消息
		m.processing = true
		m.status = "Thinking..."
		cmds = append(cmds, m.sendToAgent(msg.Content))

	case ResponseMsg:
		// 处理 Agent 响应
		m.processing = false
		if msg.Done {
			m.status = "Ready"
		}
		m.msgView.AppendToLastMessage(msg.Content)

	case ToolCallMsg:
		// 显示工具调用
		m.status = "Using tool..."
		m.msgView.AddToolCall(msg.Name, msg.Arguments)

	case ToolResultMsg:
		// 显示工具结果
		m.msgView.AddToolResult(msg.Name, msg.Result, msg.Error)
		m.status = "Ready"

	case ErrorMsg:
		m.errorMsg = msg.Err.Error()
		m.processing = false
		m.status = "Error"
		log.Error("TUI error: %v", msg.Err)
		m.msgView.AddError(m.errorMsg)
	}

	// 更新子组件
	editorModel, editorCmd := m.editor.Update(msg)
	m.editor = editorModel.(*editor.EditorComponent)
	if editorCmd != nil {
		cmds = append(cmds, editorCmd)
	}

	msgViewModel, msgViewCmd := m.msgView.Update(msg)
	m.msgView = msgViewModel.(*messages.MessageView)
	if msgViewCmd != nil {
		cmds = append(cmds, msgViewCmd)
	}

	return m, tea.Batch(cmds...)
}

// handleGlobalKeys 处理全局快捷键
func (m *appModel) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	keyStr := msg.String()

	switch keyStr {
	// 滚动快捷键
	case "ctrl+shift+end":
		m.msgView.ScrollToBottom()
		return nil
	case "ctrl+shift+home":
		m.msgView.ScrollToTop()
		return nil
	case "pgup":
		m.msgView.PageUp()
		return nil
	case "pgdown":
		m.msgView.PageDown()
		return nil
	case "ctrl+u":
		m.msgView.HalfPageUp()
		return nil
	case "ctrl+d":
		m.msgView.HalfPageDown()
		return nil
	}

	// 其他全局快捷键
	switch msg.Type {
	case tea.KeyF1:
		m.showHelp = !m.showHelp
		return nil
	case tea.KeyCtrlC:
		if m.processing {
			// 取消当前操作
			return nil
		}
		m.quitting = true
		return tea.Quit
	}

	return nil
}

// View 渲染视图
func (m appModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.showHelp {
		return m.renderHelp()
	}

	// 布局：消息视图 + 编辑器 + 状态栏
	// 计算各部分高度
	editorHeight := 5
	statusHeight := 1
	msgViewHeight := m.height - editorHeight - statusHeight - 2

	if msgViewHeight < 10 {
		msgViewHeight = 10
	}

	// 设置组件尺寸（如果变化）
	if m.width > 0 {
		m.msgView.SetSize(m.width, msgViewHeight)
		m.editor.SetSize(m.width, editorHeight)
	}

	// 组合视图
	var content string
	content = m.msgView.View()

	// 错误消息
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Padding(0, 1)
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			errorStyle.Render("Error: "+m.errorMsg),
		)
	}

	// 编辑器
	editorView := m.editor.View()

	// 状态栏
	statusView := m.renderStatusBar()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		editorView,
		statusView,
	)
}

// renderStatusBar 渲染状态栏
func (m appModel) renderStatusBar() string {
	// 左侧状态
	var statusColor string
	switch m.status {
	case "Ready":
		statusColor = "82"
	case "Thinking...", "Using tool...":
		statusColor = "135"
	case "Error":
		statusColor = "203"
	default:
		statusColor = "248"
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Width(20)

	statusStr := statusStyle.Render("● " + m.status)

	// 中间模型
	modelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Width(m.width - 45)

	modelStr := modelStyle.Render(m.model)

	// 右侧 Token
	tokenStr := ""
	if m.tokenUsage != nil {
		total := m.tokenUsage.InputTokens + m.tokenUsage.OutputTokens
		tokenStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))
		tokenStr = tokenStyle.Render(fmt.Sprintf("Tokens: %d", total))
	}

	// 组合
	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("237")).
		Padding(0, 1).
		Width(m.width)

	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusStr,
		modelStr,
		tokenStr,
	)

	return barStyle.Render(content)
}

// resizeComponents 调整组件尺寸
func (m *appModel) resizeComponents() {
	editorHeight := 5
	statusHeight := 1
	msgViewHeight := m.height - editorHeight - statusHeight - 2

	if msgViewHeight < 10 {
		msgViewHeight = 10
	}

	m.editor.SetSize(m.width, editorHeight)
	m.msgView.SetSize(m.width, msgViewHeight)
}

// sendToAgent 发送消息给 Agent
func (m *appModel) sendToAgent(input string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		log.Debug("Sending user input to agent: %s", input)

		// 添加用户消息到视图
		m.msgView.AddUserMessage(input)

		// 调用 Agent
		err := m.agent.ProcessUserInput(ctx, input)
		if err != nil {
			log.Error("Agent processing error: %v", err)
			return ErrorMsg{Err: err}
		}

		log.Debug("Agent processing completed")
		return ResponseMsg{Done: true}
	}
}

// setupSubscriptions 设置事件订阅
func (m *appModel) setupSubscriptions() {
	log.Debug("Setting up event subscriptions")

	// 订阅 Agent 响应
	m.pubsub.Subscribe(pubsub.EventTypeAgentResponse, func(ctx context.Context, event pubsub.Event) {
		if content, ok := event.Payload.(string); ok {
			log.Debug("Received agent response: %d chars", len(content))
			m.msgView.AppendToLastMessage(content)
		}
	})

	// 订阅工具事件
	m.pubsub.Subscribe(pubsub.EventTypeToolStart, func(ctx context.Context, event pubsub.Event) {
		if info, ok := event.Payload.(map[string]string); ok {
			log.Debug("Tool started: %s", info["name"])
			m.msgView.AddToolCall(info["name"], info["arguments"])
		}
	})

	// 订阅 Token 使用事件
	m.pubsub.Subscribe(pubsub.EventTypeTokenUsage, func(ctx context.Context, event pubsub.Event) {
		if usage, ok := event.Payload.(*types.TokenUsage); ok {
			log.Debug("Token usage: input=%d, output=%d", usage.InputTokens, usage.OutputTokens)
			m.tokenUsage = usage
		}
	})

	// 订阅错误事件
	m.pubsub.Subscribe(pubsub.EventTypeAgentError, func(ctx context.Context, event pubsub.Event) {
		if err, ok := event.Payload.(error); ok {
			log.Error("Agent error: %v", err)
			m.msgView.AddError(err.Error())
		}
	})
}

// renderHelp 渲染帮助界面
func (m appModel) renderHelp() string {
	help := `
Qoder CLI Help

Keyboard Shortcuts:
  Ctrl+C      Quit (or cancel if processing)
  Enter       Send message
  Shift+Enter Insert new line
  Tab         Insert spaces
  Esc         Cancel / Blur
  F1          Toggle help
  PgUp/PgDown Scroll messages
  Ctrl+U/D    Half page up/down
  Ctrl+Shift+Home/End  Scroll to top/bottom

Commands:
  Type your message and press Enter to send
`
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(m.width - 4)

	return style.Render(help)
}

// 消息类型定义

// InitMsg 初始化消息
type InitMsg struct{}

// ErrorMsg 错误消息
type ErrorMsg struct {
	Err error
}

// ResponseMsg Agent 响应消息
type ResponseMsg struct {
	Content string
	Done    bool
}

// ToolCallMsg 工具调用消息
type ToolCallMsg struct {
	Name      string
	Arguments string
}

// ToolResultMsg 工具结果消息
type ToolResultMsg struct {
	Name   string
	Result string
	Error  error
}
