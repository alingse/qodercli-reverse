package tui

import (
	"fmt"
	"os"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
	"github.com/alingse/qodercli-reverse/decompiled/tui/app"
)

// Run 运行 TUI 模式
func Run(flags *utils.Flags, systemPrompt string) error {
	log.Info("Starting qodercli in TUI mode")

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

	ps := pubsub.New()
	_ = ps // 保留 ps 用于未来扩展

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

	// 启动 TUI
	opts := app.Options{
		Config: cfg,
		Agent:  ag,
		PubSub: ps,
	}

	if err := app.Run(opts); err != nil {
		log.Error("TUI error: %v", err)
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
