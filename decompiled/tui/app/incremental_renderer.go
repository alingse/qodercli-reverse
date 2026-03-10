// Package app 增量渲染器 - 实现类似 npm/docker 的输出模式
package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// IncrementalRenderer 增量渲染器
// 实现消息历史向上滚动，底部固定输入区域的效果
type IncrementalRenderer struct {
	lastMessageCount int
	terminalWidth    int
	terminalHeight   int
}

// NewIncrementalRenderer 创建增量渲染器
func NewIncrementalRenderer() *IncrementalRenderer {
	return &IncrementalRenderer{}
}

// RenderMessages 渲染消息（只输出新增的消息）
func (r *IncrementalRenderer) RenderMessages(messages []string, currentCount int) {
	// 只输出新增的消息
	if currentCount > r.lastMessageCount {
		for i := r.lastMessageCount; i < currentCount; i++ {
			fmt.Fprintln(os.Stdout, messages[i])
		}
		r.lastMessageCount = currentCount
	}
}

// RenderFixedBottom 渲染底部固定区域（编辑器 + 状态栏）
// 使用 ANSI 转义码将光标移动到底部
func (r *IncrementalRenderer) RenderFixedBottom(editorView, statusView string, height int) {
	// 保存光标位置
	fmt.Print("\033[s")

	// 移动到底部
	lines := strings.Count(editorView, "\n") + 2 // +2 for status bar
	fmt.Printf("\033[%d;0H", r.terminalHeight-lines)

	// 清除到屏幕底部
	fmt.Print("\033[J")

	// 渲染编辑器和状态栏
	fmt.Print(editorView)
	fmt.Print("\n")
	fmt.Print(statusView)

	// 恢复光标位置
	fmt.Print("\033[u")

	// 刷新输出
	os.Stdout.Sync()
}

// ClearFixedBottom 清除底部固定区域
func (r *IncrementalRenderer) ClearFixedBottom(lines int) {
	// 移动到底部
	fmt.Printf("\033[%d;0H", r.terminalHeight-lines)
	// 清除到屏幕底部
	fmt.Print("\033[J")
	os.Stdout.Sync()
}

// UpdateTerminalSize 更新终端尺寸
func (r *IncrementalRenderer) UpdateTerminalSize(width, height int) {
	r.terminalWidth = width
	r.terminalHeight = height
}

// RenderIncrementalMessage 渲染增量消息（流式输出）
func (r *IncrementalRenderer) RenderIncrementalMessage(content string) {
	// 直接输出，不换行
	fmt.Print(content)
	os.Stdout.Sync()
}

// NewLine 输出换行
func (r *IncrementalRenderer) NewLine() {
	fmt.Println()
}

// RenderWithStyle 使用 lipgloss 样式渲染
func (r *IncrementalRenderer) RenderWithStyle(content string, style lipgloss.Style) {
	fmt.Println(style.Render(content))
	os.Stdout.Sync()
}
