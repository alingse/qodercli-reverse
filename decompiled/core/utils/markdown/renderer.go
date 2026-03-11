// Package markdown markdown 终端渲染器封装
package markdown

import (
	"os"
	"sync"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
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

	// 输出详细的调试信息
	log.Debugf("markdown.NewRenderer: width=%d, requested style=%s", width, style)

	// 检测终端背景
	hasDarkBg := lipgloss.HasDarkBackground()
	log.Debugf("markdown lipgloss.HasDarkBackground() = %v", hasDarkBg)

	// 输出终端配置信息（可能影响背景检测）
	log.Debugf("markdown COLORFGBG env = %q", getEnv("COLORFGBG"))
	log.Debugf("markdown TERM_PROGRAM env = %q", getEnv("TERM_PROGRAM"))
	log.Debugf("markdown TERM env = %q", getEnv("TERM"))
	log.Debugf("markdown ITERM_PROFILE env = %q", getEnv("ITERM_PROFILE"))

	// 确保宽度合理
	if width <= 0 {
		width = 80
	}

	// 使用高级配色方案
	// 参考 GitHub 的暗色主题配色，更加舒适和专业
	trueVal := true
	falseVal := false

	// 高级配色方案
	// 文本：柔和的白色 (252) 而不是刺眼的纯白
	// 标题：带点青色 (86) 让标题更突出
	// 强调：紫色 (141) 让关键词醒目
	// 链接：蓝色 (75) 符合习惯
	textColor := "252"      // 柔和白 - 正文
	headingColor := "86"    // 青色 - 标题
	emphColor := "141"      // 淡紫 - 强调
	linkColor := "75"       // 蓝色 - 链接
	codeColor := "228"      // 暖黄 - 代码

	boldStyle := ansi.StylePrimitive{Bold: &trueVal, Color: &headingColor}
	textStyle := ansi.StylePrimitive{Color: &textColor}
	emphStyle := ansi.StylePrimitive{Italic: &trueVal, Color: &emphColor}
	strongStyle := ansi.StylePrimitive{Bold: &trueVal, Color: &emphColor}
	linkStyle := ansi.StylePrimitive{Color: &linkColor, Underline: &falseVal}
	codeStyle := ansi.StylePrimitive{Color: &codeColor}

	log.Debugf("markdown: using premium color scheme - text:%s heading:%s emph:%s link:%s code:%s",
		textColor, headingColor, emphColor, linkColor, codeColor)

	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(ansi.StyleConfig{
			Document:   ansi.StyleBlock{StylePrimitive: textStyle},
			Paragraph:  ansi.StyleBlock{StylePrimitive: textStyle},
			BlockQuote: ansi.StyleBlock{StylePrimitive: emphStyle},

			Heading:    ansi.StyleBlock{StylePrimitive: boldStyle},
			H1:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H2:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H3:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H4:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H5:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H6:         ansi.StyleBlock{StylePrimitive: boldStyle},

			Emph:       emphStyle,
			Strong:     strongStyle,

			Link:       linkStyle,
			LinkText:   linkStyle,

			List:       ansi.StyleList{},
			CodeBlock:  ansi.StyleCodeBlock{StyleBlock: ansi.StyleBlock{StylePrimitive: codeStyle}},
		}),
	)

	if err != nil {
		log.Errorf("markdown: failed to create renderer: %v", err)
		return nil, err
	}

	log.Debugf("markdown: renderer created successfully")

	return &Renderer{
		renderer: r,
		width:    width,
		style:    style,
	}, nil
}

// getEnv 获取环境变量的辅助函数
func getEnv(key string) string {
	return os.Getenv(key)
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
	// 使用与 NewRenderer 相同的高级配色方案
	trueVal := true
	falseVal := false

	textColor := "252"
	headingColor := "86"
	emphColor := "141"
	linkColor := "75"
	codeColor := "228"

	boldStyle := ansi.StylePrimitive{Bold: &trueVal, Color: &headingColor}
	textStyle := ansi.StylePrimitive{Color: &textColor}
	emphStyle := ansi.StylePrimitive{Italic: &trueVal, Color: &emphColor}
	strongStyle := ansi.StylePrimitive{Bold: &trueVal, Color: &emphColor}
	linkStyle := ansi.StylePrimitive{Color: &linkColor, Underline: &falseVal}
	codeStyle := ansi.StylePrimitive{Color: &codeColor}

	newRenderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(ansi.StyleConfig{
			Document:   ansi.StyleBlock{StylePrimitive: textStyle},
			Paragraph:  ansi.StyleBlock{StylePrimitive: textStyle},
			BlockQuote: ansi.StyleBlock{StylePrimitive: emphStyle},

			Heading:    ansi.StyleBlock{StylePrimitive: boldStyle},
			H1:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H2:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H3:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H4:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H5:         ansi.StyleBlock{StylePrimitive: boldStyle},
			H6:         ansi.StyleBlock{StylePrimitive: boldStyle},

			Emph:       emphStyle,
			Strong:     strongStyle,

			Link:       linkStyle,
			LinkText:   linkStyle,

			List:       ansi.StyleList{},
			CodeBlock:  ansi.StyleCodeBlock{StyleBlock: ansi.StyleBlock{StylePrimitive: codeStyle}},
		}),
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
