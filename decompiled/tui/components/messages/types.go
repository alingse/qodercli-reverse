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

// getStringValue safely gets a string value from a map, supporting multiple types
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case []byte:
			return string(v)
		default:
			// try to convert to string
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// UserMessage user message
type UserMessage struct {
	Content string
	MsgTime time.Time
}

func (m *UserMessage) Type() MessageType    { return MsgTypeUser }
func (m *UserMessage) Timestamp() time.Time { return m.MsgTime }
func (m *UserMessage) Render() string {
	// style: > user input content (white prefix)
	// use plain text prefix to avoid lipgloss width calculation line-wrap issues
	return fmt.Sprintf("> %s", strings.TrimSpace(m.Content))
}

// String implements fmt.Stringer
func (m *UserMessage) String() string {
	return strings.TrimSpace(m.Content)
}

// AssistantMessage assistant message - supports Markdown rendering
type AssistantMessage struct {
	Content     string
	Rendered    string // cached rendered content
	NeedsRender bool   // whether re-render is needed
	MsgTime     time.Time
}

func (m *AssistantMessage) Type() MessageType    { return MsgTypeAssistant }
func (m *AssistantMessage) Timestamp() time.Time { return m.MsgTime }

// Render returns raw content (for MessageView unified handling)
func (m *AssistantMessage) Render() string {
	return m.Content
}

// RenderMarkdown renders markdown content
// called by MessageView, passing the markdown renderer
func (m *AssistantMessage) RenderMarkdown(renderer *markdown.Renderer) string {
	// if content is empty, return directly
	if strings.TrimSpace(m.Content) == "" {
		return ""
	}

	// if already rendered and no re-render needed, return cache
	if !m.NeedsRender && m.Rendered != "" {
		return m.Rendered
	}

	// if no renderer, return raw content
	if renderer == nil {
		m.Rendered = m.Content
		m.NeedsRender = false
		return m.Rendered
	}

	// render markdown
	rendered, err := renderer.Render(m.Content)
	if err != nil {
		// fall back to raw content on render failure
		m.Rendered = m.Content
	} else {
		m.Rendered = rendered
	}
	m.NeedsRender = false
	return m.Rendered
}

// AppendContent appends content and marks as needing re-render
func (m *AssistantMessage) AppendContent(content string) {
	m.Content += content
	m.NeedsRender = true
}

// String implements fmt.Stringer
func (m *AssistantMessage) String() string {
	return strings.TrimSpace(m.Content)
}

// TokenUsageMessage token usage stats - shown at end of message in original
type TokenUsageMessage struct {
	InputTokens  int
	OutputTokens int
	MsgTime      time.Time
}

func (m *TokenUsageMessage) Type() MessageType    { return MsgTypeSystem }
func (m *TokenUsageMessage) Timestamp() time.Time { return m.MsgTime }
func (m *TokenUsageMessage) Render() string {
	// no color set, use terminal default color, only italic style
	style := lipgloss.NewStyle().
		Italic(true)

	return style.Render(fmt.Sprintf(
		"Input token usage: %d | Output token usage: %d",
		m.InputTokens, m.OutputTokens))
}

// String implements fmt.Stringer
func (m *TokenUsageMessage) String() string {
	return m.Render()
}

// SystemMessage system message
type SystemMessage struct {
	Content string
	MsgTime time.Time
}

func (m *SystemMessage) Type() MessageType    { return MsgTypeSystem }
func (m *SystemMessage) Timestamp() time.Time { return m.MsgTime }
func (m *SystemMessage) Render() string {
	// no color set, use terminal default color, only italic style
	style := lipgloss.NewStyle().
		Italic(true)
	return style.Render(fmt.Sprintf("*System: %s*", m.Content))
}

// ToolCall tool call message - shown as collapsed box in original
type ToolCall struct {
	Name      string
	Arguments string
	MsgTime   time.Time
}

func (m *ToolCall) Type() MessageType    { return MsgTypeTool }
func (m *ToolCall) Timestamp() time.Time { return m.MsgTime }
func (m *ToolCall) Render() string {
	// Bash tool uses special icon and color
	if m.Name == "Bash" {
		// Bash shows white by default (executing), result changes by success/failure
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")) // white
		contentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))

		return fmt.Sprintf("%s Bash\n%s",
			headerStyle.Render("⏺"),
			contentStyle.Render(m.Arguments))
	}

	// other tools: ▶ Tool: name (blue)
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

// ToolResult tool result message
type ToolResult struct {
	Name    string
	Result  string
	Error   error
	MsgTime time.Time
}

func (m *ToolResult) Type() MessageType    { return MsgTypeTool }
func (m *ToolResult) Timestamp() time.Time { return m.MsgTime }
func (m *ToolResult) Render() string {
	// Bash tool result uses ⏺ icon
	if m.Name == "Bash" || strings.Contains(m.Name, "bash") {
		var iconStyle lipgloss.Style
		var icon string

		if m.Error != nil {
			iconStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("203")) // red
			icon = "⏺"
		} else {
			iconStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")) // green
			icon = "⏺"
		}

		contentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("248"))

		// truncate overly long result
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

	// other tool results
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
		// truncate overly long result
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

// ErrorMessage error message
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

// BashInfo Bash command info - shown as collapsed box in original
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
			statusColor = "203" // red
		} else {
			statusIcon = "⏺"
			statusColor = "82" // green
		}
	} else {
		statusIcon = "⏺"
		statusColor = "255" // white (executing)
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor))

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	// format: ⏺ bash command content
	content := fmt.Sprintf("%s %s",
		statusStyle.Render(statusIcon),
		commandStyle.Render(m.Command))

	// if there is output, show on next line
	if m.Completed && m.Output != "" {
		output := m.Output
		if len(output) > 500 {
			output = output[:500] + "\n... (truncated)"
		}
		content += "\n" + output
	}

	return content
}

// CommandInfo command info (non-Bash)
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

// CompactResult compaction result
type CompactResult struct {
	Content string
	IsError bool
	MsgTime time.Time
}

func (m *CompactResult) Type() MessageType    { return MsgTypeCompact }
func (m *CompactResult) Timestamp() time.Time { return m.MsgTime }
func (m *CompactResult) Render() string {
	// no color set, use terminal default color, only italic style
	style := lipgloss.NewStyle().
		Italic(true)
	return style.Render(fmt.Sprintf("[Compacted: %s]", m.Content))
}

// LogItem log item
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

// WelcomeMessage welcome message - shown on TUI startup
// follows official architecture message type system
type WelcomeMessage struct {
	Version string
	Cwd     string
	MsgTime time.Time
}

func (m *WelcomeMessage) Type() MessageType    { return MsgTypeSystem }
func (m *WelcomeMessage) Timestamp() time.Time { return m.MsgTime }
func (m *WelcomeMessage) Render() string {
	// use lipgloss to draw bordered welcome info
	// reference official TUI implementation docs style system
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("135")). // purple border (official theme color)
		Width(56)                                // fixed width, consistent with official

	// welcome box content
	welcomeContent := fmt.Sprintf("✦ Welcome to Qoder CLI! %s\n\ncwd: %s",
		m.Version, m.Cwd)

	// tips info (no border)
	tipsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")). // gray
		MarginTop(1)

	tips := "Tips for getting started:\n\n" +
		"1. Ask questions, edit files, or run commands.\n" +
		"2. Be specific for the best results.\n" +
		"3. Type /help for more information."

	return borderStyle.Render(welcomeContent) + "\n" + tipsStyle.Render(tips)
}

// ToolCallInfo unified tool call info display
// all tools (Bash, Read, Write, Update, etc.) use this type for display
type ToolCallInfo struct {
	ID        string    // tool call unique ID
	Name      string    // tool name
	Arguments string    // tool arguments (JSON format)
	Output    string    // execution output
	Completed bool      // whether completed
	IsError   bool      // whether execution errored
	MsgTime   time.Time // message time
}

func (m *ToolCallInfo) Type() MessageType    { return MsgTypeTool }
func (m *ToolCallInfo) Timestamp() time.Time { return m.MsgTime }
func (m *ToolCallInfo) Render() string {
	var statusIcon string
	var statusColor string

	if m.Completed {
		if m.IsError {
			statusIcon = "⏺"
			statusColor = "203" // red - execution failed
		} else {
			statusIcon = "⏺"
			statusColor = "82" // green - execution succeeded
		}
	} else {
		statusIcon = "⏺"
		statusColor = "255" // white - executing
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor))

	// tool name style
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")). // bright white
		Bold(true)

	// arguments style
	argsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")) // gray

	// output style
	outputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")) // bright white

	// parse arguments, extract key info for display
	displayContent := ""
	var argsMap map[string]interface{}
	isTodoWrite := m.Name == "TodoWrite"
	todoLines := []string{}

	if err := json.Unmarshal([]byte(m.Arguments), &argsMap); err == nil && len(argsMap) > 0 {
		// TodoWrite special handling: show task list
		if isTodoWrite {
			if todos, ok := argsMap["todos"].([]interface{}); ok && len(todos) > 0 {
				for _, todo := range todos {
					if todoMap, ok := todo.(map[string]interface{}); ok {
						content, _ := todoMap["content"].(string)
						status, _ := todoMap["status"].(string)
						if content != "" {
							// use [x] for completed, [ ] for in-progress/pending
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

		// first try to extract specific fields by tool type
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

		// if specific field extraction is empty, try extracting any non-empty string value
		// but skip fields that may contain large text (content, output, etc.)
		if displayContent == "" && !isTodoWrite {
			skipFields := map[string]bool{
				"content": true, "output": true, "result": true,
				"stdout": true, "stderr": true, "data": true,
			}
			for k, v := range argsMap {
				if !skipFields[k] {
					if s, ok := v.(string); ok && s != "" && len(s) < 100 {
						// only use short strings as display content
						displayContent = fmt.Sprintf("%s: %s", k, s)
						break
					}
				}
			}
		}

		// if still empty and not TodoWrite, show compact JSON version
		if displayContent == "" && !isTodoWrite {
			if compactJSON, err := json.Marshal(argsMap); err == nil {
				displayContent = string(compactJSON)
			}
		}
	} else if m.Arguments != "" && m.Arguments != "{}" {
		// JSON parse failed or empty, display raw arguments directly (remove braces)
		displayContent = strings.TrimSpace(m.Arguments)
		displayContent = strings.TrimPrefix(displayContent, "{")
		displayContent = strings.TrimSuffix(displayContent, "}")
		displayContent = strings.TrimSpace(displayContent)
	}

	// if displayContent is empty and not TodoWrite, show ellipsis
	if displayContent == "" && !isTodoWrite {
		displayContent = "..."
	} else if !isTodoWrite && len([]rune(displayContent)) > 256 {
		// if content exceeds 256 chars, truncate and show ...
		runes := []rune(displayContent)
		displayContent = string(runes[:253]) + "..."
	}

	// build content
	var content string
	if isTodoWrite && len(todoLines) > 0 {
		// TodoWrite: show task list, one task per line
		content = fmt.Sprintf("%s %s", statusStyle.Render(statusIcon), nameStyle.Render(m.Name))
		content += "\n" + argsStyle.Render(strings.Join(todoLines, "\n"))
	} else {
		// other tools: format ⏺ ToolName (arguments)
		content = fmt.Sprintf("%s %s (%s)",
			statusStyle.Render(statusIcon),
			nameStyle.Render(m.Name),
			argsStyle.Render(displayContent))
	}

	// only show output (stderr) on error; do not show stdout on success
	if m.Completed && m.IsError && m.Output != "" {
		output := m.Output
		// limit output lines
		lines := strings.Split(output, "\n")
		if len(lines) > 10 {
			output = strings.Join(lines[:10], "\n") + "\n... (truncated)"
		}
		// limit total length
		if len(output) > 500 {
			output = output[:500] + "\n... (truncated)"
		}
		content += "\n" + outputStyle.Render(output)
	}

	return content
}

// TodoListMessage todo list message - shows task progress
type TodoListMessage struct {
	Todos    []state.Todo // task list
	OldTodos []state.Todo // task list before update (optional)
	Updated  bool         // whether this is an update operation
	MsgTime  time.Time
}

func (m *TodoListMessage) Type() MessageType    { return MsgTypeTodoList }
func (m *TodoListMessage) Timestamp() time.Time { return m.MsgTime }

// Render renders the Todo list
func (m *TodoListMessage) Render() string {
	if len(m.Todos) == 0 {
		return ""
	}

	var sb strings.Builder

	// title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("135")).
		Bold(true)

	// stats
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

	// title
	if m.Updated {
		sb.WriteString(titleStyle.Render("📝 Todo List Updated"))
	} else {
		sb.WriteString(titleStyle.Render("📝 Todo List"))
	}
	sb.WriteString("\n")

	// progress bar
	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	completedPercent := 0
	if totalCount > 0 {
		completedPercent = (completedCount * 100) / totalCount
	}

	// simple progress bar
	barWidth := 20
	filled := (completedPercent * barWidth) / 100
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	sb.WriteString(progressStyle.Render(fmt.Sprintf("Progress: [%s] %d%% (%d/%d)\n", bar, completedPercent, completedCount, totalCount)))
	sb.WriteString("\n")

	// task list
	for _, todo := range m.Todos {
		var icon string
		var iconColor string
		var statusText string

		switch state.TodoStatus(todo.Status) {
		case state.TodoStatusCompleted, state.TodoStatusDone:
			icon = "✅"
			iconColor = "82" // green
			statusText = "completed"
		case state.TodoStatusInProgress:
			icon = "🔄"
			iconColor = "75" // blue
			statusText = "in_progress"
		case state.TodoStatusPending:
			icon = "⬜"
			iconColor = "248" // gray
			statusText = "pending"
		case state.TodoStatusCancelled:
			icon = "❌"
			iconColor = "203" // red
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

		// task item
		sb.WriteString(fmt.Sprintf("%s %s ",
			iconStyle.Render(icon),
			contentStyle.Render(todo.Content)))

		// if in_progress, show activeForm
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

// String implements fmt.Stringer
func (m *TodoListMessage) String() string {
	return m.Render()
}
