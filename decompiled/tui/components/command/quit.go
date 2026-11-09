// Package command TUI command handling
package command

import (
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	tea "github.com/charmbracelet/bubbletea"
)

// QuitCommand /quit command
type QuitCommand struct {
	agent *agent.Agent
}

// NewQuitCommand creates quit command
func NewQuitCommand(agent *agent.Agent) *QuitCommand {
	return &QuitCommand{
		agent: agent,
	}
}

// Execute executes the command
func (c *QuitCommand) Execute() tea.Cmd {
	// get stats
	stats := c.agent.GetState().GetStats()

	// build stats message
	statsMsg := fmt.Sprintf("\n=== Session Statistics ===\n"+
		"Total Tokens: %d (Input: %d, Output: %d)\n"+
		"Tool Calls: %d\n"+
		"Assistant Replies: %d\n"+
		"=========================\n",
		stats.TotalTokens, stats.TotalInputTokens, stats.TotalOutputTokens,
		stats.ToolCallCount, stats.AssistantReplies)

	// print stats
	fmt.Print(statsMsg)

	// return quit command
	return tea.Quit
}
