package messages

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
	"github.com/alingse/qodercli-reverse/decompiled/core/utils/markdown"
	"github.com/charmbracelet/lipgloss"
)

// getStringValue 安全地从 map 中获取字符串值，支持多种类型
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case []byte:
			return string(v)
		default:
			// 尝试转换为字符串
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// UserMessage 用户消息
type UserMessage struct {
	Content string
	MsgTime time.Time
}

func (m *UserMessage) Type() MessageType    { return MsgTypeUser }
func (m *UserMessage) Timestamp() time.Time { return m.MsgTime }
func (m *UserMessage) Render() string {
	// 样式：> 用户输入内容（前缀白色）
	// 使用纯文本前缀，避免 lipgloss 宽度计算导致换行问题
	return fmt.Sprintf("> %s", strings.TrimSpace(m.Content))
}

// String 实现 fmt.Stringer 接口
func (m *UserMessage) String() string {
	return strings.TrimSpace(m.Content)
}

// AssistantMessage 助手消息 - 支持 Markdown 渲染
type AssistantMessage struct {
	Content     string
	Rendered    string // 缓存渲染后的内容
	NeedsRender bool   // 是否需要重新渲染
	MsgTime     time.Time
}

func (m *AssistantMessage) Type() MessageType    { return MsgTypeAssistant }
func (m *AssistantMessage) Timestamp() time.Time { return m.MsgTime }

// Render 返回原始内容（供 MessageView 统一处理）
func (m *AssistantMessage) Render() string {
	return m.Content
}

// RenderMarkdown 渲染 markdown 内容
// 由 MessageView 调用，传入 markdown 渲染器
func (m *AssistantMessage) RenderMarkdown(renderer *markdown.Renderer) string {
	// 如果内容为空，直接返回
	if strings.TrimSpace(m.Content) == "" {
		return ""
	}

	// 如果已经渲染过且不需要重新渲染，返回缓存
	if !m.NeedsRender && m.Rendered != "" {
		return m.Rendered
	}

	// 如果没有渲染器，返回原始内容
	if renderer == nil {
		m.Rendered = m.Content
		m.NeedsRender = false
		return m.Rendered
	}

	// 渲染 markdown
	rendered, err := renderer.Render(m.Content)
	if err != nil {
		// 渲染失败时回退到原始内容
		m.Rendered = m.Content
	} else {
		m.Rendered = rendered
	}
	m.NeedsRender = false
	return m.Rendered
}

// AppendContent 追加内容并标记需要重新渲染
func (m *AssistantMessage) AppendContent(content string) {
	m.Content += content
	m.NeedsRender = true
}

// String 实现 fmt.Stringer 接口
func (m *AssistantMessage) String() string {
	return strings.TrimSpace(m.Content)
}

// TokenUsageMessage Token 使用统计 - 原版在消息末尾显示
type TokenUsageMessage struct {
	InputTokens  int
	OutputTokens int
	MsgTime      time.Time
}

func (m *TokenUsageMessage) Type() MessageType    { return MsgTypeSystem }
func (m *TokenUsageMessage) Timestamp() time.Time { return m.MsgTime }
func (m *TokenUsageMessage) Render() string {
	// 不设置颜色，使用终端默认颜色，只保留斜体样式
	style := lipgloss.NewStyle().
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

func (m *SystemMessage) Type() MessageType    { return MsgTypeSystem }
func (m *SystemMessage) Timestamp() time.Time { return m.MsgTime }
func (m *SystemMessage) Render() string {
	// 不设置颜色，使用终端默认颜色，只保留斜体样式
	style := lipgloss.NewStyle().
		Italic(true)
	return style.Render(fmt.Sprintf("*System: %s*", m.Content))
}

// ToolCall 工具调用消息 - 原版以折叠框显示
type ToolCall struct {
	Name      string
	Arguments string
	MsgTime   time.Time
}

func (m *ToolCall) Type() MessageType    { return MsgTypeTool }
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
	Name    string
	Result  string
	Error   error
	MsgTime time.Time
}

func (m *ToolResult) Type() MessageType    { return MsgTypeTool }
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
	ErrStr  string
	MsgTime time.Time
}

func (m *ErrorMessage) Type() MessageType    { return MsgTypeError }
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
	Command   string
	Output    string
	IsError   bool
	Completed bool
	MsgTime   time.Time
}

func (m *BashInfo) Type() MessageType    { return MsgTypeBash }
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

func (m *CommandInfo) Type() MessageType    { return MsgTypeCommand }
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

func (m *CompactResult) Type() MessageType    { return MsgTypeCompact }
func (m *CompactResult) Timestamp() time.Time { return m.MsgTime }
func (m *CompactResult) Render() string {
	// 不设置颜色，使用终端默认颜色，只保留斜体样式
	style := lipgloss.NewStyle().
		Italic(true)
	return style.Render(fmt.Sprintf("[Compacted: %s]", m.Content))
}

// LogItem 日志项
type LogItem struct {
	Level   string
	Msg     string
	MsgTime time.Time
}

func (m *LogItem) Type() MessageType    { return MsgTypeLog }
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

func (m *WelcomeMessage) Type() MessageType    { return MsgTypeSystem }
func (m *WelcomeMessage) Timestamp() time.Time { return m.MsgTime }
func (m *WelcomeMessage) Render() string {
	// 使用 lipgloss 绘制带边框的欢迎信息
	// 参考官方 TUI 实现文档中的样式系统
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("135")). // 紫色边框（官方主题色）
		Width(56)                                // 固定宽度，与官方一致

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

func (m *ToolCallInfo) Type() MessageType    { return MsgTypeTool }
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
	displayContent := ""
	var argsMap map[string]interface{}
	isTodoWrite := m.Name == "TodoWrite"
	todoLines := []string{}

	if err := json.Unmarshal([]byte(m.Arguments), &argsMap); err == nil && len(argsMap) > 0 {
		// TodoWrite 特殊处理：显示任务列表
		if isTodoWrite {
			if todos, ok := argsMap["todos"].([]interface{}); ok && len(todos) > 0 {
				for _, todo := range todos {
					if todoMap, ok := todo.(map[string]interface{}); ok {
						content, _ := todoMap["content"].(string)
						status, _ := todoMap["status"].(string)
						if content != "" {
							// 使用 [x] 表示完成，[ ] 表示进行中/待办
							checkmark := "[ ]"
							if status == "completed" || status == "done" {
								checkmark = "[x]"
							}
							todoLines = append(todoLines, fmt.Sprintf("  %s %s", checkmark, content))
						}
					}
				}
			}
		}

		// 首先尝试根据工具类型提取特定字段
		switch m.Name {
		case "Bash":
			if cmd := getStringValue(argsMap, "command"); cmd != "" {
				displayContent = cmd
			}
		case "Read", "ReadFile":
			if path := getStringValue(argsMap, "file_path"); path != "" {
				displayContent = path
			}
		case "Write", "WriteFile":
			if path := getStringValue(argsMap, "file_path"); path != "" {
				displayContent = path
			}
		case "Update", "UpdateFile":
			if path := getStringValue(argsMap, "file_path"); path != "" {
				displayContent = path
			}
		case "Grep":
			if pattern := getStringValue(argsMap, "pattern"); pattern != "" {
				displayContent = pattern
			}
		case "Glob":
			if pattern := getStringValue(argsMap, "pattern"); pattern != "" {
				displayContent = pattern
			}
		}

		// 如果特定字段提取为空，尝试提取任何非空字符串值
		// 但跳过可能包含大量文本的字段（如 content, output 等）
		if displayContent == "" && !isTodoWrite {
			skipFields := map[string]bool{
				"content": true, "output": true, "result": true,
				"stdout": true, "stderr": true, "data": true,
			}
			for k, v := range argsMap {
				if !skipFields[k] {
					if s, ok := v.(string); ok && s != "" && len(s) < 100 {
						// 只使用短字符串作为显示内容
						displayContent = fmt.Sprintf("%s: %s", k, s)
						break
					}
				}
			}
		}

		// 如果还是为空且不是 TodoWrite，显示 JSON 的精简版本
		if displayContent == "" && !isTodoWrite {
			if compactJSON, err := json.Marshal(argsMap); err == nil {
				displayContent = string(compactJSON)
			}
		}
	} else if m.Arguments != "" && m.Arguments != "{}" {
		// JSON 解析失败或为空，直接显示原始参数（去掉花括号）
		displayContent = strings.TrimSpace(m.Arguments)
		displayContent = strings.TrimPrefix(displayContent, "{")
		displayContent = strings.TrimSuffix(displayContent, "}")
		displayContent = strings.TrimSpace(displayContent)
	}

	// 如果 displayContent 为空且不是 TodoWrite，显示省略号
	if displayContent == "" && !isTodoWrite {
		displayContent = "..."
	} else if !isTodoWrite && len([]rune(displayContent)) > 256 {
		// 如果内容超过256个字符，截断并显示 ...
		runes := []rune(displayContent)
		displayContent = string(runes[:253]) + "..."
	}

	// 构建内容
	var content string
	if isTodoWrite && len(todoLines) > 0 {
		// TodoWrite：显示任务列表，每行一个任务
		content = fmt.Sprintf("%s %s", statusStyle.Render(statusIcon), nameStyle.Render(m.Name))
		content += "\n" + argsStyle.Render(strings.Join(todoLines, "\n"))
	} else {
		// 其他工具：格式 ⏺ ToolName (arguments)
		content = fmt.Sprintf("%s %s (%s)",
			statusStyle.Render(statusIcon),
			nameStyle.Render(m.Name),
			argsStyle.Render(displayContent))
	}

	// 只有在出错时才显示输出（stderr），成功时不显示 stdout
	if m.Completed && m.IsError && m.Output != "" {
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

// TodoListMessage Todo 列表消息 - 显示任务进度
type TodoListMessage struct {
	Todos    []state.Todo // 任务列表
	OldTodos []state.Todo // 更新前的任务列表（可选）
	Updated  bool         // 是否是更新操作
	MsgTime  time.Time
}

func (m *TodoListMessage) Type() MessageType    { return MsgTypeTodoList }
func (m *TodoListMessage) Timestamp() time.Time { return m.MsgTime }

// Render 渲染 Todo 列表
func (m *TodoListMessage) Render() string {
	if len(m.Todos) == 0 {
		return ""
	}

	var sb strings.Builder

	// 标题样式
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("135")).
		Bold(true)

	// 统计信息
	completedCount := 0
	inProgressCount := 0
	pendingCount := 0
	cancelledCount := 0

	for _, todo := range m.Todos {
		switch state.TodoStatus(todo.Status) {
		case state.TodoStatusCompleted, state.TodoStatusDone:
			completedCount++
		case state.TodoStatusInProgress:
			inProgressCount++
		case state.TodoStatusPending:
			pendingCount++
		case state.TodoStatusCancelled:
			cancelledCount++
		}
	}

	totalCount := len(m.Todos)

	// 标题
	if m.Updated {
		sb.WriteString(titleStyle.Render("📝 Todo List Updated"))
	} else {
		sb.WriteString(titleStyle.Render("📝 Todo List"))
	}
	sb.WriteString("\n")

	// 进度条
	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	completedPercent := 0
	if totalCount > 0 {
		completedPercent = (completedCount * 100) / totalCount
	}

	// 简单的进度条
	barWidth := 20
	filled := (completedPercent * barWidth) / 100
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	sb.WriteString(progressStyle.Render(fmt.Sprintf("Progress: [%s] %d%% (%d/%d)\n", bar, completedPercent, completedCount, totalCount)))
	sb.WriteString("\n")

	// 任务列表
	for _, todo := range m.Todos {
		var icon string
		var iconColor string
		var statusText string

		switch state.TodoStatus(todo.Status) {
		case state.TodoStatusCompleted, state.TodoStatusDone:
			icon = "✅"
			iconColor = "82" // 绿色
			statusText = "completed"
		case state.TodoStatusInProgress:
			icon = "🔄"
			iconColor = "75" // 蓝色
			statusText = "in_progress"
		case state.TodoStatusPending:
			icon = "⬜"
			iconColor = "248" // 灰色
			statusText = "pending"
		case state.TodoStatusCancelled:
			icon = "❌"
			iconColor = "203" // 红色
			statusText = "cancelled"
		default:
			icon = "⬜"
			iconColor = "248"
			statusText = todo.Status
		}

		iconStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(iconColor))

		contentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(iconColor)).
			Italic(true)

		// 任务项
		sb.WriteString(fmt.Sprintf("%s %s ",
			iconStyle.Render(icon),
			contentStyle.Render(todo.Content)))

		// 如果是 in_progress，显示 activeForm
		if state.TodoStatus(todo.Status) == state.TodoStatusInProgress && todo.ActiveForm != "" {
			activeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("75")).
				Italic(true)
			sb.WriteString(activeStyle.Render(fmt.Sprintf("(%s)", todo.ActiveForm)))
		}

		sb.WriteString(" ")
		sb.WriteString(statusStyle.Render(fmt.Sprintf("[%s]", statusText)))
		sb.WriteString("\n")
	}

	return sb.String()
}

// String 实现 fmt.Stringer 接口
func (m *TodoListMessage) String() string {
	return m.Render()
}
