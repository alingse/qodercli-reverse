// Package app TUI application launcher
package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
)

// Options application startup options
type Options struct {
	Config   *config.Config
	Agent    *agent.Agent
	PubSub   *pubsub.PubSub
	ResumeID string
	Debug    bool // enable debug mode and streaming output tracing
}

// Run starts the TUI application
func Run(opts Options) error {
	// create application model
	model := New(opts.Config, opts.Agent, opts.PubSub, opts.Debug)

	// if there is a resume session ID
	if opts.ResumeID != "" {
		ctx := context.Background()
		opts.PubSub.Publish(ctx, pubsub.Event{
			Type:    pubsub.EventTypeSessionResume,
			Payload: opts.ResumeID,
		})
	}

	// create Bubble Tea program
	// do not use AltScreen so output can scroll with the terminal
	p := tea.NewProgram(model)

	// run program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
