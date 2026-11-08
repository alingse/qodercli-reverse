// Package editor interactive editor component
// 1:1 restore original qodercli-reverse EditorComponent
package editor

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
)

// Msg editor message type
type Msg interface {
	tea.Msg
}

// SendMsg send message command
type SendMsg struct {
	Content     string
	Attachments []string
}

// ClearMsg clear input
type ClearMsg struct{}

// Mode editor mode
type Mode int

const (
	ModeInsert Mode = iota
	ModeNormal
	ModeVisual
)

// EditorComponent editor component - 1:1 restore original
type EditorComponent struct {
	// core textarea
	textarea textarea.Model

	// size
	width  int
	height int

	// state
	mode    Mode
	focused bool
	enabled bool

	// feature handlers
	attachmentHandler *AttachmentHandler
	historyHandler    *HistoryHandler
	inputCache        *InputCache

	// config
	placeholder string

	// event publishing
	pubsub *pubsub.PubSub

	// cache last line count, used to detect line count changes
	lastLineCount int
}

// NewEditorComponent creates a new editor component
func NewEditorComponent(ps *pubsub.PubSub) *EditorComponent {
	ta := textarea.New()
	ta.Placeholder = "Ask anything..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 100000
	// official style: use "> " as prompt
	ta.Prompt = "> "
	// set styles
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	// text style: white
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	// prompt style: gray
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	// cursor style
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()

	// fix: modify KeyMap to distinguish Enter and Shift+Enter
	// Enter sends message, Shift+Enter inserts newline
	ta.KeyMap.InsertNewline = key.NewBinding(
		key.WithKeys("shift+enter"),
		key.WithHelp("shift+enter", "insert newline"),
	)

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
		// set default size
		width:         80,
		height:        3,
		lastLineCount: 1, // initially 1 line
	}

	// set textarea default size
	ec.textarea.SetWidth(76) // 80 - 4 (border and padding)
	ec.textarea.SetHeight(1) // single-line input

	return ec
}

// Init initializes
func (ec *EditorComponent) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		ec.historyHandler.Init(),
	)
}

// Update updates state - core message handling
func (ec *EditorComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !ec.enabled {
		return ec, nil
	}

	var cmds []tea.Cmd
	// mark whether message was handled by handleKeyMsg (to prevent passing to textarea)
	keyHandled := false

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ec.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		cmd, handled := ec.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		keyHandled = handled

	case tea.MouseMsg:
		// handle mouse events
		if msg.Type == tea.MouseLeft {
			ec.Focus()
		}

	case ClearMsg:
		ec.resetInputs()
		return ec, nil
	}

	// update textarea (only pass when key message not handled)
	if ec.focused && !keyHandled {
		newTextarea, cmd := ec.textarea.Update(msg)
		ec.textarea = newTextarea
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// detect line count changes, sync textarea height
	// this ensures editor height responds immediately when user adds/removes newlines
	ec.syncHeight()

	return ec, tea.Batch(cmds...)
}

// handleKeyMsg handles keyboard messages - original core logic
// return value: (command, handled) - if handled, message is not passed to textarea
func (ec *EditorComponent) handleKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type {
	case tea.KeyEsc:
		return ec.handleEscape(), true

	case tea.KeyEnter:
		// check Shift+Enter - not handled, let textarea insert newline
		if msg.String() == "shift+enter" {
			return nil, false
		}
		// Enter sends message - handled, not passed to textarea
		return ec.sendMessage(), true

	case tea.KeyCtrlC:
		if ec.textarea.Value() == "" {
			return tea.Quit, true
		}

	case tea.KeyUp:
		if ec.atTopOfInput() {
			return ec.handleHistoryNavigation(-1), true
		}

	case tea.KeyDown:
		if ec.atBottomOfInput() {
			return ec.handleHistoryNavigation(1), true
		}

	case tea.KeyTab:
		// Tab: insert spaces - handled
		ec.textarea.InsertString("    ")
		return nil, true

	case tea.KeyPgUp, tea.KeyPgDown:
		// page scroll - not handled, passed to parent component
		return nil, false
	}

	return nil, false
}

// handleEscape handles Esc key
func (ec *EditorComponent) handleEscape() tea.Cmd {
	ec.resetEscState()
	return nil
}

// resetEscState resets Esc state
func (ec *EditorComponent) resetEscState() {
	ec.textarea.Blur()
	ec.focused = false
}

// resetInputs resets input
func (ec *EditorComponent) resetInputs() {
	ec.textarea.Reset()
	ec.textarea.Focus()
	ec.focused = true
	ec.attachmentHandler.Reset()
	// reset textarea height to single line (1 line content)
	ec.textarea.SetHeight(1)
	// reset line count cache
	ec.lastLineCount = 1
}

// sendMessage sends message - core functionality
func (ec *EditorComponent) sendMessage() tea.Cmd {
	content := strings.TrimSpace(ec.textarea.Value())
	if content == "" && !ec.attachmentHandler.HasAttachments() {
		return nil
	}

	// cache input
	ec.inputCache.SetInput(content)

	// add history record
	ec.historyHandler.AddHistory(content)

	// build message
	msg := SendMsg{
		Content:     content,
		Attachments: ec.attachmentHandler.GetAttachments(),
	}

	// publish event
	if ec.pubsub != nil {
		ec.pubsub.Publish(context.Background(), pubsub.Event{
			Type:    pubsub.EventTypeUserInput,
			Payload: msg,
		})
	}

	// clear input
	ec.resetInputs()

	return func() tea.Msg {
		return msg
	}
}

// handleHistoryNavigation history navigation
func (ec *EditorComponent) handleHistoryNavigation(direction int) tea.Cmd {
	history := ec.historyHandler.Navigate(direction)
	if history != "" {
		ec.textarea.SetValue(history)
		// move cursor to end
		ec.textarea.CursorEnd()
	}
	return nil
}

// atTopOfInput whether at top of input box
func (ec *EditorComponent) atTopOfInput() bool {
	// check if cursor is on first line
	return ec.textarea.Line() == 0
}

// atBottomOfInput whether at bottom of input box
func (ec *EditorComponent) atBottomOfInput() bool {
	// get total line count
	lineCount := ec.textarea.LineCount()
	return ec.textarea.Line() >= lineCount-1
}

// GetLineCount gets current line count
func (ec *EditorComponent) GetLineCount() int {
	return ec.textarea.LineCount()
}

// syncHeight syncs textarea height with content line count
// updates immediately when line count changes
func (ec *EditorComponent) syncHeight() {
	currentLines := ec.textarea.LineCount()

	// if line count unchanged, skip
	if currentLines == ec.lastLineCount {
		return
	}

	ec.lastLineCount = currentLines

	// limit content lines to 1-5 range
	contentLines := currentLines
	if contentLines > 5 {
		contentLines = 5
	}
	if contentLines < 1 {
		contentLines = 1
	}

	// set textarea internal height
	ec.textarea.SetHeight(contentLines)
}

// GetPreferredHeight gets preferred height (content lines + border + padding)
func (ec *EditorComponent) GetPreferredHeight() int {
	// content lines (max 5)
	contentLines := ec.textarea.LineCount()
	if contentLines > 5 {
		contentLines = 5
	}
	// at least 1 content line
	if contentLines < 1 {
		contentLines = 1
	}
	// total height = content lines + border (2 lines)
	totalHeight := contentLines + 2
	// limit max height to 7 lines (5 content + 2 border)
	if totalHeight > 7 {
		totalHeight = 7
	}
	return totalHeight
}

// View renders view
func (ec *EditorComponent) View() string {
	if !ec.focused {
		ec.textarea.Blur()
	} else {
		ec.textarea.Focus()
	}

	// build editor view
	var sb strings.Builder

	// render attachments (if any)
	if ec.attachmentHandler.HasAttachments() {
		sb.WriteString(ec.attachmentHandler.Render())
		sb.WriteString("\n")
	}

	// render textarea - official style: rounded border
	borderColor := lipgloss.Color("240") // gray border
	if ec.focused {
		borderColor = lipgloss.Color("255") // white border when focused
	}

	// use single-line style
	editorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1). // left/right padding 1
		Width(ec.width)

	sb.WriteString(editorStyle.Render(ec.textarea.View()))

	return sb.String()
}

// SetSize sets size
func (ec *EditorComponent) SetSize(width, height int) {
	ec.width = width
	ec.height = height
	// subtract border (2 chars) and padding (2 chars)
	ec.textarea.SetWidth(width - 4)
	// textarea max height is 5 lines (auto-expands for multiline)
	// limit internal height to 1-5 lines
	innerHeight := height - 2 // subtract border
	if innerHeight < 1 {
		innerHeight = 1
	}
	if innerHeight > 5 {
		innerHeight = 5
	}
	ec.textarea.SetHeight(innerHeight)
}

// Focus focuses
func (ec *EditorComponent) Focus() {
	ec.focused = true
	ec.textarea.Focus()
}

// Blur blurs
func (ec *EditorComponent) Blur() {
	ec.focused = false
	ec.textarea.Blur()
}

// GetInput gets current input
func (ec *EditorComponent) GetInput() string {
	return ec.textarea.Value()
}

// SetContent sets content
func (ec *EditorComponent) SetContent(content string) {
	ec.textarea.SetValue(content)
}

// IsEmpty whether empty
func (ec *EditorComponent) IsEmpty() bool {
	return strings.TrimSpace(ec.textarea.Value()) == ""
}

// HasRunningShells whether has running shells
func (ec *EditorComponent) HasRunningShells() bool {
	// simplified implementation
	return false
}

// CancelRunningShells cancels running shells
func (ec *EditorComponent) CancelRunningShells() {
	// simplified implementation
}

// addAttachment adds attachment
func (ec *EditorComponent) addAttachment(path string) error {
	return ec.attachmentHandler.AddAttachment(path)
}
