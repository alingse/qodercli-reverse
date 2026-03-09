// Package messages 消息列表组件
// 显示对话历史，支持滚动和搜索
package messages

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// 配色方案
var (
	msgColorUserBg      = lipgloss.Color("#5B8DEF")  // 用户消息背景 - 蓝色
	msgColorUserText    = lipgloss.Color("#FFFFFF")  // 用户消息文本 - 白色
	msgColorAssistant   = lipgloss.Color("#E5E7EB")  // 助手消息 - 浅灰
msgColorSystem    = lipgloss.Color("#9CA3AF")  // 系统消息 - 灰色
	msgColorTool        = lipgloss.Color("#00C853")  // 工具消息 - 绿色
)

// Role 消息角色样式
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// Message 显示消息
type Message struct {
	Role    Role
	Content string
	ToolCalls []types.ToolCall
}

// Component 消息列表组件
type Component struct {
	viewport    viewport.Model
	messages    []Message
	width       int
	height      int
	showTimestamps bool
}

// New 创建新的消息列表组件
func New() *Component {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	return &Component{
		viewport: vp,
		messages: make([]Message, 0),
	}
}

// Init 初始化组件
func (c *Component) Init() tea.Cmd {
	return nil
}

// Update 更新组件状态
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.SetSize(msg.Width, msg.Height)
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
	c.renderContent()
	return c.viewport.View()
}

// AddMessage 添加消息
func (c *Component) AddMessage(role Role, content string) {
	msg := Message{
		Role:    role,
		Content: content,
	}
	c.messages = append(c.messages, msg)
	c.renderContent()
	c.viewport.GotoBottom()
}

// AddToolMessage 添加工具消息
func (c *Component) AddToolMessage(toolName, result string) {
	msg := Message{
		Role:    RoleTool,
		Content: fmt.Sprintf("[%s] %s", toolName, result),
	}
	c.messages = append(c.messages, msg)
	c.renderContent()
	c.viewport.GotoBottom()
}

// UpdateLastMessage 更新最后一条消息
func (c *Component) UpdateLastMessage(content string) {
	if len(c.messages) == 0 {
		return
	}
	c.messages[len(c.messages)-1].Content = content
	c.renderContent()
}

// Clear 清空消息
func (c *Component) Clear() {
	c.messages = make([]Message, 0)
	c.viewport.SetContent("")
}

// SetSize 设置组件尺寸
func (c *Component) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.viewport.Width = width
	c.viewport.Height = height
	c.renderContent()
}

// UpdateContent 更新内容（用于 tea.Cmd）
func (c *Component) UpdateContent(content string) tea.Cmd {
	return func() tea.Msg {
		c.UpdateLastMessage(content)
		return nil
	}
}

// renderContent 渲染内容到 viewport
func (c *Component) renderContent() {
	var sb strings.Builder

	for i, msg := range c.messages {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(c.formatMessage(msg))
	}

	c.viewport.SetContent(sb.String())
}

// formatMessage 格式化单条消息
func (c *Component) formatMessage(msg Message) string {
	var style lipgloss.Style
	var prefix string

	switch msg.Role {
	case RoleUser:
		style = lipgloss.NewStyle().
			Foreground(msgColorUserText).
			Background(msgColorUserBg).
			Padding(0, 1)
		prefix = "You"

	case RoleAssistant:
		style = lipgloss.NewStyle().
			Foreground(msgColorAssistant)
		prefix = "Assistant"

	case RoleSystem:
		style = lipgloss.NewStyle().
			Foreground(msgColorSystem).
			Italic(true)
		prefix = "System"

	case RoleTool:
		style = lipgloss.NewStyle().
			Foreground(msgColorTool).
			Border(lipgloss.HiddenBorder()).
			Padding(0, 1)
		prefix = "Tool"
	}

	header := style.Render(fmt.Sprintf("%s:", prefix))
	content := msg.Content

	// 包装内容
	contentStyle := lipgloss.NewStyle().
		Width(c.width - 4).
		PaddingLeft(2)

	if msg.Role == RoleUser {
		return fmt.Sprintf("%s\n%s", header, contentStyle.Render(content))
	}

	return fmt.Sprintf("%s\n%s", header, contentStyle.Render(content))
}

// GetMessages 获取所有消息
func (c *Component) GetMessages() []Message {
	return c.messages
}

// ScrollDown 向下滚动
func (c *Component) ScrollDown() {
	c.viewport.LineDown(3)
}

// ScrollUp 向上滚动
func (c *Component) ScrollUp() {
	c.viewport.LineUp(3)
}
