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
			toolCallNames[call.ID] = call.Name
			log.Debug("Tool call: %s", call.Name)
			formattedArgs := formatToolCallArgs(call.Name, call.Arguments)
			if formattedArgs != "" {
				log.Debug("  ⎿ %s: %s", call.Name, formattedArgs)
			}
		},
		func(result *types.ToolResult) {
			toolName := toolCallNames[result.ToolCallID]
			log.Debug("Tool result: %s", result.ToolCallID)
			formattedResult := formatToolResult(toolName, result.Content, result.IsError)
			if formattedResult != "" {
				log.Debug("  ⎿ %s result: %s", toolName, formattedResult)
			}
			delete(toolCallNames, result.ToolCallID)
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

	// 设置新一轮 LLM 生成开始回调：重置已打印字符计数，并在不同 generation 之间添加换行
	ag.SetGenerationStartCallback(func() {
		log.Debug("[PrintMode] New generation started, lastPrintedRuneCount=%d", lastPrintedRuneCount)
		// 如果之前有输出，添加换行来分隔不同的 generation
		if lastPrintedRuneCount > 0 && !quiet {
			fmt.Println()
		}
		lastPrintedRuneCount = 0
	})

	ctx := context.Background()
	if err := ag.ProcessUserInput(ctx, input); err != nil {
		log.Error("Error processing input: %v", err)
		return fmt.Errorf("error processing input: %w", err)
	}

	return nil
}
