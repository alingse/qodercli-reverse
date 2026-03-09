// Package status 状态栏组件
// 显示当前状态、Token 用量、模型信息等
package status

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// 配色方案
var (
	statusColorIdle       = lipgloss.Color("#00C853")  // 空闲 - 绿色
	statusColorConnecting = lipgloss.Color("#FFB74D")  // 连接中 - 橙色
	statusColorThinking   = lipgloss.Color("#7D56F4")  // 思考中 - 紫色
	statusColorTool       = lipgloss.Color("#5B8DEF")  // 工具执行 - 蓝色
	statusColorError      = lipgloss.Color("#EF5350")  // 错误 - 红色
	statusColorBg         = lipgloss.Color("#1F2937")  // 状态栏背景 - 深蓝灰
	statusColorText       = lipgloss.Color("#E5E7EB")  // 文本 - 浅灰
	statusColorMuted      = lipgloss.Color("#9CA3AF")  // 弱化文本 - 灰色
	statusColorVim        = lipgloss.Color("#FFB74D")  // Vim 模式 - 橙色
)

// Status 应用状态
type Status int

const (
	StatusIdle Status = iota
	StatusConnecting
	StatusThinking
	StatusToolExecuting
	StatusError
)

// Component 状态栏组件
type Component struct {
	width       int
	height      int
	status      Status
	tokenUsage  *types.TokenUsage
	model       string
	mode        string
	message     string
	lastUpdated time.Time
}

// New 创建新的状态栏组件
func New() *Component {
	return &Component{
		status:      StatusIdle,
		model:       "auto",
		mode:        "normal",
		lastUpdated: time.Now(),
	}
}

// Init 初始化组件
func (c *Component) Init() tea.Cmd {
	return nil
}

// Update 更新组件状态
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

// View 渲染视图
func (c *Component) View() string {
	if c.height <= 0 {
		c.height = 1
	}

	// 左侧：状态指示器
	leftContent := c.renderStatus()

	// 中间：模型信息
	centerContent := c.renderModelInfo()

	// 右侧：Token 用量
	rightContent := c.renderTokenUsage()

	// 组合状态栏
	leftWidth := 20
	rightWidth := 25
	centerWidth := c.width - leftWidth - rightWidth - 4

	if centerWidth < 10 {
		centerWidth = 10
	}

	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Align(lipgloss.Left)

	centerStyle := lipgloss.NewStyle().
		Width(centerWidth).
		Align(lipgloss.Center)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Align(lipgloss.Right)

	barStyle := lipgloss.NewStyle().
		Background(statusColorBg).
		Foreground(statusColorText).
		Width(c.width).
		Padding(0, 1)

	content := fmt.Sprintf("%s %s %s",
		leftStyle.Render(leftContent),
		centerStyle.Render(centerContent),
		rightStyle.Render(rightContent))

	return barStyle.Render(content)
}

// SetSize 设置组件尺寸
func (c *Component) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetStatus 设置状态
func (c *Component) SetStatus(status Status) {
	c.status = status
	c.lastUpdated = time.Now()
}

// SetModel 设置模型
func (c *Component) SetModel(model string) {
	c.model = model
}

// SetMode 设置模式
func (c *Component) SetMode(mode string) {
	c.mode = mode
}

// SetMessage 设置状态消息
func (c *Component) SetMessage(msg string) {
	c.message = msg
}

// UpdateTokenUsage 更新 Token 用量
func (c *Component) UpdateTokenUsage(usage *types.TokenUsage) {
	c.tokenUsage = usage
	c.lastUpdated = time.Now()
}

// renderStatus 渲染状态指示器
func (c *Component) renderStatus() string {
	switch c.status {
	case StatusIdle:
		return lipgloss.NewStyle().
			Foreground(statusColorIdle).
			Render("● Ready")

	case StatusConnecting:
		return lipgloss.NewStyle().
			Foreground(statusColorConnecting).
			Render("◐ Connecting...")

	case StatusThinking:
		return lipgloss.NewStyle().
			Foreground(statusColorThinking).
			Render("◐ Thinking...")

	case StatusToolExecuting:
		return lipgloss.NewStyle().
			Foreground(statusColorTool).
			Render("◐ Tool...")

	case StatusError:
		return lipgloss.NewStyle().
			Foreground(statusColorError).
			Render("● Error")

	default:
		return "Ready"
	}
}

// renderModelInfo 渲染模型信息
func (c *Component) renderModelInfo() string {
	var modeIndicator string
	if c.mode == "vim" {
		modeIndicator = lipgloss.NewStyle().
			Foreground(statusColorVim).
			Render("[VIM] ")
	}

	modelStyle := lipgloss.NewStyle().
		Foreground(statusColorMuted)

	return fmt.Sprintf("%s%s", modeIndicator, modelStyle.Render(c.model))
}

// renderTokenUsage 渲染 Token 用量
func (c *Component) renderTokenUsage() string {
	if c.tokenUsage == nil {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(statusColorMuted)

	total := c.tokenUsage.InputTokens + c.tokenUsage.OutputTokens
	return style.Render(fmt.Sprintf("Tokens: %d", total))
}

// GetStatus 获取当前状态
func (c *Component) GetStatus() Status {
	return c.status
}

// IsProcessing 是否正在处理
func (c *Component) IsProcessing() bool {
	return c.status == StatusThinking || c.status == StatusToolExecuting
}
