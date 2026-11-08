// Package messages message display component
// 1:1 restore original MessageView
package messages

import (
	"fmt"
	"strings"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
	"github.com/alingse/qodercli-reverse/decompiled/core/utils/markdown"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MessageType message type
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

// Message message interface
type Message interface {
	Type() MessageType
	Render() string
	Timestamp() time.Time
}

// MessageView message view component - supports infinite scroll and Markdown rendering
type MessageView struct {
	viewport viewport.Model
	messages []Message
	renderer *markdown.Renderer

	width  int
	height int

	// pending messages
	pendingMessages []Message

	// auto scroll
	autoScroll bool

	// cached rendered content
	cachedContent string
}

// NewMessageView creates a new message view
func NewMessageView() *MessageView {
	vp := viewport.New(80, 20)
	vp.SetContent("")
	// enable mouse support so scrolling works
	vp.MouseWheelEnabled = true

	// create markdown renderer, auto-detect terminal theme
	renderer, _ := markdown.NewRenderer(76, "")

	return &MessageView{
		viewport:        vp,
		messages:        make([]Message, 0),
		pendingMessages: make([]Message, 0),
		renderer:        renderer,
		autoScroll:      true,
	}
}

// Init initializes
func (mv *MessageView) Init() tea.Cmd {
	return nil
}

// Update updates state
func (mv *MessageView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mv.SetSize(msg.Width, msg.Height)
	}

	// update viewport
	newVp, cmd := mv.viewport.Update(msg)
	mv.viewport = newVp

	return mv, cmd
}

// View renders view - returns full content directly, no viewport clipping
func (mv *MessageView) View() string {
	// return cached full content directly, let terminal scroll naturally
	return mv.cachedContent
}

// AddMessage adds a message
func (mv *MessageView) AddMessage(msg Message) {
	mv.messages = append(mv.messages, msg)
	// render content to cache
	mv.renderContent()
}

// AddUserMessage adds a user message
func (mv *MessageView) AddUserMessage(content string) {
	msg := &UserMessage{
		Content: content,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AddAssistantMessage adds an assistant message
func (mv *MessageView) AddAssistantMessage(content string) {
	msg := &AssistantMessage{
		Content: content,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AddSystemMessage adds a system message (for stats, etc.)
func (mv *MessageView) AddSystemMessage(content string) {
	msg := &AssistantMessage{
		Content: content,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AppendToLastMessage appends to the last message
func (mv *MessageView) AppendToLastMessage(content string) {
	if len(mv.messages) == 0 {
		mv.AddAssistantMessage(content)
		return
	}

	lastMsg := mv.messages[len(mv.messages)-1]
	if am, ok := lastMsg.(*AssistantMessage); ok {
		am.AppendContent(content) // use AppendContent method, auto-mark needs re-render
		// re-render content
		mv.renderContent()
	} else {
		mv.AddAssistantMessage(content)
	}
}

// AddBashInfo adds Bash command info
func (mv *MessageView) AddBashInfo(command string) int {
	msg := &BashInfo{
		Command: command,
		MsgTime: time.Now(),
		ID:      len(mv.messages),
	}
	mv.AddMessage(msg)
	return msg.ID
}

// UpdateBashResult updates Bash result
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

// AddPendingBash adds a pending Bash command (white ⏺)
// note: deprecated, use AddPendingTool instead
func (mv *MessageView) AddPendingBash(command string) {
	mv.AddPendingTool(fmt.Sprintf("bash-%d", len(mv.messages)), "Bash", command)
}

// CompleteBash completes a Bash command, updating status
// note: deprecated, use CompleteTool instead
func (mv *MessageView) CompleteBash(output string, isError bool) {
	// find the last unfinished tool message
	mv.CompleteTool("", "Bash", output, isError)
}

// AddPendingTool adds a pending tool call (white ⏺)
// all tool calls use this method
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

// CompleteTool completes a tool call, updating status
// toolID: unique tool call id
// toolName: tool name (for display)
// output: execution result
// isError: whether execution errored
func (mv *MessageView) CompleteTool(toolID, toolName, output string, isError bool) {
	// find corresponding pending tool message
	for i := len(mv.messages) - 1; i >= 0; i-- {
		if toolInfo, ok := mv.messages[i].(*ToolCallInfo); ok {
			// match ID (if provided) or last unfinished same-named tool
			if !toolInfo.Completed {
				if toolID == "" || toolInfo.ID == toolID {
					toolInfo.Output = output
					toolInfo.IsError = isError
					toolInfo.Completed = true
					// re-render
					mv.renderContent()
					break
				}
			}
		}
	}
}

// AddToolCall adds a tool call
func (mv *MessageView) AddToolCall(name, arguments string) {
	msg := &ToolCall{
		Name:      name,
		Arguments: arguments,
		MsgTime:   time.Now(),
	}
	mv.AddMessage(msg)
}

// AddToolResult adds a tool result
func (mv *MessageView) AddToolResult(name, result string, err error) {
	msg := &ToolResult{
		Name:    name,
		Result:  result,
		Error:   err,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AddError adds an error message
func (mv *MessageView) AddError(err string) {
	msg := &ErrorMessage{
		ErrStr:  err,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// AddTokenUsage adds token usage stats
func (mv *MessageView) AddTokenUsage(inputTokens, outputTokens int) {
	msg := &TokenUsageMessage{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		MsgTime:      time.Now(),
	}
	mv.AddMessage(msg)
}

// Clear clears messages
func (mv *MessageView) Clear() {
	mv.messages = make([]Message, 0)
	mv.viewport.SetContent("")
}

// SetSize sets size
func (mv *MessageView) SetSize(width, height int) {
	mv.width = width
	mv.height = height

	// save current scroll position (YOffset) and check if at bottom
	oldYOffset := mv.viewport.YOffset
	wasAtBottom := mv.viewport.AtBottom()

	mv.viewport.Width = width
	mv.viewport.Height = height

	// restore scroll position to avoid content loss
	// but ensure it does not exceed new content range
	mv.viewport.YOffset = oldYOffset

	// update markdown renderer size
	if mv.renderer != nil {
		mv.renderer.SetSize(width - 4) // leave margin
	}

	// re-render content to fit new width
	mv.renderContent()

	// if previously at bottom and auto scroll enabled, stay at bottom
	if wasAtBottom && mv.autoScroll {
		mv.viewport.GotoBottom()
	}
}

// renderContent renders content to cache
func (mv *MessageView) renderContent() {
	var sb strings.Builder

	for i, msg := range mv.messages {
		if i > 0 {
			prevMsg := mv.messages[i-1]
			// add blank line only between different message types
			// no blank line between same type (e.g. multiple stream appends)
			if msg.Type() != prevMsg.Type() {
				sb.WriteString("\n\n")
			} else {
				// add spacing between same type messages
				// extra blank line for tool call messages to distinguish
				if msg.Type() == MsgTypeTool {
					sb.WriteString("\n\n")
				} else {
					// only one newline between other same-type messages
					sb.WriteString("\n")
				}
			}
		}

		// choose render method by message type
		switch m := msg.(type) {
		case *AssistantMessage:
			// use markdown renderer to render assistant message
			rendered := m.RenderMarkdown(mv.renderer)
			sb.WriteString(rendered)
		case *UserMessage:
			// user message uses simple format
			sb.WriteString(mv.renderUserMessage(m))
		case *LogItem:
			// log message supports Markdown rendering (keep original color prefix)
			sb.WriteString(mv.renderLogMessage(m))
		case *SystemMessage, *ErrorMessage, *CompactResult:
			// system, error, compact result support Markdown rendering
			sb.WriteString(mv.renderWithMarkdown(msg.Render()))
		case *ToolResult:
			// tool result supports Markdown rendering
			sb.WriteString(mv.renderToolResultWithMarkdown(m))
		default:
			// other message types use default render
			sb.WriteString(msg.Render())
		}
	}

	// save to cache for View() method
	mv.cachedContent = sb.String()
}

// renderUserMessage renders user message
func (mv *MessageView) renderUserMessage(m *UserMessage) string {
	// use lipgloss to style user message
	prefixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	return fmt.Sprintf("%s %s",
		prefixStyle.Render(">"),
		strings.TrimSpace(m.Content))
}

// renderLogMessage renders log message, supports Markdown
func (mv *MessageView) renderLogMessage(m *LogItem) string {
	// keep original color level prefix
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

	prefixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color))

	// use markdown renderer to render message content
	var content string
	if mv.renderer != nil {
		rendered, err := mv.renderer.Render(m.Msg)
		if err == nil {
			content = rendered
		} else {
			content = m.Msg
		}
	} else {
		content = m.Msg
	}

	// add ANSI reset sequence before content to avoid style pollution
	// \x1b[0m resets all styles to default
	return prefixStyle.Render(fmt.Sprintf("[%s]", m.Level)) + " \x1b[0m" + content
}

// renderWithMarkdown renders content with Markdown renderer
// used for system messages, error messages, etc. needing Markdown support
func (mv *MessageView) renderWithMarkdown(content string) string {
	if mv.renderer != nil {
		rendered, err := mv.renderer.Render(content)
		if err == nil {
			return rendered
		}
	}
	return content
}

// renderToolResultWithMarkdown renders tool result, supports Markdown
func (mv *MessageView) renderToolResultWithMarkdown(m *ToolResult) string {
	// first get basic rendered content (tool name, status, etc.)
	baseContent := m.Render()

	// if there is result content and not a Bash tool, try Markdown rendering
	if m.Result != "" && m.Name != "Bash" && !strings.Contains(m.Name, "bash") {
		var renderedResult string
		if mv.renderer != nil {
			rendered, err := mv.renderer.Render(m.Result)
			if err == nil {
				renderedResult = rendered
			} else {
				renderedResult = m.Result
			}
		} else {
			renderedResult = m.Result
		}

		// return base content + Markdown-rendered result
		return baseContent + "\n" + renderedResult
	}

	return baseContent
}

// ScrollToBottom scroll to bottom (no-op, using terminal native scroll)
func (mv *MessageView) ScrollToBottom() {
	// no longer needed, content scrolls naturally
}

// ScrollToTop scroll to top (no-op, using terminal native scroll)
func (mv *MessageView) ScrollToTop() {
	// no longer needed, use terminal scrollbar
}

// PageUp page up (no-op, using terminal native scroll)
func (mv *MessageView) PageUp() {
	// no longer needed, use terminal scrollbar
}

// PageDown page down (no-op, using terminal native scroll)
func (mv *MessageView) PageDown() {
	// no longer needed, use terminal scrollbar
}

// HalfPageUp half page up (no-op, using terminal native scroll)
func (mv *MessageView) HalfPageUp() {
	// no longer needed, use terminal scrollbar
}

// HalfPageDown half page down (no-op, using terminal native scroll)
func (mv *MessageView) HalfPageDown() {
	// no longer needed, use terminal scrollbar
}

// GetLastMessage gets the last message
func (mv *MessageView) GetLastMessage() Message {
	if len(mv.messages) == 0 {
		return nil
	}
	return mv.messages[len(mv.messages)-1]
}

// AddTodoList adds a Todo list message
func (mv *MessageView) AddTodoList(todos []state.Todo) {
	msg := &TodoListMessage{
		Todos:   todos,
		Updated: false,
		MsgTime: time.Now(),
	}
	mv.AddMessage(msg)
}

// UpdateTodoList updates a Todo list message
func (mv *MessageView) UpdateTodoList(todos []state.Todo, oldTodos []state.Todo) {
	// check if an existing TodoListMessage exists
	var existingMsg *TodoListMessage
	for i := len(mv.messages) - 1; i >= 0; i-- {
		if tlm, ok := mv.messages[i].(*TodoListMessage); ok {
			existingMsg = tlm
			// remove old message from list
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

// GetTodoList gets current Todo list (if exists)
func (mv *MessageView) GetTodoList() []state.Todo {
	for i := len(mv.messages) - 1; i >= 0; i-- {
		if tlm, ok := mv.messages[i].(*TodoListMessage); ok {
			return tlm.Todos
		}
	}
	return nil
}

// GetAllMessagesText gets all messages as plain text (for terminal output on exit)
func (mv *MessageView) GetAllMessagesText() string {
	var sb strings.Builder
	for i, msg := range mv.messages {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		switch m := msg.(type) {
		case *UserMessage:
			sb.WriteString(fmt.Sprintf("👤 %s", m.Content))
		case *AssistantMessage:
			sb.WriteString(fmt.Sprintf("🤖 %s", m.Content))
		case *ToolCallInfo:
			status := "⏺"
			if m.Completed {
				if m.IsError {
					status = "❌"
				} else {
					status = "✓"
				}
			}
			sb.WriteString(fmt.Sprintf("%s %s", status, m.Name))
			if m.Output != "" {
				sb.WriteString(fmt.Sprintf("\n  %s", m.Output))
			}
		case *ErrorMessage:
			sb.WriteString(fmt.Sprintf("❌ Error: %s", m.ErrStr))
		case *WelcomeMessage:
			sb.WriteString(fmt.Sprintf("🚀 Qoder CLI %s\n📁 %s", m.Version, m.Cwd))
		default:
			sb.WriteString(msg.Render())
		}
	}
	return sb.String()
}

// AtTop checks if at top
func (mv *MessageView) AtTop() bool {
	return mv.viewport.AtTop()
}

// AtBottom checks if at bottom
func (mv *MessageView) AtBottom() bool {
	return mv.viewport.AtBottom()
}
