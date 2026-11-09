package print

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// StreamEvent is a single newline-delimited JSON event emitted in stream-json mode.
type StreamEvent struct {
	Type     string `json:"type"`
	Content  string `json:"content,omitempty"`
	ToolName string `json:"tool_name,omitempty"`
	ToolID   string `json:"tool_id,omitempty"`
	IsError  bool   `json:"is_error,omitempty"`
	Error    string `json:"error,omitempty"`
}

func emitEvent(ev StreamEvent) {
	b, err := json.Marshal(ev)
	if err != nil {
		return
	}
	fmt.Println(string(b))
}

// RunStreamJSON runs non-interactive print mode emitting newline-delimited JSON events.
func RunStreamJSON(input string, flags *utils.Flags, systemPrompt string, planMode bool, goalMode bool, goal string) error {
	log.Info("Starting qodercli-reverse in stream-json print mode")

	cfg := utils.LoadConfig(flags)
	prov, modelOverride, err := utils.CreateProvider()
	if err != nil {
		emitEvent(StreamEvent{Type: "error", Error: err.Error()})
		return fmt.Errorf("failed to create provider: %w", err)
	}
	if modelOverride != "" {
		cfg.Model = modelOverride
	}

	if systemPrompt == "" {
		workDir := flags.Workspace
		if workDir == "" {
			workDir, _ = os.Getwd()
		}
		systemPrompt, _ = utils.BuildSystemPromptAuto(workDir, flags.WithClaudeConfig)
		if systemPrompt == "" {
			systemPrompt = utils.GetDefaultSystemPrompt()
		}
	}

	agentCfg := &agent.Config{
		Model:          cfg.Model,
		MaxTokens:      flags.MaxTokens,
		Temperature:    flags.Temperature,
		MaxTurns:       flags.MaxTurns,
		PermissionMode: permission.Mode(flags.PermissionMode),
		SystemPrompt:   systemPrompt,
		PlanMode:       planMode,
		GoalMode:       goalMode,
		Goal:           goal,
	}

	ag, err := agent.NewAgent(agentCfg, prov)
	if err != nil {
		emitEvent(StreamEvent{Type: "error", Error: err.Error()})
		return fmt.Errorf("failed to create agent: %w", err)
	}

	emitEvent(StreamEvent{Type: "start", Content: input})

	ag.SetCallbacks(
		func(msg *types.Message) {
			for _, part := range msg.Content {
				if part.Type == "text" && part.Text != "" {
					emitEvent(StreamEvent{Type: "text", Content: part.Text})
				}
			}
		},
		func(call *types.ToolCall) {
			emitEvent(StreamEvent{Type: "tool_call", ToolName: call.Name, ToolID: call.ID, Content: string(call.Arguments)})
		},
		func(result *types.ToolResult) {
			emitEvent(StreamEvent{Type: "tool_result", ToolID: result.ToolCallID, Content: result.Content, IsError: result.IsError})
		},
		func(err error) {
			if err != nil {
				emitEvent(StreamEvent{Type: "error", Error: err.Error()})
			}
		},
		func(reason types.FinishReason) {
			emitEvent(StreamEvent{Type: "finish", Content: string(reason)})
		},
	)

	ctx := context.Background()
	if err := ag.ProcessUserInput(ctx, input); err != nil {
		emitEvent(StreamEvent{Type: "error", Error: err.Error()})
		return fmt.Errorf("error processing input: %w", err)
	}
	_ = strings.TrimSpace
	return nil
}
