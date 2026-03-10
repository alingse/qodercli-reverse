// Package app TUI 应用主入口
// 遵循 Bubble Tea 架构：所有状态变更必须通过 Update 循环
package app

import (
	"context"
	"os"
	"strings"
	"sync/atomic"
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

	// 上次设置的组件尺寸
	lastEditorWidth  int
	lastEditorHeight int

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

	// 流式输出追踪
	lastResponseLen int32 
	
	// 流式预览缓冲区 - 用于在 View 中实时显示正在生成的 AI 回复
	streamingBuffer string

	// 事件通道
	eventChan chan tea.Msg

	// 工具调用 ID 到名称的映射
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
		eventChan:     make(chan tea.Msg, 100),
		toolCallMap:   make(map[string]string),
	}

	// 初始化子组件
	m.editor = editor.NewEditorComponent(ps)
	m.msgView = messages.NewMessageView()

	// 设置 Agent 回调
	m.setupAgentCallbacks()

	return m
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
		
		// 2. 启动 Agent
		cmds = append(cmds, m.handleUserInput(msg.Content))

	case AgentStreamMsg:
		// 更新预览区
		m.streamingBuffer += msg.Content
		cmds = append(cmds, m.waitForEvents())

	case AgentToolStartMsg:
		m.status = "Using tool..."
		m.toolCallMap[msg.ID] = msg.Name
		
		// 打印工具调用开始
		toolMsg := &messages.ToolCallInfo{
			ID:        msg.ID,
			Name:      msg.Name,
			Arguments: msg.Arguments,
			Completed: false,
			MsgTime:   time.Now(),
		}
		cmds = append(cmds, tea.Printf(toolMsg.Render()))
		cmds = append(cmds, m.waitForEvents())

	case AgentToolResultMsg:
		toolName := m.toolCallMap[msg.ToolCallID]
		// 打印工具执行结果
		toolResultMsg := &messages.ToolCallInfo{
			ID:        msg.ToolCallID,
			Name:      toolName,
			Output:    msg.Content,
			IsError:   msg.IsError,
			Completed: true,
			MsgTime:   time.Now(),
		}
		cmds = append(cmds, tea.Printf(toolResultMsg.Render()))
		
		m.status = "Thinking..."
		cmds = append(cmds, m.waitForEvents())

	case AgentErrorMsg:
		m.errorMsg = msg.Err.Error()
		m.processing = false
		m.status = "Error"
		
		errMsg := &messages.ErrorMessage{
			ErrStr:  m.errorMsg,
			MsgTime: time.Now(),
		}
		cmds = append(cmds, tea.Printf(errMsg.Render()))
		
		m.streamingBuffer = ""
		atomic.StoreInt32(&m.lastResponseLen, 0)
		cmds = append(cmds, m.waitForEvents())

	case AgentFinishMsg:
		// 任务完成：将预览内容固化
		if m.streamingBuffer != "" {
			assistantMsg := &messages.AssistantMessage{
				Content: m.streamingBuffer,
				MsgTime: time.Now(),
			}
			cmds = append(cmds, tea.Printf(assistantMsg.Render()))
			m.streamingBuffer = "" 
		}
		
		m.processing = false
		m.status = "Ready"
		atomic.StoreInt32(&m.lastResponseLen, 0)
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

		atomic.StoreInt32(&m.lastResponseLen, 0)
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
	if len(parts) == 0 { return nil }
	
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
		func(msg *types.Message) {
			var fullContent strings.Builder
			for _, part := range msg.Content {
				if part.Type == "text" && part.Text != "" {
					fullContent.WriteString(part.Text)
				}
			}
			fullStr := fullContent.String()
			runes := []rune(fullStr)
			lastLen := atomic.LoadInt32(&m.lastResponseLen)
			if len(runes) > int(lastLen) {
				newRunes := runes[lastLen:]
				atomic.StoreInt32(&m.lastResponseLen, int32(len(runes)))
				select {
				case m.eventChan <- AgentStreamMsg{Content: string(newRunes)}:
				default:
				}
			}
		},
		func(call *types.ToolCall) {
			select {
			case m.eventChan <- AgentToolStartMsg{ID: call.ID, Name: call.Name, Arguments: call.Arguments}:
			default:
			}
		},
		func(result *types.ToolResult) {
			select {
			case m.eventChan <- AgentToolResultMsg{ToolCallID: result.ToolCallID, Content: result.Content, IsError: result.IsError}:
			default:
			}
		},
		func(err error) {
			select {
			case m.eventChan <- AgentErrorMsg{Err: err}:
			default:
			}
		},
		func(reason types.FinishReason) {
			select {
			case m.eventChan <- AgentFinishMsg{Reason: reason}:
			default:
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
	if editorHeight < 3 { editorHeight = 3 }
	if editorHeight > 7 { editorHeight = 7 }
	
	if m.width > 0 {
		m.editor.SetSize(m.width, editorHeight)
	}

	var sections []string

	// 1. 正在流式输出的预览区
	if m.streamingBuffer != "" {
		preview := &messages.AssistantMessage{
			Content: m.streamingBuffer,
			MsgTime: time.Now(),
		}
		previewStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Italic(true).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("240"))
		
		sections = append(sections, previewStyle.Render(preview.Render()))
	}

	// 2. 编辑器
	sections = append(sections, m.editor.View())

	// 3. 状态栏
	sections = append(sections, m.renderStatusBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m appModel) renderStatusBar() string {
	var statusColor string
	switch m.status {
	case "Ready": statusColor = "82"
	case "Thinking...", "Using tool...": statusColor = "135"
	case "Error": statusColor = "203"
	default: statusColor = "248"
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Width(20)
	statusStr := statusStyle.Render("● " + m.status)
	
	modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Width(m.width - 40)
	modelStr := modelStyle.Render(m.model)

	barStyle := lipgloss.NewStyle().Background(lipgloss.Color("237")).Padding(0, 1).Width(m.width)
	content := lipgloss.JoinHorizontal(lipgloss.Left, statusStr, modelStr)
	return barStyle.Render(content)
}

func (m appModel) renderHelp() string {
	help := "\nQoder CLI Help\n\nCtrl+C Quit\nEnter Send\nF1 Toggle help\n"
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Width(m.width - 4)
	return style.Render(help)
}

// === 消息类型定义 ===

type InitMsg struct{}
type StartedMsg struct{}
type AgentStreamMsg struct { Content string }
type AgentToolStartMsg struct { ID, Name, Arguments string }
type AgentToolResultMsg struct { ToolCallID, Content string; IsError bool }
type AgentErrorMsg struct { Err error }
type AgentFinishMsg struct { Reason types.FinishReason }
type QuitMsg struct{}
