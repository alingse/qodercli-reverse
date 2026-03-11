// Package app TUI 应用主入口
// 遵循 Bubble Tea 架构：所有状态变更必须通过 Update 循环
package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
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
	msgView *messages.MessageView // 保留用于格式化消息

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
	status string
	model  string

	// === 回合管理 ===
	// 当前回合内容（在 View 中临时渲染）
	currentTurn *TurnContent

	// 是否有活跃的回合
	hasTurn bool

	// 事件通道
	eventChan chan tea.Msg

	// 回调序列化 channel - 确保 Agent 回调按顺序处理
	callbackChan chan func()
}

// === 回合管理数据结构 ===

// TurnStatus 回合状态
type TurnStatus int

const (
	TurnStatusStreaming TurnStatus = iota // 正在流式输出文本
	TurnStatusTooling                     // 正在执行工具
	TurnStatusCompleted                   // 回合完成
	TurnStatusError                       // 回合出错
)

// ToolCallStatus 工具调用状态
type ToolCallStatus int

const (
	ToolCallStatusPending ToolCallStatus = iota // 等待执行
	ToolCallStatusRunning                       // 执行中
	ToolCallStatusCompleted                     // 执行成功
	ToolCallStatusError                         // 执行失败
)

// TurnToolCall 表示回合中的一个工具调用
type TurnToolCall struct {
	ID        string
	Name      string
	Arguments string
	Output    string
	IsError   bool
	Status    ToolCallStatus
	StartTime time.Time
	EndTime   time.Time
}

// TurnContent 表示一个完整的 Agent 回合内容
type TurnContent struct {
	// 流式文本内容（Assistant 消息）
	StreamingText strings.Builder

	// 工具调用列表（按顺序）
	ToolCalls []TurnToolCall

	// 回合状态
	Status TurnStatus

	// 开始时间
	StartTime time.Time

	// 已持久化的文本长度
	PersistedTextLength int
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
		eventChan:     make(chan tea.Msg, 100),
		callbackChan:  make(chan func(), 1000), // 增大缓冲区避免阻塞
		hasTurn:       false,
		currentTurn:   nil,
	}

	// 初始化子组件
	m.editor = editor.NewEditorComponent(ps)
	m.msgView = messages.NewMessageView()

	// 启动回调序列化 goroutine - 确保 Agent 回调按顺序处理
	go m.serializeCallbacks()

	// 设置 Agent 回调
	m.setupAgentCallbacks()

	return m
}

// serializeCallbacks 序列化回调处理 goroutine
// 所有 Agent 回调先发送到这里，然后按顺序发送到 eventChan
func (m *appModel) serializeCallbacks() {
	for fn := range m.callbackChan {
		fn()
	}
}

// Init 初始化 Bubble Tea 程序
func (m appModel) Init() tea.Cmd {
	log.Debug("Initializing TUI components")
	m.setDefaultSize()

	return tea.Batch(
		m.editor.Init(),
		m.waitForEvents(),
		func() tea.Msg {
			return InitMsg{}
		},
	)
}

// setDefaultSize 设置默认尺寸
func (m *appModel) setDefaultSize() {
	m.width = 80
	m.height = 24
}

// Update 处理消息更新
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.handleGlobalKeys(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case InitMsg:
		// 获取当前目录
		cwd, _ := os.Getwd()
		welcome := &messages.WelcomeMessage{
			Version: version.GetVersion(),
			Cwd:     cwd,
			MsgTime: time.Now(),
		}
		cmds = append(cmds, tea.Printf(welcome.Render()))
		cmds = append(cmds, m.waitForEvents())

	case editor.SendMsg:
		m.processing = true
		m.status = "Thinking..."

		// 1. 将用户消息持久化到终端
		userMsg := &messages.UserMessage{
			Content: msg.Content,
			MsgTime: time.Now(),
		}
		cmds = append(cmds, tea.Printf(userMsg.Render()))

		// 2. 初始化新回合
		m.hasTurn = true
		m.currentTurn = &TurnContent{
			Status:    TurnStatusStreaming,
			StartTime: time.Now(),
		}

		// 3. 启动 Agent
		cmds = append(cmds, m.handleUserInput(msg.Content))

	case AgentStreamMsg:
		// 更新当前回合的流式文本（完整内容，不是增量）
		if m.hasTurn && m.currentTurn != nil {
			// 跳过已持久化的部分，只显示新增的文本
			if len(msg.Content) > m.currentTurn.PersistedTextLength {
				newContent := msg.Content[m.currentTurn.PersistedTextLength:]
				m.currentTurn.StreamingText.Reset()
				m.currentTurn.StreamingText.WriteString(newContent)
			} else {
				// 如果没有新内容，清空 StreamingText
				m.currentTurn.StreamingText.Reset()
			}
			m.currentTurn.Status = TurnStatusStreaming
		}
		cmds = append(cmds, m.waitForEvents())

	case AgentToolStartMsg:
		m.status = "Using tool..."

		if m.hasTurn && m.currentTurn != nil {
			// 1. 先持久化之前的流式文本（如果有）
			if m.currentTurn.StreamingText.Len() > 0 {
				content := m.currentTurn.StreamingText.String()
				assistantMsg := &messages.AssistantMessage{
					Content: content,
					MsgTime: time.Now(),
				}
				cmds = append(cmds, tea.Printf(assistantMsg.Render()))

				// 记录已持久化的文本长度
				m.currentTurn.PersistedTextLength += len(content)

				// 清空流式文本缓冲区
				m.currentTurn.StreamingText.Reset()
			}

			// 2. 添加工具调用到当前回合
			m.currentTurn.Status = TurnStatusTooling
			m.currentTurn.ToolCalls = append(m.currentTurn.ToolCalls, TurnToolCall{
				ID:        msg.ID,
				Name:      msg.Name,
				Arguments: msg.Arguments,
				Status:    ToolCallStatusRunning,
				StartTime: time.Now(),
			})
		}
		cmds = append(cmds, m.waitForEvents())

	case AgentToolResultMsg:
		if m.hasTurn && m.currentTurn != nil {
			// 查找对应的工具调用
			for i := range m.currentTurn.ToolCalls {
				if m.currentTurn.ToolCalls[i].ID == msg.ToolCallID {
					// 更新工具调用信息
					m.currentTurn.ToolCalls[i].Output = msg.Content
					m.currentTurn.ToolCalls[i].IsError = msg.IsError
					if msg.IsError {
						m.currentTurn.ToolCalls[i].Status = ToolCallStatusError
					} else {
						m.currentTurn.ToolCalls[i].Status = ToolCallStatusCompleted
					}
					m.currentTurn.ToolCalls[i].EndTime = time.Now()

					// 立即持久化这个工具调用
					toolMsg := &messages.ToolCallInfo{
						ID:        m.currentTurn.ToolCalls[i].ID,
						Name:      m.currentTurn.ToolCalls[i].Name,
						Arguments: m.currentTurn.ToolCalls[i].Arguments,
						Output:    m.currentTurn.ToolCalls[i].Output,
						IsError:   m.currentTurn.ToolCalls[i].IsError,
						Completed: true,
						MsgTime:   m.currentTurn.ToolCalls[i].EndTime,
					}
					cmds = append(cmds, tea.Printf(toolMsg.Render()))

					// 从 ToolCalls 中移除（因为已经持久化了）
					m.currentTurn.ToolCalls = append(
						m.currentTurn.ToolCalls[:i],
						m.currentTurn.ToolCalls[i+1:]...,
					)
					break
				}
			}
		}

		m.status = "Thinking..."
		cmds = append(cmds, m.waitForEvents())

	case AgentErrorMsg:
		m.errorMsg = msg.Err.Error()
		m.processing = false
		m.status = "Error"

		// 错误时持久化当前回合的部分内容
		if m.hasTurn && m.currentTurn != nil {
			// 持久化已有的流式文本
			if m.currentTurn.StreamingText.Len() > 0 {
				assistantMsg := &messages.AssistantMessage{
					Content: m.currentTurn.StreamingText.String(),
					MsgTime: time.Now(),
				}
				cmds = append(cmds, tea.Printf(assistantMsg.Render()))
			}

			// 持久化正在运行的工具调用（如果有）
			// 注意：已完成的工具调用应该已经在 ToolResult 时持久化了
			// 这里只需要持久化还在运行中的工具调用
			for _, toolCall := range m.currentTurn.ToolCalls {
				if toolCall.Status == ToolCallStatusRunning {
					toolMsg := &messages.ToolCallInfo{
						ID:        toolCall.ID,
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
						Output:    "Interrupted by error",
						IsError:   true,
						Completed: true,
						MsgTime:   time.Now(),
					}
					cmds = append(cmds, tea.Printf(toolMsg.Render()))
				}
			}
		}

		// 持久化错误消息
		errMsg := &messages.ErrorMessage{
			ErrStr:  m.errorMsg,
			MsgTime: time.Now(),
		}
		cmds = append(cmds, tea.Printf(errMsg.Render()))

		// 清理回合
		m.hasTurn = false
		m.currentTurn = nil
		cmds = append(cmds, m.waitForEvents())

	case AgentFinishMsg:
		// 回合完成：持久化剩余内容
		if m.hasTurn && m.currentTurn != nil {
			m.currentTurn.Status = TurnStatusCompleted

			// 只需要持久化剩余的流式文本（如果有）
			// 工具调用已经在 ToolResult 时持久化了
			if m.currentTurn.StreamingText.Len() > 0 {
				assistantMsg := &messages.AssistantMessage{
					Content: m.currentTurn.StreamingText.String(),
					MsgTime: time.Now(),
				}
				cmds = append(cmds, tea.Printf(assistantMsg.Render()))
			}

			// 清理当前回合
			m.hasTurn = false
			m.currentTurn = nil
		}

		m.processing = false
		m.status = "Ready"
		cmds = append(cmds, m.waitForEvents())

	case QuitMsg:
		m.quitting = true
		return m, tea.Quit
	}

	// 更新子组件
	editorModel, editorCmd := m.editor.Update(msg)
	m.editor = editorModel.(*editor.EditorComponent)
	if editorCmd != nil {
		cmds = append(cmds, editorCmd)
	}

	return m, tea.Batch(cmds...)
}

// handleUserInput 处理用户输入
func (m *appModel) handleUserInput(input string) tea.Cmd {
	return func() tea.Msg {
		if strings.HasPrefix(input, "/") {
			return m.handleCommand(input)
		}

		go func() {
			ctx := context.Background()
			m.agent.ProcessUserInput(ctx, input)
		}()
		return StartedMsg{}
	}
}

// handleCommand 处理斜杠命令
func (m *appModel) handleCommand(input string) tea.Msg {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	switch command {
	case "/quit", "/exit":
		return QuitMsg{}
	default:
		go func() {
			ctx := context.Background()
			m.agent.ProcessUserInput(ctx, input)
		}()
		return StartedMsg{}
	}
}

// waitForEvents 监听 Agent 事件通道
func (m *appModel) waitForEvents() tea.Cmd {
	return func() tea.Msg {
		return <-m.eventChan
	}
}

// setupAgentCallbacks 设置 Agent 回调
func (m *appModel) setupAgentCallbacks() {
	m.agent.SetCallbacks(
		// onMessage: 处理流式文本（完整内容）
		func(msg *types.Message) {
			var fullContent strings.Builder
			for _, part := range msg.Content {
				if part.Type == "text" && part.Text != "" {
					fullContent.WriteString(part.Text)
				}
			}
			fullStr := fullContent.String()

			// 通过序列化 channel 发送
			m.callbackChan <- func() {
				m.eventChan <- AgentStreamMsg{Content: fullStr}
			}
		},
		// onToolCall: 工具调用开始
		func(call *types.ToolCall) {
			c := *call
			m.callbackChan <- func() {
				m.eventChan <- AgentToolStartMsg{
					ID:        c.ID,
					Name:      c.Name,
					Arguments: c.Arguments,
				}
			}
		},
		// onToolResult: 工具调用结果
		func(result *types.ToolResult) {
			r := *result
			m.callbackChan <- func() {
				m.eventChan <- AgentToolResultMsg{
					ToolCallID: r.ToolCallID,
					Content:    r.Content,
					IsError:    r.IsError,
				}
			}
		},
		// onError: 错误处理
		func(err error) {
			e := err
			m.callbackChan <- func() {
				m.eventChan <- AgentErrorMsg{Err: e}
			}
		},
		// onFinish: 回合完成
		func(reason types.FinishReason) {
			r := reason
			m.callbackChan <- func() {
				m.eventChan <- AgentFinishMsg{Reason: r}
			}
		},
	)
}

// handleGlobalKeys 处理全局快捷键
func (m *appModel) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyF1:
		m.showHelp = !m.showHelp
		return nil
	case tea.KeyCtrlC:
		if m.processing {
			return nil
		}
		return tea.Quit
	}
	return nil
}

// View 渲染视图
func (m appModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}
	if m.quitting {
		return ""
	}

	editorHeight := m.editor.GetPreferredHeight()
	if editorHeight < 3 {
		editorHeight = 3
	}
	if editorHeight > 7 {
		editorHeight = 7
	}

	if m.width > 0 {
		m.editor.SetSize(m.width, editorHeight)
	}

	var sections []string

	// 1. 当前回合预览区（如果有活跃回合）
	if m.hasTurn && m.currentTurn != nil {
		sections = append(sections, m.renderCurrentTurn())
	}

	// 2. 空行分隔
	sections = append(sections, "")

	// 3. 编辑器
	sections = append(sections, m.editor.View())

	// 4. 状态栏
	sections = append(sections, m.renderStatusBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderCurrentTurn 渲染当前回合的完整内容（临时预览）
func (m appModel) renderCurrentTurn() string {
	if m.currentTurn == nil {
		return ""
	}

	var sections []string

	// 1. 渲染流式文本（如果有）
	if m.currentTurn.StreamingText.Len() > 0 {
		content := m.currentTurn.StreamingText.String()

		// 限制预览区的高度，避免占用太多屏幕空间
		// 最多显示 10 行，超出部分用 "..." 表示
		lines := strings.Split(content, "\n")
		maxLines := 10
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
			content = "...\n" + strings.Join(lines, "\n")
		}

		// 使用灰色斜体样式表示正在生成的内容
		previewStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Italic(true)

		// 直接渲染内容，添加 ⏺ 前缀
		sections = append(sections, previewStyle.Render(fmt.Sprintf("⏺ %s", strings.TrimSpace(content))))
	}

	// 2. 渲染工具调用（如果有）
	for _, toolCall := range m.currentTurn.ToolCalls {
		toolMsg := &messages.ToolCallInfo{
			ID:        toolCall.ID,
			Name:      toolCall.Name,
			Arguments: toolCall.Arguments,
			Output:    toolCall.Output,
			IsError:   toolCall.IsError,
			Completed: toolCall.Status == ToolCallStatusCompleted || toolCall.Status == ToolCallStatusError,
			MsgTime:   time.Now(),
		}

		// 使用不同的颜色表示不同状态
		var color string
		switch toolCall.Status {
		case ToolCallStatusPending:
			color = "248" // 灰色
		case ToolCallStatusRunning:
			color = "255" // 白色
		case ToolCallStatusCompleted:
			color = "82" // 绿色
		case ToolCallStatusError:
			color = "203" // 红色
		default:
			color = "255" // 默认白色
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		sections = append(sections, style.Render(toolMsg.Render()))
	}

	return strings.Join(sections, "\n")
}

func (m appModel) renderStatusBar() string {
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

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Width(20)
	statusStr := statusStyle.Render("● " + m.status)

	modelWidth := m.width - 40
	if modelWidth < 0 {
		modelWidth = 0
	}
	modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Width(modelWidth)
	modelStr := modelStyle.Render(m.model)

	barStyle := lipgloss.NewStyle().Background(lipgloss.Color("237")).Padding(0, 1).Width(m.width)
	content := lipgloss.JoinHorizontal(lipgloss.Left, statusStr, modelStr)
	return barStyle.Render(content)
}

func (m appModel) renderHelp() string {
	help := "\nQoder CLI Help\n\nCtrl+C Quit\nEnter Send\nF1 Toggle help\n"
	helpWidth := m.width - 4
	if helpWidth < 20 {
		helpWidth = 20
	}
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Width(helpWidth)
	return style.Render(help)
}

// === 消息类型定义 ===

type InitMsg struct{}
type StartedMsg struct{}
type AgentStreamMsg struct{ Content string }
type AgentToolStartMsg struct{ ID, Name, Arguments string }
type AgentToolResultMsg struct {
	ToolCallID, Content string
	IsError             bool
}
type AgentErrorMsg struct{ Err error }
type AgentFinishMsg struct{ Reason types.FinishReason }
type QuitMsg struct{}
