// Package markdown markdown 终端渲染器封装
package markdown

import (
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Renderer markdown 终端渲染器
type Renderer struct {
	renderer *glamour.TermRenderer
	mu       sync.RWMutex
	width    int
	style    string
}

// NewRenderer 创建新的 markdown 渲染器
func NewRenderer(width int, style string) (*Renderer, error) {
	if style == "" {
		style = DetectStyle()
	}

	// 确保宽度合理
	if width <= 0 {
		width = 80
	}

	// 对于深色背景，使用 ASCII 样式避免颜色问题
	// ASCII 样式只使用基本格式（加粗、下划线），不使用颜色
	var r *glamour.TermRenderer
	var err error

	if style == "dark" {
		// 使用 ASCII 样式：没有颜色，只有格式
		r, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("ascii"),
		)
	} else {
		// 使用标准主题
		r, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle(style),
		)
	}

	if err != nil {
		return nil, err
	}

	return &Renderer{
		renderer: r,
		width:    width,
		style:    style,
	}, nil
}

// Render 渲染 markdown 为终端可显示的格式
func (r *Renderer) Render(content string) (string, error) {
	if r == nil {
		return content, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return content, nil
	}

	return r.renderer.Render(content)
}

// RenderBytes 渲染 markdown 字节数组
func (r *Renderer) RenderBytes(content []byte) (string, error) {
	return r.Render(string(content))
}

// SetSize 更新渲染器尺寸
func (r *Renderer) SetSize(width int) error {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.width == width {
		return nil
	}

	if width <= 0 {
		width = 80
	}

	r.width = width
	// 禁用 WithWordWrap 以避免中文换行问题
	newRenderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(r.style),
		// glamour.WithWordWrap(width), // 禁用：会导致中文换行时字符破坏
	)
	if err != nil {
		return err
	}

	r.renderer = newRenderer
	return nil
}

// SetStyle 切换主题
func (r *Renderer) SetStyle(style string) error {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.style == style {
		return nil
	}

	r.style = style
	// 禁用 WithWordWrap 以避免中文换行问题
	newRenderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		// glamour.WithWordWrap(r.width), // 禁用：会导致中文换行时字符破坏
	)
	if err != nil {
		return err
	}

	r.renderer = newRenderer
	return nil
}

// GetStyle 返回当前主题
func (r *Renderer) GetStyle() string {
	if r == nil {
		return "dark"
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.style
}

// GetWidth 返回当前宽度
func (r *Renderer) GetWidth() int {
	if r == nil {
		return 80
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.width
}

// DetectStyle 检测终端主题并返回合适的 glamour 样式
func DetectStyle() string {
	if lipgloss.HasDarkBackground() {
		return "dark"
	}
	return "light"
}

// GetAvailableStyles 返回所有可用的主题列表
// glamour 内置主题: dark, light, dracula, ascii, notty
// 官方二进制中还包含许多自定义主题，可通过配置文件加载
func GetAvailableStyles() []string {
	return []string{
		"dark",
		"light",
		"dracula",
		"ascii",
		"notty",
	}
}

// IsValidStyle 检查主题是否有效
func IsValidStyle(style string) bool {
	for _, s := range GetAvailableStyles() {
		if s == style {
			return true
		}
	}
	return false
}
