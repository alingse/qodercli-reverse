// Package messages 消息显示组件
// 1:1 还原原版 MessageView
package messages

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
)

// MessageType 消息类型
type MessageType int

const (
	MsgTypeUser MessageType = iota
	MsgTypeAssistant
	MsgTypeSystem
	MsgTypeTool
	MsgTypeBash
	MsgTypeCommand
	MsgTypeError
	MsgTypeCompact
	MsgTypeLog
)

// Message 消息接口
type Message interface {
	Type() MessageType
	Render() string
	Timestamp() time.Time
}

// MessageView 消息视图组件 - 1:1 还原原版
type MessageView struct {
	viewport viewport.Model
	messages []Message
	renderer *glamour.TermRenderer

	width  int
	height int

	// 待处理的消息
	pendingMessages []Message

	// 自动滚动
	autoScroll bool
}

// NewMessageView 创建新的消息视图
func NewMessageView() *MessageView {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(76),
	)

	return &MessageView{
		viewport:        vp,
		messages:        make([]Message, 0),
		pendingMessages: make([]Message, 0),
		renderer:        renderer,
		autoScroll:      true,
	}
}

// Init 初始化
func (mv *MessageView) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (mv *MessageView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mv.SetSize(msg.Width, msg.Height)

	case tea.MouseMsg:
		// 处理鼠标滚轮
		switch msg.Type {
		case tea.MouseWheelUp:
			mv.viewport.LineUp(3)
			mv.autoScroll = false
		case tea.MouseWheelDown:
			mv.viewport.LineDown(3)
			// 检查是否在底部
			if mv.viewport.AtBottom() {
				mv.autoScroll = true
			}
		}
	}

	// 更新 viewport
	newVp, cmd := mv.viewport.Update(msg)
	mv.viewport = newVp

	return mv, cmd
}

// View 渲染视图
func (mv *MessageView) View() string {
	mv.renderContent()
	return mv.viewport.View()
}

// AddMessage 添加消息
func (mv *MessageView) AddMessage(msg Message) {
	mv.messages = append(mv.messages, msg)
	if mv.autoScroll {
		mv.viewport.GotoBottom()
	}
}

// AddUserMessage 添加用户消息
func (mv *MessageView) AddUserMessage(content string) {
	msg := &UserMessage{
		Content: content,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AddAssistantMessage 添加助手消息
func (mv *MessageView) AddAssistantMessage(content string) {
	msg := &AssistantMessage{
		Content: content,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AppendToLastMessage 追加到最后一条消息
func (mv *MessageView) AppendToLastMessage(content string) {
	if len(mv.messages) == 0 {
		mv.AddAssistantMessage(content)
		return
	}

	lastMsg := mv.messages[len(mv.messages)-1]
	if am, ok := lastMsg.(*AssistantMessage); ok {
		am.Content += content
		if mv.autoScroll {
			mv.viewport.GotoBottom()
		}
	} else {
		mv.AddAssistantMessage(content)
	}
}

// AddBashInfo 添加 Bash 命令信息
func (mv *MessageView) AddBashInfo(command string) int {
	msg := &BashInfo{
		Command: command,
		MsgTime: time.Now(),
		ID:      len(mv.messages),
	}
	mv.AddMessage(msg)
	return msg.ID
}

// UpdateBashResult 更新 Bash 结果
func (mv *MessageView) UpdateBashResult(id int, output string, isError bool) {
	for _, msg := range mv.messages {
		if bashInfo, ok := msg.(*BashInfo); ok && bashInfo.ID == id {
			bashInfo.Output = output
			bashInfo.IsError = isError
			bashInfo.Completed = true
			break
		}
	}
}

// AddToolCall 添加工具调用
func (mv *MessageView) AddToolCall(name, arguments string) {
	msg := &ToolCall{
		Name:      name,
		Arguments: arguments,
		MsgTime:   time.Now(),
	}
	mv.AddMessage(msg)
}

// AddToolResult 添加工具结果
func (mv *MessageView) AddToolResult(name, result string, err error) {
	msg := &ToolResult{
		Name:    name,
		Result:  result,
		Error:   err,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AddError 添加错误消息
func (mv *MessageView) AddError(err string) {
	msg := &ErrorMessage{
		ErrStr:  err,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AddTokenUsage 添加 Token 使用统计
func (mv *MessageView) AddTokenUsage(inputTokens, outputTokens int) {
	msg := &TokenUsageMessage{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		MsgTime:      time.Now(),
	}
	mv.AddMessage(msg)
}

// Clear 清空消息
func (mv *MessageView) Clear() {
	mv.messages = make([]Message, 0)
	mv.viewport.SetContent("")
}

// SetSize 设置尺寸
func (mv *MessageView) SetSize(width, height int) {
	mv.width = width
	mv.height = height
	mv.viewport.Width = width
	mv.viewport.Height = height

	// 更新 glamour 渲染器
	mv.renderer, _ = glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width-4),
	)
}

// renderContent 渲染内容
func (mv *MessageView) renderContent() {
	var sb strings.Builder

	for i, msg := range mv.messages {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(msg.Render())
	}

	content := sb.String()
	if mv.renderer != nil {
		rendered, err := mv.renderer.Render(content)
		if err == nil {
			content = rendered
		}
	}

	mv.viewport.SetContent(content)
}

// ScrollToBottom 滚动到底部
func (mv *MessageView) ScrollToBottom() {
	mv.viewport.GotoBottom()
	mv.autoScroll = true
}

// ScrollToTop 滚动到顶部
func (mv *MessageView) ScrollToTop() {
	mv.viewport.GotoTop()
	mv.autoScroll = false
}

// PageUp 向上翻页
func (mv *MessageView) PageUp() {
	mv.viewport.HalfPageUp()
	mv.autoScroll = false
}

// PageDown 向下翻页
func (mv *MessageView) PageDown() {
	mv.viewport.HalfPageDown()
	if mv.viewport.AtBottom() {
		mv.autoScroll = true
	}
}

// HalfPageUp 向上半页
func (mv *MessageView) HalfPageUp() {
	mv.viewport.HalfPageUp()
	mv.autoScroll = false
}

// HalfPageDown 向下半页
func (mv *MessageView) HalfPageDown() {
	mv.viewport.HalfPageDown()
	if mv.viewport.AtBottom() {
		mv.autoScroll = true
	}
}

// GetLastMessage 获取最后一条消息
func (mv *MessageView) GetLastMessage() Message {
	if len(mv.messages) == 0 {
		return nil
	}
	return mv.messages[len(mv.messages)-1]
}
