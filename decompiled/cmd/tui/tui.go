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

// Run runs TUI mode
func Run(flags *utils.Flags, systemPrompt string, planMode bool) error {
	log.Info("Starting qodercli-reverse in TUI mode")
	if flags.Debug {
		log.Info("Debug mode enabled - streaming trace will be saved to /tmp")
	}

	cfg := utils.LoadConfig(flags)
	prov, modelOverride, err := utils.CreateProvider()
	if err != nil {
		log.Error("Failed to create provider: %v", err)
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// If model specified via env var, use it
	if modelOverride != "" {
		cfg.Model = modelOverride
	}

	// If no system prompt provided, auto-build (official behavior)
	if systemPrompt == "" {
		var buildErr error
		workDir := flags.Workspace
		if workDir == "" {
			workDir, _ = os.Getwd()
		}
		systemPrompt, buildErr = utils.BuildSystemPromptAuto(workDir, flags.WithClaudeConfig)
		if buildErr != nil {
			log.Error("Failed to build system prompt: %v", buildErr)
			// Use default prompt
			systemPrompt = utils.GetDefaultSystemPrompt()
		}
	}

	ps := pubsub.New()
	_ = ps // keep ps for future extension

	agentCfg := &agent.Config{
		Model:          cfg.Model,
		MaxTokens:      flags.MaxTokens,
		Temperature:    flags.Temperature,
		MaxTurns:       flags.MaxTurns,
		PermissionMode: permission.Mode(flags.PermissionMode),
		SystemPrompt:   systemPrompt,
		PlanMode:       planMode,
	}

	ag, err := agent.NewAgent(agentCfg, prov)
	if err != nil {
		log.Error("Failed to create agent: %v", err)
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Start TUI
	opts := app.Options{
		Config: cfg,
		Agent:  ag,
		PubSub: ps,
		Debug:  flags.Debug,
	}

	if err := app.Run(opts); err != nil {
		log.Error("TUI error: %v", err)
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
