// Package command TUI 命令处理
package command

import (
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	tea "github.com/charmbracelet/bubbletea"
)

// QuitCommand /quit 命令
type QuitCommand struct {
	agent *agent.Agent
}

// NewQuitCommand 创建 quit 命令
func NewQuitCommand(agent *agent.Agent) *QuitCommand {
	return &QuitCommand{
		agent: agent,
	}
}

// Execute 执行命令
func (c *QuitCommand) Execute() tea.Cmd {
	// 获取统计信息
	stats := c.agent.GetState().GetStats()

	// 构建统计信息消息
	statsMsg := fmt.Sprintf("\n=== Session Statistics ===\n"+
		"Total Tokens: %d (Input: %d, Output: %d)\n"+
		"Tool Calls: %d\n"+
		"Assistant Replies: %d\n"+
		"=========================\n",
		stats.TotalTokens, stats.TotalInputTokens, stats.TotalOutputTokens,
		stats.ToolCallCount, stats.AssistantReplies)

	// 打印统计信息
	fmt.Print(statsMsg)

	// 返回退出命令
	return tea.Quit
}
