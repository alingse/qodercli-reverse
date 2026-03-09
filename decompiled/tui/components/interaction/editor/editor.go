// Package editor 交互式编辑器组件
// 1:1 还原原版 qodercli 的 EditorComponent
package editor

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/textarea"

	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
)

// Msg 编辑器消息类型
type Msg interface {
	tea.Msg
}

// SendMsg 发送消息命令
type SendMsg struct {
	Content     string
	Attachments []string
}

// ClearMsg 清空输入
type ClearMsg struct{}

// Mode 编辑器模式
type Mode int

const (
	ModeInsert Mode = iota
	ModeNormal
	ModeVisual
)

// EditorComponent 编辑器组件 - 1:1 还原原版
type EditorComponent struct {
	// 核心 textarea
	textarea textarea.Model

	// 尺寸
	width  int
	height int

	// 状态
	mode      Mode
	focused   bool
	enabled   bool

	// 功能处理器
	attachmentHandler *AttachmentHandler
	historyHandler    *HistoryHandler
	inputCache        *InputCache

	// 配置
	placeholder string

	// 事件发布
	pubsub *pubsub.PubSub
}

// NewEditorComponent 创建新的编辑器组件
func NewEditorComponent(ps *pubsub.PubSub) *EditorComponent {
	ta := textarea.New()
	ta.Placeholder = "Ask anything..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 100000
	// 官方样式：使用 "> " 作为 prompt
	ta.Prompt = "> "
	// 设置样式
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	// 文本样式：白色
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	// Prompt 样式：灰色
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	// 光标样式
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()

	ec := &EditorComponent{
		textarea:          ta,
		mode:              ModeInsert,
		focused:           true,
		enabled:           true,
		placeholder:       "Ask anything...",
		pubsub:            ps,
		attachmentHandler: NewAttachmentHandler(),
		historyHandler:    NewHistoryHandler(),
		inputCache:        NewInputCache(),
		// 设置默认尺寸
		width:  80,
		height: 3,
	}

	// 设置 textarea 的默认尺寸
	ec.textarea.SetWidth(76) // 80 - 4 (边框和 padding)
	ec.textarea.SetHeight(1) // 单行输入

	return ec
}

// Init 初始化
func (ec *EditorComponent) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		ec.historyHandler.Init(),
	)
}

// Update 更新状态 - 核心消息处理
func (ec *EditorComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !ec.enabled {
		return ec, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ec.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		cmd := ec.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.MouseMsg:
		// 处理鼠标事件
		if msg.Type == tea.MouseLeft {
			ec.Focus()
		}

	case ClearMsg:
		ec.resetInputs()
		return ec, nil
	}

	// 更新 textarea
	if ec.focused {
		newTextarea, cmd := ec.textarea.Update(msg)
		ec.textarea = newTextarea
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return ec, tea.Batch(cmds...)
}

// handleKeyMsg 处理键盘消息 - 原版核心逻辑
func (ec *EditorComponent) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		return ec.handleEscape()

	case tea.KeyEnter:
		// 检查 Shift+Enter
		if msg.String() == "shift+enter" {
			return nil // 让 textarea 处理换行
		}
		// Enter 发送消息
		return ec.sendMessage()

	case tea.KeyCtrlC:
		if ec.textarea.Value() == "" {
			return tea.Quit
		}

	case tea.KeyUp:
		if ec.atTopOfInput() {
			return ec.handleHistoryNavigation(-1)
		}

	case tea.KeyDown:
		if ec.atBottomOfInput() {
			return ec.handleHistoryNavigation(1)
		}

	case tea.KeyTab:
		// Tab: 插入空格
		ec.textarea.InsertString("    ")
		return nil

	case tea.KeyPgUp, tea.KeyPgDown:
		// 页面滚动 - 传递给父组件
		return nil
	}

	return nil
}

// handleEscape 处理 Esc 键
func (ec *EditorComponent) handleEscape() tea.Cmd {
	ec.resetEscState()
	return nil
}

// resetEscState 重置 Esc 状态
func (ec *EditorComponent) resetEscState() {
	ec.textarea.Blur()
	ec.focused = false
}

// resetInputs 重置输入
func (ec *EditorComponent) resetInputs() {
	ec.textarea.Reset()
	ec.textarea.Focus()
	ec.focused = true
	ec.attachmentHandler.Reset()
	// 重置 textarea 高度为单行（1行内容 + 边框）
	ec.textarea.SetHeight(1)
}

// sendMessage 发送消息 - 核心功能
func (ec *EditorComponent) sendMessage() tea.Cmd {
	content := strings.TrimSpace(ec.textarea.Value())
	if content == "" && !ec.attachmentHandler.HasAttachments() {
		return nil
	}

	// 缓存输入
	ec.inputCache.SetInput(content)

	// 添加历史记录
	ec.historyHandler.AddHistory(content)

	// 构建消息
	msg := SendMsg{
		Content:     content,
		Attachments: ec.attachmentHandler.GetAttachments(),
	}

	// 发布事件
	if ec.pubsub != nil {
		ec.pubsub.Publish(context.Background(), pubsub.Event{
			Type:    pubsub.EventTypeUserInput,
			Payload: msg,
		})
	}

	// 清空输入
	ec.resetInputs()

	return func() tea.Msg {
		return msg
	}
}

// handleHistoryNavigation 历史记录导航
func (ec *EditorComponent) handleHistoryNavigation(direction int) tea.Cmd {
	history := ec.historyHandler.Navigate(direction)
	if history != "" {
		ec.textarea.SetValue(history)
		// 将光标移到末尾
		ec.textarea.CursorEnd()
	}
	return nil
}

// atTopOfInput 是否在输入框顶部
func (ec *EditorComponent) atTopOfInput() bool {
	// 检查光标是否在第一行
	return ec.textarea.Line() == 0
}

// atBottomOfInput 是否在输入框底部
func (ec *EditorComponent) atBottomOfInput() bool {
	// 获取总行数
	lineCount := ec.textarea.LineCount()
	return ec.textarea.Line() >= lineCount-1
}

// GetLineCount 获取当前行数
func (ec *EditorComponent) GetLineCount() int {
	return ec.textarea.LineCount()
}

// GetPreferredHeight 获取首选高度（根据内容行数 + 边框 + padding）
func (ec *EditorComponent) GetPreferredHeight() int {
	// 内容行数（最多5行）
	contentLines := ec.textarea.LineCount()
	if contentLines > 5 {
		contentLines = 5
	}
	// 最少1行内容
	if contentLines < 1 {
		contentLines = 1
	}
	// 总高度 = 内容行数 + 边框(2行)
	totalHeight := contentLines + 2
	// 限制最大高度为7行（5行内容+2行边框）
	if totalHeight > 7 {
		totalHeight = 7
	}
	return totalHeight
}

// View 渲染视图
func (ec *EditorComponent) View() string {
	if !ec.focused {
		ec.textarea.Blur()
	} else {
		ec.textarea.Focus()
	}

	// 构建编辑器视图
	var sb strings.Builder

	// 渲染附件（如果有）
	if ec.attachmentHandler.HasAttachments() {
		sb.WriteString(ec.attachmentHandler.Render())
		sb.WriteString("\n")
	}

	// 渲染 textarea - 官方样式：圆角边框
	borderColor := lipgloss.Color("240") // 灰色边框
	if ec.focused {
		borderColor = lipgloss.Color("255") // 聚焦时白色边框
	}

	// 使用单行样式
	editorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1). // 左右内边距 1
		Width(ec.width)

	sb.WriteString(editorStyle.Render(ec.textarea.View()))

	return sb.String()
}

// SetSize 设置尺寸
func (ec *EditorComponent) SetSize(width, height int) {
	ec.width = width
	ec.height = height
	// 减去边框（2字符）和 padding（2字符）
	ec.textarea.SetWidth(width - 4)
	// textarea 最大高度为 5 行（多行输入时自动扩展）
	// 限制内部高度为 1-5 行
	innerHeight := height - 2 // 减去边框
	if innerHeight < 1 {
		innerHeight = 1
	}
	if innerHeight > 5 {
		innerHeight = 5
	}
	ec.textarea.SetHeight(innerHeight)
}

// Focus 聚焦
func (ec *EditorComponent) Focus() {
	ec.focused = true
	ec.textarea.Focus()
}

// Blur 失焦
func (ec *EditorComponent) Blur() {
	ec.focused = false
	ec.textarea.Blur()
}

// GetInput 获取当前输入
func (ec *EditorComponent) GetInput() string {
	return ec.textarea.Value()
}

// SetContent 设置内容
func (ec *EditorComponent) SetContent(content string) {
	ec.textarea.SetValue(content)
}

// IsEmpty 是否为空
func (ec *EditorComponent) IsEmpty() bool {
	return strings.TrimSpace(ec.textarea.Value()) == ""
}

// HasRunningShells 是否有运行中的 shell
func (ec *EditorComponent) HasRunningShells() bool {
	// 简化实现
	return false
}

// CancelRunningShells 取消运行中的 shell
func (ec *EditorComponent) CancelRunningShells() {
	// 简化实现
}

// addAttachment 添加附件
func (ec *EditorComponent) addAttachment(path string) error {
	return ec.attachmentHandler.AddAttachment(path)
}
