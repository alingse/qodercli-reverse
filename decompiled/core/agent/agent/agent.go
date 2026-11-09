// Package agent Agent core implementation
// Decompiled from qodercli-reverse v0.1.29
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/tools/security"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/tools"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// tool default config constants
const (
	DefaultReadToolMaxBytes = 32 * 1024 * 1024 // Read tool default max read bytes (32MB)
	DefaultBashTimeout      = 5 * time.Minute  // Bash tool default timeout
)

// Agent AI Agent
type Agent struct {
	// config
	config *Config

	// Provider
	provider provider.Client

	// tools
	toolRegistry *tools.Registry
	toolExecutor *tools.Executor

	// state
	state *state.State

	// Todo state
	todoState     state.TodoState
	sessionMemory *state.SessionMemory

	// permissions
	permissionCoordinator *permission.Coordinator

	// Hook
	hooks *HookManager

	// control
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex

	// callbacks
	onMessage         func(msg *types.Message)
	onToolCall        func(call *types.ToolCall)
	onToolResult      func(result *types.ToolResult)
	onError           func(err error)
	onFinish          func(reason types.FinishReason)
	onGenerationStart func()      // new round LLM generation start callback
	onTurnComplete    func() bool // turn complete callback, returns true if new message inserted needs immediate handling
}

// Config Agent config
type Config struct {
	SystemPrompt          string
	Model                 string
	MaxTokens             int
	Temperature           float64
	MaxTurns              int
	AllowedTools          []string
	DisallowedTools       []string
	PermissionMode        permission.Mode
	ThinkLevel            provider.ThinkLevel
	SubagentMode          bool
	AgentName             string
	EnableEnhancedCompact bool // enable enhanced compact feature
	PlanMode              bool // when true, agent outputs a plan before executing tools
}

// NewAgent creates Agent
func NewAgent(config *Config, provider provider.Client) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// create session memory and Todo state
	sessionMemory := state.NewSessionMemory()
	// Load any previously saved session
	if err := sessionMemory.LoadSession(sessionMemory.GetSession()); err != nil {
		fmt.Fprintf(os.Stderr, "Note: no previous session found, starting fresh\n")
	}

	todoState := state.NewTodoState(sessionMemory)

	agent := &Agent{
		config:                config,
		provider:              provider,
		toolRegistry:          tools.NewRegistry(),
		state:                 state.New(),
		todoState:             todoState,
		sessionMemory:         sessionMemory,
		permissionCoordinator: permission.NewCoordinator(config.PermissionMode),
		hooks:                 NewHookManager(),
		ctx:                   ctx,
		cancel:                cancel,
	}

	// initialize tool executor
	agent.toolExecutor = tools.NewExecutor(agent.toolRegistry)

	// set permissions
	agent.permissionCoordinator.SetAllowedTools(config.AllowedTools)
	agent.permissionCoordinator.SetDisallowedTools(config.DisallowedTools)

	// register built-in tools
	if err := agent.registerBuiltinTools(); err != nil {
		return nil, fmt.Errorf("register builtin tools: %w", err)
	}

	return agent, nil
}

// registerBuiltinTools registers built-in tools
func (a *Agent) registerBuiltinTools() error {
	// file tools
	if err := a.toolRegistry.Register(tools.NewReadTool(DefaultReadToolMaxBytes)); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewWriteTool()); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewEditTool()); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewDeleteFileTool()); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewGlobTool()); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewGrepTool(tools.GetRipgrepPath())); err != nil {
		return err
	}

	// Bash tools
	shellManager := tools.NewDefaultShellManager()
	if err := a.toolRegistry.Register(tools.NewBashTool(shellManager, DefaultBashTimeout)); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewBashOutputTool(shellManager)); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewKillBashTool(shellManager)); err != nil {
		return err
	}

	// TodoWrite tools
	if err := a.toolRegistry.Register(tools.NewTodoWriteTool(a.todoState)); err != nil {
		return err
	}

	// Security scan tool
	ws := "."
	if err := a.toolRegistry.Register(security.NewScanner(ws)); err != nil {
		log.Warn("Failed to register security scanner: %v", err)
	}

	return nil
}

// SetCallbacks sets callbacks
func (a *Agent) SetCallbacks(
	onMessage func(msg *types.Message),
	onToolCall func(call *types.ToolCall),
	onToolResult func(result *types.ToolResult),
	onError func(err error),
	onFinish func(reason types.FinishReason),
	onTurnComplete ...func() bool, // optional param: turn complete callback
) {
	a.onMessage = onMessage
	a.onToolCall = onToolCall
	a.onToolResult = onToolResult
	a.onError = onError
	a.onFinish = onFinish
	if len(onTurnComplete) > 0 {
		a.onTurnComplete = onTurnComplete[0]
	}
}

// SetGenerationStartCallback sets new round LLM generation start callback
func (a *Agent) SetGenerationStartCallback(callback func()) {
	a.onGenerationStart = callback
}

// SetAskCallback sets permission ask callback
func (a *Agent) SetAskCallback(callback func(ctx context.Context, req *permission.Request) (permission.Decision, error)) {
	a.permissionCoordinator.SetAskCallback(callback)
}

// RegisterTool registers tool
func (a *Agent) RegisterTool(tool tools.Tool) error {
	return a.toolRegistry.Register(tool)
}

// AddHook adds hook
func (a *Agent) AddHook(hookType HookType, hook Hook) {
	a.hooks.Add(hookType, hook)
}

// SendMessage sends message
func (a *Agent) SendMessage(ctx context.Context, content string, attachments ...*types.Attachment) error {
	// build user message
	msg := &types.Message{
		Role:    types.RoleUser,
		Content: []types.ContentPart{{Type: "text", Text: content}},
	}

	// add attachments
	for _, att := range attachments {
		if att.MediaType != "" {
			msg.Content = append(msg.Content, types.ContentPart{
				Type: "image",
				ImageSource: &types.ImageSource{
					Type:      "base64",
					MediaType: att.MediaType,
					Data:      string(att.Data),
				},
			})
		}
	}

	// add to state
	a.state.AddMessage(msg)

	// trigger hook
	if err := a.hooks.Execute(HookTypeUserPromptSubmit, &HookContext{
		Agent:   a,
		Message: msg,
	}); err != nil {
		return fmt.Errorf("hook error: %w", err)
	}

	// start generation
	return a.generate(ctx)
}

// generate generates response
func (a *Agent) generate(ctx context.Context) error {
	turnCount := 0
	var planOutputted bool

	for {
		// check max turns
		if a.config.MaxTurns > 0 && turnCount >= a.config.MaxTurns {
			log.Debug("Max turns reached: %d/%d", turnCount, a.config.MaxTurns)
			if a.onFinish != nil {
				a.onFinish(types.FinishReasonLength)
			}
			return fmt.Errorf("max turns reached: %d", a.config.MaxTurns)
		}
		turnCount++

		log.Debug("Starting turn %d, current messages count: %d", turnCount, len(a.state.GetMessages()))

		// build request
		req := &provider.ModelRequest{
			Model:        a.config.Model,
			Messages:     a.state.GetMessages(),
			Tools:        a.toolRegistry.ToToolInfo(),
			MaxTokens:    a.config.MaxTokens,
			Temperature:  a.config.Temperature,
			Stream:       true,
			SystemPrompt: a.config.SystemPrompt,
			ThinkLevel:   a.config.ThinkLevel,
		}

		// streaming request
		log.Debug("Calling provider.Stream() for turn %d", turnCount)
		eventChan := a.provider.Stream(ctx, req)

		// handle events
		assistantMsg := &types.Message{
			Role: types.RoleAssistant,
		}
		var currentToolCall *types.ToolCall
		var toolCalls []types.ToolCall
		var finishReason types.FinishReason

		log.Debug("Processing events for turn %d", turnCount)
		for event := range eventChan {
			switch event.Type {
			case provider.EventTypeMessageStart:
				// message start - mark new round LLM generation start
				log.Debug("Event: MessageStart")
				if a.onGenerationStart != nil {
					a.onGenerationStart()
				}

			case provider.EventTypeContentBlockDelta:
				// content delta - accumulate to last text ContentPart
				if len(assistantMsg.Content) > 0 && assistantMsg.Content[len(assistantMsg.Content)-1].Type == "text" {
					// if last ContentPart is text type, append to it
					assistantMsg.Content[len(assistantMsg.Content)-1].Text += event.Content
				} else {
					// otherwise create new ContentPart
					assistantMsg.Content = append(assistantMsg.Content, types.ContentPart{
						Type: "text",
						Text: event.Content,
					})
				}
				if a.onMessage != nil {
					a.onMessage(assistantMsg)
				}
				log.Debug("Event: ContentBlockDelta, content length: %d", len(event.Content))

			case provider.EventTypeThinkingDelta:
				// thinking delta - accumulate to last thinking ContentPart
				if len(assistantMsg.Content) > 0 && assistantMsg.Content[len(assistantMsg.Content)-1].Type == "thinking" {
					// if last ContentPart is thinking type, append to it
					assistantMsg.Content[len(assistantMsg.Content)-1].Thinking += event.Thinking
				} else {
					// otherwise create new ContentPart
					assistantMsg.Content = append(assistantMsg.Content, types.ContentPart{
						Type:     "thinking",
						Thinking: event.Thinking,
					})
				}
				log.Debug("Event: ThinkingDelta")

			case provider.EventTypeToolUseStart:
				// tool use start
				currentToolCall = &types.ToolCall{
					ID:   generateToolCallID(),
					Name: event.ToolUse.Name,
				}
				log.Debug("Event: ToolUseStart, tool name: %s", event.ToolUse.Name)

			case provider.EventTypeToolUseDelta:
				// tool use delta
				if currentToolCall != nil {
					currentToolCall.Arguments += event.ToolCall.Arguments
				}
				log.Debug("Event: ToolUseDelta, arguments length: %d", len(currentToolCall.Arguments))

			case provider.EventTypeToolUseStop:
				// tool use end
				if currentToolCall != nil {
					toolCalls = append(toolCalls, *currentToolCall)
					log.Debug("Event: ToolUseStop, total tool calls so far: %d", len(toolCalls))
					currentToolCall = nil
				}

			case provider.EventTypeMessageStop:
				// message end
				log.Debug("Event: MessageStop")

			case provider.EventTypeMessageDelta:
				// message delta (includes finish_reason and token usage)
				finishReason = event.FinishReason
				// update token usage stats
				if event.TokenUsage != nil {
					a.state.UpdateTokenUsage(event.TokenUsage.InputTokens, event.TokenUsage.OutputTokens)
					log.Debug("Event: MessageDelta, finish reason: %s, tokens: %d input + %d output",
						finishReason, event.TokenUsage.InputTokens, event.TokenUsage.OutputTokens)
				} else {
					log.Debug("Event: MessageDelta, finish reason: %s", finishReason)
				}

			case provider.EventTypeError:
				// error
				log.Debug("Event: Error, message: %s", event.Error.Message)
				if a.onError != nil {
					a.onError(fmt.Errorf("provider error: %s", event.Error.Message))
				}
				return fmt.Errorf("provider error: %s", event.Error.Message)
			}
		}

		log.Debug("Event channel closed for turn %d, tool calls: %d, finish reason: %s", turnCount, len(toolCalls), finishReason)

		// add tool calls
		if len(toolCalls) > 0 {
			assistantMsg.ToolCalls = toolCalls
			log.Debug("Added %d tool calls to assistant message", len(toolCalls))
		}

		// save assistant message
		a.state.AddMessage(assistantMsg)
		// increment assistant reply count
		a.state.IncrementAssistantReplies()
		log.Debug("Added assistant message to state, total messages now: %d", len(a.state.GetMessages()))

		// Plan mode: if enabled and plan not yet outputted, output the plan text
		// and finish without executing any tools
		if a.config.PlanMode && !planOutputted {
			planOutputted = true
			// extract plan text from assistant message
			planText := ""
			for _, part := range assistantMsg.Content {
				if part.Type == "text" {
					planText += part.Text + "\n"
				}
			}
			if planText != "" {
				log.Debug("Plan mode: outputting plan (length: %d)", len(planText))
				// we can optionally log or store the plan; here we just finish
				// the plan is available in the state for TUI/display
			}
			if a.onFinish != nil {
				a.onFinish(types.FinishReasonStop)
			}
			return nil
		}

		// if no tool calls, finish
		if len(toolCalls) == 0 {
			log.Debug("No tool calls, finishing with reason: %s", finishReason)
			if a.onFinish != nil {
				a.onFinish(finishReason)
			}
			return nil
		}

		// execute tool calls
		log.Debug("Executing %d tool calls for turn %d", len(toolCalls), turnCount)
		for i, tc := range toolCalls {
			log.Debug("Executing tool call %d/%d: %s", i+1, len(toolCalls), tc.Name)
			// increment tool call count
			a.state.IncrementToolCallCount()
			if err := a.executeToolCall(ctx, &tc); err != nil {
				log.Debug("Tool call %d failed: %v", i+1, err)
				return err
			}
			log.Debug("Tool call %d completed successfully", i+1)
		}
		log.Debug("All tool calls completed for turn %d, will continue to next turn", turnCount)

		// turn complete callback: give TUI chance to insert queued user messages
		if a.onTurnComplete != nil {
			if shouldBreak := a.onTurnComplete(); shouldBreak {
				log.Debug("Turn %d interrupted by queued user message, will restart generation", turnCount)
				// reset turn count, new message starts new session
				turnCount = 0
			}
		}
	}
}

// executeToolCall executes tool call
func (a *Agent) executeToolCall(ctx context.Context, tc *types.ToolCall) error {
	// trigger callback
	if a.onToolCall != nil {
		a.onToolCall(tc)
	}

	// build permission request
	permReq := &permission.Request{
		ToolName:  tc.Name,
		ToolInput: tc.Arguments,
	}

	// parse params for more info
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Arguments), &args); err == nil {
		if path, ok := args["file_path"].(string); ok {
			permReq.FilePath = path
		}
		if cmd, ok := args["command"].(string); ok {
			permReq.Command = cmd
		}
		if url, ok := args["url"].(string); ok {
			permReq.URL = url
		}
	}

	// check permissions
	permResult, err := a.permissionCoordinator.Check(ctx, permReq)
	if err != nil {
		return fmt.Errorf("permission check error: %w", err)
	}

	if permResult.Decision == permission.DecisionDeny {
		result := &types.ToolResult{
			ToolCallID: tc.ID,
			Name:       tc.Name,
			Content:    fmt.Sprintf("Permission denied: %s", permResult.Reason),
			IsError:    true,
		}
		a.state.AddToolResult(result)
		if a.onToolResult != nil {
			a.onToolResult(result)
		}
		return nil
	}

	// PreToolUse Hook
	hookCtx := &HookContext{
		Agent:    a,
		ToolCall: tc,
	}
	if err := a.hooks.Execute(HookTypePreToolUse, hookCtx); err != nil {
		return fmt.Errorf("pre-tool-use hook error: %w", err)
	}

	// execute tool
	toolCall := &tools.ToolCall{
		ID:        tc.ID,
		Name:      tc.Name,
		Arguments: []byte(tc.Arguments),
	}

	result, err := a.toolExecutor.Execute(ctx, toolCall)
	if err != nil {
		result = &tools.ToolResult{
			ToolCallID: tc.ID,
			Name:       tc.Name,
			Content:    err.Error(),
			IsError:    true,
		}
	}

	// PostToolUse Hook
	hookCtx.ToolResult = result
	if err := a.hooks.Execute(HookTypePostToolUse, hookCtx); err != nil {
		return fmt.Errorf("post-tool-use hook error: %w", err)
	}

	// convert result
	toolResult := &types.ToolResult{
		ToolCallID: result.ToolCallID,
		Name:       tc.Name,
		Content:    result.Content,
		IsError:    result.IsError,
	}

	// add to state
	a.state.AddToolResult(toolResult)

	// trigger callback
	if a.onToolResult != nil {
		a.onToolResult(toolResult)
	}

	return nil
}

// Stop stops Agent
func (a *Agent) Stop() {
	a.cancel()

	// trigger stop hook
	a.hooks.Execute(HookTypeAgentStop, &HookContext{
		Agent: a,
	})
}

// GetState gets state
func (a *Agent) GetState() *state.State {
	return a.state
}

// GetMessages gets messages
func (a *Agent) GetMessages() []types.Message {
	return a.state.GetMessages()
}

// ClearMessages clears messages
func (a *Agent) ClearMessages() {
	a.state.ClearMessages()
}

// InjectUserMessage inject user message (for inserting queued messages between turns)
// unlike SendMessage, this method doesn't trigger generate, only adds message to state
func (a *Agent) InjectUserMessage(content string) {
	msg := &types.Message{
		Role:    types.RoleUser,
		Content: []types.ContentPart{{Type: "text", Text: content}},
	}
	a.state.AddMessage(msg)
	log.Debug("Injected user message: %s", content)
}

// ProcessUserInput handles user input (TUI entry)
func (a *Agent) ProcessUserInput(ctx context.Context, input string) error {
	// check if special command
	if strings.HasPrefix(input, "/") {
		return a.handleCommand(ctx, input)
	}

	// send normal message
	return a.SendMessage(ctx, input)
}

// handleCommand handles slash command
func (a *Agent) handleCommand(ctx context.Context, cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := parts[0]

	switch command {
	case "/clear":
		a.ClearMessages()
		return nil

	case "/compact":
		// compact context
		return a.compactContext(ctx)

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// compactContext compact context
func (a *Agent) compactContext(ctx context.Context) error {
	// get current messages
	messages := a.state.GetMessages()
	if len(messages) <= 2 {
		log.Debug("Too few messages to compact: %d", len(messages))
		return nil // too few messages, no need to compact
	}

	log.Debug("Starting context compaction with %d messages", len(messages))

	// simple implementation
	compressed := simpleCompact(messages)

	// update state
	a.state.SetMessages(compressed)
	log.Debug("Context compaction completed: %d -> %d messages", len(messages), len(compressed))

	return nil
}

// simpleCompact simple context compaction (backward compatible)
func simpleCompact(messages []types.Message) []types.Message {
	if len(messages) <= 2 {
		return messages
	}

	// keep system messages and recent user-assistant dialog
	var compressed []types.Message
	var summary strings.Builder

	// keep first system message (if any)
	if len(messages) > 0 && messages[0].Role == types.RoleSystem {
		compressed = append(compressed, messages[0])
	}

	// generate summary (simplified implementation)
	summary.WriteString("Previous conversation summary:\n")
	for _, msg := range messages[1 : len(messages)-2] {
		if msg.Role == types.RoleUser {
			summary.WriteString("- User asked about: ")
			for _, part := range msg.Content {
				if part.Type == "text" {
					text := part.Text
					if len(text) > 50 {
						text = text[:50] + "..."
					}
					summary.WriteString(text + "\n")
					break
				}
			}
		}
	}

	// add summary message
	compressed = append(compressed, types.Message{
		Role:    types.RoleSystem,
		Content: []types.ContentPart{{Type: "text", Text: summary.String()}},
	})

	// keep last two messages (user input and assistant reply)
	compressed = append(compressed, messages[len(messages)-2:]...)

	return compressed
}

// min returns smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateToolCallID generates tool call ID
func generateToolCallID() string {
	return fmt.Sprintf("call_%d", time.Now().UnixNano())
}

// HookType hook type
type HookType string

const (
	HookTypePreToolUse       HookType = "pre_tool_use"
	HookTypePostToolUse      HookType = "post_tool_use"
	HookTypeAgentStop        HookType = "agent_stop"
	HookTypeSessionStart     HookType = "session_start"
	HookTypeUserPromptSubmit HookType = "user_prompt_submit"
)

// Hook hook interface
type Hook interface {
	Execute(ctx *HookContext) error
}

// HookFunc hook function
type HookFunc func(ctx *HookContext) error

// Execute executes hook
func (f HookFunc) Execute(ctx *HookContext) error {
	return f(ctx)
}

// HookContext hook context
type HookContext struct {
	Agent      *Agent
	Message    *types.Message
	ToolCall   *types.ToolCall
	ToolResult *tools.ToolResult
}

// HookManager hook manager
type HookManager struct {
	hooks map[HookType][]Hook
	mu    sync.RWMutex
}

// NewHookManager creates hook manager
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[HookType][]Hook),
	}
}

// Add adds hook
func (m *HookManager) Add(hookType HookType, hook Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks[hookType] = append(m.hooks[hookType], hook)
}

// Execute executes hook
func (m *HookManager) Execute(hookType HookType, ctx *HookContext) error {
	m.mu.RLock()
	hooks := m.hooks[hookType]
	m.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.Execute(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GetTodoState gets Todo state
func (a *Agent) GetTodoState() state.TodoState {
	return a.todoState
}

// GetTodoListText gets Todo list text description
func (a *Agent) GetTodoListText() string {
	if a.todoState == nil {
		return "Todo list not initialized."
	}
	return a.todoState.TodosToText()
}
