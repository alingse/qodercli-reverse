// Package messages 消息显示组件
// 1:1 还原原版 MessageView
package messages

import (
	"fmt"
	"strings"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	MsgTypeTodoList
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
	// 启用鼠标支持，确保可以滚动
	vp.MouseWheelEnabled = true

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
	// 先渲染内容，再滚动到底部
	mv.renderContent()
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

// AddSystemMessage 添加系统消息（用于统计信息等）
func (mv *MessageView) AddSystemMessage(content string) {
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
		// 重新渲染内容
		mv.renderContent()
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

// AddPendingBash 添加待执行的 Bash 命令（白色 ⏺）
// 注意：此方法已弃用，请使用 AddPendingTool
func (mv *MessageView) AddPendingBash(command string) {
	mv.AddPendingTool(fmt.Sprintf("bash-%d", len(mv.messages)), "Bash", command)
}

// CompleteBash 完成 Bash 命令，更新状态
// 注意：此方法已弃用，请使用 CompleteTool
func (mv *MessageView) CompleteBash(output string, isError bool) {
	// 找到最后一个未完成的工具消息
	mv.CompleteTool("", "Bash", output, isError)
}

// AddPendingTool 添加待执行的工具调用（白色 ⏺）
// 所有工具调用统一使用此方法添加
func (mv *MessageView) AddPendingTool(toolID, toolName, arguments string) {
	msg := &ToolCallInfo{
		ID:        toolID,
		Name:      toolName,
		Arguments: arguments,
		Completed: false,
		MsgTime:   time.Now(),
	}
	mv.AddMessage(msg)
}

// CompleteTool 完成工具调用，更新状态
// toolID: 工具调用唯一标识
// toolName: 工具名称（用于显示）
// output: 执行结果
// isError: 是否执行出错
func (mv *MessageView) CompleteTool(toolID, toolName, output string, isError bool) {
	// 找到对应的待处理工具消息
	for i := len(mv.messages) - 1; i >= 0; i-- {
		if toolInfo, ok := mv.messages[i].(*ToolCallInfo); ok {
			// 匹配 ID（如果提供）或匹配最后一个未完成的同名工具
			if !toolInfo.Completed {
				if toolID == "" || toolInfo.ID == toolID {
					toolInfo.Output = output
					toolInfo.IsError = isError
					toolInfo.Completed = true
					// 重新渲染以更新图标颜色
					mv.renderContent()
					if mv.autoScroll {
						mv.viewport.GotoBottom()
					}
					break
				}
			}
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

	// 保存当前滚动位置（YOffset）
	oldYOffset := mv.viewport.YOffset

	mv.viewport.Width = width
	mv.viewport.Height = height

	// 恢复滚动位置，确保内容不会丢失
	// 但需要确保不超过新的内容范围
	mv.viewport.YOffset = oldYOffset

	// 更新 glamour 渲染器
	mv.renderer, _ = glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width-4),
	)

	// 重新渲染内容以适应新的宽度
	mv.renderContent()
}

// renderContent 渲染内容
func (mv *MessageView) renderContent() {
	var sb strings.Builder

	for i, msg := range mv.messages {
		if i > 0 {
			prevMsg := mv.messages[i-1]
			// 只在不同类型的消息之间添加空行
			// 同类型消息（如流式输出的多次追加）之间不添加
			if msg.Type() != prevMsg.Type() {
				sb.WriteString("\n\n")
			} else {
				// 同类型消息之间添加间隔
				// 对于工具调用消息，添加额外空行以便区分
				if msg.Type() == MsgTypeTool {
					sb.WriteString("\n\n")
				} else {
					// 其他同类型消息之间只添加一个换行
					sb.WriteString("\n")
				}
			}
		}
		sb.WriteString(msg.Render())
	}

	content := sb.String()
	// 不再使用 glamour 渲染，因为消息已经通过 lipgloss 格式化
	// glamour 会把 ANSI 转义序列当作普通文本，导致控制符显示出来
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

// AddTodoList 添加 Todo 列表消息
func (mv *MessageView) AddTodoList(todos []state.Todo) {
	msg := &TodoListMessage{
		Todos:   todos,
		Updated: false,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// UpdateTodoList 更新 Todo 列表消息
func (mv *MessageView) UpdateTodoList(todos []state.Todo, oldTodos []state.Todo) {
	// 查找是否有现有的 TodoListMessage
	var existingMsg *TodoListMessage
	for i := len(mv.messages) - 1; i >= 0; i-- {
		if tlm, ok := mv.messages[i].(*TodoListMessage); ok {
			existingMsg = tlm
			// 从列表中移除旧的消息
			mv.messages = append(mv.messages[:i], mv.messages[i+1:]...)
			break
		}
	}

	msg := &TodoListMessage{
		Todos:    todos,
		OldTodos: oldTodos,
		Updated:  existingMsg != nil,
		MsgTime:  time.Now(),
	}
	mv.AddMessage(msg)
}

// GetTodoList 获取当前的 Todo 列表（如果存在）
func (mv *MessageView) GetTodoList() []state.Todo {
	for i := len(mv.messages) - 1; i >= 0; i-- {
		if tlm, ok := mv.messages[i].(*TodoListMessage); ok {
			return tlm.Todos
		}
	}
	return nil
}
