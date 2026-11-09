// Package app TUI application entry point
// Follows Bubble Tea architecture: all state changes must go through the Update loop
package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
	"github.com/alingse/qodercli-reverse/decompiled/core/utils/markdown"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/interaction/editor"
	"github.com/alingse/qodercli-reverse/decompiled/tui/components/messages"
	"github.com/alingse/qodercli-reverse/decompiled/version"
)

// channel buffer size constants
const (
	DefaultEventChanSize    = 100  // event channel buffer size
	DefaultCallbackChanSize = 1000 // callback channel buffer size
)

// status text constants
const (
	StatusReady     = "Ready"
	StatusThinking  = "Thinking..."
	StatusUsingTool = "Using tool..."
	StatusError     = "Error"
)

// color code constants (Lipgloss ANSI colors)
const (
	ColorGreen     = "82"  // green - ready/completed status
	ColorPurple    = "135" // purple - processing status
	ColorRed       = "203" // red - error status
	ColorGray      = "248" // gray - pending/default text
	ColorGrayLight = "246" // light gray - preview text
	ColorWhite     = "255" // white - highlighted text/default
	ColorDarkBg    = "237" // dark background
)

// UI size constants
const (
	StatusBarHeight      = 1
	MinHelpWidth         = 20
	HelpBorderWidth      = 4
	StatusBarLeftWidth   = 20
	StatusBarRightMargin = 40
)

// appModel TUI application model - follows Bubble Tea reactive architecture
type appModel struct {
	// core components
	editor  *editor.EditorComponent
	msgView *messages.MessageView // kept for formatting messages

	// Markdown renderer
	markdownRenderer *markdown.Renderer

	// config
	config *config.Config

	// Agent
	agent *agent.Agent

	// event system
	pubsub *pubsub.PubSub

	// size
	width  int
	height int

	// status
	quitting      bool
	processing    bool
	showHelp      bool
	errorMsg      string
	sessionActive bool

	// status bar
	status string
	model  string

	// === turn management ===
	// current turn content (temporarily rendered in View)
	currentTurn *TurnContent

	// whether there is an active turn
	hasTurn bool

	// event channel
	eventChan chan tea.Msg

	// callback serialization channel - ensures Agent callbacks are processed in order
	callbackChan chan func()

	// streaming output tracker (debug)
	streamingTracker *StreamingTracker
}

// === turn management data structures ===

// TurnStatus turn status
type TurnStatus int

const (
	TurnStatusStreaming TurnStatus = iota // streaming text output
	TurnStatusTooling                     // executing tool
	TurnStatusCompleted                   // turn completed
	TurnStatusError                       // turn errored
)

// ToolCallStatus tool call status
type ToolCallStatus int

const (
	ToolCallStatusPending   ToolCallStatus = iota // waiting to execute
	ToolCallStatusRunning                         // executing
	ToolCallStatusCompleted                       // succeeded
	ToolCallStatusError                           // failed
)

// TurnToolCall represents a tool call within a turn
type TurnToolCall struct {
	ID        string
	Name      string
	Arguments string
	Output    string
	IsError   bool
	Status    ToolCallStatus
	StartTime time.Time
	EndTime   time.Time
}

// TurnContent represents a complete Agent turn
type TurnContent struct {
	// streaming text content (Assistant message)
	StreamingText strings.Builder

	// tool call list (in order)
	ToolCalls []TurnToolCall

	// turn status
	Status TurnStatus

	// start time
	StartTime time.Time

	// persisted text length
	PersistedTextLength int
}

// New creates a new TUI application model
func New(cfg *config.Config, ag *agent.Agent, ps *pubsub.PubSub, debug bool) *appModel {
	// create streaming tracker (if debug enabled)
	tracker, err := NewStreamingTracker(debug)
	if err != nil {
		log.Warnf("Failed to create streaming tracker: %v", err)
		tracker, _ = NewStreamingTracker(false) // create disabled tracker
	}

	m := &appModel{
		config:           cfg,
		agent:            ag,
		pubsub:           ps,
		sessionActive:    true,
		status:           StatusReady,
		model:            cfg.Model,
		eventChan:        make(chan tea.Msg, DefaultEventChanSize),
		callbackChan:     make(chan func(), DefaultCallbackChanSize), // larger buffer to avoid blocking
		hasTurn:          false,
		currentTurn:      nil,
		streamingTracker: tracker,
	}

	// init sub-components
	m.editor = editor.NewEditorComponent(ps)
	m.msgView = messages.NewMessageView()

	// init Markdown renderer
	renderer, err := markdown.NewRenderer(76, cfg.Markdown.Style)
	if err != nil {
		log.Warnf("Failed to create markdown renderer: %v", err)
		// use default renderer
		renderer, _ = markdown.NewRenderer(76, "dark")
	}
	m.markdownRenderer = renderer

	// start callback serialization goroutine - ensures Agent callbacks processed in order
	go m.serializeCallbacks()

	// set Agent callbacks
	m.setupAgentCallbacks()

	return m
}

// serializeCallbacks callback serialization goroutine
// all Agent callbacks are sent here first, then dispatched to eventChan in order
func (m *appModel) serializeCallbacks() {
	for fn := range m.callbackChan {
		fn()
	}
}

// Init initializes the Bubble Tea program
func (m appModel) Init() tea.Cmd {
	log.Debug("Initializing TUI components")
	m.setDefaultSize()

	return tea.Batch(
		m.editor.Init(),
		m.waitForEvents(),
		func() tea.Msg {
			return InitMsg{}
		},
	)
}

// setDefaultSize sets default size
func (m *appModel) setDefaultSize() {
	m.width = 80
	m.height = 24
}

// Update handles message updates
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.handleGlobalKeys(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case InitMsg:
		// get current directory
		cwd, _ := os.Getwd()
		welcome := &messages.WelcomeMessage{
			Version: version.GetVersion(),
			Cwd:     cwd,
			MsgTime: time.Now(),
		}
		cmds = append(cmds, tea.Printf("%s", welcome.Render()))
		cmds = append(cmds, m.waitForEvents())

	case editor.SendMsg:
		m.processing = true
		m.status = StatusThinking

		// 1. persist user message to terminal
		userMsg := &messages.UserMessage{
			Content: msg.Content,
			MsgTime: time.Now(),
		}
		cmds = append(cmds, tea.Printf("%s", userMsg.Render()))

		// 2. init new turn
		m.hasTurn = true
		m.currentTurn = &TurnContent{
			Status:    TurnStatusStreaming,
			StartTime: time.Now(),
		}

		// 3. start Agent
		cmds = append(cmds, m.handleUserInput(msg.Content))

	case AgentGenerationStartMsg:
		// new LLM generation started: reset persisted counter
		if m.hasTurn && m.currentTurn != nil {
			log.Debug("[GenerationStart] New generation started, resetting PersistedTextLength")
			m.currentTurn.PersistedTextLength = 0
		}
		cmds = append(cmds, m.waitForEvents())

	case AgentStreamMsg:
		// update current turn streaming text (full content, not delta)
		if m.hasTurn && m.currentTurn != nil {
			// use rune slice to correctly handle UTF-8 multibyte chars
			contentRunes := []rune(msg.Content)
			contentRuneCount := len(contentRunes)

			// debug log: output full content and key lengths
			log.Debugf("[StreamMsg] BEGIN - FullContent runes=%d, PersistedRuneLen=%d",
				contentRuneCount, m.currentTurn.PersistedTextLength)

			// show first 50 chars of raw content (debug)
			if len(contentRunes) > 0 {
				prefixLen := 50
				if len(contentRunes) < prefixLen {
					prefixLen = len(contentRunes)
				}
				log.Debugf("[StreamMsg] FullContent prefix=%q", string(contentRunes[:prefixLen]))
			}

			// skip persisted part, show only new text
			// use rune index instead of byte index for UTF-8
			if contentRuneCount > m.currentTurn.PersistedTextLength {
				// use rune slice to get new content
				newRunes := contentRunes[m.currentTurn.PersistedTextLength:]
				newContent := string(newRunes)
				log.Debugf("[StreamMsg] newContent runes=%d, bytes=%d",
					len(newRunes), len(newContent))

				// show first 50 chars of new content (debug)
				if len(newRunes) > 0 {
					prefixLen := 50
					if len(newRunes) < prefixLen {
						prefixLen = len(newRunes)
					}
					log.Debugf("[StreamMsg] newContent prefix=%q", string(newRunes[:prefixLen]))
				}

				// log to tracker
				m.streamingTracker.LogStreamMsg(
					msg.Content,
					m.currentTurn.PersistedTextLength,
					newContent,
					len(newRunes),
				)

				m.currentTurn.StreamingText.Reset()
				m.currentTurn.StreamingText.WriteString(newContent)
			} else {
				// if no new content, clear StreamingText
				log.Debugf("[StreamMsg] No new content, clearing StreamingText")
				m.currentTurn.StreamingText.Reset()
			}
			m.currentTurn.Status = TurnStatusStreaming
			log.Debugf("[StreamMsg] END - StreamingText.Len=%d", m.currentTurn.StreamingText.Len())
		}
		cmds = append(cmds, m.waitForEvents())

	case AgentToolStartMsg:
		m.status = StatusUsingTool

		if m.hasTurn && m.currentTurn != nil {
			// 1. persist previous streaming text first (if any)
			if m.currentTurn.StreamingText.Len() > 0 {
				content := m.currentTurn.StreamingText.String()
				contentRuneCount := len([]rune(content))
				contentByteLen := len(content)

				log.Debugf("[ToolStart] BEGIN - Persisting streaming text")
				log.Debugf("[ToolStart] content runes=%d, bytes=%d", contentRuneCount, contentByteLen)

				// show first 50 chars of persisted content (debug)
				contentRunes := []rune(content)
				if len(contentRunes) > 0 {
					prefixLen := 50
					if len(contentRunes) < prefixLen {
						prefixLen = len(contentRunes)
					}
					log.Debugf("[ToolStart] content prefix=%q", string(contentRunes[:prefixLen]))
				}

				log.Debugf("[ToolStart] PersistedRuneLen before=%d", m.currentTurn.PersistedTextLength)

				// log to tracker (before persist)
				m.streamingTracker.LogToolStart(
					content,
					m.currentTurn.PersistedTextLength,
					m.currentTurn.PersistedTextLength+contentRuneCount,
				)

				cmds = append(cmds, tea.Printf("%s", m.renderAssistantMessage(content)))

				// record persisted text length (using rune count)
				m.currentTurn.PersistedTextLength += contentRuneCount
				log.Debugf("[ToolStart] PersistedRuneLen after=%d (added %d runes)",
					m.currentTurn.PersistedTextLength, contentRuneCount)

				// clear streaming text buffer
				m.currentTurn.StreamingText.Reset()
				log.Debugf("[ToolStart] END - StreamingText cleared")
			} else {
				log.Debugf("[ToolStart] No streaming text to persist (StreamingText.Len=%d)",
					m.currentTurn.StreamingText.Len())
			}

			// 2. add tool call to current turn
			m.currentTurn.Status = TurnStatusTooling
			m.currentTurn.ToolCalls = append(m.currentTurn.ToolCalls, TurnToolCall{
				ID:        msg.ID,
				Name:      msg.Name,
				Arguments: msg.Arguments,
				Status:    ToolCallStatusRunning,
				StartTime: time.Now(),
			})
		}
		cmds = append(cmds, m.waitForEvents())

	case AgentToolResultMsg:
		if m.hasTurn && m.currentTurn != nil {
			// find corresponding tool call
			for i := range m.currentTurn.ToolCalls {
				if m.currentTurn.ToolCalls[i].ID == msg.ToolCallID {
					// update tool call info
					m.currentTurn.ToolCalls[i].Output = msg.Content
					m.currentTurn.ToolCalls[i].IsError = msg.IsError
					if msg.IsError {
						m.currentTurn.ToolCalls[i].Status = ToolCallStatusError
					} else {
						m.currentTurn.ToolCalls[i].Status = ToolCallStatusCompleted
					}
					m.currentTurn.ToolCalls[i].EndTime = time.Now()

					// persist this tool call immediately
					toolMsg := &messages.ToolCallInfo{
						ID:        m.currentTurn.ToolCalls[i].ID,
						Name:      m.currentTurn.ToolCalls[i].Name,
						Arguments: m.currentTurn.ToolCalls[i].Arguments,
						Output:    m.currentTurn.ToolCalls[i].Output,
						IsError:   m.currentTurn.ToolCalls[i].IsError,
						Completed: true,
						MsgTime:   m.currentTurn.ToolCalls[i].EndTime,
					}
					cmds = append(cmds, tea.Printf("%s", toolMsg.Render()))

					// remove from ToolCalls (already persisted)
					m.currentTurn.ToolCalls = append(
						m.currentTurn.ToolCalls[:i],
						m.currentTurn.ToolCalls[i+1:]...,
					)
					break
				}
			}
		}

		m.status = StatusThinking
		cmds = append(cmds, m.waitForEvents())

	case AgentErrorMsg:
		m.errorMsg = msg.Err.Error()
		m.processing = false
		m.status = StatusError

		// on error, persist partial turn content
		if m.hasTurn && m.currentTurn != nil {
			// persist existing streaming text
			if m.currentTurn.StreamingText.Len() > 0 {
				content := m.currentTurn.StreamingText.String()
				cmds = append(cmds, tea.Printf("%s", m.renderAssistantMessage(content)))
			}

			// persist running tool calls (if any)
			// note: completed tool calls should have been persisted at ToolResult
			// here we only need to persist still-running tool calls
			for _, toolCall := range m.currentTurn.ToolCalls {
				if toolCall.Status == ToolCallStatusRunning {
					toolMsg := &messages.ToolCallInfo{
						ID:        toolCall.ID,
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
						Output:    "Interrupted by error",
						IsError:   true,
						Completed: true,
						MsgTime:   time.Now(),
					}
					cmds = append(cmds, tea.Printf("%s", toolMsg.Render()))
				}
			}
		}

		// persist error message
		errMsg := &messages.ErrorMessage{
			ErrStr:  m.errorMsg,
			MsgTime: time.Now(),
		}
		cmds = append(cmds, tea.Printf("%s", errMsg.Render()))

		// cleanup turn
		m.hasTurn = false
		m.currentTurn = nil
		cmds = append(cmds, m.waitForEvents())

	case AgentFinishMsg:
		// turn completed: persist remaining content
		if m.hasTurn && m.currentTurn != nil {
			m.currentTurn.Status = TurnStatusCompleted

			// only need to persist remaining streaming text (if any)
			// tool calls were already persisted at ToolResult
			if m.currentTurn.StreamingText.Len() > 0 {
				content := m.currentTurn.StreamingText.String()
				cmds = append(cmds, tea.Printf("%s", m.renderAssistantMessage(content)))
			}

			// cleanup current turn
			m.hasTurn = false
			m.currentTurn = nil
		}

		m.processing = false
		m.status = StatusReady
		cmds = append(cmds, m.waitForEvents())

	case QuitMsg:
		m.quitting = true
		// cleanup tracker
		if trackerFile, err := m.streamingTracker.Close(); err == nil && trackerFile != "" {
			log.Infof("Streaming trace saved to: %s", trackerFile)
		}
		return m, tea.Quit
	}

	// update sub-components
	editorModel, editorCmd := m.editor.Update(msg)
	m.editor = editorModel.(*editor.EditorComponent)
	if editorCmd != nil {
		cmds = append(cmds, editorCmd)
	}

	return m, tea.Batch(cmds...)
}

// handleUserInput handles user input
func (m *appModel) handleUserInput(input string) tea.Cmd {
	return func() tea.Msg {
		if strings.HasPrefix(input, "/") {
			return m.handleCommand(input)
		}

		go func() {
			ctx := context.Background()
			m.agent.ProcessUserInput(ctx, input)
		}()
		return StartedMsg{}
	}
}

// handleCommand handles slash commands
func (m *appModel) handleCommand(input string) tea.Msg {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]
	switch command {
	case "/quit", "/exit":
		return QuitMsg{}
	default:
		go func() {
			ctx := context.Background()
			m.agent.ProcessUserInput(ctx, input)
		}()
		return StartedMsg{}
	}
}

// waitForEvents listens to Agent event channel
func (m *appModel) waitForEvents() tea.Cmd {
	return func() tea.Msg {
		return <-m.eventChan
	}
}

// setupAgentCallbacks sets Agent callbacks
func (m *appModel) setupAgentCallbacks() {
	m.agent.SetCallbacks(
		// onMessage: handle streaming text (full content)
		func(msg *types.Message) {
			var fullContent strings.Builder
			for _, part := range msg.Content {
				if part.Type == "text" && part.Text != "" {
					fullContent.WriteString(part.Text)
				}
			}
			fullStr := fullContent.String()

			// send via serialization channel
			m.callbackChan <- func() {
				m.eventChan <- AgentStreamMsg{Content: fullStr}
			}
		},
		// onToolCall: tool call started
		func(call *types.ToolCall) {
			c := *call
			m.callbackChan <- func() {
				m.eventChan <- AgentToolStartMsg{
					ID:        c.ID,
					Name:      c.Name,
					Arguments: c.Arguments,
				}
			}
		},
		// onToolResult: tool call result
		func(result *types.ToolResult) {
			r := *result
			m.callbackChan <- func() {
				m.eventChan <- AgentToolResultMsg{
					ToolCallID: r.ToolCallID,
					Content:    r.Content,
					IsError:    r.IsError,
				}
			}
		},
		// onError: error handling
		func(err error) {
			e := err
			m.callbackChan <- func() {
				m.eventChan <- AgentErrorMsg{Err: e}
			}
		},
		// onFinish: turn completed
		func(reason types.FinishReason) {
			r := reason
			m.callbackChan <- func() {
				m.eventChan <- AgentFinishMsg{Reason: r}
			}
		},
	)

	// onGenerationStart: new LLM generation started
	m.agent.SetGenerationStartCallback(func() {
		m.callbackChan <- func() {
			m.eventChan <- AgentGenerationStartMsg{}
		}
	})
}

// handleGlobalKeys handles global shortcuts
func (m *appModel) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyF1:
		m.showHelp = !m.showHelp
		return nil
	case tea.KeyCtrlC:
		if m.processing {
			return nil
		}
		return tea.Quit
	}
	return nil
}

// View renders the view
func (m appModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}
	if m.quitting {
		return ""
	}

	editorHeight := m.editor.GetPreferredHeight()
	if editorHeight < 3 {
		editorHeight = 3
	}
	if editorHeight > 7 {
		editorHeight = 7
	}

	if m.width > 0 {
		m.editor.SetSize(m.width, editorHeight)
	}

	var sections []string

	// 1. current turn preview area (if active turn)
	if m.hasTurn && m.currentTurn != nil {
		sections = append(sections, m.renderCurrentTurn())
	}

	// 2. blank line separator
	sections = append(sections, "")

	// 3. editor
	sections = append(sections, m.editor.View())

	// 4. status bar
	sections = append(sections, m.renderStatusBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderCurrentTurn renders full current turn content (temporary preview)
func (m appModel) renderCurrentTurn() string {
	if m.currentTurn == nil {
		return ""
	}

	var sections []string

	// 1. render streaming text (if any)
	if m.currentTurn.StreamingText.Len() > 0 {
		content := m.currentTurn.StreamingText.String()
		contentRunes := len([]rune(content))
		log.Debugf("[RenderCurrentTurn] StreamingText runes=%d, bytes=%d",
			contentRunes, len(content))

		// show first 50 chars
		if contentRunes > 0 {
			runes := []rune(content)
			prefixLen := 50
			if contentRunes < prefixLen {
				prefixLen = contentRunes
			}
			log.Debugf("[RenderCurrentTurn] StreamingText prefix=%q", string(runes[:prefixLen]))
		}

		// limit preview area height to avoid taking too much screen space
		// at most 10 lines, excess shown as "..."
		lines := strings.Split(content, "\n")
		maxLines := 10
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
			content = "...\n" + strings.Join(lines, "\n")
		}

		// use gray italic style for content being generated
		previewStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGrayLight)).
			Italic(true)

		// render content directly, add ⏺ prefix
		sections = append(sections, previewStyle.Render(fmt.Sprintf("⏺ %s", strings.TrimSpace(content))))
	}

	// 2. render tool calls (if any)
	for _, toolCall := range m.currentTurn.ToolCalls {
		toolMsg := &messages.ToolCallInfo{
			ID:        toolCall.ID,
			Name:      toolCall.Name,
			Arguments: toolCall.Arguments,
			Output:    toolCall.Output,
			IsError:   toolCall.IsError,
			Completed: toolCall.Status == ToolCallStatusCompleted || toolCall.Status == ToolCallStatusError,
			MsgTime:   time.Now(),
		}

		// use different colors for different statuses
		var color string
		switch toolCall.Status {
		case ToolCallStatusPending:
			color = ColorGray // gray
		case ToolCallStatusRunning:
			color = ColorWhite // white
		case ToolCallStatusCompleted:
			color = ColorGreen // green
		case ToolCallStatusError:
			color = ColorRed // red
		default:
			color = ColorWhite // default white
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		sections = append(sections, style.Render(toolMsg.Render()))
	}

	return strings.Join(sections, "\n")
}

func (m appModel) renderStatusBar() string {
	var statusColor string
	switch m.status {
	case StatusReady:
		statusColor = ColorGreen
	case StatusThinking, StatusUsingTool:
		statusColor = ColorPurple
	case StatusError:
		statusColor = ColorRed
	default:
		statusColor = ColorGray
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Width(StatusBarLeftWidth)
	statusStr := statusStyle.Render("● " + m.status)

	modelWidth := m.width - StatusBarRightMargin
	if modelWidth < 0 {
		modelWidth = 0
	}
	modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray)).Width(modelWidth)
	modelStr := modelStyle.Render(m.model)

	barStyle := lipgloss.NewStyle().Background(lipgloss.Color(ColorDarkBg)).Padding(0, 1).Width(m.width)
	content := lipgloss.JoinHorizontal(lipgloss.Left, statusStr, modelStr)
	return barStyle.Render(content)
}

// renderAssistantMessage renders assistant message with Markdown support
func (m appModel) renderAssistantMessage(content string) string {
	// debug log: record content before render
	log.Debugf("[Render] Input content runes=%d, bytes=%d", len([]rune(content)), len(content))

	// if no renderer or Markdown disabled, return raw content
	if m.markdownRenderer == nil || !m.config.Markdown.Enabled {
		// log to tracker (unrendered)
		m.streamingTracker.LogRender(content, content, "none")
		return content
	}

	// create AssistantMessage and render with renderer
	assistantMsg := &messages.AssistantMessage{
		Content: content,
		MsgTime: time.Now(),
	}

	// use markdown renderer for Markdown rendering
	// RenderMarkdown returns only a string value, no error
	rendered := assistantMsg.RenderMarkdown(m.markdownRenderer)

	// debug log: record content after render
	log.Debugf("[Render] Output content runes=%d, bytes=%d", len([]rune(rendered)), len(rendered))
	if len([]rune(rendered)) > 0 {
		// check first 10 chars
		renderedRunes := []rune(rendered)
		prefix := ""
		if len(renderedRunes) > 10 {
			prefix = string(renderedRunes[:10])
		} else {
			prefix = string(renderedRunes)
		}
		log.Debugf("[Render] Output prefix=%q", prefix)
	}

	// log to tracker
	style := m.markdownRenderer.GetStyle()
	m.streamingTracker.LogRender(content, rendered, style)

	return rendered
}

func (m appModel) renderHelp() string {
	help := "\nQoder CLI Help\n\nCtrl+C Quit\nEnter Send\nF1 Toggle help\n"
	helpWidth := m.width - HelpBorderWidth
	if helpWidth < MinHelpWidth {
		helpWidth = MinHelpWidth
	}
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Width(helpWidth)
	return style.Render(help)
}

// === message type definitions ===

type InitMsg struct{}
type StartedMsg struct{}
type AgentStreamMsg struct{ Content string }
type AgentGenerationStartMsg struct{} // new: new LLM generation started
type AgentToolStartMsg struct{ ID, Name, Arguments string }
type AgentToolResultMsg struct {
	ToolCallID, Content string
	IsError             bool
}
type AgentErrorMsg struct{ Err error }
type AgentFinishMsg struct{ Reason types.FinishReason }
type QuitMsg struct{}
