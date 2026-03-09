// Package editor 消息编辑器组件
// 支持多行输入、Vim 模式、语法高亮
package editor

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/textarea"
)

// 配色方案
var (
	editorColorBorder      = lipgloss.Color("#7D56F4")  // 边框 - 紫色
	editorColorBorderVim   = lipgloss.Color("#00C853")  // Vim 模式边框 - 绿色
	editorColorPlaceholder = lipgloss.Color("#6B7280")  // 占位符 - 灰色
)

// Mode 编辑器模式
type Mode int

const (
	ModeInsert Mode = iota
	ModeNormal
	ModeVisual
)

// Component 编辑器组件
type Component struct {
	textarea textarea.Model
	mode     Mode
	width    int
	height   int
	focused  bool
}

// New 创建新的编辑器组件
func New() *Component {
	ta := textarea.New()
	ta.Placeholder = "Type your message... (Press Enter to send, Shift+Enter for new line)"
	ta.ShowLineNumbers = false
	ta.CharLimit = 10000

	return &Component{
		textarea: ta,
		mode:     ModeInsert,
		focused:  true,
	}
}

// Init 初始化组件
func (c *Component) Init() tea.Cmd {
	return textarea.Blink
}

// Update 更新组件状态
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 处理特殊按键
		switch msg.Type {
		case tea.KeyCtrlD:
			// 发送消息
			return c, nil

		case tea.KeyTab:
			// 缩进
			c.textarea.InsertString("    ")
			return c, nil

		case tea.KeyShiftTab:
			// 减少缩进
			return c, nil
		}
	}

	// 更新 textarea
	newTextarea, cmd := c.textarea.Update(msg)
	c.textarea = newTextarea
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

// View 渲染视图
func (c *Component) View() string {
	if !c.focused {
		c.textarea.Blur()
	} else {
		c.textarea.Focus()
	}

	// 样式
	borderColor := editorColorBorder
	if c.mode == ModeNormal {
		borderColor = editorColorBorderVim
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(c.width)

	return style.Render(c.textarea.View())
}

// GetInput 获取当前输入内容
func (c *Component) GetInput() string {
	return strings.TrimSpace(c.textarea.Value())
}

// Clear 清空输入
func (c *Component) Clear() {
	c.textarea.Reset()
}

// SetSize 设置组件尺寸
func (c *Component) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.textarea.SetWidth(width - 4)
	c.textarea.SetHeight(height - 2)
}

// Focus 聚焦组件
func (c *Component) Focus() {
	c.focused = true
	c.textarea.Focus()
}

// Blur 失焦组件
func (c *Component) Blur() {
	c.focused = false
	c.textarea.Blur()
}

// SetMode 设置编辑模式
func (c *Component) SetMode(mode Mode) {
	c.mode = mode
}

// GetMode 获取当前模式
func (c *Component) GetMode() Mode {
	return c.mode
}

// InsertText 插入文本
func (c *Component) InsertText(text string) {
	c.textarea.InsertString(text)
}

// SetPlaceholder 设置占位符
func (c *Component) SetPlaceholder(placeholder string) {
	c.textarea.Placeholder = placeholder
}
