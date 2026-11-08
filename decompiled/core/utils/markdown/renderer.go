// Package markdown terminal markdown renderer wrapper
package markdown

import (
	"os"
	"sync"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
)

// Renderer terminal markdown renderer
type Renderer struct {
	renderer *glamour.TermRenderer
	mu       sync.RWMutex
	width    int
	style    string
}

// NewRenderer creates a new markdown renderer
func NewRenderer(width int, style string) (*Renderer, error) {
	if style == "" {
		style = DetectStyle()
	}

	// Print detailed debug info
	log.Debugf("markdown.NewRenderer: width=%d, requested style=%s", width, style)

	// Detect terminal background
	hasDarkBg := lipgloss.HasDarkBackground()
	log.Debugf("markdown lipgloss.HasDarkBackground() = %v", hasDarkBg)

	// Print terminal config info (may affect background detection)
	log.Debugf("markdown COLORFGBG env = %q", getEnv("COLORFGBG"))
	log.Debugf("markdown TERM_PROGRAM env = %q", getEnv("TERM_PROGRAM"))
	log.Debugf("markdown TERM env = %q", getEnv("TERM"))
	log.Debugf("markdown ITERM_PROFILE env = %q", getEnv("ITERM_PROFILE"))

	// Ensure reasonable width
	if width <= 0 {
		width = 80
	}

	// Use advanced color scheme
	// Reference GitHub dark theme palette, more comfortable and professional
	trueVal := true
	falseVal := false

	// Advanced color scheme
	// Text: soft white (252) instead of harsh pure white
	// Headings: slight cyan (86) to make headings stand out
	// Emphasis: purple (141) to highlight keywords
	// Links: blue (75) per convention
	textColor := "252"   // soft white - body text
	headingColor := "86" // cyan - headings
	emphColor := "141"   // light purple - emphasis
	linkColor := "75"    // blue - links
	codeColor := "228"   // warm yellow - code

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

			Heading: ansi.StyleBlock{StylePrimitive: boldStyle},
			H1:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H2:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H3:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H4:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H5:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H6:      ansi.StyleBlock{StylePrimitive: boldStyle},

			Emph:   emphStyle,
			Strong: strongStyle,

			Link:     linkStyle,
			LinkText: linkStyle,

			List:      ansi.StyleList{},
			CodeBlock: ansi.StyleCodeBlock{StyleBlock: ansi.StyleBlock{StylePrimitive: codeStyle}},
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

// getEnv helper to get an environment variable
func getEnv(key string) string {
	return os.Getenv(key)
}

// Render renders markdown into a terminal-displayable format
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

// RenderBytes renders markdown byte slice
func (r *Renderer) RenderBytes(content []byte) (string, error) {
	return r.Render(string(content))
}

// SetSize updates renderer dimensions
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

	// Disable WithWordWrap to avoid Chinese line-wrap issues
	// Use the same advanced color scheme as NewRenderer
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

			Heading: ansi.StyleBlock{StylePrimitive: boldStyle},
			H1:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H2:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H3:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H4:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H5:      ansi.StyleBlock{StylePrimitive: boldStyle},
			H6:      ansi.StyleBlock{StylePrimitive: boldStyle},

			Emph:   emphStyle,
			Strong: strongStyle,

			Link:     linkStyle,
			LinkText: linkStyle,

			List:      ansi.StyleList{},
			CodeBlock: ansi.StyleCodeBlock{StyleBlock: ansi.StyleBlock{StylePrimitive: codeStyle}},
		}),
	)
	if err != nil {
		return err
	}

	r.renderer = newRenderer
	return nil
}

// SetStyle switches theme
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
	// Disable WithWordWrap to avoid Chinese line-wrap issues
	newRenderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		// glamour.WithWordWrap(r.width), // disabled: breaks Chinese text when wrapping
	)
	if err != nil {
		return err
	}

	r.renderer = newRenderer
	return nil
}

// GetStyle returns the current theme
func (r *Renderer) GetStyle() string {
	if r == nil {
		return "dark"
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.style
}

// GetWidth returns the current width
func (r *Renderer) GetWidth() int {
	if r == nil {
		return 80
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.width
}

// DetectStyle detects the terminal theme and returns an appropriate glamour style
func DetectStyle() string {
	if lipgloss.HasDarkBackground() {
		return "dark"
	}
	return "light"
}

// GetAvailableStyles returns the list of all available themes
// glamour built-in themes: dark, light, dracula, ascii, notty
// The official binary also contains many custom themes loadable via config file
func GetAvailableStyles() []string {
	return []string{
		"dark",
		"light",
		"dracula",
		"ascii",
		"notty",
	}
}

// IsValidStyle checks whether a theme is valid
func IsValidStyle(style string) bool {
	for _, s := range GetAvailableStyles() {
		if s == style {
			return true
		}
	}
	return false
}
