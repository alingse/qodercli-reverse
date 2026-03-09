package messages

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// UserMessage 用户消息
type UserMessage struct {
	Content string
	MsgTime time.Time
}

func (m *UserMessage) Type() MessageType   { return MsgTypeUser }
func (m *UserMessage) Timestamp() time.Time { return m.MsgTime }
func (m *UserMessage) Render() string {
	// 原版格式：直接显示内容，无标记
	return m.Content
}

// String 实现 fmt.Stringer 接口
func (m *UserMessage) String() string {
	return m.Content
}

// AssistantMessage 助手消息
type AssistantMessage struct {
	Content string
	MsgTime time.Time
}

func (m *AssistantMessage) Type() MessageType  { return MsgTypeAssistant }
func (m *AssistantMessage) Timestamp() time.Time { return m.MsgTime }
func (m *AssistantMessage) Render() string {
	// 原版格式：直接显示内容
	return m.Content
}

// String 实现 fmt.Stringer 接口
func (m *AssistantMessage) String() string {
	return m.Content
}

// TokenUsageMessage Token 使用统计 - 原版在消息末尾显示
type TokenUsageMessage struct {
	InputTokens int
	OutputTokens int
	MsgTime      time.Time
}

func (m *TokenUsageMessage) Type() MessageType   { return MsgTypeSystem }
func (m *TokenUsageMessage) Timestamp() time.Time { return m.MsgTime }
func (m *TokenUsageMessage) Render() string {
	// 原版格式：Input token usage: xxx | Output token usage: xxx
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")). // 灰色
		Italic(true)

	return style.Render(fmt.Sprintf(
		"Input token usage: %d | Output token usage: %d",
		m.InputTokens, m.OutputTokens))
}

// String 实现 fmt.Stringer 接口
func (m *TokenUsageMessage) String() string {
	return m.Render()
}

// SystemMessage 系统消息
type SystemMessage struct {
	Content string
	MsgTime time.Time
}

func (m *SystemMessage) Type() MessageType   { return MsgTypeSystem }
func (m *SystemMessage) Timestamp() time.Time { return m.MsgTime }
func (m *SystemMessage) Render() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Italic(true)
	return style.Render(fmt.Sprintf("*System: %s*", m.Content))
}

// ToolCall 工具调用消息 - 原版以折叠框显示
type ToolCall struct {
	Name      string
	Arguments string
	MsgTime   time.Time
}

func (m *ToolCall) Type() MessageType   { return MsgTypeTool }
func (m *ToolCall) Timestamp() time.Time { return m.MsgTime }
func (m *ToolCall) Render() string {
	// 原版格式：▶ Tool: name (蓝色)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("75")).
		Bold(true)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	return fmt.Sprintf("%s %s\n%s",
		headerStyle.Render("▶ Tool:"),
		m.Name,
		contentStyle.Render(m.Arguments))
}

// ToolResult 工具结果消息
type ToolResult struct {
	Name       string
	Result     string
	Error     error
	MsgTime    time.Time
}

func (m *ToolResult) Type() MessageType   { return MsgTypeTool }
func (m *ToolResult) Timestamp() time.Time { return m.MsgTime }
func (m *ToolResult) Render() string {
	var headerStyle lipgloss.Style
	var icon string

	if m.Error != nil {
		headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)
		icon = "✗"
	} else {
		headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
		icon = "✓"
	}

	var content string
	if m.Error != nil {
		content = fmt.Sprintf("%s %s: %s",
			headerStyle.Render(icon),
			m.Name,
			m.Error.Error())
	} else {
		// 截断过长的结果
		result := m.Result
		lines := strings.Split(result, "\n")
		if len(lines) > 20 {
			result = strings.Join(lines[:20], "\n") + "\n... (truncated)"
		}
		content = fmt.Sprintf("%s %s: %s",
			headerStyle.Render(icon),
			m.Name,
			result)
	}

	return content
}

// ErrorMessage 错误消息
type ErrorMessage struct {
	ErrStr string
	MsgTime time.Time
}

func (m *ErrorMessage) Type() MessageType   { return MsgTypeError }
func (m *ErrorMessage) Timestamp() time.Time { return m.MsgTime }
func (m *ErrorMessage) Render() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("203")).
		Background(lipgloss.Color("17")).
		Padding(0, 1)
	return style.Render(fmt.Sprintf("Error: %s", m.ErrStr))
}

// BashInfo Bash 命令信息 - 原版以折叠框显示
type BashInfo struct {
	ID        int
	Command  string
	Output    string
	IsError  bool
	Completed bool
	MsgTime   time.Time
}

func (m *BashInfo) Type() MessageType   { return MsgTypeBash }
func (m *BashInfo) Timestamp() time.Time { return m.MsgTime }
func (m *BashInfo) Render() string {
	var statusIcon string
	var statusColor string

	if m.Completed {
		if m.IsError {
			statusIcon= "✗"
			statusColor = "203"
		} else {
			statusIcon = "✓"
			statusColor = "82"
		}
	} else {
		statusIcon = "◐"
		statusColor = "215"
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor))

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("75")).
		Bold(true)

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	content := fmt.Sprintf("%s %s $ %s",
		statusStyle.Render(statusIcon),
		headerStyle.Render("Bash"),
		commandStyle.Render(m.Command))

	if m.Completed && m.Output != "" {
		output := m.Output
		if len(output) > 500 {
			output = output[:500] + "\n... (truncated)"
		}
		content += "\n" + output
	}

	return content
}

// CommandInfo 命令信息（非 Bash）
type CommandInfo struct {
	Name    string
	Args    []string
	MsgTime time.Time
}

func (m *CommandInfo) Type() MessageType   { return MsgTypeCommand }
func (m *CommandInfo) Timestamp() time.Time { return m.MsgTime }
func (m *CommandInfo) Render() string {
	return fmt.Sprintf("**Command:** %s %s", m.Name, strings.Join(m.Args, " "))
}

// CompactResult 压缩结果
type CompactResult struct {
	Content string
	IsError bool
	MsgTime time.Time
}

func (m *CompactResult) Type() MessageType   { return MsgTypeCompact }
func (m *CompactResult) Timestamp() time.Time { return m.MsgTime }
func (m *CompactResult) Render() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		Italic(true)
	return style.Render(fmt.Sprintf("[Compacted: %s]", m.Content))
}

// LogItem 日志项
type LogItem struct {
	Level   string
	Msg     string
	MsgTime time.Time
}

func (m *LogItem) Type() MessageType   { return MsgTypeLog }
func (m *LogItem) Timestamp() time.Time { return m.MsgTime }
func (m *LogItem) Render() string {
	var color string
	switch m.Level {
	case "DEBUG":
		color = "248"
	case "INFO":
		color = "75"
	case "WARN":
		color = "215"
	case "ERROR":
		color = "203"
	default:
		color = "248"
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color))
	return style.Render(fmt.Sprintf("[%s] %s", m.Level, m.Msg))
}

// WelcomeMessage 欢迎消息 - 在 TUI 启动时显示
// 遵循官方架构的消息类型系统
type WelcomeMessage struct {
	Version string
	Cwd     string
	MsgTime time.Time
}

func (m *WelcomeMessage) Type() MessageType   { return MsgTypeSystem }
func (m *WelcomeMessage) Timestamp() time.Time { return m.MsgTime }
func (m *WelcomeMessage) Render() string {
	// 使用 lipgloss 绘制带边框的欢迎信息
	// 参考官方 TUI 实现文档中的样式系统
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("135")). // 紫色边框（官方主题色）
		Width(56) // 固定宽度，与官方一致

	// 欢迎框内容
	welcomeContent := fmt.Sprintf("✦ Welcome to Qoder CLI! %s\n\ncwd: %s",
		m.Version, m.Cwd)

	// 提示信息（不带边框）
	tipsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")). // 灰色
		MarginTop(1)

	tips := "Tips for getting started:\n\n" +
		"1. Ask questions, edit files, or run commands.\n" +
		"2. Be specific for the best results.\n" +
		"3. Type /help for more information."

	return borderStyle.Render(welcomeContent) + "\n" + tipsStyle.Render(tips)
}
