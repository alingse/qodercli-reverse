// Package app TUI 应用主入口
// 遵循 Bubble Tea 架构：所有状态变更必须通过 Update 循环
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/interaction/editor"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/messages"
	"github.com/alingse/qodercli-reverse/decompiled/version"
)

// appModel TUI 应用模型 - 遵循 Bubble Tea 响应式架构
type appModel struct {
	// 核心组件
	editor  *editor.EditorComponent
	msgView *messages.MessageView

	// 配置
	config *config.Config

	// Agent
	agent *agent.Agent

	// 事件系统（保留用于其他模块通信）
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
	status     string
	model      string
	tokenUsage *types.TokenUsage

	// 流式输出追踪（使用 atomic 保证并发安全）
	lastResponseLen int32 // 已输出的响应长度，用于增量显示

	// 事件通道 - 用于从 Agent 回调安全地发送消息到 Update 循环
	// 这是实现真正流式 UI 的关键：Agent 在后台运行，通过 channel 发送事件
	eventChan chan tea.Msg

	// 工具调用 ID 到名称的映射，用于跟踪工具执行状态
	toolCallMap map[string]string
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
		eventChan:     make(chan tea.Msg, 100), // 缓冲通道，避免阻塞 Agent
		toolCallMap:   make(map[string]string), // 工具调用 ID 到名称的映射
	}

	// 初始化子组件
	m.editor = editor.NewEditorComponent(ps)
	m.msgView = messages.NewMessageView()

	// 设置 Agent 回调，将流式输出转换为 channel 消息
	// 这是关键：回调在 Agent goroutine 中执行，通过 channel 发送到 Update 循环
	m.setupAgentCallbacks()

	return m
}

// Init 初始化 Bubble Tea 程序
func (m appModel) Init() tea.Cmd {
	log.Debug("Initializing TUI components")
	
	// 设置默认尺寸（在收到 WindowSizeMsg 之前使用）
	m.setDefaultSize()
	
	return tea.Batch(
		m.editor.Init(),
		m.msgView.Init(),
		m.waitForEvents(), // 启动事件监听，等待 Agent 回调
		func() tea.Msg {
			return InitMsg{}
		},
	)
}

// setDefaultSize 设置默认尺寸（在收到 WindowSizeMsg 之前使用）
func (m *appModel) setDefaultSize() {
	// 使用合理的默认值
	m.width = 80
	m.height = 24
	m.resizeComponents()
}

// Update 处理消息更新 - 这是唯一可以修改状态的地方
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

		// 添加欢迎消息
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "unknown"
			log.Warn("Failed to get current working directory: %v", err)
		}

		welcomeMsg := &messages.WelcomeMessage{
			Version: version.GetVersion(),
			Cwd:     cwd,
			MsgTime: time.Now(),
		}
		m.msgView.AddMessage(welcomeMsg)

		m.pubsub.Publish(context.Background(), pubsub.Event{
			Type:    pubsub.EventTypeSessionStart,
			Payload: m.config,
		})

	case editor.SendMsg:
		// 处理编辑器发送的消息
		m.processing = true
		m.status = "Thinking..."
		// 先添加用户消息，然后启动 Agent（非阻塞）
		cmds = append(cmds, m.handleUserInput(msg.Content))

	// === Agent 流式事件处理 ===
	// 这些消息从 Agent 回调通过 eventChan 发送过来

	case AgentStreamMsg:
		// 流式内容增量 - 追加到助手消息
		// 如果还没有助手消息（第一个字符），先创建
		if m.msgView.GetLastMessage() == nil || m.msgView.GetLastMessage().Type() != messages.MsgTypeAssistant {
			m.msgView.AddAssistantMessage(msg.Content)
		} else {
			m.msgView.AppendToLastMessage(msg.Content)
		}
		// 继续监听下一个事件
		cmds = append(cmds, m.waitForEvents())

	case AgentToolStartMsg:
		// 工具调用开始
		m.status = "Using tool..."
		// 保存工具 ID 到名称的映射
		m.toolCallMap[msg.ID] = msg.Name
		// 添加待执行的工具调用（所有工具统一使用 PendingTool 显示）
		m.msgView.AddPendingTool(msg.ID, msg.Name, msg.Arguments)
		cmds = append(cmds, m.waitForEvents())

	case AgentToolResultMsg:
		// 工具调用结果
		// 根据工具调用 ID 查找工具名称
		toolName := m.toolCallMap[msg.ToolCallID]
		if toolName == "" {
			toolName = msg.ToolCallID // 备用：使用 ID 作为名称
		}
		
		// 特殊处理 TodoWrite 工具
		if toolName == "TodoWrite" && !msg.IsError {
			// 解析工具参数中的 todos
			var params struct {
				Todos []state.Todo `json:"todos"`
			}
			if err := json.Unmarshal([]byte(msg.Content), &params); err == nil && len(params.Todos) > 0 {
				// 获取旧的 todos
				oldTodos := m.msgView.GetTodoList()
				// 更新 Todo 列表显示
				m.msgView.UpdateTodoList(params.Todos, oldTodos)
			} else {
				// 如果解析失败，回退到普通工具显示
				m.msgView.CompleteTool(msg.ToolCallID, toolName, msg.Content, msg.IsError)
			}
		} else {
			// 更新工具调用状态（所有工具统一处理）
			m.msgView.CompleteTool(msg.ToolCallID, toolName, msg.Content, msg.IsError)
		}
		m.status = "Thinking..."
		cmds = append(cmds, m.waitForEvents())

	case AgentErrorMsg:
		// Agent 错误
		m.errorMsg = msg.Err.Error()
		m.processing = false
		m.status = "Error"
		log.Error("Agent error: %v", msg.Err)
		m.msgView.AddError(m.errorMsg)
		// 重置状态
		atomic.StoreInt32(&m.lastResponseLen, 0)
		cmds = append(cmds, m.waitForEvents())

	case AgentFinishMsg:
		// Agent 完成
		m.processing = false
		m.status = "Ready"
		log.Debug("Agent finished with reason: %s", msg.Reason)
		// 重置计数器，为下一次对话做准备
		atomic.StoreInt32(&m.lastResponseLen, 0)
		cmds = append(cmds, m.waitForEvents())

	case ResponseMsg:
		// 保留用于兼容性
		m.processing = false
		if msg.Done {
			m.status = "Ready"
		}

	case ToolCallMsg:
		// 保留用于兼容性
		m.status = "Using tool..."
		m.msgView.AddToolCall(msg.Name, msg.Arguments)

	case ToolResultMsg:
		// 保留用于兼容性
		m.msgView.AddToolResult(msg.Name, msg.Result, msg.Error)
		m.status = "Ready"

	case ErrorMsg:
		// 保留用于兼容性
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

// handleUserInput 处理用户输入 - 非阻塞启动 Agent
func (m *appModel) handleUserInput(input string) tea.Cmd {
	return func() tea.Msg {
		// 重置流式输出计数器
		atomic.StoreInt32(&m.lastResponseLen, 0)

		// 添加用户消息到视图
		m.msgView.AddUserMessage(input)

		// 注意：不再预创建空的助手消息
		// 而是在收到第一个流式内容时才创建，避免提前显示 ⏺

		// 在后台 goroutine 中运行 Agent，不阻塞 Update 循环
		go func() {
			ctx := context.Background()
			log.Debug("Starting Agent processing in background goroutine")

			err := m.agent.ProcessUserInput(ctx, input)
			if err != nil {
				log.Error("Agent processing error: %v", err)
				// 错误通过回调机制发送，这里不需要额外处理
			}

			log.Debug("Agent processing goroutine completed")
		}()

		// 立即返回，不阻塞 UI
		return StartedMsg{}
	}
}

// waitForEvents 监听 Agent 事件通道
// 这是关键方法：它返回一个 tea.Cmd，当 channel 有消息时会触发 Update
func (m *appModel) waitForEvents() tea.Cmd {
	return func() tea.Msg {
		// 阻塞等待事件，但这是在 Bubble Tea 的 runtime 管理下的阻塞
		// 不会冻结 UI，因为 Bubble Tea 会调度其他命令
		return <-m.eventChan
	}
}

// setupAgentCallbacks 设置 Agent 回调
// 回调在 Agent 的 goroutine 中执行，通过 channel 发送到 Update 循环
func (m *appModel) setupAgentCallbacks() {
	log.Debug("Setting up agent callbacks with channel-based event delivery")

	m.agent.SetCallbacks(
		// onMessage - 处理流式消息输出（只发送增量）
		func(msg *types.Message) {
			// 构建当前完整文本内容
			var fullContent strings.Builder
			for _, part := range msg.Content {
				if part.Type == "text" && part.Text != "" {
					fullContent.WriteString(part.Text)
				}
			}

			fullStr := fullContent.String()
			runes := []rune(fullStr)

			// 只发送新增的文本部分
			lastLen := atomic.LoadInt32(&m.lastResponseLen)
			if len(runes) > int(lastLen) {
				newRunes := runes[lastLen:]
				atomic.StoreInt32(&m.lastResponseLen, int32(len(runes)))

				// 发送到 channel，触发 Update 循环
				select {
				case m.eventChan <- AgentStreamMsg{Content: string(newRunes)}:
					log.Debug("Sent stream event: %d chars", len(newRunes))
				default:
					log.Warn("Event channel full, dropping message")
				}
			}
		},
		// onToolCall - 处理工具调用开始
		func(call *types.ToolCall) {
			select {
			case m.eventChan <- AgentToolStartMsg{
				ID:        call.ID,
				Name:      call.Name,
				Arguments: call.Arguments,
			}:
				log.Debug("Sent tool start event: %s (ID: %s)", call.Name, call.ID)
			default:
				log.Warn("Event channel full, dropping tool start")
			}
		},
		// onToolResult - 处理工具调用结果
		func(result *types.ToolResult) {
			select {
			case m.eventChan <- AgentToolResultMsg{
				ToolCallID: result.ToolCallID, // 使用 ToolCallID 作为标识
				Content:    result.Content,
				IsError:    result.IsError,
			}:
				log.Debug("Sent tool result event: %s", result.ToolCallID)
			default:
				log.Warn("Event channel full, dropping tool result")
			}
		},
		// onError - 处理错误
		func(err error) {
			select {
			case m.eventChan <- AgentErrorMsg{Err: err}:
				log.Debug("Sent error event: %v", err)
			default:
				log.Warn("Event channel full, dropping error")
			}
		},
		// onFinish - 处理完成
		func(reason types.FinishReason) {
			select {
			case m.eventChan <- AgentFinishMsg{Reason: reason}:
				log.Debug("Sent finish event: %s", reason)
			default:
				log.Warn("Event channel full, dropping finish")
			}
		},
	)
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

	// 计算布局尺寸
	// 编辑器高度：根据内容自适应（最少3行，最多7行）
	editorHeight := m.editor.GetPreferredHeight()
	if editorHeight < 3 {
		editorHeight = 3
	}
	if editorHeight > 7 {
		editorHeight = 7
	}
	statusHeight := 1
	msgViewHeight := m.height - editorHeight - statusHeight - 1
	if msgViewHeight < 5 {
		msgViewHeight = 5
	}

	// 设置组件尺寸（必须在 View 之前设置）
	if m.width > 0 {
		m.msgView.SetSize(m.width, msgViewHeight)
		m.editor.SetSize(m.width, editorHeight)
	}

	// 组合视图：消息视图 + 编辑器 + 状态栏
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
	// 编辑器高度根据内容动态调整（最少3行，最多7行）
	editorHeight := m.editor.GetPreferredHeight()
	if editorHeight < 3 {
		editorHeight = 3
	}
	if editorHeight > 7 {
		editorHeight = 7
	}

	statusHeight := 1
	// 减去额外的边距/边框空间
	msgViewHeight := m.height - editorHeight - statusHeight - 2

	if msgViewHeight < 5 {
		msgViewHeight = 5
	}

	m.editor.SetSize(m.width, editorHeight)
	m.msgView.SetSize(m.width, msgViewHeight)
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

// StartedMsg 表示 Agent 已开始处理
type StartedMsg struct{}

// ErrorMsg 错误消息
type ErrorMsg struct {
	Err error
}

// ResponseMsg Agent 响应消息（保留用于兼容性）
type ResponseMsg struct {
	Content string
	Done    bool
}

// ToolCallMsg 工具调用消息（保留用于兼容性）
type ToolCallMsg struct {
	Name      string
	Arguments string
}

// ToolResultMsg 工具结果消息（保留用于兼容性）
type ToolResultMsg struct {
	Name   string
	Result string
	Error  error
}

// === Agent 流式事件消息类型 ===
// 这些消息从 Agent 回调通过 eventChan 发送到 Update 循环

// AgentStreamMsg 流式内容增量
type AgentStreamMsg struct {
	Content string
}

// AgentToolStartMsg 工具调用开始
type AgentToolStartMsg struct {
	ID        string // 工具调用唯一 ID
	Name      string // 工具名称
	Arguments string // 工具参数
}

// AgentToolResultMsg 工具调用结果
type AgentToolResultMsg struct {
	ToolCallID string // 工具调用唯一 ID
	Content    string // 执行结果内容
	IsError    bool   // 是否执行出错
}

// AgentErrorMsg Agent 错误
type AgentErrorMsg struct {
	Err error
}

// AgentFinishMsg Agent 完成
type AgentFinishMsg struct {
	Reason types.FinishReason
}
