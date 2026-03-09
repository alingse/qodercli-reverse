// Package app TUI 应用主入口
// 基于 Bubble Tea 框架实现
package app

import (
	"context"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/chat"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/editor"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/messages"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/status"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode TUI 运行模式
type Mode int

const (
	ModeNormal Mode = iota
	ModeVim
	ModeHelp
)

// Model TUI 应用模型
type Model struct {
	// 配置
	config *config.Config
	mode   Mode

	// 子组件
	messageList *messages.Component
	editor      *editor.Component
	statusBar   *status.Component
	chatView    *chat.Component

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
}

// InitMsg 初始化消息
type InitMsg struct{}

// ErrorMsg 错误消息
type ErrorMsg struct {
	Error error
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

// New 创建新的 TUI 应用模型
func New(cfg *config.Config, ag *agent.Agent, ps *pubsub.PubSub) *Model {
	m := &Model{
		config:        cfg,
		mode:          ModeNormal,
		agent:         ag,
		pubsub:        ps,
		sessionActive: true,
	}

	// 初始化子组件
	m.messageList = messages.New()
	m.editor = editor.New()
	m.statusBar = status.New()
	m.chatView = chat.New()

	// 设置事件订阅
	m.setupSubscriptions()

	return m
}

// Init 初始化 Bubble Tea 程序
func (m Model) Init() tea.Cmd {
	log.Debug("Initializing TUI components")
	return tea.Batch(
		m.messageList.Init(),
		m.editor.Init(),
		m.statusBar.Init(),
		m.chatView.Init(),
		func() tea.Msg {
			return InitMsg{}
		},
	)
}

// Update 处理消息更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := m.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// 调整子组件尺寸
		cmds = append(cmds, m.resizeComponents())

	case InitMsg:
		// 初始化完成
		log.Info("TUI session started")
		m.pubsub.Publish(context.Background(), pubsub.Event{
			Type:    pubsub.EventTypeSessionStart,
			Payload: m.config,
		})

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.processing = false
		log.Error("TUI error: %v", msg.Error)

	case ResponseMsg:
		if msg.Done {
			m.processing = false
		}
		// 更新消息列表
		cmds = append(cmds, m.messageList.UpdateContent(msg.Content))

	case ToolCallMsg:
		// 显示工具调用
		cmds = append(cmds, m.chatView.ShowToolCall(msg.Name, msg.Arguments))

	case ToolResultMsg:
		// 显示工具结果
		cmds = append(cmds, m.chatView.ShowToolResult(msg.Name, msg.Result, msg.Error))

	case tea.Cmd:
		cmds = append(cmds, msg)
	}

	// 更新子组件
	msgListModel, msgCmd := m.messageList.Update(msg)
	m.messageList = msgListModel.(*messages.Component)
	if msgCmd != nil {
		cmds = append(cmds, msgCmd)
	}

	editorModel, editorCmd := m.editor.Update(msg)
	m.editor = editorModel.(*editor.Component)
	if editorCmd != nil {
		cmds = append(cmds, editorCmd)
	}

	statusModel, statusCmd := m.statusBar.Update(msg)
	m.statusBar = statusModel.(*status.Component)
	if statusCmd != nil {
		cmds = append(cmds, statusCmd)
	}

	chatModel, chatCmd := m.chatView.Update(msg)
	m.chatView = chatModel.(*chat.Component)
	if chatCmd != nil {
		cmds = append(cmds, chatCmd)
	}

	return m, tea.Batch(cmds...)
}

// View 渲染视图
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.showHelp {
		return m.renderHelp()
	}

	// 布局：消息列表 + 聊天视图 + 编辑器 + 状态栏
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		m.chatView.View(),
	)

	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		mainContent = lipgloss.JoinVertical(
			lipgloss.Left,
			mainContent,
			errorStyle.Render("Error: "+m.errorMsg),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		m.editor.View(),
		m.statusBar.View(),
	)
}

// handleKeyMsg 处理键盘消息
func (m *Model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	// 先检查是否显示帮助
	if m.showHelp && msg.Type != tea.KeyF1 {
		m.showHelp = false
		return nil
	}

	switch msg.String() {
	// 滚动快捷键
	case "ctrl+shift+end":
		m.chatView.ScrollToBottom()
		return nil
	case "ctrl+shift+home":
		m.chatView.ScrollToTop()
		return nil
	case "pgup":
		m.chatView.PageUp()
		return nil
	case "pgdown":
		m.chatView.PageDown()
		return nil
	case "ctrl+u":
		m.chatView.HalfPageUp()
		return nil
	case "ctrl+d":
		m.chatView.HalfPageDown()
		return nil
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		m.quitting = true
		return tea.Quit

	case tea.KeyEsc:
		if m.mode == ModeHelp {
			m.mode = ModeNormal
			return nil
		}

	case tea.KeyF1:
		m.showHelp = !m.showHelp
		return nil

	case tea.KeyEnter:
		if m.processing {
			return nil
		}
		input := m.editor.GetInput()
		if input == "" {
			return nil
		}

		// 检查斜杠命令
		if input[0] == '/' {
			return m.handleSlashCommand(input)
		}

		// 发送消息给 Agent
		m.processing = true
		m.editor.Clear()
		return m.sendToAgent(input)
	}

	return nil
}

// handleSlashCommand 处理斜杠命令
func (m *Model) handleSlashCommand(cmd string) tea.Cmd {
	switch cmd {
	case "/quit", "/q":
		m.quitting = true
		return tea.Quit

	case "/clear":
		m.chatView.Clear()
		m.editor.Clear()
		return nil

	case "/help":
		m.showHelp = true
		return nil

	case "/vim":
		if m.mode == ModeVim {
			m.mode = ModeNormal
		} else {
			m.mode = ModeVim
		}
		return nil

	case "/compact":
		m.pubsub.Publish(context.Background(), pubsub.Event{
			Type: pubsub.EventTypeSessionCompact,
		})
		return nil

	case "/status":
		// 显示状态
		return nil

	default:
		m.errorMsg = fmt.Sprintf("Unknown command: %s", cmd)
		return nil
	}
}

// sendToAgent 发送消息给 Agent
func (m *Model) sendToAgent(input string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		log.Debug("Sending user input to agent: %s", input)

		// 添加用户消息
		m.pubsub.Publish(ctx, pubsub.Event{
			Type:    pubsub.EventTypeUserInput,
			Payload: input,
		})

		// 调用 Agent
		err := m.agent.ProcessUserInput(ctx, input)
		if err != nil {
			log.Error("Agent processing error: %v", err)
			return ErrorMsg{Error: err}
		}

		log.Debug("Agent processing completed")
		return ResponseMsg{Done: true}
	}
}

// resizeComponents 调整组件尺寸
func (m *Model) resizeComponents() tea.Cmd {
	return func() tea.Msg {
		// 计算各组件尺寸
		editorHeight := 3
		statusHeight := 1
		chatHeight := m.height - editorHeight - statusHeight - 2

		if chatHeight < 10 {
			chatHeight = 10
		}

		m.chatView.SetSize(m.width, chatHeight)
		m.editor.SetSize(m.width, editorHeight)
		m.statusBar.SetSize(m.width, statusHeight)

		return nil
	}
}

// setupSubscriptions 设置事件订阅
func (m *Model) setupSubscriptions() {
	log.Debug("Setting up event subscriptions")

	// 订阅 Agent 事件
	m.pubsub.Subscribe(pubsub.EventTypeAgentResponse, func(ctx context.Context, event pubsub.Event) {
		if content, ok := event.Payload.(string); ok {
			log.Debug("Received agent response: %d chars", len(content))
			m.chatView.AppendContent(content)
		}
	})

	// 订阅工具事件
	m.pubsub.Subscribe(pubsub.EventTypeToolStart, func(ctx context.Context, event pubsub.Event) {
		if info, ok := event.Payload.(map[string]string); ok {
			log.Debug("Tool started: %s", info["name"])
			m.chatView.ShowToolCall(info["name"], info["arguments"])
		}
	})

	// 订阅 Token 使用事件
	m.pubsub.Subscribe(pubsub.EventTypeTokenUsage, func(ctx context.Context, event pubsub.Event) {
		if usage, ok := event.Payload.(*types.TokenUsage); ok {
			log.Debug("Token usage: input=%d, output=%d", usage.InputTokens, usage.OutputTokens)
			m.statusBar.UpdateTokenUsage(usage)
		}
	})

	// 订阅错误事件
	m.pubsub.Subscribe(pubsub.EventTypeAgentError, func(ctx context.Context, event pubsub.Event) {
		if err, ok := event.Payload.(error); ok {
			log.Error("Agent error: %v", err)
			m.chatView.ShowError(err)
		}
	})
}

// renderHelp 渲染帮助界面
func (m Model) renderHelp() string {
	help := `
Qoder CLI Help

Keyboard Shortcuts:
  Ctrl+C    Quit
  Enter     Send message
  Esc       Cancel / Close help
  F1        Toggle help

Slash Commands:
  /clear      Clear conversation
  /compact    Compact context
  /help       Show this help
  /quit       Exit
  /vim        Toggle vim mode

Press Esc to close help.
`
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(m.width - 4)

	return style.Render(help)
}
