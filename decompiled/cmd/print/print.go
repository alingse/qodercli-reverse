package print

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// Run 运行非交互打印模式
func Run(input string, flags *utils.Flags, quiet bool, systemPrompt string) error {
	log.Info("Starting qodercli in print mode")
	log.Debug("Input: %s", input)

	cfg := utils.LoadConfig(flags)
	prov, modelOverride, err := utils.CreateProvider()
	if err != nil {
		log.Error("Failed to create provider: %v", err)
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// 如果环境变量指定了模型，使用它
	if modelOverride != "" {
		cfg.Model = modelOverride
	}

	// 如果没有提供系统提示词，自动构建（官方行为）
	if systemPrompt == "" {
		var buildErr error
		workDir := flags.Workspace
		if workDir == "" {
			workDir, _ = os.Getwd()
		}
		systemPrompt, buildErr = utils.BuildSystemPromptAuto(workDir, flags.WithClaudeConfig)
		if buildErr != nil {
			log.Error("Failed to build system prompt: %v", buildErr)
			// 使用默认提示词
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
	}

	ag, err := agent.NewAgent(agentCfg, prov)
	if err != nil {
		log.Error("Failed to create agent: %v", err)
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// 简单的输出处理 - 流式增量打印
	var lastPrintedRuneCount int
	var fullText strings.Builder
	toolCallNames := make(map[string]string) // ToolCallID -> ToolName 映射

	ag.SetCallbacks(
		func(msg *types.Message) {
			if !quiet {
				// 构建完整的文本内容
				fullText.Reset()
				for _, part := range msg.Content {
					if part.Type == "text" && part.Text != "" {
						fullText.WriteString(part.Text)
					}
				}

				// 只打印新增的文本部分
				fullStr := fullText.String()
				runes := []rune(fullStr)
				if len(runes) > lastPrintedRuneCount {
					newRunes := runes[lastPrintedRuneCount:]
					fmt.Print(string(newRunes))
					lastPrintedRuneCount = len(runes)
				}
			}
		},
		func(call *types.ToolCall) {
			toolCallNames[call.ID] = call.Name // 保存映射
			log.Debug("Tool call: %s", call.Name)
			if !quiet && log.GetLevel() == log.LevelDebug {
				fmt.Fprintf(os.Stderr, "\n● %s\n", call.Name)
				formattedArgs := formatToolCallArgs(call.Name, call.Arguments)
				if formattedArgs != "" {
					fmt.Fprintf(os.Stderr, "  ⎿ %s\n", formattedArgs)
				}
			}
		},
		func(result *types.ToolResult) {
			toolName := toolCallNames[result.ToolCallID] // 获取工具名
			log.Debug("Tool result: %s", result.ToolCallID)
			if !quiet && log.GetLevel() == log.LevelDebug {
				formattedResult := formatToolResult(toolName, result.Content, result.IsError)
				if formattedResult != "" {
					fmt.Fprintf(os.Stderr, "  ⎿ %s\n", formattedResult)
				}
			}
			delete(toolCallNames, result.ToolCallID) // 清理
		},
		func(err error) {
			log.Error("Callback error: %v", err)
		},
		func(reason types.FinishReason) {
			if !quiet {
				fmt.Println()
			}
			log.Info("Request completed with finish reason: %s", reason)
		},
	)

	ctx := context.Background()
	if err := ag.ProcessUserInput(ctx, input); err != nil {
		log.Error("Error processing input: %v", err)
		return fmt.Errorf("error processing input: %w", err)
	}

	return nil
}
