// Package agent Agent 核心实现
// 反编译自 qodercli v0.1.29
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/state"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/tools"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// Agent AI Agent
type Agent struct {
	// 配置
	config *Config
	
	// Provider
	provider provider.Client
	
	// 工具
	toolRegistry *tools.Registry
	toolExecutor *tools.Executor
	
	// 状态
	state *state.State
	
	// 权限
	permissionCoordinator *permission.Coordinator
	
	// Hook
	hooks *HookManager
	
	// 控制
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	
	// 回调
	onMessage    func(msg *types.Message)
	onToolCall   func(call *types.ToolCall)
	onToolResult func(result *types.ToolResult)
	onError      func(err error)
	onFinish     func(reason types.FinishReason)
}

// Config Agent 配置
type Config struct {
	SystemPrompt     string
	Model            string
	MaxTokens        int
	Temperature      float64
	MaxTurns         int
	AllowedTools     []string
	DisallowedTools  []string
	PermissionMode   permission.Mode
	ThinkLevel       provider.ThinkLevel
	SubagentMode     bool
	AgentName        string
}

// NewAgent 创建 Agent
func NewAgent(config *Config, provider provider.Client) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	agent := &Agent{
		config:                config,
		provider:              provider,
		toolRegistry:          tools.NewRegistry(),
		state:                 state.New(),
		permissionCoordinator: permission.NewCoordinator(config.PermissionMode),
		hooks:                 NewHookManager(),
		ctx:                   ctx,
		cancel:                cancel,
	}
	
	// 初始化工具执行器
	agent.toolExecutor = tools.NewExecutor(agent.toolRegistry)
	
	// 设置权限
	agent.permissionCoordinator.SetAllowedTools(config.AllowedTools)
	agent.permissionCoordinator.SetDisallowedTools(config.DisallowedTools)
	
	// 注册内置工具
	if err := agent.registerBuiltinTools(); err != nil {
		return nil, fmt.Errorf("register builtin tools: %w", err)
	}
	
	return agent, nil
}

// registerBuiltinTools 注册内置工具
func (a *Agent) registerBuiltinTools() error {
	// 文件工具
	if err := a.toolRegistry.Register(tools.NewReadTool(32 * 1024 * 1024)); err != nil {
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
	
	// Bash 工具
	shellManager := tools.NewDefaultShellManager()
	if err := a.toolRegistry.Register(tools.NewBashTool(shellManager, 5*time.Minute)); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewBashOutputTool(shellManager)); err != nil {
		return err
	}
	if err := a.toolRegistry.Register(tools.NewKillBashTool(shellManager)); err != nil {
		return err
	}
	
	return nil
}

// SetCallbacks 设置回调
func (a *Agent) SetCallbacks(
	onMessage func(msg *types.Message),
	onToolCall func(call *types.ToolCall),
	onToolResult func(result *types.ToolResult),
	onError func(err error),
	onFinish func(reason types.FinishReason),
) {
	a.onMessage = onMessage
	a.onToolCall = onToolCall
	a.onToolResult = onToolResult
	a.onError = onError
	a.onFinish = onFinish
}

// SetAskCallback 设置权限询问回调
func (a *Agent) SetAskCallback(callback func(ctx context.Context, req *permission.Request) (permission.Decision, error)) {
	a.permissionCoordinator.SetAskCallback(callback)
}

// RegisterTool 注册工具
func (a *Agent) RegisterTool(tool tools.Tool) error {
	return a.toolRegistry.Register(tool)
}

// AddHook 添加 Hook
func (a *Agent) AddHook(hookType HookType, hook Hook) {
	a.hooks.Add(hookType, hook)
}

// SendMessage 发送消息
func (a *Agent) SendMessage(ctx context.Context, content string, attachments ...*types.Attachment) error {
	// 构建用户消息
	msg := &types.Message{
		Role:    types.RoleUser,
		Content: []types.ContentPart{{Type: "text", Text: content}},
	}
	
	// 添加附件
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
	
	// 添加到状态
	a.state.AddMessage(msg)
	
	// 触发 Hook
	if err := a.hooks.Execute(HookTypeUserPromptSubmit, &HookContext{
		Agent:   a,
		Message: msg,
	}); err != nil {
		return fmt.Errorf("hook error: %w", err)
	}
	
	// 开始生成
	return a.generate(ctx)
}

// generate 生成响应
func (a *Agent) generate(ctx context.Context) error {
	turnCount := 0
	
	for {
		// 检查最大回合数
		if a.config.MaxTurns > 0 && turnCount >= a.config.MaxTurns {
			log.Debug("Max turns reached: %d/%d", turnCount, a.config.MaxTurns)
			if a.onFinish != nil {
				a.onFinish(types.FinishReasonLength)
			}
			return fmt.Errorf("max turns reached: %d", a.config.MaxTurns)
		}
		turnCount++
		
		log.Debug("Starting turn %d, current messages count: %d", turnCount, len(a.state.GetMessages()))
		
		// 构建请求
		req := &provider.ModelRequest{
			Model:       a.config.Model,
			Messages:    a.state.GetMessages(),
			Tools:       a.toolRegistry.ToToolInfo(),
			MaxTokens:   a.config.MaxTokens,
			Temperature: a.config.Temperature,
			Stream:      true,
			SystemPrompt: a.config.SystemPrompt,
			ThinkLevel:  a.config.ThinkLevel,
		}
		
		// 流式请求
		log.Debug("Calling provider.Stream() for turn %d", turnCount)
		eventChan := a.provider.Stream(ctx, req)
		
		// 处理事件
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
				// 消息开始
				log.Debug("Event: MessageStart")
				
			case provider.EventTypeContentBlockDelta:
				// 内容增量 - 只打印新增的文本
				assistantMsg.Content = append(assistantMsg.Content, types.ContentPart{
					Type: "text",
					Text: event.Content,
				})
				if a.onMessage != nil {
					a.onMessage(assistantMsg)
				}
				log.Debug("Event: ContentBlockDelta, content length: %d", len(event.Content))
				
			case provider.EventTypeThinkingDelta:
				// 思考增量
				assistantMsg.Content = append(assistantMsg.Content, types.ContentPart{
					Type:     "thinking",
					Thinking: event.Thinking,
				})
				log.Debug("Event: ThinkingDelta")
				
			case provider.EventTypeToolUseStart:
				// 工具使用开始
				currentToolCall = &types.ToolCall{
					ID:   generateToolCallID(),
					Name: event.ToolUse.Name,
				}
				log.Debug("Event: ToolUseStart, tool name: %s", event.ToolUse.Name)
				
			case provider.EventTypeToolUseDelta:
				// 工具使用增量
				if currentToolCall != nil {
					currentToolCall.Arguments += event.ToolCall.Arguments
				}
				log.Debug("Event: ToolUseDelta, arguments length: %d", len(currentToolCall.Arguments))
				
			case provider.EventTypeToolUseStop:
				// 工具使用结束
				if currentToolCall != nil {
					toolCalls = append(toolCalls, *currentToolCall)
					log.Debug("Event: ToolUseStop, total tool calls so far: %d", len(toolCalls))
					currentToolCall = nil
				}
				
			case provider.EventTypeMessageStop:
				// 消息结束
				log.Debug("Event: MessageStop")
				
			case provider.EventTypeMessageDelta:
				// 消息增量（包含 finish_reason）
				finishReason = event.FinishReason
				log.Debug("Event: MessageDelta, finish reason: %s", finishReason)
				
			case provider.EventTypeError:
				// 错误
				log.Debug("Event: Error, message: %s", event.Error.Message)
				if a.onError != nil {
					a.onError(fmt.Errorf("provider error: %s", event.Error.Message))
				}
				return fmt.Errorf("provider error: %s", event.Error.Message)
			}
		}
		
		log.Debug("Event channel closed for turn %d, tool calls: %d, finish reason: %s", turnCount, len(toolCalls), finishReason)
		
		// 添加工具调用
		if len(toolCalls) > 0 {
			assistantMsg.ToolCalls = toolCalls
			log.Debug("Added %d tool calls to assistant message", len(toolCalls))
		}
		
		// 保存助手消息
		a.state.AddMessage(assistantMsg)
		log.Debug("Added assistant message to state, total messages now: %d", len(a.state.GetMessages()))
		
		// 如果没有工具调用，结束
		if len(toolCalls) == 0 {
			log.Debug("No tool calls, finishing with reason: %s", finishReason)
			if a.onFinish != nil {
				a.onFinish(finishReason)
			}
			return nil
		}
		
		// 执行工具调用
		log.Debug("Executing %d tool calls for turn %d", len(toolCalls), turnCount)
		for i, tc := range toolCalls {
			log.Debug("Executing tool call %d/%d: %s", i+1, len(toolCalls), tc.Name)
			if err := a.executeToolCall(ctx, &tc); err != nil {
				log.Debug("Tool call %d failed: %v", i+1, err)
				return err
			}
			log.Debug("Tool call %d completed successfully", i+1)
		}
		log.Debug("All tool calls completed for turn %d, will continue to next turn", turnCount)
	}
}

// executeToolCall 执行工具调用
func (a *Agent) executeToolCall(ctx context.Context, tc *types.ToolCall) error {
	// 触发回调
	if a.onToolCall != nil {
		a.onToolCall(tc)
	}
	
	// 构建权限请求
	permReq := &permission.Request{
		ToolName:  tc.Name,
		ToolInput: tc.Arguments,
	}
	
	// 解析参数以获取更多信息
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
	
	// 检查权限
	permResult, err := a.permissionCoordinator.Check(ctx, permReq)
	if err != nil {
		return fmt.Errorf("permission check error: %w", err)
	}
	
	if permResult.Decision == permission.DecisionDeny {
		result := &types.ToolResult{
			ToolCallID: tc.ID,
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
	
	// 执行工具
	toolCall := &tools.ToolCall{
		ID:        tc.ID,
		Name:      tc.Name,
		Arguments: []byte(tc.Arguments),
	}
	
	result, err := a.toolExecutor.Execute(ctx, toolCall)
	if err != nil {
		result = &tools.ToolResult{
			ToolCallID: tc.ID,
			Content:    err.Error(),
			IsError:    true,
		}
	}
	
	// PostToolUse Hook
	hookCtx.ToolResult = result
	if err := a.hooks.Execute(HookTypePostToolUse, hookCtx); err != nil {
		return fmt.Errorf("post-tool-use hook error: %w", err)
	}
	
	// 转换结果
	toolResult := &types.ToolResult{
		ToolCallID: result.ToolCallID,
		Content:    result.Content,
		IsError:    result.IsError,
	}
	
	// 添加到状态
	a.state.AddToolResult(toolResult)
	
	// 触发回调
	if a.onToolResult != nil {
		a.onToolResult(toolResult)
	}
	
	return nil
}

// Stop 停止 Agent
func (a *Agent) Stop() {
	a.cancel()
	
	// 触发停止 Hook
	a.hooks.Execute(HookTypeAgentStop, &HookContext{
		Agent: a,
	})
}

// GetState 获取状态
func (a *Agent) GetState() *state.State {
	return a.state
}

// GetMessages 获取消息
func (a *Agent) GetMessages() []types.Message {
	return a.state.GetMessages()
}

// ClearMessages 清空消息
func (a *Agent) ClearMessages() {
	a.state.ClearMessages()
}

// ProcessUserInput 处理用户输入（TUI 入口）
func (a *Agent) ProcessUserInput(ctx context.Context, input string) error {
	// 检查是否是特殊命令
	if strings.HasPrefix(input, "/") {
		return a.handleCommand(ctx, input)
	}

	// 发送普通消息
	return a.SendMessage(ctx, input)
}

// handleCommand 处理斜杠命令
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
		// 压缩上下文
		return a.compactContext(ctx)

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// compactContext 压缩上下文
func (a *Agent) compactContext(ctx context.Context) error {
	// 获取当前消息
	messages := a.state.GetMessages()
	if len(messages) <= 2 {
		return nil // 消息太少，无需压缩
	}

	// 保留系统消息和最近的用户-助手对话
	var compressed []types.Message
	var summary strings.Builder

	// 保留第一条系统消息（如果有）
	if len(messages) > 0 && messages[0].Role == types.RoleSystem {
		compressed = append(compressed, messages[0])
	}

	// 生成摘要（简化实现）
	summary.WriteString("Previous conversation summary:\n")
	for _, msg := range messages[1 : len(messages)-2] {
		if msg.Role == types.RoleUser {
			summary.WriteString("- User asked about: ")
			for _, part := range msg.Content {
				if part.Type == "text" {
					summary.WriteString(part.Text[:min(len(part.Text), 50)])
					if len(part.Text) > 50 {
						summary.WriteString("...")
					}
					summary.WriteString("\n")
					break
				}
			}
		}
	}

	// 添加摘要消息
	compressed = append(compressed, types.Message{
		Role:    types.RoleSystem,
		Content: []types.ContentPart{{Type: "text", Text: summary.String()}},
	})

	// 保留最后两条消息（用户输入和助手回复）
	compressed = append(compressed, messages[len(messages)-2:]...)

	// 更新状态
	a.state.SetMessages(compressed)

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateToolCallID 生成工具调用 ID
func generateToolCallID() string {
	return fmt.Sprintf("call_%d", time.Now().UnixNano())
}

// HookType Hook 类型
type HookType string

const (
	HookTypePreToolUse       HookType = "pre_tool_use"
	HookTypePostToolUse      HookType = "post_tool_use"
	HookTypeAgentStop        HookType = "agent_stop"
	HookTypeSessionStart     HookType = "session_start"
	HookTypeUserPromptSubmit HookType = "user_prompt_submit"
)

// Hook Hook 接口
type Hook interface {
	Execute(ctx *HookContext) error
}

// HookFunc Hook 函数
type HookFunc func(ctx *HookContext) error

// Execute 执行 Hook
func (f HookFunc) Execute(ctx *HookContext) error {
	return f(ctx)
}

// HookContext Hook 上下文
type HookContext struct {
	Agent      *Agent
	Message    *types.Message
	ToolCall   *types.ToolCall
	ToolResult *tools.ToolResult
}

// HookManager Hook 管理器
type HookManager struct {
	hooks map[HookType][]Hook
	mu    sync.RWMutex
}

// NewHookManager 创建 Hook 管理器
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[HookType][]Hook),
	}
}

// Add 添加 Hook
func (m *HookManager) Add(hookType HookType, hook Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks[hookType] = append(m.hooks[hookType], hook)
}

// Execute 执行 Hook
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
