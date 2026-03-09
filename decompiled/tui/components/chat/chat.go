// Package chat 聊天视图组件
// 主聊天界面，整合消息显示和交互，支持 Markdown 渲染
package chat

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
)

// MessageType 消息类型
type MessageType int

const (
	MsgTypeText MessageType = iota
	MsgTypeToolCall
	MsgTypeToolResult
	MsgTypeError
	MsgTypeSystem
)

// ChatMessage 聊天消息
type ChatMessage struct {
	Type      MessageType
	Role      string
	Content   string
	Metadata  map[string]string
	Timestamp time.Time
	ToolName  string
	ToolArgs  string
	ToolResult string
	ToolError error
}

// 配色方案
var (
	colorPrimary       = lipgloss.Color("#7D56F4")  // 主色调 - 紫色
	colorSecondary     = lipgloss.Color("#5B8DEF")  // 辅助色 - 蓝色
	colorSuccess       = lipgloss.Color("#00C853")  // 成功 - 绿色
	colorWarning       = lipgloss.Color("#FFB74D")  // 警告 - 橙色
	colorError         = lipgloss.Color("#EF5350")  // 错误 - 红色
	colorErrorBg       = lipgloss.Color("#1A1A2E")  // 错误背景 - 深蓝
	colorText          = lipgloss.Color("#E0E0E0")  // 文本 - 浅灰
	colorTextMuted     = lipgloss.Color("#6B7280")  // 弱化文本 - 灰色
	colorBorder        = lipgloss.Color("#4B5563")  // 边框 - 中灰
	colorUserBg        = lipgloss.Color("#5B8DEF")  // 用户消息背景 - 蓝色
)

// Component 聊天组件
type Component struct {
	viewport     viewport.Model
	messages     []ChatMessage
	spinner      spinner.Model
	width        int
	height       int
	isThinking   bool
	currentTool  string
	renderer     *glamour.TermRenderer
	renderedCache string
	needsRender  bool
}

// New 创建新的聊天组件
func New() *Component {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	// 初始化 Markdown 渲染器
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	return &Component{
		viewport:    vp,
		messages:    make([]ChatMessage, 0),
		spinner:     s,
		renderer:    renderer,
		needsRender: true,
	}
}

// Init 初始化组件
func (c *Component) Init() tea.Cmd {
	return c.spinner.Tick
}

// Update 更新组件状态
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.SetSize(msg.Width, msg.Height)

	case spinner.TickMsg:
		if c.isThinking {
			newSpinner, cmd := c.spinner.Update(msg)
			c.spinner = newSpinner
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.MouseMsg:
		// 处理鼠标滚轮事件
		switch msg.Type {
		case tea.MouseWheelUp:
			c.viewport.LineUp(3)
			return c, nil
		case tea.MouseWheelDown:
			c.viewport.LineDown(3)
			return c, nil
		case tea.MouseWheelLeft:
			c.viewport.ScrollLeft(3)
			return c, nil
		case tea.MouseWheelRight:
			c.viewport.ScrollRight(3)
			return c, nil
		}
	}

	// 更新 viewport
	newViewport, cmd := c.viewport.Update(msg)
	c.viewport = newViewport
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

// View 渲染视图
func (c *Component) View() string {
	if c.needsRender {
		c.renderContent()
	}

	var sb strings.Builder
	sb.WriteString(c.viewport.View())

	// 显示思考中状态
	if c.isThinking {
		sb.WriteString("\n")
		thinkingStyle := lipgloss.NewStyle().
			Foreground(colorSecondary).
			Italic(true)
		if c.currentTool != "" {
			sb.WriteString(thinkingStyle.Render(fmt.Sprintf("%s Using %s...", c.spinner.View(), c.currentTool)))
		} else {
			sb.WriteString(thinkingStyle.Render(fmt.Sprintf("%s Thinking...", c.spinner.View())))
		}
	}

	return sb.String()
}

// AddUserMessage 添加用户消息
func (c *Component) AddUserMessage(content string) {
	msg := ChatMessage{
		Type:      MsgTypeText,
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}
	c.messages = append(c.messages, msg)
	c.needsRender = true
	c.viewport.GotoBottom()
}

// AddAssistantMessage 添加助手消息
func (c *Component) AddAssistantMessage(content string) {
	msg := ChatMessage{
		Type:      MsgTypeText,
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}
	c.messages = append(c.messages, msg)
	c.needsRender = true
	c.viewport.GotoBottom()
}

// AppendContent 追加内容到最后一条消息
func (c *Component) AppendContent(content string) {
	if len(c.messages) == 0 {
		c.AddAssistantMessage(content)
		return
	}

	lastMsg := &c.messages[len(c.messages)-1]
	if lastMsg.Role == "assistant" {
		lastMsg.Content += content
	} else {
		c.AddAssistantMessage(content)
	}
	c.needsRender = true
	c.viewport.GotoBottom()
}

// ShowToolCall 显示工具调用
func (c *Component) ShowToolCall(name, arguments string) tea.Cmd {
	return func() tea.Msg {
		msg := ChatMessage{
			Type:      MsgTypeToolCall,
			Role:      "assistant",
			ToolName:  name,
			ToolArgs:  arguments,
			Timestamp: time.Now(),
		}
		c.messages = append(c.messages, msg)
		c.currentTool = name
		c.needsRender = true
		c.viewport.GotoBottom()
		return nil
	}
}

// ShowToolResult 显示工具结果
func (c *Component) ShowToolResult(name, result string, err error) tea.Cmd {
	return func() tea.Msg {
		msg := ChatMessage{
			Type:       MsgTypeToolResult,
			Role:       "tool",
			ToolName:   name,
			ToolResult: result,
			ToolError:  err,
			Timestamp:  time.Now(),
		}
		c.messages = append(c.messages, msg)
		c.currentTool = ""
		c.needsRender = true
		c.viewport.GotoBottom()
		return nil
	}
}

// ShowError 显示错误
func (c *Component) ShowError(err error) {
	msg := ChatMessage{
		Type:      MsgTypeError,
		Role:      "system",
		Content:   err.Error(),
		Timestamp: time.Now(),
	}
	c.messages = append(c.messages, msg)
	c.needsRender = true
	c.viewport.GotoBottom()
}

// Clear 清空聊天
func (c *Component) Clear() {
	c.messages = make([]ChatMessage, 0)
	c.viewport.SetContent("")
	c.renderedCache = ""
	c.needsRender = true
}

// SetSize 设置组件尺寸
func (c *Component) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Width = width
	c.viewport.Height = height - 2 // 预留状态行
	
	// 更新渲染器的换行设置
	if c.renderer != nil {
		c.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width - 4),
		)
	}
	
	c.needsRender = true
}

// SetThinking 设置思考状态
func (c *Component) SetThinking(thinking bool) {
	c.isThinking = thinking
}

// ScrollToTop 滚动到顶部
func (c *Component) ScrollToTop() {
	c.viewport.GotoTop()
}

// ScrollToBottom 滚动到底部
func (c *Component) ScrollToBottom() {
	c.viewport.GotoBottom()
}

// PageUp 向上一页
func (c *Component) PageUp() {
	c.viewport.LineUp(c.viewport.Height)
}

// PageDown 向下一页
func (c *Component) PageDown() {
	c.viewport.LineDown(c.viewport.Height)
}

// HalfPageUp 向上半页
func (c *Component) HalfPageUp() {
	c.viewport.HalfPageUp()
}

// HalfPageDown 向下半页
func (c *Component) HalfPageDown() {
	c.viewport.HalfPageDown()
}

// renderContent 渲染内容
func (c *Component) renderContent() {
	if !c.needsRender {
		return
	}

	var sb strings.Builder

	for i, msg := range c.messages {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(c.formatMessage(msg))
	}

	rendered := sb.String()
	
	// 使用 glamour 渲染 Markdown
	if c.renderer != nil {
		if r, err := c.renderer.Render(rendered); err == nil {
			rendered = r
		}
	}

	c.viewport.SetContent(rendered)
	c.renderedCache = rendered
	c.needsRender = false
}

// formatMessage 格式化消息
func (c *Component) formatMessage(msg ChatMessage) string {
	var sb strings.Builder

	switch msg.Type {
	case MsgTypeText:
		sb.WriteString(c.formatTextMessage(msg))

	case MsgTypeToolCall:
		sb.WriteString(c.formatToolCall(msg))

	case MsgTypeToolResult:
		sb.WriteString(c.formatToolResult(msg))

	case MsgTypeError:
		sb.WriteString(c.formatError(msg))
	}

	return sb.String()
}

// formatTextMessage 格式化文本消息
func (c *Component) formatTextMessage(msg ChatMessage) string {
	var prefix string

	switch msg.Role {
	case "user":
		prefix = "**You:**\n"

	case "assistant":
		prefix = "**Assistant:**\n"
	}

	return prefix + msg.Content
}

// formatToolCall 格式化工具调用
func (c *Component) formatToolCall(msg ChatMessage) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(c.width - 4)

	content := fmt.Sprintf("%s %s\n%s",
		headerStyle.Render("▶ Tool:"),
		msg.ToolName,
		msg.ToolArgs)

	return boxStyle.Render(content)
}

// formatToolResult 格式化工具结果
func (c *Component) formatToolResult(msg ChatMessage) string {
	var headerStyle lipgloss.Style
	var borderColor lipgloss.Color

	if msg.ToolError != nil {
		headerStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)
		borderColor = colorError
	} else {
		headerStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)
		borderColor = colorSuccess
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(c.width - 4)

	var content string
	if msg.ToolError != nil {
		content = fmt.Sprintf("%s %s\n%s",
			headerStyle.Render("✗ Error:"),
			msg.ToolName,
			msg.ToolError.Error())
	} else {
		// 截断过长的结果
		result := msg.ToolResult
		lines := strings.Split(result, "\n")
		if len(lines) > 20 {
			result = strings.Join(lines[:20], "\n") + "\n... (truncated)"
		}
		content = fmt.Sprintf("%s %s\n%s",
			headerStyle.Render("✓ Result:"),
			msg.ToolName,
			result)
	}

	return boxStyle.Render(content)
}

// formatError 格式化错误
func (c *Component) formatError(msg ChatMessage) string {
	errorStyle := lipgloss.NewStyle().
		Foreground(colorError).
		Background(colorErrorBg).
		Padding(0, 1).
		Width(c.width - 4)

	return errorStyle.Render(fmt.Sprintf("Error: %s", msg.Content))
}
