package messages

import (
	"encoding/json"
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
	// 样式：> 用户输入内容（前缀白色）
	prefixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")) // 白色
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")) // 亮白色

	return fmt.Sprintf("%s %s",
		prefixStyle.Render(">"),
		contentStyle.Render(m.Content))
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
	// 样式：⏺ 系统输出（白色），与内容在同一行
	prefixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")) // 白色

	return fmt.Sprintf("%s %s",
		prefixStyle.Render("⏺"),
		m.Content)
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
	// Bash 工具使用特殊图标和颜色
	if m.Name == "Bash" {
		// Bash 默认显示为白色（执行中），结果会根据成功/失败改变
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")) // 白色
		contentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))

		return fmt.Sprintf("%s Bash\n%s",
			headerStyle.Render("⏺"),
			contentStyle.Render(m.Arguments))
	}

	// 其他工具：▶ Tool: name (蓝色)
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
	// Bash 工具结果使用 ⏺ 图标
	if m.Name == "Bash" || strings.Contains(m.Name, "bash") {
		var iconStyle lipgloss.Style
		var icon string

		if m.Error != nil {
			iconStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("203")) // 红色
			icon = "⏺"
		} else {
			iconStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")) // 绿色
			icon = "⏺"
		}

		contentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))

		// 截断过长的结果
		result := m.Result
		if result == "" && m.Error != nil {
			result = m.Error.Error()
		}
		lines := strings.Split(result, "\n")
		if len(lines) > 20 {
			result = strings.Join(lines[:20], "\n") + "\n... (truncated)"
		}

		return fmt.Sprintf("%s %s",
			iconStyle.Render(icon),
			contentStyle.Render(result))
	}

	// 其他工具结果
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
			statusIcon = "⏺"
			statusColor = "203" // 红色
		} else {
			statusIcon = "⏺"
			statusColor = "82" // 绿色
		}
	} else {
		statusIcon = "⏺"
		statusColor = "255" // 白色（执行中）
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor))

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	// 格式：⏺ bash 命令内容
	content := fmt.Sprintf("%s %s",
		statusStyle.Render(statusIcon),
		commandStyle.Render(m.Command))

	// 如果有输出，显示在下一行
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

// ToolCallInfo 统一的工具调用信息显示
// 所有工具（Bash、Read、Write、Update 等）都使用此类型显示
type ToolCallInfo struct {
	ID        string    // 工具调用唯一 ID
	Name      string    // 工具名称
	Arguments string    // 工具参数（JSON 格式）
	Output    string    // 执行输出
	Completed bool      // 是否已完成
	IsError   bool      // 是否执行出错
	MsgTime   time.Time // 消息时间
}

func (m *ToolCallInfo) Type() MessageType   { return MsgTypeTool }
func (m *ToolCallInfo) Timestamp() time.Time { return m.MsgTime }
func (m *ToolCallInfo) Render() string {
	var statusIcon string
	var statusColor string

	if m.Completed {
		if m.IsError {
			statusIcon = "⏺"
			statusColor = "203" // 红色 - 执行失败
		} else {
			statusIcon = "⏺"
			statusColor = "82" // 绿色 - 执行成功
		}
	} else {
		statusIcon = "⏺"
		statusColor = "255" // 白色 - 执行中
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor))

	// 工具名称样式
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")). // 亮白色
		Bold(true)

	// 参数样式
	argsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")) // 灰色

	// 输出样式
	outputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")) // 亮白色

	// 解析参数，提取关键信息用于显示
	displayContent := m.Arguments
	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(m.Arguments), &argsMap); err == nil {
		// 根据不同工具类型提取关键信息
		switch m.Name {
		case "Bash":
			if cmd, ok := argsMap["command"].(string); ok && cmd != "" {
				displayContent = cmd
			}
		case "Read", "ReadFile":
			if path, ok := argsMap["file_path"].(string); ok && path != "" {
				displayContent = path
			}
		case "Write", "WriteFile":
			if path, ok := argsMap["file_path"].(string); ok && path != "" {
				displayContent = path
			}
		case "Update", "UpdateFile":
			if path, ok := argsMap["file_path"].(string); ok && path != "" {
				displayContent = path
			}
		default:
			// 对于其他工具，尝试提取第一个字符串类型的值
			for _, v := range argsMap {
				if s, ok := v.(string); ok && s != "" {
					displayContent = s
					break
				}
			}
		}
	}

	// 格式：⏺ ToolName arguments
	content := fmt.Sprintf("%s %s %s",
		statusStyle.Render(statusIcon),
		nameStyle.Render(m.Name),
		argsStyle.Render(displayContent))

	// 如果有输出，显示在下一行（限制长度）
	if m.Completed && m.Output != "" {
		output := m.Output
		// 限制输出行数
		lines := strings.Split(output, "\n")
		if len(lines) > 10 {
			output = strings.Join(lines[:10], "\n") + "\n... (truncated)"
		}
		// 限制总长度
		if len(output) > 500 {
			output = output[:500] + "\n... (truncated)"
		}
		content += "\n" + outputStyle.Render(output)
	}

	return content
}
